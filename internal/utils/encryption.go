package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

var encryptionKey []byte

func getEncryptionKey() []byte {
	if encryptionKey == nil {
		key := os.Getenv("ENCRYPTION_KEY")
		if key == "" {
			panic("ENCRYPTION_KEY environment variable is not set")
		}
		encryptionKey = []byte(key)
	}
	return encryptionKey
}

func Encrypt(plainText string) (string, error) {
	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func Decrypt(cipherText string) (string, error) {
	key := getEncryptionKey()
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, cipher := data[:nonceSize], data[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, cipher, nil)
	if err != nil {
		return "", err
	}
	return string(plainText), nil
}
