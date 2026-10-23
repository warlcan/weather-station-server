package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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

	token, err := services.SignInOrRegister(req.Email, req.Code)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(model.VerifyCodeResponse{Token: nil})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.VerifyCodeResponse{
		Token: &token,
	})
}

func ListStationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	userUid, err := services.GetUserIdByToken(token)
	if err != nil || userUid == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	stations, err := services.ListUserStations(userUid)
	if err != nil {
		log.Printf("ListStations error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(model.ListStationsResponse{Stations: stations})
}

func ListMeasurementsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	userUid, err := services.GetUserIdByToken(token)
	if err != nil || userUid == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idsParam := r.URL.Query().Get("station_uids")
	if idsParam == "" {
		http.Error(w, "station_uids is required", http.StatusBadRequest)
		return
	}

	parts := strings.Split(idsParam, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			http.Error(w, "invalid station_uids", http.StatusBadRequest)
			return
		}
		ids = append(ids, v)
	}

	result, err := services.ListMeasurements(userUid, ids)
	if err != nil {
		log.Printf("ListMeasurements error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(model.ListMeasurementsResponse{Measurements: result})
}
