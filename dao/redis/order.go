package redis

import (
	"Project1_Shop/models"
	"errors"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func SaveOrder(Order *models.Order) error {
	OrderKey := "order:" + strconv.FormatInt(Order.OrderID, 10)
	UserKey := "user:order:" + strconv.FormatInt(Order.UserID, 10)
	pipe := RDB.Pipeline()
	pipe.HSet(ctx, OrderKey, map[string]interface{}{
		"user_id":     Order.UserID,
		"total_price": Order.TotalPrice,
		"status":      Order.Status,
		"created_at":  Order.CreatedAt,
	})
	time := models.TimeParse(Order.CreatedAt)
	if time == -1 {
		return errors.New("time Error")
	}
	pipe.ZAdd(ctx, UserKey, redis.Z{
		Score:  float64(time),
		Member: Order.OrderID,
	})
	pipe.Expire(ctx, OrderKey, RandTTL(OrderTime))
	_, err := pipe.Exec(ctx)
	return err
}

func SaveOrderItems(orderItems []*models.OrderItem) error {
	pipe := RDB.Pipeline()
	for _, item := range orderItems {
		key := "order:items:" + strconv.FormatInt(item.ID, 10)
		SKey := "items:order:" + strconv.FormatInt(item.OrderID, 10)
		pipe.HSet(ctx, key, map[string]interface{}{
			"order_id": item.OrderID,
			"book_id":  item.BookID,
			"price":    item.Price,
			"quantity": item.Quantity,
		})
		pipe.Expire(ctx, key, RandTTL(OrderTime))
		pipe.SAdd(ctx, SKey, item.ID)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func GetUserOrdersInfo(UserID int64) ([]*models.Order, []int64, error) {
	UserKey := "user:order:" + strconv.FormatInt(UserID, 10)
	pipe := RDB.Pipeline()
	Ans, err := pipe.ZRevRangeWithScores(ctx, UserKey, 0, -1).Result()
	if err != nil {
		return nil, nil, err
	}
	var OrderIDs []int64
	for _, z := range Ans {
		OrderIDs = append(OrderIDs, z.Member.(int64))
	}
	cmds := make([]*redis.MapStringStringCmd, 0, len(OrderIDs))
	for _, id := range OrderIDs {
		key := "order:" + strconv.FormatInt(id, 10)
		cmds = append(cmds, pipe.HGetAll(ctx, key))
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, nil, err
	}
	var missIDs []int64
	var Orders []*models.Order
	for i, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil || len(data) == 0 {
			missIDs = append(missIDs, OrderIDs[i])
			continue
		}
		order := parseOrder(data)
		if order != nil {
			order.OrderID = OrderIDs[i]
			Orders = append(Orders, order)
		}
	}
	return Orders, missIDs, nil
}

func parseOrder(data map[string]string) *models.Order {
	if len(data) == 0 {
		return nil
	}
	UserID, _ := strconv.ParseInt(data["user_id"], 10, 64)
	TotalPrice, _ := strconv.ParseInt(data["total_price"], 10, 64)
	Status, _ := strconv.ParseInt(data["status"], 10, 8)
	return &models.Order{
		UserID:     UserID,
		TotalPrice: TotalPrice,
		Status:     int8(Status),
	}
}

func SetUserOrdersInfo(Orders []*models.Order, missIDs []int64) error {
	if len(missIDs) == 0 || len(Orders) == 0 {
		return nil
	}
	missMap := make(map[int64]*models.Order, len(missIDs))
	for _, order := range Orders {
		for _, missID := range missIDs {
			if order.OrderID == missID {
				missMap[missID] = order
				break
			}
		}
	}
	if len(missMap) == 0 {
		return nil
	}
	pipe := RDB.Pipeline()
	for orderID, order := range missMap {
		OrderKey := "order:" + strconv.FormatInt(orderID, 10)
		UserKey := "user:order:" + strconv.FormatInt(order.UserID, 10)
		pipe.HSet(ctx, OrderKey, map[string]interface{}{
			"user_id":     order.UserID,
			"total_price": order.TotalPrice,
			"status":      order.Status,
			"created_at":  order.CreatedAt,
		})
		time := models.TimeParse(order.CreatedAt)
		if time != -1 {
			pipe.ZAdd(ctx, UserKey, redis.Z{
				Score:  float64(time),
				Member: orderID,
			})
		}
		pipe.Expire(ctx, OrderKey, RandTTL(OrderTime))
	}
	_, err := pipe.Exec(ctx)
	return err
}

func GetOrderItemsInfo(OrderViews []*models.OrderView) {
	for _, order := range OrderViews {
		pipe := RDB.Pipeline()
		SKey := "items:order:" + strconv.FormatInt(order.OrderID, 10)
		Res, err := pipe.SMembers(ctx, SKey).Result()
		if err != nil {
			continue
		}
		var Items []*models.OrderItem
		for _, IDStr := range Res {
			ID, _ := strconv.ParseInt(IDStr, 10, 64)
			key := "order:items:" + IDStr
			data, _ := pipe.HGetAll(ctx, key).Result()
			BookID, _ := strconv.ParseInt(data["book_id"], 10, 64)
			Price, _ := strconv.ParseInt(data["price"], 10, 64)
			Quantity, _ := strconv.ParseInt(data["quantity"], 10, 64)
			Item := &models.OrderItem{
				ID:       ID,
				OrderID:  order.OrderID,
				BookID:   BookID,
				Price:    Price,
				Quantity: Quantity,
			}
			Items = append(Items, Item)
		}
		_, err = pipe.Exec(ctx)
		if err == nil {
			order.Items = Items
		}
	}
}
