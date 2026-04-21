package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
	"strconv"

	"go.uber.org/zap"
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
	err = redis.AddSaleRank(book.BookID, 0)
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

func GetBooksByTitle(Title string) ([]*models.BookCache, error) {
	z, err, _ := redis.G.Do(Title, func() (interface{}, error) {
		ids, err := redis.GetBookIDsByTitle(Title)
		if err != nil || len(ids) <= 0 {
			Books, err := mysql.GetBooksByTitle(Title)
			var BookCache []*models.BookCache
			for _, Book := range Books {
				BookCache = append(BookCache, BookToCache(Book))
			}
			if err != nil {
				return nil, err
			}
			go func() {
				for _, Book := range BookCache {
					_ = redis.SetBookCache(Book, Book.Score)
				}
			}()
			return BookCache, nil
		}
		Books, err := GetBooksByIDs(ids)
		if err != nil {
			return nil, err
		}
		return Books, nil
	})
	if err != nil {
		return nil, err
	}
	if z != nil {
		return z.([]*models.BookCache), nil
	}
	return nil, nil
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

func GetBookPriceByIDs(BookIDs []int64) (map[int64]int64, error) {
	bookPriceMap := make(map[int64]int64)
	Books, err := GetBooksByIDs(BookIDs)
	if err != nil {
		return nil, err
	}
	for _, book := range Books {
		bookPriceMap[book.BookID] = book.Price
	}
	return bookPriceMap, nil
}

func GetTopSaleList() ([]*models.ListBook, models.ResCode) {
	results, err := redis.GetSaleList()
	if err != nil || len(results) <= 0 {
		Books, _, err := mysql.GetBooksPageBySale(1)
		if err != nil {
			return nil, models.CodeMySQLError
		}
		var AnsList []*models.ListBook
		for _, Book := range Books {
			T := BookListToCache(Book)
			_ = redis.SetBookSummary(T)
			err = redis.AddSaleRank(Book.BookID, Book.Sales)
			if err != nil {
				continue
			}
			AnsList = append(AnsList, T)
		}
		return AnsList, models.CodeSuccess
	}
	var Ans []*models.ListBook
	for _, z := range results {
		var BookID int64
		switch v := z.Member.(type) {
		case int64:
			BookID = v
		case string:
			var err error
			BookID, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				zap.L().Error("ParseInt failed", zap.String("member", v), zap.Error(err))
				continue
			}
		default:
			zap.L().Error("unexpected member type", zap.Any("value", v))
			continue
		}
		bookCache, err := redis.GetBookByBookID(BookID)
		if err != nil {
			book, err := mysql.GetBookByID(BookID)
			if err != nil {
				return nil, models.CodeMySQLError
			}
			T := BookListToCache(book)
			_ = redis.SetBookCache(BookToCache(book), book.Score)
			_ = redis.SetBookSummary(T)
			Ans = append(Ans, T)
			continue
		}
		T := redis.BookCacheToListBook(bookCache)
		_ = redis.SetBookSummary(T)
		Ans = append(Ans, T)
	}
	return Ans, models.CodeSuccess
}
