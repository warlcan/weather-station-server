package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"

	"weather-server/src/model"
	"weather-server/src/repository"
	"weather-server/src/services"
)

var gatewayService = services.NewGatewayService()

func InitGatewayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.InitGatewayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.GatewayHardwareId == "" || req.UserApiToken == "" {
		http.Error(w, "GatewayHardwareId and UserApiToken are required", http.StatusBadRequest)
		return
	}

	if _, err := uuid.Parse(req.GatewayHardwareId); err != nil {
		http.Error(w, "GatewayHardwareId must be a valid UUID", http.StatusBadRequest)
		return
	}

	resp, err := gatewayService.InitGateway(&req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidToken):
			http.Error(w, "invalid user token", http.StatusUnauthorized)
		case errors.Is(err, repository.ErrGatewayAlreadyBound):
			http.Error(w, "gateway already bound to another user", http.StatusConflict)
		default:
			log.Printf("InitGateway error: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func UpdateStationsListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.UpdateStationsListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.GatewayHardwareId == "" {
		http.Error(w, "GatewayHardwareId is required", http.StatusBadRequest)
		return
	}

	if _, err := uuid.Parse(req.GatewayHardwareId); err != nil {
		http.Error(w, "GatewayHardwareId must be a valid UUID", http.StatusBadRequest)
		return
	}

	resp, err := gatewayService.UpdateStationsList(&req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrGatewayNotFound):
			http.Error(w, "gateway not found", http.StatusNotFound)
		case errors.Is(err, repository.ErrStationAlreadyBound):
			http.Error(w, "station already bound to another gateway", http.StatusConflict)
		case errors.Is(err, services.ErrInvalidStationId):
			http.Error(w, "invalid station hardware id", http.StatusBadRequest)
		default:
			log.Printf("UpdateStationsList error: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
