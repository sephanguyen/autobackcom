package models

// EventData là model chung cho dữ liệu event từ các exchange
// Có thể mở rộng thêm các trường cần thiết cho mọi loại event
// Các handler sẽ map dữ liệu event vào struct này

type EventData struct {
	EventType string      `json:"event_type"`
	UserID    string      `json:"user_id,omitempty"`
	Exchange  string      `json:"exchange"`
	Market    string      `json:"market"`
	OrderID   string      `json:"order_id,omitempty"`
	Symbol    string      `json:"symbol,omitempty"`
	Side      string      `json:"side,omitempty"`
	Price     string      `json:"price,omitempty"`
	Quantity  string      `json:"quantity,omitempty"`
	Status    string      `json:"status,omitempty"`
	Timestamp int64       `json:"timestamp,omitempty"`
	Raw       interface{} `json:"raw,omitempty"` // Lưu dữ liệu gốc nếu cần
}
