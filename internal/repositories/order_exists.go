package repositories

import (
	"time"
)

// Exists kiểm tra order đã tồn tại chưa (theo orderId, userId, exchange, market, time)
func (r *OrderRepository) Exists(orderID, userID, exchange, market string, orderTime time.Time) bool {
	// TODO: Thực hiện query DB kiểm tra trùng
	return false // giả sử chưa có, bạn cần tự hiện thực
}
