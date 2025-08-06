package repositories

import (
	"autobackcom/internal/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(client *mongo.Client, dbName, collectionName string) *UserRepository {
	return &UserRepository{
		collection: client.Database(dbName).Collection(collectionName),
	}
}

func (r *UserRepository) SaveUser(user models.User) error {
	_, err := r.collection.InsertOne(context.Background(), user)
	return err
}

func (r *UserRepository) UpdateUser(user models.User) error {
	_, err := r.collection.UpdateOne(context.Background(), bson.M{"_id": user.ID}, bson.M{"$set": user})
	return err
}

func (r *UserRepository) GetUser(userID string) (models.User, error) {
	var user models.User
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return user, err
	}
	err = r.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	return user, err
}

func (r *UserRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &users)
	return users, err
}
