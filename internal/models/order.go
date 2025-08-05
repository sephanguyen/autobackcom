package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID               string             `bson:"id"`
	UserID           primitive.ObjectID `bson:"user_id"`
	Exchange         string             `bson:"exchange"`
	Market           string             `bson:"market"`
	Symbol           string             `bson:"symbol"`
	Status           string             `bson:"status"`
	Side             string             `bson:"side"`
	Type             string             `bson:"type"`
	Price            string             `bson:"price"`
	Quantity         string             `bson:"quantity"`
	ExecutedQuantity string             `bson:"executed_quantity"`
	AvgPrice         string             `bson:"avg_price"`
	Time             time.Time          `bson:"time"`
	Commission       string             `bson:"commission"`
	CommissionAsset  string             `bson:"commission_asset"`
	OrderID          int64              `bson:"order_id"`
	OrderListId      int64              `bson:"order_list_id"`
	QuoteQuantity    string             `bson:"quote_quantity,omitempty"`
}
