package services

import (
	"autobackcom/internal/exchanges"
	"autobackcom/internal/exchanges/binance"
	"autobackcom/internal/models"
	"autobackcom/internal/utils"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ClientsInfo chứa các client cho một user, theo cặp exchange/market
type ClientsInfo struct {
	Clients   map[string]exchanges.ExchangeClient // Key: "exchange:market"
	CreatedAt time.Time
}

// ClientManagerService quản lý cache các client
type ClientManagerService struct {
	clientCache *cache.Cache
	mutexes     map[string]*sync.RWMutex // Mutex cho mỗi user.ID
	mutex       sync.RWMutex             // Khóa để quản lý mutexes map
}

// NewClientManagerService khởi tạo service
func NewClientManagerService() *ClientManagerService {
	return &ClientManagerService{
		clientCache: cache.New(24*time.Hour, 1*time.Hour),
		mutexes:     make(map[string]*sync.RWMutex),
	}
}

// getUserMutex lấy hoặc tạo mutex cho user.ID
func (s *ClientManagerService) getUserMutex(userID string) *sync.RWMutex {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if mu, exists := s.mutexes[userID]; exists {
		return mu
	}
	mu := &sync.RWMutex{}
	s.mutexes[userID] = mu
	return mu
}

// createClient tạo client dựa trên exchange và market
func createClient(exchange, market string, user models.User) (exchanges.ExchangeClient, error) {
	apiKey, err := utils.Decrypt(user.EncryptedAPIKey)
	if err != nil {
		log.Printf("Decrypt key error for user %s: %v", user.Username, err)
		return nil, err
	}
	secret, err := utils.Decrypt(user.EncryptedSecret)
	if err != nil {
		log.Printf("Decrypt secret error for user %s: %v", user.Username, err)
		return nil, err
	}
	switch exchange {
	case "binance":
		switch market {
		case "spot":
			return binance.NewBinanceSpotExchange(apiKey, secret), nil
		case "futures":
			return binance.NewBinanceFetureExchange(apiKey, secret), nil
		default:
			return nil, fmt.Errorf("unsupported market: %s for exchange: %s", market, exchange)
		}
	default:
		return nil, fmt.Errorf("unsupported exchange: %s", exchange)
	}
}

// GetOrCreateClient lấy hoặc tạo client cho user
func (s *ClientManagerService) GetOrCreateClient(user models.User) (*ClientsInfo, error) {
	cacheKey := user.ID.Hex()
	clientKey := fmt.Sprintf("%s:%s", user.Exchange, user.Market)
	userMutex := s.getUserMutex(cacheKey)

	// Kiểm tra cache với read lock
	userMutex.RLock()
	if clientPair, found := s.clientCache.Get(cacheKey); found {
		pair := clientPair.(*ClientsInfo)
		if _, exists := pair.Clients[clientKey]; exists {
			userMutex.RUnlock()
			return pair, nil
		}
	}
	userMutex.RUnlock()

	// Nâng cấp lên write lock để tạo hoặc cập nhật client
	userMutex.Lock()
	defer userMutex.Unlock()

	// Kiểm tra lại cache để tránh race condition
	if clientPair, found := s.clientCache.Get(cacheKey); found {
		pair := clientPair.(*ClientsInfo)
		if _, exists := pair.Clients[clientKey]; exists {
			return pair, nil
		}
		client, err := createClient(user.Exchange, user.Market, user)
		if err != nil {
			log.Printf("Create client error for user %s, exchange %s, market %s: %v", user.Username, user.Exchange, user.Market, err)
			return nil, err
		}
		pair.Clients[clientKey] = client
		s.clientCache.Set(cacheKey, pair, cache.DefaultExpiration)
		return pair, nil
	}

	// Tạo ClientsInfo mới
	client, err := createClient(user.Exchange, user.Market, user)
	if err != nil {
		log.Printf("Create client error for user %s, exchange %s, market %s: %v", user.Username, user.Exchange, user.Market, err)
		return nil, err
	}
	clientPair := &ClientsInfo{
		Clients:   map[string]exchanges.ExchangeClient{clientKey: client},
		CreatedAt: time.Now(),
	}
	s.clientCache.Set(cacheKey, clientPair, cache.DefaultExpiration)
	return clientPair, nil
}

// InvalidateClient xóa client cho user
func (s *ClientManagerService) InvalidateClient(userID primitive.ObjectID) {
	cacheKey := userID.Hex()
	s.clientCache.Delete(cacheKey)
	s.mutex.Lock()
	delete(s.mutexes, cacheKey)
	s.mutex.Unlock()
}

// CleanupClients xóa các client hết hạn
func (s *ClientManagerService) CleanupClients(maxAge time.Duration) {
	s.clientCache.DeleteExpired()
	s.mutex.Lock()
	for cacheKey := range s.mutexes {
		if _, found := s.clientCache.Get(cacheKey); !found {
			delete(s.mutexes, cacheKey)
		}
	}
	s.mutex.Unlock()
}

// GetClientCount trả về số lượng client trong cache
func (s *ClientManagerService) GetClientCount() int {
	return s.clientCache.ItemCount()
}
