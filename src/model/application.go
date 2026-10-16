package model

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
