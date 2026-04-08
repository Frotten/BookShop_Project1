package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
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
