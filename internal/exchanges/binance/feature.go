package binance

import (
	"autobackcom/internal/models"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BinanceFeatureExchange struct {
	client *futures.Client
}

func NewBinanceFetureExchange(apiKey, secret string, isTestnet bool) *BinanceFeatureExchange {
	futures.UseTestnet = isTestnet
	client := futures.NewClient(apiKey, secret)
	return &BinanceFeatureExchange{client: client}
}

func (b *BinanceFeatureExchange) FetchTrades(ctx context.Context, registedAccountID primitive.ObjectID, start time.Time) ([]models.Order, error) {
	// Giả sử Client đã được khởi tạo với API key/secret đã giải mã từ user
	svc := b.client.NewListAccountTradeService()
	// Có thể cần truyền thêm symbol, startTime, endTime nếu muốn lọc theo thời gian
	// svc.Symbol("BTCUSDT").StartTime(start.UnixMilli()).EndTime(end.UnixMilli())
	if !start.IsZero() {
		svc.StartTime(start.UnixMilli())
	}
	svc.Limit(1000)
	trades, err := svc.Do(ctx)
	if err != nil {
		log.Printf("Failed to fetch Binance spot trade history for user %s: %v", registedAccountID, err)
		return nil, err
	}
	var orders []models.Order
	for _, trade := range trades {

		order := models.Order{
			ID:                  fmt.Sprintf("%d", trade.ID),
			RegisteredAccountID: registedAccountID,
			Symbol:              trade.Symbol,
			OrderID:             trade.OrderID,
			Price:               trade.Price,
			Quantity:            trade.Quantity,
			QuoteQuantity:       trade.QuoteQuantity,
			Commission:          trade.Commission,
			CommissionAsset:     trade.CommissionAsset,
			Time:                time.UnixMilli(trade.Time),
			Exchange:            "binance",
			Market:              "futures",
			Side:                string(trade.Side),
			PositionSide:        string(trade.PositionSide),
		}
		orders = append(orders, order)
	}
	return orders, nil
}
