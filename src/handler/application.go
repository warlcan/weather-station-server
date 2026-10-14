package handler

import (
	"net/http"
	"weather-server/src/services"
)

func VerifyTokenHandler(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("Auth")

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
