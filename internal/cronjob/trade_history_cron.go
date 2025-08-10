package cronjob

import (
	"autobackcom/internal/services"
	"context"
	"log"
	"os"
	"strconv"

	"github.com/robfig/cron/v3"
)

// StartTradeHistoryCron sẽ chạy hàm FetchAllUserTradeHistory mỗi 30 phút
func StartTradeHistoryCron(ctx context.Context, tradeHistoryService *services.TradeHistoryService) {
	// Đọc số phút từ env, mặc định 30
	minuteStr := os.Getenv("TRADE_HISTORY_CRON_MINUTES")
	minutes := 30
	if minuteStr != "" {
		if m, err := strconv.Atoi(minuteStr); err == nil && m > 0 {
			minutes = m
		}
	}

	// Chạy ngay khi khởi động
	err := tradeHistoryService.FetchAllAccountTradeHistory(ctx)
	if err != nil {
		log.Println("[CRON] Immediate FetchAllAccountTradeHistory error:", err)
	}

	// Đặt lịch chạy mỗi N phút
	c := cron.New()
	cronSpec := "@every " + strconv.Itoa(minutes) + "m"
	c.AddFunc(cronSpec, func() {
		err := tradeHistoryService.FetchAllAccountTradeHistory(ctx)
		if err != nil {
			log.Println("[CRON] FetchAllUserTradeHistory error:", err)
		}
	})
	c.Start()
}
