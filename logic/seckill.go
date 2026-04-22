package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
	"Project1_Shop/pkg/mq"
	"time"

	"go.uber.org/zap"
)

func DoSeckill(userID, productID int64) models.ResCode {
	spView, err := redis.GetSeckillProductFromCache(productID)
	if err != nil || spView == nil {
		sp, err := mysql.GetSeckillProductByID(productID)
		if err != nil || sp.Status != 1 {
			return models.CodeSeckillNotFound
		}
		_ = redis.CacheSeckillProduct(sp)
		spView = &models.SeckillProductView{
			ID:           sp.ID,
			BookID:       sp.BookID,
			Title:        sp.Title,
			SeckillPrice: sp.SeckillPrice,
			OrigPrice:    sp.OrigPrice,
			StartTime:    sp.StartTime.Format(models.TimeParseLayout),
			EndTime:      sp.EndTime.Format(models.TimeParseLayout),
			Status:       sp.Status,
		}
	}
	now := time.Now()
	startTime, err := time.ParseInLocation(models.TimeParseLayout, spView.StartTime, time.Local)
	if err != nil || now.Before(startTime) {
		return models.CodeSeckillNotFound
	}
	endTime, err := time.ParseInLocation(models.TimeParseLayout, spView.EndTime, time.Local)
	if err != nil || now.After(endTime) {
		return models.CodeSeckillEnded
	}
	dedupTTL := time.Until(endTime) + time.Hour
	luaResult, err := redis.TrySeckill(userID, productID, dedupTTL)
	if err != nil {
		zap.L().Error("DoSeckill TrySeckill failed",
			zap.Int64("user_id", userID),
			zap.Int64("product_id", productID),
			zap.Error(err),
		)
		return models.CodeServerBusy
	}
	switch luaResult {
	case 1:
		return models.CodeSeckillDuplicate
	case 2:
		return models.CodeSeckillSoldOut
	}
	so := &models.SeckillOrder{
		UserID:    userID,
		ProductID: productID,
		BookID:    spView.BookID,
		Price:     spView.SeckillPrice,
		Status:    0,
		CreatedAt: time.Now().Format(models.TimeParseLayout),
	}
	if err := mysql.CreateSeckillOrder(so); err != nil {
		_ = redis.RecoverSeckillStock(productID)
		_ = redis.RevokeSeckillUser(userID, productID)
		zap.L().Error("DoSeckill CreateSeckillOrder failed", zap.Error(err))
		return models.CodeServerBusy
	}
	msg := &models.SeckillMsg{
		UserID:    userID,
		ProductID: productID,
	}
	if err := mq.PublishSeckillOrder(msg); err != nil {
		zap.L().Error("DoSeckill PublishSeckillOrder failed",
			zap.Int64("seckill_order_id", so.ID),
			zap.Error(err),
		)
		return doSeckillSync(so, spView)
	}

	return models.CodeSuccess
}

func doSeckillSync(so *models.SeckillOrder, spView *models.SeckillProductView) models.ResCode {
	rows := mysql.ReduceStockByBookID(spView.BookID, 1)
	if rows <= 0 {
		_ = mysql.SetSeckillOrderStatus(so.ID, 0, 2)
		_ = redis.RevokeSeckillUser(so.UserID, so.ProductID)
		_ = redis.RecoverSeckillStock(so.ProductID)
		return models.CodeSeckillSoldOut
	}
	_ = redis.ReduceStockByBookID(spView.BookID, 1)

	order := &models.Order{
		UserID:     so.UserID,
		Status:     0,
		TotalPrice: so.Price,
		CreatedAt:  time.Now().Format(models.TimeParseLayout),
	}
	if err := mysql.CreateOrder(order); err != nil {
		_ = mysql.RecoverStockByBookID(spView.BookID, 1)
		_ = redis.UpdateBookCacheStock(spView.BookID, 1)
		return models.CodeServerBusy
	}
	items := []*models.OrderItem{
		{OrderID: order.OrderID, BookID: spView.BookID, Price: so.Price, Quantity: 1},
	}
	_ = mysql.CreateOrderItems(items)
	_ = redis.SaveOrder(order)
	_ = redis.SaveOrderItems(items)
	_ = mysql.SetSeckillOrderStatus(so.ID, order.OrderID, 1)
	_ = mq.PublishOrderPending(order.OrderID)
	return models.CodeSuccess
}

func GetActiveSeckillList() ([]*models.SeckillProductView, models.ResCode) {
	ids, err := redis.GetActiveSeckillIDs()
	if err != nil || len(ids) == 0 {
		return loadSeckillFromMySQL()
	}
	var list []*models.SeckillProductView
	for _, id := range ids {
		view, err := redis.GetSeckillProductFromCache(id)
		if err != nil || view == nil {
			sp, err := mysql.GetSeckillProductByID(id)
			if err != nil {
				continue
			}
			_ = redis.CacheSeckillProduct(sp)
			stock, _ := redis.GetSeckillStock(id)
			view = &models.SeckillProductView{
				ID:           sp.ID,
				BookID:       sp.BookID,
				Title:        sp.Title,
				SeckillPrice: sp.SeckillPrice,
				OrigPrice:    sp.OrigPrice,
				Stock:        stock,
				StartTime:    sp.StartTime.Format(models.TimeParseLayout),
				EndTime:      sp.EndTime.Format(models.TimeParseLayout),
				Status:       sp.Status,
			}
		}
		now := time.Now()
		endT, _ := time.ParseInLocation(models.TimeParseLayout, view.EndTime, time.Local)
		startT, _ := time.ParseInLocation(models.TimeParseLayout, view.StartTime, time.Local)
		if now.After(endT) || now.Before(startT) {
			continue
		}
		list = append(list, view)
	}
	if len(list) == 0 {
		return loadSeckillFromMySQL()
	}
	return list, models.CodeSuccess
}

func loadSeckillFromMySQL() ([]*models.SeckillProductView, models.ResCode) {
	sps, err := mysql.GetSeckillProducts()
	if err != nil {
		return nil, models.CodeMySQLError
	}
	var list []*models.SeckillProductView
	now := time.Now()
	for _, sp := range sps {
		if now.Before(sp.StartTime) || now.After(sp.EndTime) {
			continue
		}
		_ = redis.CacheSeckillProduct(sp)
		stock := sp.Stock
		if s, err := redis.GetSeckillStock(sp.ID); err == nil {
			stock = s
		}
		list = append(list, &models.SeckillProductView{
			ID:           sp.ID,
			BookID:       sp.BookID,
			Title:        sp.Title,
			SeckillPrice: sp.SeckillPrice,
			OrigPrice:    sp.OrigPrice,
			Stock:        stock,
			StartTime:    sp.StartTime.Format(models.TimeParseLayout),
			EndTime:      sp.EndTime.Format(models.TimeParseLayout),
			Status:       sp.Status,
		})
	}
	return list, models.CodeSuccess
}

func AdminCreateSeckill(param *models.SeckillParam) models.ResCode {
	book, err := GetBookByID(param.BookID)
	if err != nil {
		return models.CodeBookNotExist
	}
	startTime, err := time.ParseInLocation(models.TimeParseLayout, param.StartTime, time.Local)
	if err != nil {
		return models.CodeInvalidParam
	}
	endTime, err := time.ParseInLocation(models.TimeParseLayout, param.EndTime, time.Local)
	if err != nil {
		return models.CodeInvalidParam
	}
	if endTime.Before(startTime) || endTime.Before(time.Now()) {
		return models.CodeInvalidParam
	}
	sp := &models.SeckillProduct{
		BookID:       param.BookID,
		Title:        book.Title,
		SeckillPrice: param.SeckillPrice,
		OrigPrice:    book.Price,
		Stock:        param.Stock,
		StartTime:    startTime,
		EndTime:      endTime,
		Status:       1,
	}
	if err := mysql.CreateSeckillProduct(sp); err != nil {
		zap.L().Error("AdminCreateSeckill failed", zap.Error(err))
		return models.CodeMySQLError
	}
	if err := redis.InitSeckillStock(sp.ID, sp.Stock, sp.EndTime); err != nil {
		zap.L().Error("AdminCreateSeckill InitSeckillStock failed", zap.Error(err))
	}
	_ = redis.CacheSeckillProduct(sp)
	zap.L().Info("AdminCreateSeckill success",
		zap.Int64("product_id", sp.ID),
		zap.String("title", sp.Title),
		zap.String("end_time", param.EndTime),
	)
	return models.CodeSuccess
}

func AdminDownSeckill(productID int64) models.ResCode {
	if err := mysql.DownSeckillProduct(productID); err != nil {
		return models.CodeMySQLError
	}
	_ = redis.RemoveSeckillFromActive(productID)
	zap.L().Info("AdminDownSeckill success", zap.Int64("product_id", productID))
	return models.CodeSuccess
}

func GetSeckillDetail(productID int64) (*models.SeckillProductView, models.ResCode) {
	view, err := redis.GetSeckillProductFromCache(productID)
	if err == nil && view != nil {
		return view, models.CodeSuccess
	}
	sp, err := mysql.GetSeckillProductByID(productID)
	if err != nil {
		return nil, models.CodeSeckillNotFound
	}
	_ = redis.CacheSeckillProduct(sp)
	stock, _ := redis.GetSeckillStock(sp.ID)
	return &models.SeckillProductView{
		ID:           sp.ID,
		BookID:       sp.BookID,
		Title:        sp.Title,
		SeckillPrice: sp.SeckillPrice,
		OrigPrice:    sp.OrigPrice,
		Stock:        stock,
		StartTime:    sp.StartTime.Format(models.TimeParseLayout),
		EndTime:      sp.EndTime.Format(models.TimeParseLayout),
		Status:       sp.Status,
	}, models.CodeSuccess
}
