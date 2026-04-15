package mysql

import (
	"Project1_Shop/models"

	"gorm.io/gorm/clause"
)

func GetBooksPageByScore(page int64) ([]*models.Book, int64, error) {
	var Books []*models.Book
	var TotalPage int64
	DB.Model(&models.Book{}).Count(&TotalPage)
	offset := (page - 1) * models.PageSize
	err := DB.Order("score DESC").Limit(models.PageSize).Offset(int(offset)).Find(&Books).Error //从高到低对分数排序后分页查询（加上Where还能筛选）
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

func GetBooksByIDs(IDs []int64) ([]*models.Book, error) {
	var Books []*models.Book
	err := DB.Where("book_id IN ?", IDs).Find(&Books).Error
	return Books, err
}

func GetBooksByTitle(Title string) ([]*models.Book, error) {
	var Books []*models.Book
	err := DB.Where("title LIKE ?", "%"+Title+"%").Find(&Books).Error
	return Books, err
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

func CheckRate(p *models.UserRateBook) bool {
	var Count int64
	DB.Model(&models.UserRateBook{}).Where("book_id = ? AND user_id = ?", p.BookID, p.UserID).Count(&Count)
	if Count > 0 {
		return true
	}
	return false
}

func UpdateRateBook(rateBook *models.RateBook) error {
	result := DB.Save(rateBook)
	return result.Error
}

func GetBeforeBookScore(BookID, UserID int64) (int64, error) {
	var Temp models.UserRateBook
	result := DB.Where("book_id = ? AND user_id = ?", BookID, UserID).First(&Temp)
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

func GetTagByName(name string) (*models.Tag, error) {
	tag := models.Tag{Name: name}
	err := DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).Create(&tag).Error
	if err != nil {
		return nil, err
	}
	// 如果是已存在，tag.ID 可能为 0，需要查一次
	if tag.ID == 0 {
		err = DB.Where("name = ?", name).First(&tag).Error
	}
	return &tag, err
}

func GetTagsByNames(names []string) ([]models.Tag, error) {
	var tags []models.Tag
	if err := DB.Where("name IN ?", names).Find(&tags).Error; err != nil {
		return nil, err
	}
	existing := make(map[string]models.Tag)
	for _, t := range tags {
		existing[t.Name] = t
	}
	var toCreate []models.Tag
	for _, name := range names {
		if _, ok := existing[name]; !ok {
			toCreate = append(toCreate, models.Tag{Name: name})
		}
	}
	if len(toCreate) > 0 {
		if err := DB.Create(&toCreate).Error; err != nil {
			return nil, err
		}
		tags = append(tags, toCreate...)
	}
	return tags, nil
}

func AddBookTag(bookID, tagID int64) error {
	return DB.Create(&models.BookTag{
		BookID: bookID,
		TagID:  tagID,
	}).Error
}

func GetTagsByBookID(bookID int64) ([]models.Tag, error) {
	var tags []models.Tag
	err := DB.Model(&models.Tag{}).
		Select("tags.id, tags.name").
		Joins("JOIN book_tags ON book_tags.tag_id = tags.id").
		Where("book_tags.book_id = ?", bookID).
		Scan(&tags).Error
	return tags, err
}

func DeleteBookToTag(BookID int64) error {
	return DB.Where("book_id = ?", BookID).Delete(&models.BookTag{}).Error
}

func GetUserRatingByUserID(UserID int64) ([]*models.UserRateBook, error) {
	var URB []*models.UserRateBook
	err := DB.Where("user_id = ?", UserID).Find(&URB).Error
	return URB, err
}
