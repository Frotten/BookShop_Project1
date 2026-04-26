package mysql

import (
	"Project1_Shop/models"
)

func CreateSeckillProduct(sp *models.SeckillProduct) error {
	return DB.Create(sp).Error
}

func GetSeckillProducts() ([]*models.SeckillProduct, error) {
	var list []*models.SeckillProduct
	err := DB.Where("status = 1").Order("start_time ASC").Find(&list).Error
	return list, err
}

func GetSeckillProductByID(id int64) (*models.SeckillProduct, error) {
	var sp models.SeckillProduct
	err := DB.First(&sp, id).Error
	if err != nil {
		return nil, err
	}
	return &sp, nil
}

func DownSeckillProduct(id int64) error {
	return DB.Model(&models.SeckillProduct{}).
		Where("id = ?", id).
		Update("status", 0).Error
}

func CreateSeckillOrder(so *models.SeckillOrder) error {
	return DB.Create(so).Error
}

func SetSeckillOrderStatus(id int64, orderID int64, status int8) error {
	return DB.Model(&models.SeckillOrder{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"order_id": orderID,
			"status":   status,
		}).Error
}

func GetSeckillOrderByID(id int64) (*models.SeckillOrder, error) {
	var so models.SeckillOrder
	err := DB.First(&so, id).Error
	if err != nil {
		return nil, err
	}
	return &so, nil
}
