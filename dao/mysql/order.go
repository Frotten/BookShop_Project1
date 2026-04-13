package mysql

import (
	"Project1_Shop/models"

	"gorm.io/gorm"
)

func CreateOrder(Order *models.Order) error {
	return DB.Create(Order).Error
}

func CreateOrderItems(OrderItems []*models.OrderItem) error {
	return DB.Create(&OrderItems).Error
}

func ReduceStockByBookID(BookID, Quantity int64) int64 {
	return DB.Model(&models.Book{}).
		Where("book_id = ? AND stock >= ?", BookID, Quantity).
		Update("stock", gorm.Expr("stock - ?", Quantity)).RowsAffected
}

func RecoverStockByBookID(BookID, Quantity int64) error {
	return DB.Model(&models.Book{}).
		Where("book_id = ?", BookID).
		Update("stock", gorm.Expr("stock + ?", Quantity)).Error
}

func GetUserOrdersInfo(UserID int64) ([]*models.Order, error) {
	var Ans []*models.Order
	err := DB.Where("user_id = ?", UserID).Order("order_id DESC").Find(&Ans).Error
	return Ans, err
}

func GetOrderByIDAndUser(orderID, userID int64) (*models.Order, error) {
	var o models.Order
	err := DB.Where("order_id = ? AND user_id = ?", orderID, userID).First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func GetOrderByID(orderID int64) (*models.Order, error) {
	var o models.Order
	err := DB.Where("order_id = ?", orderID).First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func GetOrderItemsByOrderID(orderID int64) ([]*models.OrderItem, error) {
	var items []*models.OrderItem
	err := DB.Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}

func ConfirmOrderAtomic(orderID, userID int64) (rowsAffected int64, err error) {
	res := DB.Model(&models.Order{}).
		Where("order_id = ? AND user_id = ? AND status = ?", orderID, userID, 0).
		Update("status", 1)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func SetCancelStatusByID(orderID int64) (rowsAffected int64, err error) {
	res := DB.Model(&models.Order{}).
		Where("order_id = ? AND status = ?", orderID, 0).
		Update("status", -1)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}
