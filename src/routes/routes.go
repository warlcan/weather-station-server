package routes

import (
	"net/http"
	"weather-server/src/handler"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/app/auth/verify-token", handler.VerifyTokenHandler)

	mux.HandleFunc("POST /api/app/auth/send-code", handler.SendCodeHandler)
	mux.HandleFunc("POST /api/app/auth/verify-code", handler.VerifyCodeHandler)

	mux.HandleFunc("POST /api/gateway/init", handler.InitGatewayHandler)
	mux.HandleFunc("POST /api/gateway/list/update", handler.UpdateStationsListHandler)
	mux.HandleFunc("POST /api/gateway/measure", handler.StoreMeasureHandler)

	mux.HandleFunc("GET /api/app/stations", handler.ListStationsHandler)
	mux.HandleFunc("GET /api/app/measurements", handler.ListMeasurementsHandler)

	return mux
}
