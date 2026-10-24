package model

import "time"

type InitGatewayRequest struct {
	GatewayHardwareId string `json:"GatewayHardwareId"`
	UserApiToken      string `json:"UserApiToken"`
}

type InitGatewayResponse struct {
	GatewayUid string `json:"gateway_uid"`
}

type UpdateStationsListRequest struct {
	GatewayHardwareId        string   `json:"GatewayHardwareId"`
	MetheostationHardwareIds []string `json:"MetheostationHardwareIds"`
}

type UpdateStationsListResponse struct {
	Added int `json:"added"`
	Total int `json:"total"`
}

type MeasureRequest struct {
	GatewayHardwareId        string    `json:"GatewayHardwareId"`
	WeatherStationHardwareId string    `json:"WeatherStationHardwareId"`
	Time                     time.Time `json:"Time"`
	Temperature              float32   `json:"Temperature"`
	Humidity                 float32   `json:"Humidity"`
	Pressure                 float32   `json:"Pressure"`
	VoltageLevel             float32   `json:"VoltageLevel"`
	ErrorMask                int16     `json:"ErrorMask"`
}

type MeasureResponse struct {
	Ok bool `json:"ok"`
}
