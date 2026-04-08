package redis

import (
	"Project1_Shop/models"
	"strconv"
)

func SaveOrder(Order *models.Order) error {
	OrderKey := "order:" + strconv.FormatInt(Order.OrderID, 10)
	RDB.HSet(ctx, OrderKey, map[string]interface{}{
		"user_id":     Order.UserID,
		"total_price": Order.TotalPrice,
		"status":      Order.Status,
		"created_at":  Order.CreatedAt,
	})
	return RDB.Expire(ctx, OrderKey, RandTTL(OrderTime)).Err()
}

func SaveOrderItems(orderItems []*models.OrderItem) error {
	pipe := RDB.Pipeline()
	for _, item := range orderItems {
		key := "order:items:" + strconv.FormatInt(item.ID, 10)
		pipe.HSet(ctx, key, map[string]interface{}{
			"order_id": item.OrderID,
			"book_id":  item.BookID,
			"price":    item.Price,
			"quantity": item.Quantity,
		})
		pipe.Expire(ctx, key, RandTTL(OrderTime))
	}
	_, err := pipe.Exec(ctx)
	return err
}
