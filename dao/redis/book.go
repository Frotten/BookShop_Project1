package redis

import (
	"Project1_Shop/models"
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

const BufferTime = time.Minute * 3
const ListTime = time.Hour * 24 * 7

func CheckEmpty(key string) bool {
	EmptyKey := GetEmpty(key)
	Emp, err := RDB.Get(ctx, EmptyKey).Result()
	if Emp == "__NULL__" || err == nil {
		return false
	}
	return true
}

func SetEmpty(key string) {
	EmptyKey := GetEmpty(key)
	RDB.Set(ctx, EmptyKey, "__NULL__", BufferTime)
}

func GetEmpty(key string) string {
	return "empty:" + key
}

func RandTTL(T time.Duration) time.Duration {
	return T + time.Duration(rand.Intn(600))*time.Second
}

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

func UpdateUserRate(UserID, BookID, Score int64) error {
	key := "user:rating:" + strconv.FormatInt(UserID, 10)
	field := strconv.FormatInt(BookID, 10)
	return RDB.HSet(ctx, key, field, Score).Err()
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

func UpdateBook(BookID int64, AllScore, Count int64) error {
	key := "book:" + strconv.FormatInt(BookID, 10)
	ok := CheckEmpty(key)
	if !ok {
		return errors.New("already exist empty key")
	}
	Score := models.WeightedCalculation(AllScore, Count)
	_, err := RDB.HGet(ctx, key, "score").Result()
	if err != nil {
		SetEmpty(key)
		return err
	}
	_, err = RDB.HSet(ctx, key, "score", Score).Result()
	if err != nil {
		return err
	}
	return nil
}

func GetScoreList() ([]redis.Z, error) {
	key := "book:rank:score"
	results, err := RDB.ZRevRangeWithScores(ctx, key, 0, 9).Result()
	if err != nil {
		return nil, err
	}
	return results, nil
}

func GetBookSummaryByBookID(BookID int64) (*models.ListBook, error) {
	key := "book:summary:" + strconv.FormatInt(BookID, 10)
	ok := CheckEmpty(key)
	if !ok {
		return nil, errors.New("already exist empty")
	}
	data, err := RDB.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		SetEmpty(key)
		return &models.ListBook{
			BookID: -1,
		}, err
	}
	Sales, err := strconv.ParseInt(data["sales"], 10, 64)
	if err != nil {
		return nil, err
	}
	Score, err := strconv.ParseFloat(data["score"], 64)
	if err != nil {
		return nil, err
	}
	var tags []string
	err = json.Unmarshal([]byte(data["tags"]), &tags)
	return &models.ListBook{
		BookID: BookID,
		Title:  data["title"],
		Sales:  Sales,
		Score:  Score,
		Tags:   tags,
	}, nil
}

func SetBookSummary(List *models.ListBook) error {
	key := "book:summary:" + strconv.FormatInt(List.BookID, 10)
	defer RDB.Del(ctx, GetEmpty(key))
	_, err := RDB.HSet(ctx, key, List).Result()
	RDB.Expire(ctx, key, RandTTL(ListTime))
	return err
}

func GetBookByBookID(BookID int64) (*models.BookCache, error) {
	key := "book:" + strconv.FormatInt(BookID, 10)
	ok := CheckEmpty(key)
	if !ok {
		return nil, errors.New("already exist empty")
	}
	data, err := RDB.HGetAll(ctx, key).Result()
	if err != nil {
		return &models.BookCache{
			BookID: -1,
		}, err
	}
	Sales, err := strconv.ParseInt(data["sales"], 10, 64)
	if err != nil {
		return nil, err
	}
	Stock, err := strconv.ParseInt(data["stock"], 10, 64)
	if err != nil {
		return nil, err
	}
	Price, err := strconv.ParseFloat(data["price"], 64)
	if err != nil {
		return nil, err
	}
	Score, err := strconv.ParseFloat(data["score"], 64)
	if err != nil {
		return nil, err
	}
	var tags []string
	err = json.Unmarshal([]byte(data["tags"]), &tags)
	return &models.BookCache{
		BookID:     BookID,
		Title:      data["title"],
		Sales:      Sales,
		Score:      Score,
		Author:     data["author"],
		Publisher:  data["publisher"],
		Stock:      Stock,
		Price:      Price,
		CoverImage: data["cover_image"],
		Tags:       tags,
	}, nil
}

func NewScoreAndRank(UserID, BookID, Score int64) error {
	err := CacheBookScoreCount(BookID, 1)
	if err != nil {
		return err
	}
	err = CacheBookScoreSum(BookID, Score)
	if err != nil {
		return err
	}
	err = UpdateUserRate(UserID, BookID, Score)
	if err != nil {
		return err
	}
	AllScore, Count, err := GetAllScoreAndCount(BookID)
	if err != nil {
		return err
	}
	err = AddScoreRank(BookID, AllScore, Count)
	if err != nil {
		return err
	}
	err = UpdateBook(BookID, AllScore, Count)
	if err != nil {
		return err
	}
	return nil
}

func UpdateScoreAndRank(UserID, BookID, Score int64) error {
	BeforeScore, err := GetBeforeBookScore(UserID, BookID)
	if err != nil {
		return err
	}
	err = CacheBookScoreSum(BookID, Score-BeforeScore) //评分更新需要先获取用户之前的评分，然后计算新的评分差值，再更新Redis中的评分总和
	if err != nil {
		return err
	}
	err = UpdateUserRate(UserID, BookID, Score)
	if err != nil {
		return err
	}
	AllScore, Count, err := GetAllScoreAndCount(BookID)
	if err != nil {
		return err
	}
	err = AddScoreRank(BookID, AllScore, Count)
	if err != nil {
		return err
	}
	err = UpdateBook(BookID, AllScore, Count)
	if err != nil {
		return err
	}
	return nil
}
