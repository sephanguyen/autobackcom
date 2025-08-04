package exchanges

import (
	"autobackcom/internal/models"
	"context"
	"fmt"

	"github.com/adshao/go-binance/v2"
)

type BinanceExchange struct {
	client *binance.Client
}

func NewBinanceExchange(apiKey, secret string) *BinanceExchange {
	client := binance.NewClient(apiKey, secret)
	return &BinanceExchange{client: client}
}

func (b *BinanceExchange) GetOrders() ([]models.Order, error) {
	orders, err := b.client.NewListOrdersService().Do(context.Background())
	if err != nil {
		return nil, err
	}
	var internalOrders []models.Order
	for _, o := range orders {
		internalOrders = append(internalOrders, models.Order{
			ID:     fmt.Sprintf("%d", o.OrderID),
			Symbol: o.Symbol,
			Status: string(o.Status),
		})
	}
	return internalOrders, nil
}
