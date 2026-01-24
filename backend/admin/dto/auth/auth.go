package auth

type AuthResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    AuthData `json:"data"`
}
type AuthData struct {
	PartnerID  int64  `json:"partnerID"`
	PartnerKey string `json:"partnerKey"`
	Redirect   string `json:"redirect"`
	IsSandbox  bool   `json:"isSandbox"`
}
