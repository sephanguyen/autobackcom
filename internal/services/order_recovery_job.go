package services

import (
	"autobackcom/internal/exchanges"
	"autobackcom/internal/models"
	"autobackcom/internal/repositories"
	"context"
	"time"
)

type OrderRecoveryJob struct {
	UserRepo  *repositories.UserRepository
	OrderRepo *repositories.OrderRepository
	Streamers map[string]map[string]exchanges.StreamHandler // sàn/market
}

func (j *OrderRecoveryJob) Run(ctx context.Context) {
	users, _ := j.UserRepo.GetAllUsers()
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	for i := range users {
		user := &users[i]
		for exchange, markets := range j.Streamers {
			for market, handler := range markets {
				go func(u *models.User, ex, mk string, h exchanges.StreamHandler) {
					orders, err := h.FetchOrders(ctx, u, start, end)
					if err == nil {
						for _, order := range orders {
							if !j.OrderRepo.Exists(order.ID, u.ID.Hex(), ex, mk, order.Time) {
								j.OrderRepo.SaveOrder(models.Order{
									ID:       order.ID,
									UserID:   u.ID,
									Exchange: ex,
									Market:   mk,
									Time:     order.Time,
									// Symbol, Status, Side, Type, Price, Quantity, ExecutedQuantity, AvgPrice, Commission, CommissionAsset: map nếu exchanges.Order có các trường này
								})
							}
						}
					}
				}(user, exchange, market, handler)
			}
		}
	}
}
