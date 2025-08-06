package services

import (
	"autobackcom/internal/repositories"
	"context"
	"log"
	"time"
)

type TradeHistoryService struct {
	userRepository  *repositories.UserRepository
	orderRepository *repositories.OrderRepository
	clientManager   *ClientManagerService
}

func NewTradeHistoryService(userRepository *repositories.UserRepository, orderRepository *repositories.OrderRepository, clientManager *ClientManagerService) *TradeHistoryService {
	return &TradeHistoryService{
		userRepository:  userRepository,
		orderRepository: orderRepository,
		clientManager:   clientManager,
	}
}

// Cronjob function
func (s *TradeHistoryService) FetchAllUserTradeHistory(ctx context.Context, start, end time.Time) error {
	users, err := s.userRepository.GetAllUsers(ctx)
	if err != nil {
		return err
	}

	for _, user := range users {
		clientsInfo, err := s.clientManager.GetOrCreateClient(user)
		if err != nil {
			log.Printf("Get client error %s: %v", user.Username, err)
			continue
		}
		for _, client := range clientsInfo.Clients {
			orders, err := client.FetchTrades(ctx, user.ID, start, end)
			if err != nil {
				log.Println("FetchOrders error for user", user.Username, ":", err)
				continue
			}
			// Save orders to DB
			err = s.orderRepository.SaveOrders(orders)
			if err != nil {
				log.Println("Save orders error:", err)
			}
		}

	}
	return nil
}
