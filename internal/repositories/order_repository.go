package repositories

import (
	"autobackcom/internal/models"
	"context"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (r *OrderRepository) SaveOrders(orders []models.Order) error {
	if len(orders) == 0 {
		return nil
	}

	models := make([]mongo.WriteModel, len(orders))
	for i, order := range orders {
		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{
				"orderID":  order.OrderID,
				"exchange": order.Exchange,
				"market":   order.Market,
			}).
			SetUpdate(bson.M{"$set": order}).
			SetUpsert(true)
		models[i] = model
	}

	opts := options.BulkWrite().SetOrdered(false) // Unordered để tiếp tục khi có lỗi
	result, err := r.collection.BulkWrite(context.Background(), models, opts)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":      err,
			"orderCount": len(orders),
		}).Error("Failed to save orders")
		if bulkErr, ok := err.(mongo.BulkWriteException); ok {
			for _, writeErr := range bulkErr.WriteErrors {
				logrus.WithFields(logrus.Fields{
					"index": writeErr.Index,
					"code":  writeErr.Code,
					"msg":   writeErr.Message,
				}).Error("BulkWrite error for order")
			}
		}
		return err
	}

	logrus.WithFields(logrus.Fields{
		"inserted": result.InsertedCount,
		"updated":  result.ModifiedCount,
		"orders":   len(orders),
	}).Info("Successfully saved orders")
	return nil
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
