package mysql

import "Project1_Shop/models"

func CreateOrder(Order *models.Order) error {
	return DB.Create(&Order).Error
}

func CreateOrderItems(OrderItems []*models.OrderItem) error {
	return DB.Create(&OrderItems).Error
}
