package redis

import (
	"Project1_Shop/models"
	"errors"
	"strconv"
)

func SetCart(Cart *models.Cart) error {
	key := "cart:" + strconv.FormatInt(Cart.UserID, 10)
	field := strconv.FormatInt(Cart.BookID, 10)
	value := strconv.FormatInt(Cart.Quantity, 10)
	return RDB.HSet(ctx, key, field, value).Err()
}

func SetCartList(UserID int64, cartList []models.CartView) {
	key := "cart:" + strconv.FormatInt(UserID, 10)
	data := make(map[string]interface{})
	for _, cart := range cartList {
		field := strconv.FormatInt(cart.BookID, 10)
		value := strconv.FormatInt(cart.Quantity, 10)
		data[field] = value
	}
	RDB.HMSet(ctx, key, data)
}

func GetCartRaw(UserID int64) ([]int64, map[int64]int64, error) {
	key := "cart:" + strconv.FormatInt(UserID, 10)
	data, err := RDB.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, nil, err
	}
	if len(data) == 0 {
		return nil, nil, errors.New("cache miss")
	}
	var ids []int64
	quantityMap := make(map[int64]int64)
	for field, value := range data {
		id, _ := strconv.ParseInt(field, 10, 64)
		qty, _ := strconv.ParseInt(value, 10, 64)

		ids = append(ids, id)
		quantityMap[id] = qty
	}
	return ids, quantityMap, nil
}

func DelCartItem(UserID, BookID int64) error {
	key := "cart:" + strconv.FormatInt(UserID, 10)
	field := strconv.FormatInt(BookID, 10)
	return RDB.HDel(ctx, key, field).Err()
}
