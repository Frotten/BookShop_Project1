package redis

import (
	"Project1_Shop/models"
	"strconv"
)

func SaveOrder(Order *models.Order) {
	OrderKey := "order:" + strconv.FormatInt(Order.OrderID, 10)
	RDB.HSet(ctx, OrderKey, map[string]interface{}{
		"user_id":     Order.UserID,
		"total_price": Order.TotalPrice,
		"status":      Order.Status,
		"created_at":  Order.CreatedAt,
	})
	RDB.Expire(ctx, OrderKey, OrderTime)
}
