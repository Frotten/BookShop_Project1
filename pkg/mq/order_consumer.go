package mq

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"

	"go.uber.org/zap"
)

func StartOrderExpiredConsumer() error {
	return StartConsumer(OrderExpiredQueue, handleOrderExpired)
}

func StartOrderShippingConsumer() error {
	return StartConsumer(OrderShippingQueue, handleOrderShipping)
}

func handleOrderExpired(orderID int64) bool {
	zap.L().Info("order expired, start cancel", zap.Int64("order_id", orderID))
	order, err := mysql.GetOrderByID(orderID)
	if err != nil {
		zap.L().Error("GetOrderByID failed", zap.Int64("order_id", orderID), zap.Error(err))
		return true
	}
	if order.Status != 0 {
		zap.L().Info("order already processed, skip cancel",
			zap.Int64("order_id", orderID),
			zap.Int8("status", order.Status),
		)
		return true
	}
	rows, err := mysql.SetCancelStatusByID(orderID)
	if err != nil {
		zap.L().Error("SetCancelStatusByID failed", zap.Int64("order_id", orderID), zap.Error(err))
		return false
	}
	if rows == 0 {
		zap.L().Info("order status changed by concurrent request, skip",
			zap.Int64("order_id", orderID),
		)
		return true
	}
	items, err := mysql.GetOrderItemsByOrderID(orderID)
	if err != nil {
		zap.L().Error("GetOrderItemsByOrderID failed", zap.Int64("order_id", orderID), zap.Error(err))
		return true
	}
	for _, item := range items {
		if err := mysql.RecoverStockByBookID(item.BookID, item.Quantity); err != nil {
			zap.L().Error("RecoverStockByBookID failed",
				zap.Int64("order_id", orderID),
				zap.Int64("book_id", item.BookID),
				zap.Error(err),
			)
		}
		if err := redis.UpdateBookCacheStock(item.BookID, item.Quantity); err != nil {
			_ = redis.DeleteBookCache(item.BookID)
			zap.L().Warn("UpdateBookCacheStock failed, cache deleted",
				zap.Int64("book_id", item.BookID),
				zap.Error(err),
			)
		}
		_ = redis.DeleteOrder(item.OrderID)
		_ = redis.DeleteOrderItem(item.ID)
	}
	zap.L().Info("order expired and cancelled successfully", zap.Int64("order_id", orderID))
	return true
}

func handleOrderShipping(orderID int64) bool {
	zap.L().Info("order confirmed, entered shipping queue",
		zap.Int64("order_id", orderID),
	)
	// TODO: 后续可在此处对接仓库系统 / 发货通知 / 短信推送等
	return true
}
