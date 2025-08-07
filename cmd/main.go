package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"autobackcom/internal/cronjob"
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

	// Load env trước khi build container
	_ = godotenv.Load()

	c, err := di.BuildContainer()
	if err != nil {
		logrus.Fatal(err)
	}
	var client *mongo.Client
	var appHandlers *di.AppHandlers
	err = c.Invoke(func(cl *mongo.Client, ah *di.AppHandlers) {
		client = cl
		appHandlers = ah
	})
	if err != nil {
		logrus.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	http.HandleFunc("/register", appHandlers.RegisterHandler)
	http.Handle("/orders", appHandlers.GetOrdersHandler)
	http.HandleFunc("/fetch-trades-all-user", appHandlers.FetchAllTradesHandler)

	// Đăng ký cronjob lấy trade history định kỳ
	err = c.Invoke(func(ths *services.TradeHistoryService) {
		go func() {
			importCron := func() {
				cronjobPkg := "autobackcom/internal/cronjob"
				_ = cronjobPkg // chỉ để gợi ý import nếu IDE hỗ trợ
			}
			importCron() // chỉ để IDE gợi ý import nếu cần
			// Gọi hàm thực tế
			cronjob.StartTradeHistoryCron(context.Background(), ths)
		}()
	})
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
