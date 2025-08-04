package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"autobackcom/internal/api"
	"autobackcom/internal/services"

	"autobackcom/internal/di"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/sirupsen/logrus"
)

func main() {
	// Cấu hình logger
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	levelStr := os.Getenv("LOG_LEVEL")
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(level)
	}

	// Khởi tạo ứng dụng với dig
	c, err := di.BuildContainer()
	_ = godotenv.Load()
	if err != nil {
		logrus.Fatal(err)
	}
	var client *mongo.Client
	var exchangeService *services.ExchangeService
	var appHandlers di.AppHandlers
	err = c.Invoke(func(cl *mongo.Client, es *services.ExchangeService, ah di.AppHandlers) {
		client = cl
		exchangeService = es
		appHandlers = ah
	})
	if err != nil {
		logrus.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	exchangeService.StartAllUserStreams()
	http.HandleFunc("/register", appHandlers.RegisterHandler)
	http.Handle("/orders", api.JWTAuthMiddleware(http.HandlerFunc(appHandlers.GetOrdersHandler)))
	logrus.Info("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
