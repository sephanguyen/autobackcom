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
	Price            float64            `bson:"price"`
	Quantity         float64            `bson:"quantity"`
	ExecutedQuantity float64            `bson:"executed_quantity"`
	AvgPrice         float64            `bson:"avg_price"`
	Time             time.Time          `bson:"time"`
	Commission       float64            `bson:"commission"`
	CommissionAsset  string             `bson:"commission_asset"`
}
