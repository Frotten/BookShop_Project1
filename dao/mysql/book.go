package mysql

import (
	"Project1_Shop/models"
)

const PageSize = 8

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

func UpdateBook(book *models.Book) error {
	result := DB.Save(book)
	return result.Error
}

func GetRateBookByID(ID int64) (*models.RateBook, error) {
	var rateBook models.RateBook
	result := DB.Where("book_id = ?", ID).First(&rateBook)
	if result.RowsAffected == 0 {
		DB.Create(&models.RateBook{
			BookID:     ID,
			ScoreCount: 0,
			Score:      0,
			Sale:       0,
		})
		result = DB.Where("book_id = ?", ID).First(&rateBook)
		if result.RowsAffected == 0 {
			return nil, result.Error
		}
		return &rateBook, result.Error
	}
	return &rateBook, result.Error
}

func UpdateRateBook(rateBook *models.RateBook) error {
	result := DB.Save(rateBook)
	return result.Error
}

func GetBeforeBookScore(p *models.UserRateBook) (int64, error) {
	var Temp models.UserRateBook
	result := DB.Where("book_id = ? AND user_id = ?", p.BookID, p.UserID).First(&Temp)
	return Temp.Score, result.Error
}

func UpdateUserRate(p *models.UserRateBook) error {
	result := DB.Save(p)
	return result.Error
}

func UpdateBookScore(RB *models.RateBook) error {
	Count := RB.ScoreCount
	Score := RB.Score
	Ans := models.WeightedCalculation(Score, Count)
	AnsInt := int64(Ans * 100)
	return DB.Model(&models.Book{}).Where("book_id = ?", RB.BookID).Update("score", AnsInt).Error
}
