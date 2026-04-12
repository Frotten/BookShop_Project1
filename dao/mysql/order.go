package mysql

import "Project1_Shop/models"

func CreateOrder(Order *models.Order) error {
	return DB.Create(&Order).Error
}

func CreateOrderItems(OrderItems []*models.OrderItem) error {
	return DB.Create(&OrderItems).Error
}

func GetUserOrdersInfo(UserID int64) ([]*models.Order, error) {
	var Ans []*models.Order
	err := DB.Where("user_id = ?", UserID).Find(&Ans).Error
	return Ans, err
}
