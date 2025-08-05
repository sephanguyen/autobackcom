package exchanges

import (
	"autobackcom/internal/models"
	"context"
	"time"
)

type ExchangeClient interface {
	FetchTrades(ctx context.Context, user *models.User, start, end time.Time) ([]models.Order, error)
}
