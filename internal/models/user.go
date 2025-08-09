package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID                  primitive.ObjectID `bson:"_id"`
	Username            string             `bson:"username"`
	Exchange            string             `bson:"exchange"`
	Market              string             `bson:"market"`
	EncryptedAPIKey     string             `bson:"encrypted_api_key"`
	EncryptedSecret     string             `bson:"encrypted_secret"`
	EncryptedPassphrase string             `bson:"encrypted_passphrase,omitempty"` // Thêm cho OKX
	ListenKey           string             `bson:"listen_key,omitempty"`
	IsTestnet           bool               `bson:"is_testnet"`
}
