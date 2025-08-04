package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"autobackcom/internal/models"
	"autobackcom/internal/repositories"
	"autobackcom/internal/services"
	"autobackcom/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func GenerateToken(userID string) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logrus.Error("Missing authorization header")
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			logrus.WithField("error", err).Error("Invalid token")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok {
			logrus.Error("Invalid token claims")
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "userID", claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RegisterHandler(userRepo *repositories.UserRepository, exchangeService *services.ExchangeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string   `json:"username"`
			Exchange string   `json:"exchange"`
			Markets  []string `json:"markets"`
			APIKey   string   `json:"api_key"`
			Secret   string   `json:"secret"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			logrus.WithField("error", err).Error("Invalid request")
			http.Error(w, "Yêu cầu không hợp lệ", http.StatusBadRequest)
			return
		}
		encryptedAPIKey, err := utils.Encrypt(req.APIKey)
		if err != nil {
			logrus.WithField("error", err).Error("Encryption error")
			http.Error(w, "Lỗi mã hóa", http.StatusInternalServerError)
			return
		}
		encryptedSecret, err := utils.Encrypt(req.Secret)
		if err != nil {
			logrus.WithField("error", err).Error("Encryption error")
			http.Error(w, "Lỗi mã hóa", http.StatusInternalServerError)
			return
		}
		user := models.User{
			ID:              primitive.NewObjectID(),
			Username:        req.Username,
			Exchange:        req.Exchange,
			Markets:         req.Markets,
			EncryptedAPIKey: encryptedAPIKey,
			EncryptedSecret: encryptedSecret,
		}
		err = userRepo.SaveUser(user)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"user":  user.Username,
				"error": err,
			}).Error("Failed to save user")
			http.Error(w, "Lỗi cơ sở dữ liệu", http.StatusInternalServerError)
			return
		}
		token, err := GenerateToken(user.ID.Hex())
		if err != nil {
			logrus.WithField("error", err).Error("Failed to generate token")
			http.Error(w, "Không thể tạo token", http.StatusInternalServerError)
			return
		}
		go exchangeService.RegisterAndStartStream(&user)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"token": token})
		logrus.WithField("user", user.Username).Info("User registered successfully")
	}
}

func GetOrdersHandler(userRepo *repositories.UserRepository, orderRepo *repositories.OrderRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("userID").(string)
		if !ok {
			logrus.Error("User ID not found in context")
			http.Error(w, "Không tìm thấy ID người dùng", http.StatusUnauthorized)
			return
		}
		id, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			logrus.WithField("error", err).Error("Invalid user ID")
			http.Error(w, "ID người dùng không hợp lệ", http.StatusBadRequest)
			return
		}
		orders, err := orderRepo.GetOrdersByUserID(id)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"user":  userID,
				"error": err,
			}).Error("Failed to get orders")
			http.Error(w, "Lỗi lấy danh sách lệnh", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(orders)
	}
}
