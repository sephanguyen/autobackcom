package exchanges

import (
	"autobackcom/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"
)

type BinanceFuturesStreamHandler struct {
	client     *futures.Client
	listenDone chan struct{}
}

func (h *BinanceFuturesStreamHandler) FetchOrders(ctx context.Context, user *models.User, start, end time.Time) ([]models.Order, error) {
	// TODO: Gọi API Binance Futures lấy orders theo khoảng thời gian
	return nil, nil
}

func (h *BinanceFuturesStreamHandler) Authenticate(ctx context.Context, apiKey, secret string) (string, error) {
	futures.UseTestnet = true

	h.client = futures.NewClient(apiKey, secret)
	listenKey, err := h.client.NewStartUserStreamService().Do(ctx)
	if err != nil {
		logrus.WithField("error", err).Error("Failed to get Binance futures listen key")
		return "", err
	}
	return listenKey, nil
}

func (h *BinanceFuturesStreamHandler) Connect(ctx context.Context, listenKey string) error {
	return nil
}

func (h *BinanceFuturesStreamHandler) Listen(ctx context.Context, handler func(event map[string]interface{})) error {
	h.listenDone = make(chan struct{})

	// Lấy listenKey từ Authenticate
	listenKey := ""
	if h.client != nil {
		lk, err := h.client.NewStartUserStreamService().Do(ctx)
		if err != nil {
			logrus.WithField("error", err).Error("Failed to get Binance futures listen key for Listen")
			return err
		}
		listenKey = lk
	}
	fmt.Println("Starting Binance futures stream with listenKey:", listenKey)
	wsStopFunc, _, err := futures.WsUserDataServe(
		listenKey,
		func(event *futures.WsUserDataEvent) {
			if event != nil {
				fmt.Printf("[Binance Futures Event] %+v\n", event)
				handler(map[string]interface{}{
					"EventType": event.Event,
					"Data":      event,
				})
			}
		},
		func(err error) {
			logrus.WithField("error", err).Error("Binance futures WebSocket error")
		},
	)
	if err != nil {
		logrus.WithField("error", err).Error("Failed to start Binance futures WebSocket")
		return err
	}

	// Goroutine tự động renew listenKey mỗi 30 phút
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if h.client != nil {
					err := h.client.NewKeepaliveUserStreamService().ListenKey(listenKey).Do(ctx)
					if err != nil {
						logrus.WithField("error", err).Error("Failed to renew Binance futures listen key")
					} else {
						logrus.Info("Binance futures listen key renewed")
					}
				}
			case <-h.listenDone:
				close(wsStopFunc)
				logrus.Warn("Binance futures WebSocket closed")
				return
			}
		}
	}()
	return nil
}

func (h *BinanceFuturesStreamHandler) Close() error {
	if h.listenDone != nil {
		close(h.listenDone)
	}
	return nil
}
