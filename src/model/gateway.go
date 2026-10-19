package model

type InitGatewayRequest struct {
	GatewayHardwareId string `json:"GatewayHardwareId"`
	UserApiToken      string `json:"UserApiToken"`
}

type InitGatewayResponse struct {
	GatewayUid string `json:"gateway_uid"`
}
