package repositories

import (
	"autobackcom/internal/models"
	"context"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(client *mongo.Client, dbName, collectionName string) *OrderRepository {
	return &OrderRepository{
		collection: client.Database(dbName).Collection(collectionName),
	}
}

func (r *OrderRepository) SaveOrder(order models.Order) error {
	_, err := r.collection.InsertOne(context.Background(), order)
	if err != nil {
		logrus.WithField("error", err).Error("Failed to save order")
	}
	return err
}

func (r *OrderRepository) GetOrdersByUserID(userID primitive.ObjectID) ([]models.Order, error) {
	var orders []models.Order
	cursor, err := r.collection.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		logrus.WithField("error", err).Error("Failed to find orders by userID")
		return nil, err
	}
	err = cursor.All(context.Background(), &orders)
	if err != nil {
		logrus.WithField("error", err).Error("Failed to decode orders")
		return nil, err
	}
	return orders, nil
}
