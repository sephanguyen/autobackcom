package exchanges

import (
	"autobackcom/internal/models"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ExchangeClient interface {
	FetchTrades(ctx context.Context, userID primitive.ObjectID, start, end time.Time) ([]models.Order, error)
}
