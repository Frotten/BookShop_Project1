package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
)

func AddBookToCart(CartParam *models.CartParam) models.ResCode {
	Cart := &models.Cart{
		UserID:   CartParam.UserID,
		BookID:   CartParam.BookID,
		Quantity: CartParam.Quantity,
	}
	err := mysql.AddBookToCart(Cart)
	if err != nil {
		return models.CodeMySQLError
	}
	err = redis.SetCart(Cart)
	if err != nil {
		return models.CodeRedisError
	}
	return models.CodeSuccess
}

func buildCartView(
	UserID int64,
	bookList []*models.BookCache,
	quantityMap map[int64]int64,
) []models.CartView {
	var cartViews []models.CartView
	for _, book := range bookList {
		cartViews = append(cartViews, models.CartView{
			UserID:     UserID,
			BookID:     book.BookID,
			Title:      book.Title,
			Price:      book.Price,
			Stock:      book.Stock,
			Author:     book.Author,
			CoverImage: book.CoverImage,
			Quantity:   quantityMap[book.BookID],
		})
	}
	return cartViews
}

func GetCartList(UserID int64) ([]models.CartView, models.ResCode) {
	ids, quantityMap, err := redis.GetCartRaw(UserID)
	if err == nil {
		bookList, err := GetBooksByIDs(ids)
		if err != nil {
			return nil, models.CodeServerBusy
		}
		return buildCartView(UserID, bookList, quantityMap), models.CodeSuccess
	}
	cartList, err := mysql.GetCartList(UserID)
	if err != nil {
		return nil, models.CodeMySQLError
	}
	go redis.SetCartList(UserID, cartList)
	return cartList, models.CodeSuccess
}

func UpdateCartItem(CartParam *models.CartParam) models.ResCode {
	err := mysql.UpdateCartItem(CartParam)
	if err != nil {
		return models.CodeMySQLError
	}
	_ = redis.DelCartItem(CartParam.UserID, CartParam.BookID)
	return models.CodeSuccess
}
