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

func GetBooksPageByScore(page int64) ([]*models.Book, int64, error) {
	var Books []*models.Book
	var TotalPage int64
	DB.Model(&models.Book{}).Count(&TotalPage)
	offset := (page - 1) * PageSize
	err := DB.Order("score DESC").Limit(PageSize).Offset(int(offset)).Find(&Books).Error //从高到低对分数排序后分页查询（加上Where还能筛选）
	return Books, TotalPage, err
}

func AddBook(book *models.Book) error {
	result := DB.Create(book)
	return result.Error
}

func ExistBook(ID int64) bool {
	var Book models.Book
	result := DB.Where("book_id = ?", ID).First(&Book)
	if result.RowsAffected == 0 {
		return false
	}
	return true
}

func ExistBookByInfo(Title, Author, Publisher string) bool {
	var Book models.Book
	result := DB.Where("title = ? AND author = ? AND publisher = ?", Title, Author, Publisher).First(&Book)
	if result.RowsAffected == 0 {
		return false
	}
	return true
}

func GetBookByID(ID int64) (*models.Book, error) {
	var Book models.Book
	result := DB.First(&Book, ID)
	if result.RowsAffected == 0 {
		return nil, result.Error
	}
	return &Book, result.Error
}

func DeleteBook(ID int64) error {
	result := DB.Where("book_id = ?", ID).Delete(&models.Book{})
	return result.Error
}
