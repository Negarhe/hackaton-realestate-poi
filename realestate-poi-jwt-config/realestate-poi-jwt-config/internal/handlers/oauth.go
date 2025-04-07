package handlers

import (
	"fmt"
	"log"
	"net/http"

	"git.divar.cloud/divar/girls-hackathon/realestate-poi/internal/services"
	"git.divar.cloud/divar/girls-hackathon/realestate-poi/utils"
)

type OauthResourceType string

const (
	POST_ADDON_CREATE  OauthResourceType = "POST_ADDON_CREATE"
	USER_PHONE         OauthResourceType = "USER_PHONE"
	USER_ID            OauthResourceType = "USER_ID"
	OFFLINE_ACCESS     OauthResourceType = "offline_access"
	DefaultRedirectURL                   = "https://divar.ir/"
)

type Scope struct {
	resourceType OauthResourceType
	resourceID   string
}

type oAuthHandler struct {
	oauthService *services.OAuthService
	kenarService *services.KenarService
	store        *utils.SessionStore
	jwt          *utils.JWTManager
}

func NewOAuthHandler(store *utils.SessionStore, serv *services.OAuthService, kenar *services.KenarService, jwt *utils.JWTManager) *oAuthHandler {
	if store == nil {
		log.Fatal("cookie store can not be nil")
	}
	if serv == nil {
		log.Fatal("oauth service can not be nil")
	}

	return &oAuthHandler{
		store:        store,
		oauthService: serv,
		kenarService: kenar,
		jwt:          jwt,
	}
}

func (h *oAuthHandler) buildScopes(postToken string, isBuyer bool) []string {
	var oauthScopes []Scope

	if isBuyer {
		// For buyers, only include USER_ID scope
		oauthScopes = []Scope{
			{resourceType: USER_ID},
		}
	} else {
		// For non-buyers, include all scopes
		oauthScopes = []Scope{
			{resourceType: POST_ADDON_CREATE, resourceID: postToken},
			{resourceType: USER_ID},
			{resourceType: OFFLINE_ACCESS},
		}
	}

	var scopes []string
	for _, scope := range oauthScopes {
		if scope.resourceID != "" {
			scopes = append(scopes, fmt.Sprintf("%s.%s", scope.resourceType, scope.resourceID))
		} else {
			scopes = append(scopes, string(scope.resourceType))
		}
	}
	return scopes
}

// func (h *oAuthHandler) buildScopes(postToken string) []string {
// 	oauthScopes := []Scope{
// 		{resourceType: POST_ADDON_CREATE, resourceID: postToken},
// 		{resourceType: USER_ID},
// 		{resourceType: OFFLINE_ACCESS},
// 	}

// 	var scopes []string
// 	for _, scope := range oauthScopes {
// 		if scope.resourceID != "" {
// 			scopes = append(scopes, fmt.Sprintf("%s.%s", scope.resourceType, scope.resourceID))
// 		} else {
// 			scopes = append(scopes, string(scope.resourceType))
// 		}
// 	}
// 	return scopes
// }

func (h *oAuthHandler) AddonOauth(w http.ResponseWriter, r *http.Request) {
	log.Println("AddonOauth called")

	postToken := r.URL.Query().Get("post_token")
	return_url := r.URL.Query().Get("return_url")

	isBuyer := return_url == ""

	if postToken == "" {
		http.Error(w, "post_token is required", http.StatusBadRequest)
		return
	}
	session, err := h.store.GetExistingSession(w, r)
	if err != nil || session.PostToken != postToken || session.IsBuyer != isBuyer {
		log.Println("new person. new post!")
		session, err = h.store.CreateNewSession(w, r, postToken, return_url, isBuyer)
		if err != nil {
			http.Error(w, "Failed to create session: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// creating scopes
	scopes := h.buildScopes(postToken, isBuyer)
	redirect_url := h.oauthService.GenerateAuthURL(scopes, session.State)
	http.Redirect(w, r, redirect_url, http.StatusFound)

}

// call back
func (h *oAuthHandler) validateCallback(r *http.Request) (string, string, error) {

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		return "", "", fmt.Errorf("code and state are required")
	}
	return code, state, nil
}

func (h *oAuthHandler) OauthCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("OauthCallback called")
	code, state, err := h.validateCallback(r)
	if err != nil {
		log.Printf("Invalid callback parameters: %v", err)
		http.Redirect(w, r, DefaultRedirectURL, http.StatusSeeOther)
		return
	}

	//get existing session
	session, err := h.store.GetExistingSession(w, r)
	if err != nil {
		http.Error(w, "Failed to get session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//check if state is the same as the one in the session
	if session.State != state {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}
	//sending code to get the token
	token, err := h.oauthService.ExchangeToken(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	userDetail, err := h.kenarService.GetUserDetail(token.AccessToken)
	if err != nil {
		utils.HanleError(w, r, http.StatusInternalServerError,
			"خطا در دریافت اطلاعات کاربری",
			"امکان دریافت اطلاعات شما وجود ندارد. لطفا بعدا تلاش کنید", err.Error())
		return
	}
	// if err != nil {
	// 	http.Error(w, "Failed to get user information: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	properyDetail, err := h.kenarService.GetPropertyDetail(r.Context(), session.PostToken)
	if err != nil {
		http.Error(w, "Failed to get coordinates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the user is the owner of this post
	_, err = h.oauthService.IsUserPostOwner(r.Context(), userDetail.UserId, session.PostToken)
	if err != nil {
		log.Printf("Error checking post ownership: %v", err)
		// Continue even if we can't check ownership
	}

	transactionData := &services.Transaction{
		PropertyDetail: properyDetail,
		UserDetail:     userDetail,
		IsBuyer:        session.IsBuyer,
	}

	// Only include token info for  (sellers)
	if !session.IsBuyer {
		transactionData.TokenInfo = &services.TokenInfo{
			RefreshToken: token.RefreshToken,
			AccessToken:  token.AccessToken,
			ExpiresIn:    token.Expiry,
		}
	}

	err = h.oauthService.RegisterAuthData(r.Context(), transactionData)
	if err != nil {
		http.Error(w, "Failed to register auth data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jwtToken, err := h.jwt.CreateJwtToken(userDetail.UserId)
	if err != nil {
		http.Error(w, "Error creating jwt token", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization_Token",
		Value:    jwtToken,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   86400,
	})
	var url string
	if session.IsBuyer {
		url = fmt.Sprintf("/api/buyer/landing?post_token=%s&return_url=%s", session.PostToken, "https://open-platform-redirect.divar.ir/completion")
	} else {
		url = fmt.Sprintf("/api/seller/landing?post_token=%s&return_url=%s", session.PostToken, session.ReturnUrl)
	}
	// url := fmt.Sprintf("/api/main?post_token=%s&return_url=%s", session.PostToken, session.ReturnUrl)
	http.Redirect(w, r, url, http.StatusSeeOther)

}
