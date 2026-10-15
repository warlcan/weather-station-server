package handler

import (
	"net/http"
	"weather-server/src/services"
)

func VerifyTokenHandler(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("Token")

	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !services.IsTokenValid(token) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func SendCodeHandler(w http.ResponseWriter, r *http.Request) {
	email := r.Header.Get("Email")

	if email == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if services.SendCode(email) {

	}
}
