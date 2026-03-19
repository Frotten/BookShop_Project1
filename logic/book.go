package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
	"encoding/json"
	"sync"
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
		Price:      book.Price,
		Score:      book.Score,
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

func RateNewBook(p *models.UserRateBook) models.ResCode {
	errCh := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(1)
	go func(p *models.UserRateBook) {
		err := redis.CacheBookScoreCount(p.BookID, 1)
		if err != nil {
			errCh <- err
			wg.Done()
			return
		}
		err = redis.CacheBookScoreSum(p.BookID, p.Score)
		if err != nil {
			errCh <- err
			wg.Done()
			return
		}
		err = redis.UpdateUserRate(p)
		if err != nil {
			errCh <- err
			wg.Done()
			return
		}
		errCh <- nil
		wg.Done()
	}(p)
	wg.Add(1)
	go func(p *models.UserRateBook) {
		RB, err := mysql.GetRateBookByID(p.BookID) //Wrong
		if err != nil {
			errCh <- err
			wg.Done()
			return
		}
		RB.ScoreCount++
		RB.Score += p.Score
		//更新MySQL数据库中的评分信息
		err = mysql.UpdateRateBook(RB)
		if err != nil {
			errCh <- err
			wg.Done()
			return
		}
		err = mysql.UpdateUserRate(p)
		if err != nil {
			errCh <- err
			wg.Done()
			return
		}
		errCh <- nil
		wg.Done()
	}(p)
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return models.CodeServerBusy
		}
	}
	return models.CodeSuccess
}

func RateUpdateBook(p *models.UserRateBook) models.ResCode {
	errCh := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(1)
	go func(p *models.UserRateBook) {
		err := redis.CacheBookScoreCount(p.BookID, 0) //评分更新不改变评分数量
		if err != nil {
			errCh <- err
			wg.Done()
			return
		}
		BeforeScore, err := redis.GetBeforeBookScore(p.UserID, p.BookID)
		if err != nil {
			errCh <- err
			wg.Done()
			return
		}
		err = redis.CacheBookScoreSum(p.BookID, p.Score-BeforeScore) //评分更新需要先获取用户之前的评分，然后计算新的评分差值，再更新Redis中的评分总和
		if err != nil {
			errCh <- err
			wg.Done()
			return
		}
		err = redis.UpdateUserRate(p)
		if err != nil {
			errCh <- err
			wg.Done()
			return
		}
		errCh <- nil
		wg.Done()
	}(p)
	wg.Add(1)
	go func(p *models.UserRateBook) { //MySQL部分的更新
		RB, err := mysql.GetRateBookByID(p.BookID)
		if err != nil {
			errCh <- err
			wg.Done()
			return
		}
		BeforeScore, err := mysql.GetBeforeBookScore(p)
		if err != nil {
			errCh <- err
			wg.Done()
			return
		}
		RB.Score = RB.Score + p.Score - BeforeScore
		//更新MySQL数据库中的评分信息
		err = mysql.UpdateRateBook(RB)
		if err != nil {
			errCh <- err
			wg.Done()
			return
		}
		err = mysql.UpdateUserRate(p)
		if err != nil {
			errCh <- err
			wg.Done()
			return
		}
		errCh <- nil
		wg.Done()
	}(p)
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return models.CodeServerBusy
		}
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
