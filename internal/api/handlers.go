package api

import (
	"autobackcom/internal/api/dto"
	"autobackcom/internal/models"
	"autobackcom/internal/repositories"
	"autobackcom/internal/services"
	"autobackcom/internal/utils"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logrus.Error("Missing authorization header")
			c.JSON(401, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			logrus.WithField("error", err).Error("Invalid token")
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok {
			logrus.Error("Invalid token claims")
			c.JSON(401, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
		c.Set("userID", claims.Subject)
		c.Next()
	}
}

// RegisterHandler godoc
// @Summary Đăng ký tài khoản giao dịch
// @Description Đăng ký tài khoản để lấy lịch sử giao dịch
// @Tags registered_accounts
// @Accept json
// @Produce json
// @Param body body dto.RegisterRequest true "Thông tin đăng ký"
// @Success 201 {object} dto.APIResponse{data=dto.RegisterResponse}
// @Failure 400,500 {object} dto.APIResponse
// @Router /register [post]
func RegisterHandler(userRepo *repositories.RegisteredAccountRepository, tradeHistoryService *services.TradeHistoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logrus.WithField("error", err).Error("Invalid request")
			c.JSON(400, utils.Error("Yêu cầu không hợp lệ"))
			return
		}
		encryptedAPIKey, err := utils.Encrypt(req.APIKey)
		if err != nil {
			logrus.WithField("error", err).Error("Encryption error")
			c.JSON(500, utils.Error("Lỗi mã hóa"))
			return
		}
		encryptedSecret, err := utils.Encrypt(req.Secret)
		if err != nil {
			logrus.WithField("error", err).Error("Encryption error")
			c.JSON(500, utils.Error("Lỗi mã hóa"))
			return
		}
		account := models.RegisteredAccount{
			ID:              primitive.NewObjectID(),
			Username:        req.Username,
			Exchange:        req.Exchange,
			Market:          req.Market,
			EncryptedAPIKey: encryptedAPIKey,
			EncryptedSecret: encryptedSecret,
			IsTestnet:       req.IsTestnet,
		}
		err = userRepo.SaveRegisteredAccount(account)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"user":  account.Username,
				"error": err,
			}).Error("Failed to save user")
			c.JSON(500, utils.Error("Lỗi cơ sở dữ liệu"))
			return
		}
		resp := dto.RegisterResponse{RegisteredAccountID: account.ID.Hex(), Status: "ok"}
		c.JSON(201, utils.Success(resp))
		logrus.WithField("user", account.Username).Info("User registered successfully")
	}
}

// FetchAllTradesForUser godoc
// @Summary Lấy toàn bộ lịch sử giao dịch của các tài khoản đã đăng ký
// @Description Trigger lấy lịch sử giao dịch cho tất cả registered_accounts
// @Tags trades
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.FetchAllTradesResponse}
// @Failure 500 {object} dto.APIResponse
// @Router /fetch-trades-all-user [post]
func FetchAllTradesForUser(tradeHistoryService *services.TradeHistoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := tradeHistoryService.FetchAllAccountTradeHistory(c.Request.Context())
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("Failed to fetch trades of users")
			c.JSON(500, utils.Error("Lỗi lấy danh sách lệnh"))
			return
		}
		resp := dto.FetchAllTradesResponse{Status: "ok"}
		c.JSON(200, utils.Success(resp))
	}
}

// GetOrdersHandler godoc
// @Summary Lấy danh sách lệnh của tài khoản
// @Description Lấy danh sách order theo registered_account_id
// @Tags orders
// @Accept json
// @Produce json
// @Param body body dto.GetOrdersRequest true "ID tài khoản đã đăng ký"
// @Success 200 {object} dto.APIResponse{data=dto.GetOrdersResponse}
// @Failure 400,500 {object} dto.APIResponse
// @Router /orders [post]
func GetOrdersHandler(userRepo *repositories.RegisteredAccountRepository, orderRepo *repositories.OrderRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.GetOrdersRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logrus.WithField("error", err).Error("Invalid request")
			c.JSON(400, utils.Error("Yêu cầu không hợp lệ"))
			return
		}
		id, err := primitive.ObjectIDFromHex(req.RegisteredAccountID)
		if err != nil {
			logrus.WithField("error", err).Error("Invalid registered account ID")
			c.JSON(400, utils.Error("ID tài khoản không hợp lệ"))
			return
		}
		orders, err := orderRepo.GetOrdersByUserID(id)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"registered_account_id": req.RegisteredAccountID,
				"error":                 err,
			}).Error("Failed to get orders")
			c.JSON(500, utils.Error("Lỗi lấy danh sách lệnh"))
			return
		}
		resp := dto.GetOrdersResponse{
			Status: "ok",
			Data:   orders,
		}
		c.JSON(200, utils.Success(resp))
	}
}
