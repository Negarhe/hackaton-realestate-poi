package utils

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"text/template"
)

type ErrorPageData struct {
	Title          string
	Message        string
	Description    string
	StatusCode     int
	TechnicalError string
}

func HanleError(w http.ResponseWriter, r *http.Request, statusCode int, message string, description string, technicalError string) {
	errorURL := fmt.Sprintf("/error?code=%d&message=%s&description=%s&technical=%s",
		statusCode,
		url.QueryEscape(message),
		url.QueryEscape(description),
		url.QueryEscape(technicalError),
	)
	log.Println("27 error handler.go")
	http.Redirect(w, r, errorURL, http.StatusSeeOther)
	return
}

func RenderErrorPage(w http.ResponseWriter, r *http.Request) {
	log.Println("Rendering")
	statusCodestr := r.URL.Query().Get("code")
	message := r.URL.Query().Get("message")
	description := r.URL.Query().Get("description")
	technicalError := r.URL.Query().Get("technical")
	statusCode, err := strconv.Atoi(statusCodestr)
	if err != nil {
		log.Printf("couldn't convert status code to int: %v", err)
		statusCode = http.StatusInternalServerError
	}

	w.WriteHeader(statusCode)

	title := "خطا"
	switch statusCode {
	case http.StatusBadRequest:
		title = "درخواست نامعتبر"
	case http.StatusInternalServerError:
		title = "خطای سیستمی"
	case http.StatusUnauthorized:
		title = "خطای احراز هویت"
	case http.StatusForbidden:
		title = "دسترسی غیرمجاز"
	case http.StatusNotFound:
		title = "صفحه یافت نشد"
	}

	data := ErrorPageData{
		Title:          title,
		Message:        message,
		Description:    description,
		StatusCode:     statusCode,
		TechnicalError: technicalError,
	}

	errorTemplate, err := template.ParseFiles("./web/error.html")
	if err != nil {
		log.Println("could not parse error html" + err.Error())
		return
	}

	err = errorTemplate.Execute(w, data)
	if err != nil {
		log.Printf("Error rendering error template: %v", err)
		http.Error(w, "خطا در نمایش صفحه: "+message, http.StatusInternalServerError)
	}
	log.Println("79 error")
	return
}
