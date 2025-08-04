package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"autobackcom/internal/exchanges"
	"autobackcom/internal/models"
	"autobackcom/internal/parsers"
	"autobackcom/internal/repositories"
	"autobackcom/internal/utils"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	logrus.SetOutput(ConsoleWriter{})
}

// ConsoleWriter để đảm bảo log in ra console
type ConsoleWriter struct{}

func (ConsoleWriter) Write(p []byte) (n int, err error) {
	return fmt.Print(string(p))
}

type ExchangeService struct {
	userRepo       *repositories.UserRepository
	orderRepo      *repositories.OrderRepository
	streamHandlers map[string]map[string]exchanges.StreamHandler
}

func NewExchangeService(
	userRepo *repositories.UserRepository,
	orderRepo *repositories.OrderRepository,
	streamHandlers map[string]map[string]exchanges.StreamHandler,
) *ExchangeService {
	return &ExchangeService{
		userRepo:       userRepo,
		orderRepo:      orderRepo,
		streamHandlers: streamHandlers,
	}
}

// Đăng ký user, mở luồng stream, lấy order và lưu vào DB
func (s *ExchangeService) RegisterAndStartStream(user *models.User) error {
	// Lưu user vào DB
	fmt.Println("user", user)
	if err := s.userRepo.UpdateUser(*user); err != nil {
		logrus.WithField("user", user.ID.Hex()).Error("Failed to save user")
		return err
	}

	return s.startUserStreamsWithPool(user)
}

// Tách worker pool thành hàm riêng cho dễ đọc
func (s *ExchangeService) startUserStreamsWithPool(user *models.User) error {
	const maxWorkers = 3
	jobs := make(chan string, len(user.Markets))
	results := make(chan error, len(user.Markets))

	// Worker xử lý từng market
	worker := func() {
		for market := range jobs {
			fmt.Println("Starting stream for market:", market)
			err := s.startStreamForMarket(user, market)
			results <- err
		}
	}

	// Khởi tạo worker pool
	for w := 0; w < maxWorkers; w++ {
		go worker()
	}
	// Đưa các market vào jobs
	for _, market := range user.Markets {
		jobs <- market
	}
	close(jobs)

	// Chờ tất cả worker hoàn thành
	for i := 0; i < len(user.Markets); i++ {
		<-results
	}
	return nil
}

// Tách logic khởi tạo stream cho từng market
func (s *ExchangeService) startStreamForMarket(user *models.User, market string) error {
	handler, ok := s.streamHandlers[user.Exchange][market]
	if !ok {
		logrus.WithFields(logrus.Fields{
			"user":     user.ID.Hex(),
			"exchange": user.Exchange,
			"market":   market,
		}).Error("No stream handler found")
		return fmt.Errorf("no stream handler")
	}

	key, err := utils.Decrypt(user.EncryptedAPIKey)
	if err != nil {
		logrus.WithFields(logrus.Fields{"user": user.ID.Hex(), "error": err}).Error("Failed to decrypt API key")
	}
	secret, err := utils.Decrypt(user.EncryptedSecret)
	if err != nil {
		logrus.WithFields(logrus.Fields{"user": user.ID.Hex(), "error": err}).Error("Failed to decrypt API secret")
	}
	fmt.Println("key", key, secret, user.Exchange, market, handler, s.streamHandlers)
	fmt.Printf("handler type: %T\n", handler)

	_, err = handler.Authenticate(context.Background(), key, secret)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user":   user.ID.Hex(),
			"market": market,
			"error":  err,
		}).Error("Failed to connect stream")
		return err
	}
	go s.listenAndSaveOrder(handler, market, user)
	return nil
}

// Tách logic lắng nghe và lưu order
func (s *ExchangeService) listenAndSaveOrder(handler exchanges.StreamHandler, market string, user *models.User) {
	fmt.Println("Starting stream for user:", user.ID.Hex(), "market:", market)

	for {
		err := handler.Listen(context.Background(), func(event map[string]interface{}) {
			fmt.Println("Received event:", event)
			order := parseOrderByExchange(user, market, event)
			if order.ID != "" {
				if err := s.orderRepo.SaveOrder(order); err != nil {
					logrus.WithFields(logrus.Fields{
						"order":  order.ID,
						"user":   user.ID.Hex(),
						"market": market,
						"error":  err,
					}).Error("Failed to save order")
				}
			}
		})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"user":   user.ID.Hex(),
				"market": market,
				"error":  err,
			}).Error("Stream listen error, retrying...")
			time.Sleep(5 * time.Second)
			continue
		} else {
			// Nếu Listen thành công, thoát khỏi vòng lặp
			break
		}
	}
}

// Strategy pattern cho order parser
type OrderParser interface {
	ParseOrder(market string, user *models.User, event map[string]interface{}) models.Order
}

var orderParsers = map[string]OrderParser{
	"binance": parsers.BinanceResponseParser{},
	"okx":     parsers.OKXResponseParser{},
}

// Gom parse order theo exchange
func parseOrderByExchange(user *models.User, market string, event map[string]interface{}) models.Order {
	if parser, ok := orderParsers[user.Exchange]; ok {
		return parser.ParseOrder(market, user, event)
	}
	return models.Order{}
}

// Tự động restart lại các stream của user khi ứng dụng khởi động lại
func (s *ExchangeService) StartAllUserStreams() {
	fmt.Println("Restarting all user streams...")
	users, err := s.userRepo.GetAllUsers()
	if err != nil {
		logrus.WithField("error", err).Error("Failed to get all users for stream restart")
		return
	}
	var wg sync.WaitGroup
	for _, user := range users {
		wg.Add(1)
		go func(u models.User) {
			defer wg.Done()
			s.RegisterAndStartStream(&u)
		}(user)
	}
	go func() {
		wg.Wait()
		fmt.Println("All user streams finished.")
	}()
}

// Parse order từ event stream
// Tách parse order cho từng exchange/market để giảm độ phức tạp

// Đã chuyển sang OKXResponseParser
