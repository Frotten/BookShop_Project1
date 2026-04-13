package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
	"strconv"
	"time"
)

func GetBookPriceByIDs(BookIDs []int64) (map[int64]int64, error) {
	bookPriceMap := make(map[int64]int64)
	Books, err := GetBooksByIDs(BookIDs)
	if err != nil {
		return nil, err
	}
	for _, book := range Books {
		bookPriceMap[book.BookID] = book.Price
	}
	return bookPriceMap, nil
}

func CreateOrder(orderParam models.OrderRequest, UserID int64) models.ResCode {
	var OrderItems []*models.OrderItem
	var BookIDs []int64
	for _, item := range orderParam.Items {
		BookIDs = append(BookIDs, item.BookID)
	}
	PriceMap, err := GetBookPriceByIDs(BookIDs)
	if err != nil {
		return models.CodeServerBusy
	}
	var TotalPrice int64
	for _, item := range orderParam.Items {
		TotalPrice += PriceMap[item.BookID] * item.Quantity
	}
	Order := &models.Order{
		UserID:     UserID,
		Status:     0,
		TotalPrice: TotalPrice,
		CreatedAt:  time.Now().Format(models.TimeParseLayout),
	}
	err = mysql.CreateOrder(Order)
	if err != nil {
		return models.CodeMySQLError
	}
	err = redis.SaveOrder(Order)
	if err != nil {
		return models.CodeRedisError
	}
	for _, item := range orderParam.Items {
		OrderItems = append(OrderItems, &models.OrderItem{
			BookID:   item.BookID,
			Quantity: item.Quantity,
			Price:    PriceMap[item.BookID],
			OrderID:  Order.OrderID,
		})
	}
	err = mysql.CreateOrderItems(OrderItems)
	if err != nil {
		return models.CodeMySQLError
	}
	err = redis.SaveOrderItems(OrderItems)
	if err != nil {
		return models.CodeRedisError
	}
	return models.CodeSuccess
}

func ReduceStock(orderParam models.OrderRequest) models.ResCode {
	for _, item := range orderParam.Items {
		rows := mysql.ReduceStockByBookID(item.BookID, item.Quantity)
		if rows <= 0 {
			return models.CodeInSufficient
		}
		err := redis.ReduceStockByBookID(item.BookID, item.Quantity)
		if err != nil {
			_ = redis.DeleteBookCache(item.BookID)
			return models.CodeRedisError
		}
	}
	return models.CodeSuccess
}

func GetUserOrder(UserID int64) ([]*models.OrderView, models.ResCode) {
	z, err, _ := redis.G.Do(strconv.FormatInt(UserID, 10), func() (interface{}, error) {
		Orders, missIDs, err := redis.GetUserOrdersInfo(UserID)
		if err != nil {
			return nil, err
		}
		if len(missIDs) > 0 || len(Orders) == 0 {
			Orders, err := mysql.GetUserOrdersInfo(UserID)
			if err != nil {
				return nil, err
			}
			var Ans []*models.OrderView
			for _, order := range Orders {
				Ans = append(Ans, &models.OrderView{
					OrderID:    order.OrderID,
					UserID:     order.UserID,
					TotalPrice: order.TotalPrice,
					Status:     order.Status,
					CreatedAt:  order.CreatedAt,
				})
			}
			_ = redis.SetUserOrdersInfo(Orders, missIDs)
			return Ans, nil
		}
		var Ans []*models.OrderView
		for _, order := range Orders {
			Ans = append(Ans, &models.OrderView{
				OrderID:    order.OrderID,
				UserID:     order.UserID,
				TotalPrice: order.TotalPrice,
				Status:     order.Status,
				CreatedAt:  order.CreatedAt,
			})
		}
		return Ans, nil
	})
	if err != nil {
		return nil, models.CodeServerBusy
	}
	return z.([]*models.OrderView), models.CodeSuccess
}

func GetOrderItems(userID int64, orderViews []*models.OrderView) models.ResCode {
	for _, ov := range orderViews {
		if ov == nil || ov.UserID != userID {
			return models.CodeOrderNotExist
		}
		if redis.TryLoadOrderItemsIntoView(ov) {
			continue
		}
		items, err := mysql.GetOrderItemsByOrderID(ov.OrderID)
		if err != nil {
			return models.CodeMySQLError
		}
		for _, item := range items {
			Book, _ := GetBookByID(item.BookID)
			item.Title = Book.Title
		}
		ov.Items = items
		if len(items) > 0 {
			if err := redis.SaveOrderItems(items); err != nil {
				return models.CodeRedisError
			}
		}
	}
	return models.CodeSuccess
}

func GetOrderDetailSecure(userID, orderID int64) (*models.OrderView, models.ResCode) {
	data, err := redis.GetOrderHash(orderID)
	if err != nil {
		return nil, models.CodeRedisError
	}
	var base *models.Order
	if len(data) > 0 {
		base = redis.OrderFromRedisHash(orderID, data)
		if base == nil {
			base, err = mysql.GetOrderByIDAndUser(orderID, userID)
			if err != nil {
				return nil, models.CodeOrderNotExist
			}
			if err := redis.SaveOrder(base); err != nil {
				return nil, models.CodeRedisError
			}
		} else if base.UserID != userID {
			return nil, models.CodeOrderNotExist
		}
	} else {
		base, err = mysql.GetOrderByIDAndUser(orderID, userID)
		if err != nil {
			return nil, models.CodeOrderNotExist
		}
		if err := redis.SaveOrder(base); err != nil {
			return nil, models.CodeRedisError
		}
	}
	ov := &models.OrderView{
		OrderID:    base.OrderID,
		UserID:     base.UserID,
		TotalPrice: base.TotalPrice,
		Status:     base.Status,
		CreatedAt:  base.CreatedAt,
	}
	if redis.TryLoadOrderItemsIntoView(ov) {
		return ov, models.CodeSuccess
	}
	items, err := mysql.GetOrderItemsByOrderID(orderID)
	if err != nil {
		return nil, models.CodeMySQLError
	}
	for _, item := range items {
		Book, _ := GetBookByID(item.BookID)
		item.Title = Book.Title
	}
	ov.Items = items
	if len(items) > 0 {
		if err := redis.SaveOrderItems(items); err != nil {
			return nil, models.CodeRedisError
		}
	}
	return ov, models.CodeSuccess
}

func ConfirmOrder(userID, orderID int64) models.ResCode {
	rows, err := mysql.ConfirmOrderAtomic(orderID, userID)
	if err != nil {
		return models.CodeMySQLError
	}
	if rows > 0 {
		if err := redis.UpdateOrderStatusInCache(orderID, 1); err != nil {
			return models.CodeRedisError
		}
		return models.CodeSuccess
	}
	o, err := mysql.GetOrderByIDAndUser(orderID, userID)
	if err != nil {
		return models.CodeOrderNotExist
	}
	if o.Status != 0 {
		return models.CodeOrderAlreadyConfirmed
	}
	return models.CodeOrderNotExist
}

func CancelOrder(orderID, userID int64) models.ResCode {
	rows, err := mysql.SetCancelStatus(orderID, userID)
	if rows <= 0 {
		return models.CodeOrderNotExist
	}
	if err != nil {
		return models.CodeMySQLError
	}
	Items, err := mysql.GetOrderItemsByOrderID(orderID)
	if err != nil {
		return models.CodeMySQLError
	}
	for _, item := range Items {
		err := mysql.RecoverStockByBookID(item.BookID, item.Quantity)
		if err != nil {
			return models.CodeMySQLError
		}
		err = redis.UpdateBookCacheStock(item.BookID, item.Quantity)
		if err != nil {
			return models.CodeRedisError
		}
		_ = redis.DeleteOrder(item.OrderID)
		_ = redis.DeleteOrderItem(item.ID)
	}
	return models.CodeSuccess
}
