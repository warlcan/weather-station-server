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
	if !services.VerifyToken(token) {
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

func VerifyCodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req model.VerifeCodeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Email == "" || req.Code == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(model.VerifyCodeResponse{Token: nil})
		return
	}

	defer r.Body.Close()

	if services.VerifyCode(req.Email, req.Code) {
		generatedToken := services.GenerateUserToken(req.Email)

		response := model.VerifyCodeResponse{
			Token: &generatedToken,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	} else {
		response := model.VerifyCodeResponse{
			Token: nil,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
	}
}
