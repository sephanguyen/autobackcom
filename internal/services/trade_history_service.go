package services

import (
	"autobackcom/internal/exchanges"
	"autobackcom/internal/models"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TradeHistoryService struct {
	UserCollection  *mongo.Collection
	OrderCollection *mongo.Collection
	StreamHandler   exchanges.StreamHandler
}

func NewTradeHistoryService(userCol, orderCol *mongo.Collection, handler exchanges.StreamHandler) *TradeHistoryService {
	return &TradeHistoryService{
		UserCollection:  userCol,
		OrderCollection: orderCol,
		StreamHandler:   handler,
	}
}

// Cronjob function
func (s *TradeHistoryService) FetchAllUserTradeHistory(ctx context.Context, start, end time.Time) error {
	cursor, err := s.UserCollection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			log.Println("Decode user error:", err)
			continue
		}
		orders, err := s.StreamHandler.FetchOrders(ctx, &user, start, end)
		if err != nil {
			log.Println("FetchOrders error for user", user.Username, ":", err)
			continue
		}
		// Save orders to DB
		for _, order := range orders {
			_, err := s.OrderCollection.InsertOne(ctx, order)
			if err != nil {
				log.Println("Insert order error:", err)
			}
		}
	}
	return nil
}
