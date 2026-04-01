package mysql

import "Project1_Shop/models"

func AddBookToCart(Cart *models.Cart) error {
	return DB.Save(Cart).Error
}

func GetCartList(UserID int64) ([]models.CartView, error) {
	var CartViewList []models.CartView
	var CartList []models.Cart
	err := DB.Where("user_id = ?", UserID).Find(&CartList).Error
	if err != nil {
		return nil, err
	}
	var ids []int64
	quantityMap := make(map[int64]int64)
	for _, cart := range CartList {
		ids = append(ids, cart.BookID)
		quantityMap[cart.BookID] = cart.Quantity
	}
	bookList, err := GetBooksByIDs(ids)
	if err != nil {
		return nil, err
	}
	for _, book := range bookList {
		CartViewList = append(CartViewList, models.CartView{
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
	return CartViewList, nil
}

func UpdateCartItem(CartParam *models.CartParam) error {
	return DB.Model(&models.Cart{}).Where("user_id = ? AND book_id = ?", CartParam.UserID, CartParam.BookID).Update("quantity", CartParam.Quantity).Error
}
