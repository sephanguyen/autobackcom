package exchanges

import (
	"autobackcom/internal/models"
	"context"
	"time"
)

type StreamHandler interface {
	Authenticate(ctx context.Context, apiKey, secret string) (string, error)
	Connect(ctx context.Context, token string) error
	Listen(ctx context.Context, handler func(event map[string]interface{})) error
	Close() error
	FetchOrders(ctx context.Context, user *models.User, start, end time.Time) ([]models.Order, error)
}
