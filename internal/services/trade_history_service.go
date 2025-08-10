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
	registeredAccountRepository *repositories.RegisteredAccountRepository
	orderRepository             *repositories.OrderRepository
	clientManager               *ClientManagerService
}

func NewTradeHistoryService(registeredAccountRepository *repositories.RegisteredAccountRepository, orderRepository *repositories.OrderRepository, clientManager *ClientManagerService) *TradeHistoryService {
	return &TradeHistoryService{
		registeredAccountRepository: registeredAccountRepository,
		orderRepository:             orderRepository,
		clientManager:               clientManager,
	}
}

// Cronjob function

func (s *TradeHistoryService) FetchAllAccountTradeHistory(ctx context.Context) error {
	accounts, err := s.registeredAccountRepository.GetAllRegisteredAccounts(ctx)
	if err != nil {
		return err
	}

	accountsPoolSize := 5
	accountsPool := make(chan struct{}, accountsPoolSize)
	var accountWg sync.WaitGroup

	for _, account := range accounts {
		accountsPool <- struct{}{}
		accountWg.Add(1)
		go func(accountCopy models.RegisteredAccount) {
			defer func() {
				<-accountsPool
				accountWg.Done()
			}()
			_ = s.FetchAllTradeHistory(ctx, accountCopy)
		}(account)
	}
	accountWg.Wait()
	return nil
}

// Fetch trade history cho một account
func (s *TradeHistoryService) FetchAllTradeHistory(ctx context.Context, account models.RegisteredAccount) error {
	var start time.Time
	latestOrder, err := s.orderRepository.GetLatestOrder(ctx, account.ID, account.Exchange, account.Market)
	if err == nil && latestOrder != nil && !latestOrder.Time.IsZero() {
		start = latestOrder.Time
	}
	s.handleAccountTradeHistory(ctx, account, start)
	return nil
}

func (s *TradeHistoryService) handleAccountTradeHistory(ctx context.Context, account models.RegisteredAccount, start time.Time) {
	clientsInfo, err := s.clientManager.GetOrCreateClient(account)
	if err != nil {
		log.Printf("Get client error %s: %v", account.Username, err)
		return
		// Fetch trade history cho một account
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
			s.handleClientTradeHistory(ctx, clientCopy, account, start)
		}(client)
	}
	clientWg.Wait()
}

func (s *TradeHistoryService) handleClientTradeHistory(ctx context.Context, client exchanges.ExchangeFetcher, account models.RegisteredAccount, start time.Time) {
	orders, err := client.FetchTrades(ctx, account.ID, start)
	if err != nil {
		log.Println("FetchOrders error for account", account.Username, ":", err)
		return
	}
	err = s.orderRepository.SaveOrders(orders)
	if err != nil {
		log.Println("Save orders error:", err)
	}
}
