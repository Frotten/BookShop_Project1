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

func OrderFromRedisHash(orderID int64, data map[string]string) *models.Order {
	if len(data) == 0 {
		return nil
	}
	o := parseOrder(data)
	if o == nil {
		return nil
	}
	o.OrderID = orderID
	return o
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
		CreatedAt:  data["created_at"],
	}
}

func SetUserOrdersInfo(Orders []*models.Order, missIDs []int64) error {
	if len(Orders) == 0 {
		return nil
	}
	if len(missIDs) == 0 {
		for _, Order := range Orders {
			missIDs = append(missIDs, Order.OrderID)
		}
	}
	missSet := make(map[int64]struct{}, len(missIDs))
	for _, id := range missIDs {
		missSet[id] = struct{}{}
	}
	missMap := make(map[int64]*models.Order, len(missIDs))
	for _, order := range Orders {
		if _, ok := missSet[order.OrderID]; ok {
			missMap[order.OrderID] = order
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

func TryLoadOrderItemsIntoView(ov *models.OrderView) bool {
	SKey := "items:order:" + strconv.FormatInt(ov.OrderID, 10)
	members, err := RDB.SMembers(ctx, SKey).Result()
	if err != nil || len(members) == 0 {
		return false
	}
	pipe := RDB.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(members))
	for i, idStr := range members {
		cmds[i] = pipe.HGetAll(ctx, "order:items:"+idStr)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return false
	}
	items := make([]*models.OrderItem, 0, len(members))
	for i, idStr := range members {
		data, err := cmds[i].Result()
		if err != nil || len(data) == 0 {
			return false
		}
		id, _ := strconv.ParseInt(idStr, 10, 64)
		bookID, _ := strconv.ParseInt(data["book_id"], 10, 64)
		price, _ := strconv.ParseInt(data["price"], 10, 64)
		qty, _ := strconv.ParseInt(data["quantity"], 10, 64)
		Book, _ := GetBookByBookID(bookID)
		items = append(items, &models.OrderItem{
			ID:       id,
			OrderID:  ov.OrderID,
			BookID:   bookID,
			Price:    price,
			Quantity: qty,
			Title:    Book.Title,
		})
	}
	ov.Items = items
	return true
}

// UpdateOrderStatusInCache 与库内状态一致化；订单主档 miss 时仅更新已有 hash。
func UpdateOrderStatusInCache(orderID int64, status int8) error {
	key := "order:" + strconv.FormatInt(orderID, 10)
	return RDB.HSet(ctx, key, "status", status).Err()
}

// GetOrderHash 读取订单主档缓存（可能为空 map）。
func GetOrderHash(orderID int64) (map[string]string, error) {
	key := "order:" + strconv.FormatInt(orderID, 10)
	return RDB.HGetAll(ctx, key).Result()
}
