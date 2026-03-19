package redis

import (
	"Project1_Shop/models"
	"context"
	"errors"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func CacheBookScoreCount(bookID int64, count int64) error {
	key := "book:score_count:" + strconv.FormatInt(bookID, 10)
	return RDB.IncrBy(ctx, key, count).Err()
}

func CacheBookScoreSum(bookID int64, score int64) error {
	key := "book:score:" + strconv.FormatInt(bookID, 10)
	return RDB.IncrBy(ctx, key, score).Err()
}

func CacheBookSale(bookID int64, sales int64) error {
	key := "book:sale:" + strconv.FormatInt(bookID, 10)
	return RDB.IncrBy(ctx, key, sales).Err()
}

func CheckRate(p *models.UserRateBook) (bool, error) {
	key := "user:rating:" + strconv.FormatInt(p.UserID, 10)
	field := strconv.FormatInt(p.BookID, 10)
	_, err := RDB.HGet(ctx, key, field).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func GetBeforeBookScore(userID, bookID int64) (int64, error) {
	key := "user:rating:" + strconv.FormatInt(userID, 10)
	field := strconv.FormatInt(bookID, 10)
	Res, err := RDB.HGet(ctx, key, field).Result()
	if err != nil {
		return 0, err
	}
	Score, err := strconv.ParseInt(Res, 10, 64)
	if err != nil {
		return 0, err
	}
	return Score, nil
}

func UpdateUserRate(p *models.UserRateBook) error {
	key := "user:rating:" + strconv.FormatInt(p.UserID, 10)
	field := strconv.FormatInt(p.BookID, 10)
	return RDB.HSet(ctx, key, field, p.Score).Err()
}

func GetAllScoreAndCount(BookID int64) (int64, int64, error) {
	key1 := "book:score:" + strconv.FormatInt(BookID, 10)
	key2 := "book:score_count:" + strconv.FormatInt(BookID, 10)
	ScoreString, err := RDB.Get(ctx, key1).Result()
	if err != nil {
		return 0, 0, err
	}
	CountString, err := RDB.Get(ctx, key2).Result()
	if err != nil {
		return 0, 0, err
	}
	Score, err := strconv.ParseInt(ScoreString, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	Count, err := strconv.ParseInt(CountString, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return Score, Count, err
}

func AddScoreRank(BookID, AllScore, Count int64) error {
	key := "book:rank:score"
	Score := models.WeightedCalculation(AllScore, Count)
	return RDB.ZAdd(ctx, key, redis.Z{
		Score:  Score,
		Member: BookID,
	}).Err()
}
