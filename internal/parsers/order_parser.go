package parsers

import "autobackcom/internal/models"

// OrderParser interface cho tất cả các sàn
type OrderParser interface {
	ParseOrder(market string, user *models.User, event map[string]interface{}) models.Order
}
