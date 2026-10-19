package model

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
