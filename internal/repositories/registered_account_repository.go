package repositories

import (
	"autobackcom/internal/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RegisteredAccountRepository struct {
	collection *mongo.Collection
}

func NewRegisteredAccountRepository(client *mongo.Client, dbName, collectionName string) *RegisteredAccountRepository {
	return &RegisteredAccountRepository{
		collection: client.Database(dbName).Collection("registered_accounts"),
	}
}

func (r *RegisteredAccountRepository) SaveRegisteredAccount(account models.RegisteredAccount) error {
	_, err := r.collection.InsertOne(context.Background(), account)
	return err
}

func (r *RegisteredAccountRepository) UpdateRegisteredAccount(account models.RegisteredAccount) error {
	_, err := r.collection.UpdateOne(context.Background(), bson.M{"_id": account.ID}, bson.M{"$set": account})
	return err
}

func (r *RegisteredAccountRepository) GetRegisteredAccount(accountID string) (models.RegisteredAccount, error) {
	var account models.RegisteredAccount
	id, err := primitive.ObjectIDFromHex(accountID)
	if err != nil {
		return account, err
	}
	err = r.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&account)
	return account, err
}

func (r *RegisteredAccountRepository) GetAllRegisteredAccounts(ctx context.Context) ([]models.RegisteredAccount, error) {
	var accounts []models.RegisteredAccount
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &accounts)
	return accounts, err
}
