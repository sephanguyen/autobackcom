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
	RegisterHandler  http.HandlerFunc
	GetOrdersHandler http.HandlerFunc
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

// Provider cho các StreamHandler (dùng interface để DI chuẩn)
func ProvideBinanceSpotStreamHandler() exchanges.StreamHandler {
	return &exchanges.BinanceSpotStreamHandler{}
}

func ProvideBinanceFuturesStreamHandler() exchanges.StreamHandler {
	return &exchanges.BinanceFuturesStreamHandler{}
}

// Provider cho ExchangeService (chuẩn hóa DI, gom các handler vào map)
type ExchangeServiceDeps struct {
	dig.In
	UserRepo              *repositories.UserRepository
	OrderRepo             *repositories.OrderRepository
	BinanceSpotHandler    exchanges.StreamHandler `name:"binanceSpot"`
	BinanceFuturesHandler exchanges.StreamHandler `name:"binanceFutures"`
}

func NewExchangeService(deps ExchangeServiceDeps) *services.ExchangeService {
	streamHandlers := map[string]map[string]exchanges.StreamHandler{
		"binance": {
			"spot":    deps.BinanceSpotHandler,
			"futures": deps.BinanceFuturesHandler,
		},
	}
	fmt.Println("Creating ExchangeService with handlers:", streamHandlers)
	return services.NewExchangeService(deps.UserRepo, deps.OrderRepo, streamHandlers)
}

// Provider cho RegisterHandler
func NewRegisterHandler(userRepo *repositories.UserRepository, exchangeService *services.ExchangeService) http.HandlerFunc {
	return api.RegisterHandler(userRepo, exchangeService)
}

// Provider cho GetOrdersHandler
func NewGetOrdersHandler(userRepo *repositories.UserRepository, orderRepo *repositories.OrderRepository) http.HandlerFunc {
	return api.GetOrdersHandler(userRepo, orderRepo)
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
	c.Provide(func() exchanges.StreamHandler { return &exchanges.BinanceSpotStreamHandler{} }, dig.Name("binanceSpot"))
	c.Provide(func() exchanges.StreamHandler { return &exchanges.BinanceFuturesStreamHandler{} }, dig.Name("binanceFutures"))
	c.Provide(NewExchangeService)
	c.Provide(NewRegisterHandler)
	c.Provide(NewGetOrdersHandler)
	c.Provide(func(rh http.HandlerFunc, goh http.HandlerFunc) AppHandlers {
		return AppHandlers{RegisterHandler: rh, GetOrdersHandler: goh}
	})
	return c, nil
}
