package exchanges

import (
	"autobackcom/internal/models"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/adshao/go-binance/v2"
)

type BinanceExchange struct {
	client *binance.Client
}

func NewBinanceExchange(apiKey, secret string) *BinanceExchange {
	client := binance.NewClient(apiKey, secret)
	return &BinanceExchange{client: client}
}

func (b *BinanceExchange) FetchTrades(ctx context.Context, user *models.User, start, end time.Time) ([]models.Order, error) {
	// Giả sử Client đã được khởi tạo với API key/secret đã giải mã từ user
	svc := b.client.NewListTradesService()
	// Có thể cần truyền thêm symbol, startTime, endTime nếu muốn lọc theo thời gian
	// svc.Symbol("BTCUSDT").StartTime(start.UnixMilli()).EndTime(end.UnixMilli())
	trades, err := svc.Do(ctx)
	if err != nil {
		log.Printf("Failed to fetch Binance spot trade history for user %s: %v", user.Username, err)
		return nil, err
	}
	var orders []models.Order
	for _, trade := range trades {

		order := models.Order{
			ID:              fmt.Sprintf("%d", trade.ID),
			UserID:          user.ID,
			Symbol:          trade.Symbol,
			OrderID:         trade.OrderID,
			OrderListId:     trade.OrderListId,
			Price:           trade.Price,
			Quantity:        trade.Quantity,
			QuoteQuantity:   trade.QuoteQuantity,
			Commission:      trade.Commission,
			CommissionAsset: trade.CommissionAsset,
			Time:            time.UnixMilli(trade.Time),
		}
		orders = append(orders, order)
	}
	return orders, nil
}
