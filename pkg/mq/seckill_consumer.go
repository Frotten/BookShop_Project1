package mq

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
	"time"

	"go.uber.org/zap"
)

func StartSeckillConsumer() error {
	return StartConsumer(SeckillQueue, handleSeckillOrder)
}

func handleSeckillOrder(seckillID int64) bool {
	so, err := mysql.GetSeckillOrderByID(seckillID)
	if err != nil {
		zap.L().Error("handleSeckillOrder: GetSeckillOrderByID failed",
			zap.Int64("seckill_id", seckillID), zap.Error(err))
		return false
	}
	if so.Status != 0 {
		zap.L().Info("handleSeckillOrder: already processed", zap.Int64("id", seckillID))
		return true
	}
	sp, err := mysql.GetSeckillProductByID(so.ProductID)
	if err != nil || sp.Status != 1 {
		zap.L().Warn("handleSeckillOrder: seckill product invalid", zap.Int64("product_id", so.ProductID))
		_ = mysql.SetSeckillOrderStatus(so.ID, 0, 2)
		_ = redis.RevokeSeckillUser(so.UserID, so.ProductID)
		_ = redis.RecoverSeckillStock(so.ProductID)
		return true
	}
	rows := mysql.ReduceStockByBookID(so.BookID, 1)
	if rows <= 0 {
		zap.L().Error("handleSeckillOrder: ReduceStockByBookID insufficient",
			zap.Int64("book_id", so.BookID))
		_ = mysql.SetSeckillOrderStatus(so.ID, 0, 2)
		_ = redis.RevokeSeckillUser(so.UserID, so.ProductID)
		_ = redis.RecoverSeckillStock(so.ProductID)
		return true
	}
	_ = redis.ReduceStockByBookID(so.BookID, 1)
	order := &models.Order{
		UserID:     so.UserID,
		Status:     0, // 未支付
		TotalPrice: so.Price,
		CreatedAt:  time.Now().Format(models.TimeParseLayout),
	}
	if err := mysql.CreateOrder(order); err != nil {
		zap.L().Error("handleSeckillOrder: CreateOrder failed", zap.Error(err))
		_ = mysql.RecoverStockByBookID(so.BookID, 1)
		_ = redis.UpdateBookCacheStock(so.BookID, 1)
		return false
	}
	items := []*models.OrderItem{
		{
			OrderID:  order.OrderID,
			BookID:   so.BookID,
			Price:    so.Price,
			Quantity: 1,
		},
	}
	if err := mysql.CreateOrderItems(items); err != nil {
		zap.L().Error("handleSeckillOrder: CreateOrderItems failed", zap.Error(err))
		return false
	}
	_ = redis.SaveOrder(order)
	_ = redis.SaveOrderItems(items)
	if err := mysql.SetSeckillOrderStatus(so.ID, order.OrderID, 1); err != nil {
		zap.L().Error("handleSeckillOrder: SetSeckillOrderStatus failed", zap.Error(err))
		return false
	}
	if err := PublishOrderPending(order.OrderID); err != nil {
		zap.L().Warn("handleSeckillOrder: PublishOrderPending failed", zap.Error(err))
	}
	zap.L().Info("handleSeckillOrder: success",
		zap.Int64("seckill_order_id", so.ID),
		zap.Int64("order_id", order.OrderID),
		zap.Int64("user_id", so.UserID),
	)
	return true
}
