package parsers

import (
	"autobackcom/internal/models"
	"autobackcom/internal/utils"
)

// OKXResponseParser gom các hàm parse response từ OKX

type OKXResponseParser struct{}

func (OKXResponseParser) ParseOrder(market string, user *models.User, event map[string]interface{}) models.Order {
	var order models.Order
	if arg, ok := event["arg"].(map[string]interface{}); ok {
		if channel, ok := arg["channel"].(string); ok && channel == "orders" {
			data, ok := event["data"].([]interface{})
			if !ok || len(data) == 0 {
				return order
			}
			orderData, ok := data[0].(map[string]interface{})
			if !ok {
				return order
			}
			order = models.Order{
				ID:               utils.GetString(orderData["ordId"]),
				UserID:           user.ID,
				Market:           market,
				Symbol:           utils.GetString(orderData["instId"]),
				Status:           utils.GetString(orderData["state"]),
				Side:             utils.GetString(orderData["side"]),
				Type:             utils.GetString(orderData["ordType"]),
				Price:            utils.GetFloat(orderData["px"]),
				Quantity:         utils.GetFloat(orderData["sz"]),
				ExecutedQuantity: utils.GetFloat(orderData["fillSz"]),
				AvgPrice:         utils.GetFloat(orderData["avgPx"]),
				Time:             utils.GetTime(orderData["uTime"]),
				Commission:       utils.GetFloat(orderData["fee"]),
				CommissionAsset:  utils.GetString(orderData["feeCcy"]),
			}
		}
	}
	return order
}
