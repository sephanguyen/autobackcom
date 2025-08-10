package di

import (
	"autobackcom/internal/api"
	"autobackcom/internal/exchanges"
	"autobackcom/internal/repositories"
	"autobackcom/internal/services"
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/dig"
)

type AppHandlers struct {
	RegisterHandler       gin.HandlerFunc `name:"register"`
	GetOrdersHandler      gin.HandlerFunc `name:"getOrders"`
	FetchAllTradesHandler gin.HandlerFunc `name:"fetchAllTrades"`
}

// Provider cho MongoDB client
func NewMongoClient(uri string) (*mongo.Client, error) {
	fmt.Println("Connecting to MongoDB at", uri)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		logrus.WithField("error", err).Fatal("Failed to connect to MongoDB")
		return nil, err
	}
	return client, nil
}

// Provider cho RegisteredAccountRepository
func NewRegisteredAccountRepository(client *mongo.Client) *repositories.RegisteredAccountRepository {
	return repositories.NewRegisteredAccountRepository(client, "exchange_db", "registered_accounts")
}

// Provider cho OrderRepository
func NewOrderRepository(client *mongo.Client) *repositories.OrderRepository {
	return repositories.NewOrderRepository(client, "exchange_db", "orders")
}

// Provider cho ExchangeService (nếu cần gom fetcher vào map)
type ExchangeServiceDeps struct {
	dig.In
	RegisteredAccountRepo *repositories.RegisteredAccountRepository
	OrderRepo             *repositories.OrderRepository
	BinanceSpotFetcher    exchanges.ExchangeFetcher `name:"binanceSpot"`
	BinanceFuturesFetcher exchanges.ExchangeFetcher `name:"binanceFutures"`
}

// Provider cho RegisterHandler
func NewRegisterHandler(accountRepo *repositories.RegisteredAccountRepository, tradeHistoryService *services.TradeHistoryService) gin.HandlerFunc {
	return api.RegisterHandler(accountRepo, tradeHistoryService)
}

// Provider cho GetOrdersHandler
func NewGetOrdersHandler(accountRepo *repositories.RegisteredAccountRepository, orderRepo *repositories.OrderRepository) gin.HandlerFunc {
	return api.GetOrdersHandler(accountRepo, orderRepo)
}

func NewFetchAllTradeOfUsersHandler(tradeHistoryService *services.TradeHistoryService) gin.HandlerFunc {
	return api.FetchAllTradesForUser(tradeHistoryService)
}

func BuildContainer() (*dig.Container, error) {
	c := dig.New()
	c.Provide(func() string {
		uri := os.Getenv("MONGODB_URI")
		if uri == "" {
			logrus.Fatal("MONGODB_URI is not set in environment")
		}
		return uri
	})
	c.Provide(NewMongoClient)
	c.Provide(NewRegisteredAccountRepository)
	c.Provide(NewOrderRepository)
	c.Provide(services.NewClientManagerService)
	c.Provide(func(accountRepo *repositories.RegisteredAccountRepository, orderRepo *repositories.OrderRepository, clientManager *services.ClientManagerService) *services.TradeHistoryService {
		return services.NewTradeHistoryService(accountRepo, orderRepo, clientManager)
	})
	c.Provide(NewRegisterHandler, dig.Name("register"))
	c.Provide(NewGetOrdersHandler, dig.Name("getOrders"))
	c.Provide(NewFetchAllTradeOfUsersHandler, dig.Name("fetchAllTrades"))
	type appHandlerIn struct {
		dig.In
		RegisterHandler       gin.HandlerFunc `name:"register"`
		GetOrdersHandler      gin.HandlerFunc `name:"getOrders"`
		FetchAllTradesHandler gin.HandlerFunc `name:"fetchAllTrades"`
	}
	c.Provide(func(in appHandlerIn) *AppHandlers {
		return &AppHandlers{
			RegisterHandler:       in.RegisterHandler,
			GetOrdersHandler:      in.GetOrdersHandler,
			FetchAllTradesHandler: in.FetchAllTradesHandler,
		}
	})
	// Đăng ký cleanup cho ClientManagerService
	err := c.Invoke(func(cm *services.ClientManagerService) {
		// defer sẽ chạy khi main kết thúc
		go func() {
			<-context.Background().Done() // hoặc dùng signal thực tế nếu có
			cm.Clean()
		}()
	})
	if err != nil {
		return c, err
	}
	return c, nil
}
