package exchanges

import (
	"autobackcom/internal/models"
	"context"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"
)

type BinanceSpotStreamHandler struct {
	client     *binance.Client
	listenDone chan struct{}
}

func (h *BinanceSpotStreamHandler) FetchOrders(ctx context.Context, user *models.User, start, end time.Time) ([]models.Order, error) {
	// TODO: Gọi API Binance lấy orders theo khoảng thời gian
	return nil, nil
}

func (h *BinanceSpotStreamHandler) Authenticate(ctx context.Context, apiKey, secret string) (string, error) {
	binance.UseTestnet = true

	h.client = binance.NewClient(apiKey, secret)
	listenKey, err := h.client.NewStartUserStreamService().Do(ctx)
	if err != nil {
		logrus.WithField("error", err).Error("Failed to get Binance spot listen key")
		return "", err
	}
	return listenKey, nil
}

func (h *BinanceSpotStreamHandler) Connect(ctx context.Context, listenKey string) error {
	return nil
}

func (h *BinanceSpotStreamHandler) Listen(ctx context.Context, handler func(event map[string]interface{})) error {
	h.listenDone = make(chan struct{})

	// Lấy listenKey từ Authenticate
	listenKey := ""
	if h.client != nil {
		lk, err := h.client.NewStartUserStreamService().Do(ctx)
		if err != nil {
			logrus.WithField("error", err).Error("Failed to get Binance spot listen key for Listen")
			return err
		}
		listenKey = lk
	}

	// Khởi tạo WebSocket
	wsStopFunc, _, err := binance.WsUserDataServe(
		listenKey,
		func(event *binance.WsUserDataEvent) {
			if event != nil {
				handler(map[string]interface{}{
					"EventType": event.Event,
					"Data":      event,
				})
			}
		},
		func(err error) {
			logrus.WithField("error", err).Error("Binance spot WebSocket error")
		},
	)
	if err != nil {
		logrus.WithField("error", err).Error("Failed to start Binance spot WebSocket")
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
						logrus.WithField("error", err).Error("Failed to renew Binance spot listen key")
					} else {
						logrus.Info("Binance spot listen key renewed")
					}
				}
			case <-h.listenDone:
				close(wsStopFunc)
				logrus.Warn("Binance spot WebSocket closed")
				return
			}
		}
	}()
	return nil
}

func (h *BinanceSpotStreamHandler) Close() error {
	if h.listenDone != nil {
		close(h.listenDone)
	}
	return nil
}
