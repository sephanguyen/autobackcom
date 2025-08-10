package dto

type GetOrdersRequest struct {
	RegisteredAccountID string `json:"registeredAccountID"`
}

type GetOrdersResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}
