package model

import "time"

type SendCodeRequest struct {
	Email string `json:"email"`
}

type VerifeCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}
type VerifyCodeResponse struct {
	Token *string `json:"token"`
}

type AppWeatherStation struct {
	Uid        int64   `json:"uid"`
	HardwareId string  `json:"hardware_id"`
	Name       *string `json:"name"`
	ErrorMask  int16   `json:"error_mask"`
	GatewayUid string  `json:"gateway_uid"`
}

type ListStationsResponse struct {
	Stations []AppWeatherStation `json:"stations"`
}

type AppMeasurement struct {
	Time         time.Time `json:"time"`
	Temperature  *float32  `json:"temperature"`
	Humidity     *float32  `json:"humidity"`
	Pressure     *float32  `json:"pressure"`
	VoltageLevel *float32  `json:"voltage_level"`
}

type ListMeasurementsResponse struct {
	Measurements map[int64][]AppMeasurement `json:"measurements"`
}
