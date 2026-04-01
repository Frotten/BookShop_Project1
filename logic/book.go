package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
	"strconv"
)

func BookToCache(book *models.Book) *models.BookCache {
	Tags, err := mysql.GetTagsByBookID(book.BookID)
	if err != nil {
		return nil
	}
	var tagNames []string
	for _, t := range Tags {
		tagNames = append(tagNames, t.Name)
	}
	return &models.BookCache{
		BookID:       book.BookID,
		Title:        book.Title,
		Author:       book.Author,
		Publisher:    book.Publisher,
		Introduction: book.Introduction,
		Stock:        book.Stock,
		Price:        book.Price,
		CoverImage:   book.CoverImage,
		Tags:         tagNames,
		Sales:        book.Sales,
		Score:        book.Score,
	}
}

func GetPageBooks(Page int64) (*models.Page, error) {
	start := (Page - 1) * models.PageSize
	end := start + models.PageSize - 1
	ids, total, err := redis.GetRankIDsByScore(start, end)
	if err != nil {
		return nil, err
	}
	books, err := GetBooksByIDs(ids)
	if err == nil {
		return &models.Page{
			Page:  Page,
			Data:  books,
			Total: total,
		}, nil
	}
	SQLBooks, total, err := mysql.GetBooksPageByScore(Page)
	if err != nil {
		return nil, err
	}
	for _, Book := range SQLBooks {
		T := BookListToCache(Book)
		_ = redis.SetBookSummary(T)
		BookCache := BookToCache(Book)
		err = redis.SetBookCache(BookCache, Book.Score)
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
		return models.CodeMySQLError
	}
	for _, tagName := range p.Tags {
		tag, err := mysql.GetTagByName(tagName)
		if err != nil {
			return models.CodeMySQLError
		}
		err = mysql.AddBookTag(book.BookID, tag.ID)
		if err != nil {
			return models.CodeMySQLError
		}
	}
	BookCache := BookToCache(book)
	err = redis.SetBookCache(BookCache, 0)
	if err != nil {
		return models.CodeRedisError
	}
	err = redis.AddScoreRank(book.BookID, 0)
	if err != nil {
		return models.CodeRedisError
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
		Score:  book.Score,
		Tags:   tagNames,
	}
}

func GetBookByID(ID int64) (*models.BookCache, error) {
	bookCache, err := redis.GetBookByBookID(ID)
	if err == nil && bookCache.BookID != -1 {
		return bookCache, nil
	}
	key := strconv.FormatInt(ID, 10)
	v, err, _ := redis.G.Do(key, func() (interface{}, error) {
		bookCache, err := redis.GetBookByBookID(ID)
		if err == nil && bookCache.BookID != -1 {
			return bookCache, nil
		}
		book, err := mysql.GetBookByID(ID)
		if err != nil {
			return nil, err
		}
		cache := BookToCache(book)
		_ = redis.SetBookCache(cache, book.Score)
		return cache, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*models.BookCache), nil
}

func GetBooksByIDs(ids []int64) ([]*models.BookCache, error) {
	books, missIDs, err := redis.MGetBooks(ids)
	if err != nil {
		return nil, err
	}
	if len(missIDs) > 0 {
		dbBooks, err := mysql.GetBooksByIDs(missIDs)
		if err == nil {
			for _, b := range dbBooks {
				cacheBook := BookToCache(b)
				if cacheBook == nil {
					redis.SetEmpty("book:" + strconv.FormatInt(b.BookID, 10))
					continue
				}
				books = append(books, cacheBook)
				go func() {
					err := redis.SetBookCache(cacheBook, cacheBook.Score)
					if err != nil {
						return
					}
				}()
			}
		}
	}
	return books, nil
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
		return models.CodeRedisError
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
		return models.CodeMySQLError
	}
	err = mysql.DeleteBookToTag(B.BookID)
	if err != nil {
		return models.CodeMySQLError
	}
	for _, tag := range tags {
		err = mysql.AddBookTag(B.BookID, tag.ID)
		if err != nil {
			return models.CodeMySQLError
		}
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
		return models.CodeMySQLError
	}
	_ = redis.DeleteBookCache(book.BookID)
	return models.CodeSuccess
}
