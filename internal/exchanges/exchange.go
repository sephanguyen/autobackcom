package exchanges

import "autobackcom/internal/models"

type Exchange interface {
	GetOrders() ([]models.Order, error)
}
