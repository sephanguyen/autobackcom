package di

import (
	"autobackcom/internal/api"
	"autobackcom/internal/exchanges"
	"autobackcom/internal/repositories"
	"autobackcom/internal/services"
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/dig"
)

type AppHandlers struct {
	RegisterHandler       http.HandlerFunc `name:"register"`
	GetOrdersHandler      http.HandlerFunc `name:"getOrders"`
	FetchAllTradesHandler http.HandlerFunc `name:"fetchAllTrades"`
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

// Provider cho UserRepository
func NewUserRepository(client *mongo.Client) *repositories.UserRepository {
	return repositories.NewUserRepository(client, "exchange_db", "users")
}

// Provider cho OrderRepository
func NewOrderRepository(client *mongo.Client) *repositories.OrderRepository {
	return repositories.NewOrderRepository(client, "exchange_db", "orders")
}

// Provider cho ExchangeService (nếu cần gom fetcher vào map)
type ExchangeServiceDeps struct {
	dig.In
	UserRepo              *repositories.UserRepository
	OrderRepo             *repositories.OrderRepository
	BinanceSpotFetcher    exchanges.ExchangeFetcher `name:"binanceSpot"`
	BinanceFuturesFetcher exchanges.ExchangeFetcher `name:"binanceFutures"`
}

// Provider cho RegisterHandler
func NewRegisterHandler(userRepo *repositories.UserRepository, tradeHistoryService *services.TradeHistoryService) http.HandlerFunc {
	return api.RegisterHandler(userRepo, tradeHistoryService)
}

// Provider cho GetOrdersHandler
func NewGetOrdersHandler(userRepo *repositories.UserRepository, orderRepo *repositories.OrderRepository) http.HandlerFunc {
	return api.GetOrdersHandler(userRepo, orderRepo)
}

func NewFetchAllTradeOfUsersHandler(tradeHistoryService *services.TradeHistoryService) http.HandlerFunc {
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
	c.Provide(NewUserRepository)
	c.Provide(NewOrderRepository)
	c.Provide(services.NewClientManagerService)
	c.Provide(func(userRepo *repositories.UserRepository, orderRepo *repositories.OrderRepository, clientManager *services.ClientManagerService) *services.TradeHistoryService {
		return services.NewTradeHistoryService(userRepo, orderRepo, clientManager)
	})
	c.Provide(NewRegisterHandler, dig.Name("register"))
	c.Provide(NewGetOrdersHandler, dig.Name("getOrders"))
	c.Provide(NewFetchAllTradeOfUsersHandler, dig.Name("fetchAllTrades"))
	type appHandlerIn struct {
		dig.In
		RegisterHandler       http.HandlerFunc `name:"register"`
		GetOrdersHandler      http.HandlerFunc `name:"getOrders"`
		FetchAllTradesHandler http.HandlerFunc `name:"fetchAllTrades"`
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
