package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
)

func GetPageBooks(Page int64) (*models.Page, error) {
	RedisBooks, total, err := redis.GetBooksPageByScore(Page)
	if err == nil && total != 0 {
		Pages := &models.Page{
			Page:  Page,
			Total: total,
			Data:  RedisBooks,
		}
		return Pages, nil
	}
	SQLBooks, total, err := mysql.GetBooksPageByScore(Page)
	if err != nil {
		return nil, err
	}
	for _, Book := range SQLBooks {
		T := BookListToCache(Book)
		_ = redis.SetBookSummary(T)
		err = redis.SetBook(Book, float64(Book.Score)/100)
		if err != nil {
			continue
		}
		err = redis.AddScoreRank(Book.BookID, float64(Book.Score)/100)
		if err != nil {
			continue
		}
	}
	Pages := &models.Page{
		Page:  Page,
		Total: total,
		Data:  SQLBooks,
	}
	return Pages, nil
}

func AddBook(p *models.AddBookParam) models.ResCode {
	exist := mysql.ExistBookByInfo(p.Title, p.Author, p.Publisher)
	if exist {
		return models.CodeBookExist
	}
	book := &models.Book{
		Title:        p.Title,
		Author:       p.Author,
		Publisher:    p.Publisher,
		Introduction: p.Introduction,
		Stock:        p.Stock,
		Sales:        0,
		Price:        p.Price,
		Score:        0,
		CoverImage:   p.CoverImage,
	}
	err := mysql.AddBook(book)
	if err != nil {
		return models.CodeServerBusy
	}
	for _, tagName := range p.Tags {
		tag, err := mysql.GetTagByName(tagName)
		if err != nil {
			return models.CodeServerBusy
		}
		err = mysql.AddBookTag(book.BookID, tag.ID)
		if err != nil {
			return models.CodeServerBusy
		}
	}
	err = redis.SetBook(book, 0)
	if err != nil {
		return models.CodeServerBusy
	}
	err = redis.AddScoreRank(book.BookID, 0)
	if err != nil {
		return models.CodeServerBusy
	}
	return models.CodeSuccess
}

func BookListToCache(book *models.Book) *models.ListBook {
	var tagNames []string
	for _, t := range book.Tags {
		tagNames = append(tagNames, t.Name)
	}
	return &models.ListBook{
		BookID: book.BookID,
		Title:  book.Title,
		Sales:  book.Sales,
		Score:  float64(book.Score) / 100,
		Tags:   tagNames,
	}
}

func GetBookByID(ID int64) (*models.BookCache, error) {
	bookCache, err := redis.GetBookByBookID(ID)
	if err == nil && bookCache.BookID != -1 {
		return bookCache, nil
	}
	book, err := mysql.GetBookByID(ID)
	if err != nil {
		return nil, err
	}
	tags, err := mysql.GetTagsByBookID(ID)
	if err != nil {
		return nil, err
	}
	var tagNames []string
	for _, t := range tags {
		tagNames = append(tagNames, t.Name)
	}
	cache := &models.BookCache{
		BookID:       book.BookID,
		Title:        book.Title,
		Sales:        book.Sales,
		Score:        float64(book.Score),
		Author:       book.Author,
		Publisher:    book.Publisher,
		Introduction: book.Introduction,
		Stock:        book.Stock,
		Price:        float64(book.Price),
		CoverImage:   book.CoverImage,
		Tags:         tagNames,
	}
	_ = redis.SetBook(book, float64(book.Score))
	return cache, nil
}

func DeleteBook(ID int64) models.ResCode {
	exist := mysql.ExistBook(ID)
	if !exist {
		return models.CodeBookNotExist
	}
	_ = mysql.DeleteBookToTag(ID)
	_ = mysql.DeleteBook(ID)
	err := redis.DeleteBook(ID)
	if err != nil {
		return models.CodeServerBusy
	}
	exist = mysql.ExistBook(ID)
	if !exist {
		return models.CodeSuccess
	}
	return models.CodeServerBusy
}

func UpdateBook(B *models.UpdateBookParam) models.ResCode {
	tags, err := mysql.GetTagsByNames(B.Tags)
	if err != nil {
		return models.CodeServerBusy
	}
	book := &models.Book{
		BookID:       B.BookID,
		Title:        B.Title,
		Author:       B.Author,
		Publisher:    B.Publisher,
		Introduction: B.Introduction,
		Stock:        B.Stock,
		Price:        B.Price,
		CoverImage:   B.CoverImage,
		Tags:         tags,
	}
	if err := mysql.UpdateBook(book); err != nil {
		return models.CodeServerBusy
	}
	err = redis.UpdateBookWithOutScore(book)
	if err != nil {
		return models.CodeServerBusy
	}
	return models.CodeSuccess
}
