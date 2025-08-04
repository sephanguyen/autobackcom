#!/bin/bash
# gen_curl.sh - Script gọi API register và orders

API_URL="http://localhost:8080"

# Đăng ký user mới
curl -X POST "$API_URL/register" \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass"}'

echo "\n---"

# Lấy danh sách orders (cần JWT nếu API yêu cầu)
# Thay YOUR_JWT_TOKEN bằng token thực tế nếu cần
curl -X GET "$API_URL/orders" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"

echo "\n---"
