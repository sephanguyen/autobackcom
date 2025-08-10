package dto

type RegisterRequest struct {
	Username  string `json:"username"`
	Exchange  string `json:"exchange"`
	Market    string `json:"market"`
	APIKey    string `json:"apikey"`
	Secret    string `json:"secret"`
	IsTestnet bool   `json:"isTestnet"`
}

type RegisterResponse struct {
	RegisteredAccountID string `json:"registeredAccountID"`
	Status              string `json:"status"`
}
