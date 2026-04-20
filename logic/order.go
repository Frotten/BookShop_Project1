package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
	"Project1_Shop/pkg/mq"
	"strconv"
	"time"

	"go.uber.org/zap"
)

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
	if err = mysql.CreateOrder(Order); err != nil {
		return models.CodeMySQLError
	}
	if err = redis.SaveOrder(Order); err != nil {
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
	if err = mysql.CreateOrderItems(OrderItems); err != nil {
		return models.CodeMySQLError
	}
	if err = redis.SaveOrderItems(OrderItems); err != nil {
		return models.CodeRedisError
	}
	if err = mq.PublishOrderPending(Order.OrderID); err != nil {
		zap.L().Error("PublishOrderPending failed",
			zap.Int64("order_id", Order.OrderID),
			zap.Error(err),
		)
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

func GetOrderDetailSecure(orderID int64) (*models.OrderView, models.ResCode) {
	data, err := redis.GetOrderHash(orderID)
	if err != nil {
		return nil, models.CodeRedisError
	}
	var base *models.Order
	if len(data) > 0 {
		base = redis.OrderFromRedisHash(orderID, data)
		if base == nil {
			base, err = mysql.GetOrderByID(orderID)
			if err != nil {
				return nil, models.CodeOrderNotExist
			}
			if err := redis.SaveOrder(base); err != nil {
				return nil, models.CodeRedisError
			}
		}
	} else {
		base, err = mysql.GetOrderByID(orderID)
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

func OrderPay(orderID int64) models.ResCode { //添加统计，sale表中增加销量，以上操作可以异步完成
	rows, err := mysql.SetOrderStatus(orderID, 0, 1)
	if err != nil {
		return models.CodeMySQLError
	}
	if rows > 0 {
		if err := redis.UpdateOrderStatusInCache(orderID, 1); err != nil {
			return models.CodeRedisError
		}
		if err = redis.CacheOrderID(orderID); err != nil {
			return models.CodeRedisError
		}
		err = mq.PublishOrderPayment(orderID)
		if err != nil {
			zap.L().Error("PublishOrderPayment failed", zap.Error(err))
			return models.CodeServerBusy
		}
		return models.CodeSuccess
	}
	o, err := mysql.GetOrderByIDAndUser(orderID)
	if err != nil {
		return models.CodeOrderNotExist
	}
	if o.Status != 0 {
		return models.CodeOrderAlreadyConfirmed
	}
	return models.CodeOrderNotExist
}

func OrderCancel(orderID int64) models.ResCode {
	Order, res := GetOrderDetailSecure(orderID)
	if res != models.CodeSuccess {
		return res
	}
	if Order.Status == 1 {
		_ = redis.DeleteOrderIDFromShipCache(orderID)
	}
	if Order.Status >= 1 {
		for _, item := range Order.Items {
			err := mysql.ReduceSale(item.BookID, item.Quantity)
			if err != nil {
				return models.CodeMySQLError
			}
			_ = redis.ReduceSale(item.BookID, item.Quantity)
		}
	}
	rows, err := mysql.SetCancelStatusByID(orderID)
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

func GetShipOrder() ([]*models.OrderView, models.ResCode) {
	IDs, err := redis.GetShipOrderID()
	if err != nil || len(IDs) == 0 {
		IDs, err := mysql.GetShipOrderIDByStatus(1)
		if err != nil {
			return nil, models.CodeServerBusy
		}
		var OV []*models.OrderView
		for _, ID := range IDs {
			View, Res := GetOrderDetailSecure(ID)
			if Res != models.CodeSuccess {
				continue
			}
			OV = append(OV, View)
		}
		return OV, models.CodeSuccess
	}
	var OV []*models.OrderView
	for _, ID := range IDs {
		id, _ := strconv.ParseInt(ID, 10, 64)
		View, Res := GetOrderDetailSecure(id)
		if Res != models.CodeSuccess {
			continue
		}
		OV = append(OV, View)
	}
	return OV, models.CodeSuccess
}

func OrderShip(orderID int64) models.ResCode { //如果有具体发货操作的话应该在这补充，但目前只有一个状态变更，所以就直接改状态了
	rows, err := mysql.SetOrderStatus(orderID, 1, 2)
	if err != nil {
		return models.CodeMySQLError
	}
	if rows > 0 {
		if err := redis.UpdateOrderStatusInCache(orderID, 2); err != nil {
			return models.CodeRedisError
		}
		if err = redis.DeleteOrderIDFromShipCache(orderID); err != nil {
			return models.CodeRedisError
		}
		return models.CodeSuccess
	}
	o, err := mysql.GetOrderByIDAndUser(orderID)
	if err != nil {
		return models.CodeOrderNotExist
	}
	if o.Status != 1 {
		return models.CodeOrderAlreadyConfirmed
	}
	return models.CodeOrderNotExist
}

func OrderConfirm(orderID int64) models.ResCode {
	rows, err := mysql.SetOrderStatus(orderID, 2, 3)
	if err != nil {
		return models.CodeMySQLError
	}
	if rows > 0 {
		if err := redis.UpdateOrderStatusInCache(orderID, 3); err != nil {
			return models.CodeRedisError
		}
		return models.CodeSuccess
	}
	o, err := mysql.GetOrderByIDAndUser(orderID)
	if err != nil {
		return models.CodeOrderNotExist
	}
	if o.Status != 2 {
		return models.CodeOrderAlreadyConfirmed
	}
	return models.CodeOrderNotExist
}
