package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"git.divar.cloud/divar/girls-hackathon/realestate-poi/pkg/configs"
	"github.com/golang-jwt/jwt"
)

type JWTManager struct {
	secretKey []byte
}

func NewJWTManager(config *configs.JWTConfig) *JWTManager {
	return &JWTManager{
		secretKey: []byte(config.JwtSecret),
	}
}
func (j *JWTManager) CreateJwtToken(userId string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(j.secretKey)
	if err != nil {
		log.Println("could not create jwt token")
		return "", err
	}
	return signedToken, nil

}

// func (j *JWTManager) JWTMiddlewear(next http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		cookie, err := r.Cookie("Authorization_Token")
// 		if err != nil {
// 			if err == http.ErrNoCookie {
// 				HanleError(w, r, http.StatusUnauthorized, "خطای احراز هویت", "توکن احراز هویت ارائه نشده است", "No token provided")
// 				return
// 			}
// 			HanleError(w, r, http.StatusBadRequest, "درخواست نامعتبر", "خطا در خواندن کوکی احراز هویت", err.Error())
// 			return
// 		}

// 		// Parse token
// 		tokenString := cookie.Value
// 		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 			// Verify signing method
// 			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 			}
// 			return j.secretKey, nil
// 		})

// 		// Handle parsing errors
// 		if err != nil {
// 			ve, ok := err.(*jwt.ValidationError)
// 			if ok {
// 				if ve.Errors&jwt.ValidationErrorExpired != 0 {
// 					HanleError(w, r, http.StatusUnauthorized, "خطای احراز هویت", "توکن منقضی شده است. لطفا مجدد وارد شوید", "Token expired")
// 					return
// 				}
// 				log.Println("here")
// 				HanleError(w, r, http.StatusUnauthorized, "خطای احراز هویت", "توکن نامعتبر است", err.Error())
// 				return
// 			}
// 			HanleError(w, r, http.StatusUnauthorized, "خطای احراز هویت", "توکن نامعتبر است", err.Error())
// 			return
// 		}
// 		if !token.Valid {
// 			HanleError(w, r, http.StatusUnauthorized, "خطای احراز هویت", "توکن نامعتبر است", "Invalid token")
// 			return
// 		}

// 		if claims, ok := token.Claims.(jwt.MapClaims); ok {
// 			ctx := context.WithValue(r.Context(), "user_id", claims["user_id"])
// 			next(w, r.WithContext(ctx))
// 		} else {
// 			HanleError(w, r, http.StatusUnauthorized, "خطای احراز هویت", "توکن نامعتبر است", "Invalid token claims")
// 		}

// 	}

// }

func (j *JWTManager) JWTMiddlewear(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if it's an AJAX request
		isAjax := r.Header.Get("X-Requested-With") == "XMLHttpRequest" ||
			r.Header.Get("Accept") == "application/json"

		cookie, err := r.Cookie("Authorization_Token")
		if err != nil {
			if err == http.ErrNoCookie {
				if isAjax {
					// JSON response for AJAX
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error":   "خطای احراز هویت",
						"message": "توکن احراز هویت ارائه نشده است",
						"redirect": "/error?code=401&message=" + url.QueryEscape("خطای احراز هویت") +
							"&description=" + url.QueryEscape("توکن احراز هویت ارائه نشده است") +
							"&technical=" + url.QueryEscape("No token provided"),
					})
					return
				}
				// Regular redirect for browser
				HanleError(w, r, http.StatusUnauthorized, "خطای احراز هویت", "توکن احراز هویت ارائه نشده است", "No token provided")
				return
			}
			// Handle bad cookie
			if isAjax {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "درخواست نامعتبر",
					"message": "خطا در خواندن کوکی احراز هویت",
					"redirect": "/error?code=400&message=" + url.QueryEscape("درخواست نامعتبر") +
						"&description=" + url.QueryEscape("خطا در خواندن کوکی احراز هویت") +
						"&technical=" + url.QueryEscape(err.Error()),
				})
				return
			}
			HanleError(w, r, http.StatusBadRequest, "درخواست نامعتبر", "خطا در خواندن کوکی احراز هویت", err.Error())
			return
		}

		// Parse token
		tokenString := cookie.Value
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return j.secretKey, nil
		})

		// Handle parsing errors
		if err != nil {
			ve, ok := err.(*jwt.ValidationError)
			if ok && ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token expired
				if isAjax {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error":   "خطای احراز هویت",
						"message": "توکن منقضی شده است. لطفا مجدد وارد شوید",
						"redirect": "/error?code=401&message=" + url.QueryEscape("خطای احراز هویت") +
							"&description=" + url.QueryEscape("توکن منقضی شده است. لطفا مجدد وارد شوید") +
							"&technical=" + url.QueryEscape("Token expired"),
					})
					return
				}
				HanleError(w, r, http.StatusUnauthorized, "خطای احراز هویت", "توکن منقضی شده است. لطفا مجدد وارد شوید", "Token expired")
				return
			}
			// Other token validation errors
			if isAjax {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "خطای احراز هویت",
					"message": "توکن نامعتبر است",
					"redirect": "/error?code=401&message=" + url.QueryEscape("خطای احراز هویت") +
						"&description=" + url.QueryEscape("توکن نامعتبر است") +
						"&technical=" + url.QueryEscape(err.Error()),
				})
				return
			}
			HanleError(w, r, http.StatusUnauthorized, "خطای احراز هویت", "توکن نامعتبر است", err.Error())
			return
		}

		// Check token validity
		if !token.Valid {
			if isAjax {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "خطای احراز هویت",
					"message": "توکن نامعتبر است",
					"redirect": "/error?code=401&message=" + url.QueryEscape("خطای احراز هویت") +
						"&description=" + url.QueryEscape("توکن نامعتبر است") +
						"&technical=" + url.QueryEscape("Invalid token"),
				})
				return
			}
			HanleError(w, r, http.StatusUnauthorized, "خطای احراز هویت", "توکن نامعتبر است", "Invalid token")
			return
		}

		// Extract claims and proceed to next handler
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			ctx := context.WithValue(r.Context(), "user_id", claims["user_id"])
			next(w, r.WithContext(ctx))
			return
		} else {
			// Invalid claims
			if isAjax {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "خطای احراز هویت",
					"message": "توکن نامعتبر است",
					"redirect": "/error?code=401&message=" + url.QueryEscape("خطای احراز هویت") +
						"&description=" + url.QueryEscape("توکن نامعتبر است") +
						"&technical=" + url.QueryEscape("Invalid token claims"),
				})
				return
			}
			HanleError(w, r, http.StatusUnauthorized, "خطای احراز هویت", "توکن نامعتبر است", "Invalid token claims")
			return
		}
	}
}
