/*
@title Auto Backcom API
@version 1.0
@description API cho hệ thống Auto Backcom
@host localhost:8080
@BasePath /
@schemes http https
*/
package main

import (
	"context"
	"log"
	"os"

	_ "autobackcom/docs" // import docs để swagger serve được
	"autobackcom/internal/cronjob"
	"autobackcom/internal/di"
	"autobackcom/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.mongodb.org/mongo-driver/mongo"
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

	r := gin.Default()
	r.POST("/register", appHandlers.RegisterHandler)
	r.POST("/orders", appHandlers.GetOrdersHandler)
	r.POST("/fetch-trades-all-user", appHandlers.FetchAllTradesHandler)
	r.GET("/swagger/*any", gin.WrapF(httpSwagger.WrapHandler))
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
	log.Fatal(r.Run(":8080"))
}
