package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"weather-server/src/model"
	"weather-server/src/services"
)

func VerifyTokenHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	if !services.IsTokenValid(token) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func SendCodeHandler(w http.ResponseWriter, r *http.Request) {
	var req model.SendCodeRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if services.SendCode(req.Email) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
