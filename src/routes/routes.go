package routes

import (
	"net/http"
	"weather-server/src/handler"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/app/auth/verify-token", handler.VerifyTokenHandler)
	mux.HandleFunc("GET /api/app/auth/send-code", handler.SendCodeHandler)

	return mux
}
