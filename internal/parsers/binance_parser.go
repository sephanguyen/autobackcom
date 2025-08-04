package parsers

import (
	"autobackcom/internal/models"
	"autobackcom/internal/utils"
	"fmt"
)

// ExchangeEventParser là interface cho mọi parser của các sàn
type ExchangeEventParser interface {
	ParseEvent(user *models.User, event map[string]interface{}) (interface{}, string)
}

// BinanceResponseParser implements ExchangeEventParser
type BinanceResponseParser struct{}

func (BinanceResponseParser) ParseEvent(user *models.User, event map[string]interface{}) (interface{}, string) {
	eventType, _ := event["EventType"]
	eventTypeStr := fmt.Sprintf("%v", eventType)
	switch eventTypeStr {
	case "ORDER_TRADE_UPDATE":
		return BinanceResponseParser{}.ParseFuturesOrder(user, event), "order"
	case "TRADE_LITE":
		return BinanceResponseParser{}.ParseTrade(user, event), "trade"
	default:
		return nil, ""
	}
}

func (BinanceResponseParser) ParseOrder(market string, user *models.User, event map[string]interface{}) models.Order {
	if market == "spot" {
		return BinanceResponseParser{}.ParseSpotOrder(user, event)
	}
	if market == "futures" {
		return BinanceResponseParser{}.ParseFuturesOrder(user, event)
	}
	return models.Order{}
}

func (BinanceResponseParser) ParseSpotOrder(user *models.User, event map[string]interface{}) models.Order {
	eventType, ok := event["EventType"].(string)
	if !ok || eventType != "executionReport" {
		return models.Order{}
	}
	data, ok := event["Data"].(map[string]interface{})
	if !ok {
		return models.Order{}
	}
	return models.Order{
		ID:               utils.GetString(data["i"]),
		UserID:           user.ID,
		Market:           "spot",
		Symbol:           utils.GetString(data["s"]),
		Status:           utils.GetString(data["X"]),
		Side:             utils.GetString(data["S"]),
		Type:             utils.GetString(data["o"]),
		Price:            utils.GetFloat(data["p"]),
		Quantity:         utils.GetFloat(data["q"]),
		ExecutedQuantity: utils.GetFloat(data["z"]),
		AvgPrice:         utils.GetFloat(data["ap"]),
		Time:             utils.GetTime(data["T"]),
		Commission:       utils.GetFloat(data["n"]),
		CommissionAsset:  utils.GetString(data["N"]),
	}
}

func (BinanceResponseParser) ParseFuturesOrder(user *models.User, event map[string]interface{}) models.Order {

	data, ok := event["Data"].(map[string]interface{})
	fmt.Printf("FuturesOrder event data: %+v\n", event)
	if !ok {
		return models.Order{}
	}
	orderData, ok := data["o"].(map[string]interface{})
	if !ok {
		return models.Order{}
	}
	return models.Order{
		ID:               utils.GetString(orderData["i"]),
		UserID:           user.ID,
		Market:           "futures",
		Symbol:           utils.GetString(orderData["s"]),
		Status:           utils.GetString(orderData["X"]),
		Side:             utils.GetString(orderData["S"]),
		Type:             utils.GetString(orderData["o"]),
		Price:            utils.GetFloat(orderData["p"]),
		Quantity:         utils.GetFloat(orderData["q"]),
		ExecutedQuantity: utils.GetFloat(orderData["z"]),
		AvgPrice:         utils.GetFloat(orderData["ap"]),
		Time:             utils.GetTime(orderData["T"]),
		Commission:       utils.GetFloat(orderData["n"]),
		CommissionAsset:  utils.GetString(orderData["N"]),
	}
}

func (BinanceResponseParser) ParseTrade(user *models.User, event map[string]interface{}) models.Trade {
	// Giả sử event["Data"] là map chứa thông tin trade
	data, ok := event["Data"].(map[string]interface{})
	if !ok {
		return models.Trade{}
	}
	return models.Trade{
		ID:       utils.GetString(data["t"]),
		UserID:   user.ID,
		Symbol:   utils.GetString(data["s"]),
		Price:    utils.GetFloat(data["p"]),
		Quantity: utils.GetFloat(data["q"]),
		Time:     utils.GetTime(data["T"]),
		IsBuyer:  utils.GetBool(data["m"]),
		IsMaker:  utils.GetBool(data["M"]),
	}
}
