package mysql

import "Project1_Shop/models"

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
