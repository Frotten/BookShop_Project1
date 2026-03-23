package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
	"encoding/json"
	"strconv"

	"go.uber.org/zap"
)

func AddBook(p *models.AddBookParam) models.ResCode {
	tagsJSON, _ := json.Marshal(p.Tags)
	exist := mysql.ExistBookByInfo(p.Title, p.Author, p.Publisher)
	if exist {
		return models.CodeBookExist
	}
	book := &models.Book{
		Title:      p.Title,
		Author:     p.Author,
		Publisher:  p.Publisher,
		Stock:      p.Stock,
		Sales:      0,
		Price:      p.Price,
		Score:      0,
		CoverImage: p.CoverImage,
		Tags:       tagsJSON,
	}
	err := mysql.AddBook(book)
	if err != nil {
		return models.CodeServerBusy
	}
	return models.CodeSuccess
}

func BookListToCache(book *models.Book) (*models.ListBook, error) {
	var tags []string
	if len(book.Tags) > 0 {
		if err := json.Unmarshal(book.Tags, &tags); err != nil {
			return nil, err
		}
	}
	return &models.ListBook{
		BookID: book.BookID,
		Title:  book.Title,
		Sales:  book.Sales,
		Score:  float64(book.Score) / 100,
		Tags:   tags,
	}, nil
}

func BookToCache(book *models.Book) (*models.BookCache, error) {
	var tags []string
	if len(book.Tags) > 0 {
		if err := json.Unmarshal(book.Tags, &tags); err != nil {
			return nil, err
		}
	}
	return &models.BookCache{
		BookID:     book.BookID,
		Title:      book.Title,
		Author:     book.Author,
		Publisher:  book.Publisher,
		Stock:      book.Stock,
		Sales:      book.Sales,
		Price:      float64(book.Price) / 100,
		Score:      float64(book.Score) / 100,
		CoverImage: book.CoverImage,
		Tags:       tags,
	}, nil
}

func GetBookByID(ID int64) (*models.Book, error) {
	book, err := mysql.GetBookByID(ID)
	if err != nil {
		return nil, err
	}
	return book, nil
}

func DeleteBook(ID int64) models.ResCode {
	exist := mysql.ExistBook(ID)
	if !exist {
		return models.CodeBookNotExist
	}
	_ = mysql.DeleteBook(ID)
	exist = mysql.ExistBook(ID)
	if !exist {
		return models.CodeSuccess
	}
	return models.CodeServerBusy
}

func RateNewBook(p *models.UserRateBook) models.ResCode { //这里必须保证MySQL先成功才能去缓存Redis，不能并发
	RB, err := mysql.GetRateBookByID(p.BookID)
	if err != nil {
		return models.CodeInvalidParam
	}
	RB.ScoreCount++
	RB.Score += p.Score
	//更新MySQL数据库中的评分信息
	err = mysql.UpdateRateBook(RB)
	if err != nil {
		return models.CodeServerBusy
	}
	err = mysql.UpdateUserRate(p)
	if err != nil {
		return models.CodeServerBusy
	}
	err = mysql.UpdateBookScore(RB)
	if err != nil {
		return models.CodeServerBusy
	}
	//MySQL和Redis的分界线
	p.Op = models.RateOpNew
	select {
	case models.RateChan <- p:
	default:
		go func() {
			err := redis.NewScoreAndRank(p.UserID, p.BookID, p.Score)
			if err != nil {
				zap.L().Error("NewScoreAndUpdateRedis Failed", zap.Error(err))
			}
		}()
	}
	return models.CodeSuccess
}

func RateUpdateBook(p *models.UserRateBook) models.ResCode {
	RB, err := mysql.GetRateBookByID(p.BookID)
	BeforeScore, err := mysql.GetBeforeBookScore(p)
	if err != nil {
		return models.CodeServerBusy
	}
	RB.Score = RB.Score + p.Score - BeforeScore
	//更新MySQL数据库中的评分信息
	err = mysql.UpdateRateBook(RB)
	if err != nil {
		return models.CodeServerBusy
	}
	err = mysql.UpdateUserRate(p)
	if err != nil {
		return models.CodeServerBusy
	}
	err = mysql.UpdateBookScore(RB)
	if err != nil {
		return models.CodeServerBusy
	}
	p.Op = models.RateOpUpdate
	select {
	case models.RateChan <- p:
	default:
		go func() {
			err := redis.UpdateScoreAndRank(p.UserID, p.BookID, p.Score)
			if err != nil {
				zap.L().Error("UpdateScoreAndUpdateRedis Failed", zap.Error(err))
			}
		}()
	}
	return models.CodeSuccess
}

func RateBook(p *models.UserRateBook) models.ResCode {
	ok, err := redis.CheckRate(p)
	if err != nil {
		return models.CodeServerBusy
	}
	if !ok {
		return RateNewBook(p)
	}
	return RateUpdateBook(p)
}

func GetTopScoreList() ([]*models.ListBook, models.ResCode) {
	results, err := redis.GetScoreList()
	if err != nil {
		return nil, models.CodeServerBusy
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
		T, err := redis.GetBookSummaryByBookID(BookID)
		if err != nil {
			continue
		}
		if T.BookID == -1 {
			// 合并并发请求，只有一个会执行函数体
			v, err, _ := redis.G.Do(strconv.FormatInt(BookID, 10), func() (interface{}, error) {
				Book, err := mysql.GetBookByID(BookID)
				if err != nil {
					return nil, err
				}
				T, err := BookListToCache(Book)
				if err != nil {
					return nil, err
				}
				_ = redis.SetBookSummary(T)
				return T, nil
			})
			if err != nil {
				continue
			}
			T = v.(*models.ListBook)
		}
		if T.Score < 0 {
			continue
		}
		Ans = append(Ans, T)
	}
	return Ans, models.CodeSuccess
}
