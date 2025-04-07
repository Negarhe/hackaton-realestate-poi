package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"git.divar.cloud/divar/girls-hackathon/realestate-poi/internal/services"
	"git.divar.cloud/divar/girls-hackathon/realestate-poi/pkg/transport"
	"git.divar.cloud/divar/girls-hackathon/realestate-poi/utils"
)

type KenarHandler struct {
	store            *utils.SessionStore
	kenarService     *services.KenarService
	transportService *services.TransportService
}

func NewKenarHandler(store *utils.SessionStore, serv *services.KenarService, transportService *services.TransportService) *KenarHandler {
	return &KenarHandler{
		store:            store,
		kenarService:     serv,
		transportService: transportService,
	}
}

func (k *KenarHandler) GetPrice(w http.ResponseWriter, r *http.Request) {
	log.Println("Get price called")
	var req struct {
		PostToken string `json:"post_token"`

		Origin struct {
			Lat  float64 `json:"lat"`
			Long float64 `json:"lng"`
		} `json:"origin"`
		Destination struct {
			Lat  float64 `json:"lat"`
			Long float64 `json:"lng"`
		} `json:"destination"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println(err.Error())
		utils.HanleError(w, r, http.StatusInternalServerError, "خطا در پردازش درخواست", "فرمت درخواست نامعتبر است", err.Error())
		return
	}

	price, err := k.transportService.GetPrice(r.Context(), strconv.FormatFloat(req.Origin.Lat, 'f', -1, 64), strconv.FormatFloat(req.Origin.Long, 'f', -1, 64), strconv.FormatFloat(req.Destination.Lat, 'f', -1, 64), strconv.FormatFloat(req.Destination.Long, 'f', -1, 64))
	//even if both snapp and tapsi could not get the price, we should return 0 for both service
	if err != nil {
		// http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"snapp_price": price.SnappPrice,
		"tapsi_price": price.TapsiPrice,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.HanleError(w, r, http.StatusInternalServerError, "خطای سیستمی", "خطا در تولید پاسخ", err.Error())

		// http.Error(w, "failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (k *KenarHandler) Poi(w http.ResponseWriter, r *http.Request) {
	log.Println("Kenar called")

	userId, ok := r.Context().Value("user_id").(string)
	if !ok {
		utils.HanleError(w, r, http.StatusInternalServerError, "خطای احراز هویت", "کاربر شناسایی نشد", "User ID not found in context")
		return
	}
	log.Println("hereeeeeeeeeeeeee")
	log.Println("userId: ", userId)
	var req struct {
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"lng"`
		PostToken string  `json:"post_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HanleError(w, r, http.StatusInternalServerError, "خطا در پردازش درخواست", "فرمت درخواست نامعتبر است", err.Error())

		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("here")
	stationResult, err := k.transportService.FindNearestStation(r.Context(), userId, req.PostToken, strconv.FormatFloat(req.Latitude, 'f', -1, 64), strconv.FormatFloat(req.Longitude, 'f', -1, 64))
	if err != nil {
		http.Error(w, "failed to find nearest station: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(stationResult); err != nil {
		http.Error(w, "failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}

}

type SubwayInfo struct {
	Distance string `json:"distance"`
	Name     string `json:"name"`
	Duration string `json:"duration"`
}

// type PoiResponse struct {
// 	Subway SubwayInfo `json:"subway"`
// }

type AddToListingRequest struct {
	PostToken string                       `json:"post_token"`
	Amenity   transport.NearbyPOIsResponse `json:"amenities"`
}

func (h *KenarHandler) AddLocationWidget(w http.ResponseWriter, r *http.Request) {
	log.Println("Add location widget")

	userId, ok := r.Context().Value("user_id").(string)
	if !ok {
		utils.HanleError(w, r, http.StatusInternalServerError, "خطای احراز هویت", "کاربر شناسایی نشد", "User ID not found in context")
		// http.Error(w, "User ID not found or invalid", http.StatusInternalServerError)
		return
	}

	var req AddToListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println(err.Error())
		utils.HanleError(w, r, http.StatusBadRequest, "درخواست نامعتبر", "فرمت درخواست نامعتبر است", err.Error())
		// http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.PostToken == "" {
		utils.HanleError(w, r, http.StatusBadRequest, "درخواست نامعتبر", "فیلدهای ضروری وارد نشده‌اند", "Missing post_token field")
		// http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// sample userId until we use jwt

	log.Println("finallyyyyyyyy")
	log.Println(userId)

	err := h.kenarService.PostLocationWidget(r.Context(), userId, req.PostToken, req.Amenity)

	if err != nil {
		log.Println(err.Error())
		utils.HanleError(w, r, http.StatusInternalServerError, "خطا در ثبت ویجت", "خطا در ثبت اطلاعات مکانی", err.Error())
		// http.Error(w, "Failed to post widget: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Location widget added successfully",
	})
}

func (k *KenarHandler) GetOriginCoordinates(w http.ResponseWriter, r *http.Request) {
	// this works with database
	var req struct {
		PostToken string `json:"post_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HanleError(w, r, http.StatusBadRequest, "درخواست نامعتبر", "خطا در خواندن اطلاعات درخواست", err.Error())
		// http.Error(w, "failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	post, err := k.kenarService.GetPropertyDetail(r.Context(), req.PostToken)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.HanleError(w, r, http.StatusNotFound, "یافت نشد", "آگهی موردنظر یافت نشد", err.Error())
			// http.Error(w, "no post found", http.StatusNotFound)
			return
		}
		utils.HanleError(w, r, http.StatusInternalServerError, "خطای سیستمی", "خطا در دریافت اطلاعات آگهی", err.Error())
		// http.Error(w, "failed to get post: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(struct {
		Latitude  float64 `json:"Latitude"`
		Longitude float64 `json:"Longitude"`
		Title     string  `json:"title"`
	}{
		Title:     post.Title,
		Latitude:  post.Latitude,
		Longitude: post.Longitude,
	}); err != nil {
		utils.HanleError(w, r, http.StatusInternalServerError, "خطای سیستمی", "خطا در تولید پاسخ", err.Error())
		// http.Error(w, "failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}

}

func (k *KenarHandler) RecordPurchase(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("user_id").(string)
	if !ok {
		utils.HanleError(w, r, http.StatusInternalServerError, "خطای احراز هویت", "کاربر شناسایی نشد", "User ID not found in context")
		return
	}
	var req struct {
		PostToken string `json:"post_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.HanleError(w, r, http.StatusBadRequest, "درخواست نامعتبر", "خطا در خواندن اطلاعات درخواست", err.Error())
		return
	}
	err := k.kenarService.InsertPostPurchase(r.Context(), req.PostToken, userId)
	if err != nil {
		utils.HanleError(w, r, http.StatusInternalServerError, "خطای سیستمی", "خطا در ذخیره اطلاعات", err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "خرید با موفقیت ثبت شد",
		"data": map[string]interface{}{
			"post_token": req.PostToken,
		},
	})

}
