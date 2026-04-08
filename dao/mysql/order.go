package mysql

import "Project1_Shop/models"

func CreateOrder(Order *models.Order) error {
	return DB.Create(&Order).Error
}
