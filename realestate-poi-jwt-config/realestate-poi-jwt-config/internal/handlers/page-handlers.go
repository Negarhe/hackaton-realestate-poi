package handlers

import (
	"html/template"
	"log"
	"net/http"

	"git.divar.cloud/divar/girls-hackathon/realestate-poi/internal/services"
	"git.divar.cloud/divar/girls-hackathon/realestate-poi/utils"
)

type pageHandler struct {
	sessionStore *utils.SessionStore
	kenarService *services.KenarService
	taxiService  *services.TransportService
}

func NewPageHandler(
	sessionStore *utils.SessionStore,
	kenarService *services.KenarService,
	taxiService *services.TransportService,
) *pageHandler {
	return &pageHandler{
		sessionStore: sessionStore,
		kenarService: kenarService,
		taxiService:  taxiService,
	}
}
func (p *pageHandler) BuyerDashboardHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("here")
	postToken := r.URL.Query().Get("post_token")
	return_url := r.URL.Query().Get("return_url")

	if postToken == "" || return_url == "" {
		http.Error(w, "post_token and return_url are required", http.StatusBadRequest)
		return
	}
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get property details
	property, err := p.kenarService.GetPropertyDetail(r.Context(), postToken)
	if err != nil {
		http.Error(w, "Failed to fetch property details", http.StatusInternalServerError)
		return
	}

	hasPurchased, err := p.kenarService.CheckUserPurchase(r.Context(), postToken, userID)

	// Render buyer template with data
	data := map[string]interface{}{
		"UserID":       userID,
		"PostToken":    postToken,
		"PropertyData": property,
		"IsPurchased":  hasPurchased,
	}
	tmp, err := template.ParseFiles("./web/buyer_landing.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmp.ExecuteTemplate(w, "buyer_landing.html", data)
	return
}
func (p *pageHandler) SellerDashboardHandler(w http.ResponseWriter, r *http.Request) {
	postToken := r.URL.Query().Get("post_token")
	return_url := r.URL.Query().Get("return_url")
	log.Println(postToken)
	log.Println(return_url)

	if postToken == "" || return_url == "" {
		http.Error(w, "post_token and return_url are required", http.StatusBadRequest)
		return
	}

	tmp, err := template.ParseFiles("./web/landing.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	data := struct {
		Token        string
		RedirectLink string
	}{
		Token:        postToken,
		RedirectLink: return_url,
	}

	tmp.Execute(w, data)
	return
}
