package services

import (
	"autobackcom/internal/exchanges"
	"autobackcom/internal/models"
	"autobackcom/internal/repositories"
	"context"
	"log"
	"sync"
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

func (s *TradeHistoryService) FetchAllUserTradeHistory(ctx context.Context) error {
	users, err := s.userRepository.GetAllUsers(ctx)
	if err != nil {
		return err
	}

	userPoolSize := 5
	userPool := make(chan struct{}, userPoolSize)
	var userWg sync.WaitGroup

	for _, user := range users {
		userPool <- struct{}{}
		userWg.Add(1)
		go func(userCopy models.User) {
			defer func() {
				<-userPool
				userWg.Done()
			}()
			// Lấy order mới nhất cho user-exchange-market
			latestOrder, err := s.orderRepository.GetLatestOrder(ctx, userCopy.ID, userCopy.Exchange, userCopy.Market)
			var start time.Time
			if err == nil && latestOrder != nil && !latestOrder.Time.IsZero() {
				start = latestOrder.Time
			}
			s.handleUserTradeHistory(ctx, userCopy, start)
		}(user)
	}
	userWg.Wait()
	return nil
}

func (s *TradeHistoryService) handleUserTradeHistory(ctx context.Context, user models.User, start time.Time) {
	clientsInfo, err := s.clientManager.GetOrCreateClient(user)
	if err != nil {
		log.Printf("Get client error %s: %v", user.Username, err)
		return
	}
	clientPoolSize := 3
	clientPool := make(chan struct{}, clientPoolSize)
	var clientWg sync.WaitGroup
	for _, client := range clientsInfo.Clients {
		clientPool <- struct{}{}
		clientWg.Add(1)
		go func(clientCopy exchanges.ExchangeFetcher) {
			defer func() {
				<-clientPool
				clientWg.Done()
			}()
			s.handleClientTradeHistory(ctx, clientCopy, user, start)
		}(client)
	}
	clientWg.Wait()
}

func (s *TradeHistoryService) handleClientTradeHistory(ctx context.Context, client exchanges.ExchangeFetcher, user models.User, start time.Time) {
	orders, err := client.FetchTrades(ctx, user.ID, start)
	if err != nil {
		log.Println("FetchOrders error for user", user.Username, ":", err)
		return
	}
	err = s.orderRepository.SaveOrders(orders)
	if err != nil {
		log.Println("Save orders error:", err)
	}
}
