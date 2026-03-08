package mysql

import "Project1_Shop/models"

const PageSize = 10

func GetAllBooks() ([]*models.Book, error) {
	var Books []*models.Book
	result := DB.Find(&Books)
	if result.RowsAffected == 0 {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return Books, nil
}

func GetBooksByPage(page int) ([]*models.Book, int64, error) {
	var Books []*models.Book
	var TotalPage int64
	DB.Model(&models.Book{}).Count(&TotalPage)
	offset := (page - 1) * PageSize
	err := DB.Limit(PageSize).Offset(offset).Find(&Books).Error
	return Books, TotalPage, err
}
