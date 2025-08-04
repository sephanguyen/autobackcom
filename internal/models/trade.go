package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Trade là model lưu trade event
// Có thể mở rộng thêm trường nếu cần

type Trade struct {
	ID       string             `bson:"_id,omitempty" json:"id"`
	UserID   primitive.ObjectID `bson:"user_id" json:"user_id"`
	Symbol   string             `bson:"symbol" json:"symbol"`
	Price    float64            `bson:"price" json:"price"`
	Quantity float64            `bson:"quantity" json:"quantity"`
	Time     time.Time          `bson:"time" json:"time"`
	IsBuyer  bool               `bson:"is_buyer" json:"is_buyer"`
	IsMaker  bool               `bson:"is_maker" json:"is_maker"`
}
