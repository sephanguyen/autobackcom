package dto

type ExchangeType string

const (
	ExchangeBinance ExchangeType = "binance"
)

type MarketType string

const (
	MarketSpot    MarketType = "spot"
	MarketFutures MarketType = "futures"
)

type RegisterRequest struct {
	Username  string       `json:"username"`
	Exchange  ExchangeType `json:"exchange"`
	Market    MarketType   `json:"market"`
	APIKey    string       `json:"apikey"`
	Secret    string       `json:"secret"`
	IsTestnet bool         `json:"isTestnet"`
}

type RegisterResponse struct {
	RegisteredAccountID string `json:"registeredAccountID"`
	Status              string `json:"status"`
}

// Validate ExchangeType
func (e ExchangeType) IsValid() bool {
	switch e {
	case ExchangeBinance:
		return true
	default:
		return false
	}
}

// Chuyển đổi từ string sang ExchangeType
func ToExchangeType(s string) (ExchangeType, bool) {
	et := ExchangeType(s)
	return et, et.IsValid()
}

// Validate MarketType
func (m MarketType) IsValid() bool {
	switch m {
	case MarketSpot, MarketFutures:
		return true
	default:
		return false
	}
}

// Chuyển đổi từ string sang MarketType
func ToMarketType(s string) (MarketType, bool) {
	mt := MarketType(s)
	return mt, mt.IsValid()
}
