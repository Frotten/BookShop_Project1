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
const BookTime = time.Hour * 24 * 7 * 2

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
	UserKey := "user:rating:" + strconv.FormatInt(UserID, 10)
	BookKey := "book:rating:" + strconv.FormatInt(BookID, 10)
	pipe := RDB.Pipeline()
	field := strconv.FormatInt(BookID, 10)
	pipe.HSet(ctx, UserKey, field, Score)
	pipe.SAdd(ctx, BookKey, UserID)
	_, err := pipe.Exec(ctx)
	return err
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

func AddScoreRank(BookID int64, Score float64) error {
	key := "book:rank:score"
	return RDB.ZAdd(ctx, key, redis.Z{
		Score:  Score,
		Member: BookID,
	}).Err()
}

func SetBookCache(Book *models.BookCache, Score int64) error {
	key := "book:" + strconv.FormatInt(Book.BookID, 10)
	tagsJSON, err := json.Marshal(Book.Tags)
	if err != nil {
		return err
	}
	_, err = RDB.HSet(ctx, key, map[string]interface{}{
		"book_id":      Book.BookID,
		"title":        Book.Title,
		"author":       Book.Author,
		"publisher":    Book.Publisher,
		"introduction": Book.Introduction,
		"stock":        Book.Stock,
		"price":        Book.Price,
		"score":        Score,
		"cover_image":  Book.CoverImage,
		"tags":         tagsJSON,
	}).Result()
	if err != nil {
		return err
	}
	RDB.Del(ctx, GetEmpty(key))
	RDB.Expire(ctx, key, RandTTL(BookTime))
	return nil
}

func DeleteBook(BookID int64) error {
	BookKey := "book:rating:" + strconv.FormatInt(BookID, 64)
	userIDs, err := RDB.SMembers(ctx, BookKey).Result()
	if err != nil {
		return err
	}
	pipe := RDB.Pipeline()
	pipe.Del(ctx, "book:"+strconv.FormatInt(BookID, 10))
	pipe.Del(ctx, "book:score:"+strconv.FormatInt(BookID, 10))
	pipe.Del(ctx, "book:score_count:"+strconv.FormatInt(BookID, 10))
	pipe.Del(ctx, "book:summary:"+strconv.FormatInt(BookID, 10))
	for _, id := range userIDs {

		pipe.HDel(ctx, "user:rating:"+id, strconv.FormatInt(BookID, 10))
	}
	pipe.Del(ctx, BookKey)
	_, err = pipe.Exec(ctx)
	return err
}

func DeleteBookCache(BookID int64) error {
	key := "book:" + strconv.FormatInt(BookID, 10)
	return RDB.Del(ctx, key).Err()
}

func GetScoreList() ([]redis.Z, error) {
	key := "book:rank:score"
	results, err := RDB.ZRevRangeWithScores(ctx, key, 0, 9).Result()
	if err != nil {
		return nil, err
	}
	return results, nil
}

func GetBookSummaryByBookID(BookID int64, Score int64) (*models.ListBook, error) {
	ID := strconv.FormatInt(BookID, 10)
	key := "book:summary:" + ID
	ok := CheckEmpty(key)
	if !ok {
		return nil, errors.New("already exist empty")
	}
	v, err, _ := G.Do(ID, func() (interface{}, error) {
		data, err := RDB.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		if len(data) == 0 {
			SetEmpty(key)
			return &models.ListBook{
				BookID: -1,
			}, nil
		}
		var tags []string
		err = json.Unmarshal([]byte(data["tags"]), &tags)
		return &models.ListBook{
			BookID: BookID,
			Title:  data["title"],
			Tags:   tags,
			Score:  Score,
		}, err
	})
	if err != nil {
		return nil, err
	}
	return v.(*models.ListBook), nil
}

func SetBookSummary(List *models.ListBook) error {
	key := "book:summary:" + strconv.FormatInt(List.BookID, 10)
	defer RDB.Del(ctx, GetEmpty(key))
	tagsJson, err := json.Marshal(List.Tags)
	if err != nil {
		return err
	}
	_, err = RDB.HSet(ctx, key, map[string]interface{}{
		"book_id": List.BookID,
		"title":   List.Title,
		"sales":   List.Sales,
		"tags":    tagsJson,
	}).Result()
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
		var numErr *strconv.NumError
		if errors.As(err, &numErr) && errors.Is(numErr.Err, strconv.ErrSyntax) {
			Sales = 0
		} else {
			return nil, err
		}
	}
	Stock, err := strconv.ParseInt(data["stock"], 10, 64)
	if err != nil {
		var numErr *strconv.NumError
		if errors.As(err, &numErr) && errors.Is(numErr.Err, strconv.ErrSyntax) {
			Stock = 0
		} else {
			return nil, err
		}
	}
	Price, err := strconv.ParseInt(data["price"], 10, 64)
	if err != nil {
		return nil, err
	}
	Score, err := strconv.ParseInt(data["score"], 10, 64)
	if err != nil {
		return nil, err
	}
	var tags []string
	err = json.Unmarshal([]byte(data["tags"]), &tags)
	return &models.BookCache{
		BookID:       BookID,
		Title:        data["title"],
		Sales:        Sales,
		Score:        Score,
		Author:       data["author"],
		Publisher:    data["publisher"],
		Stock:        Stock,
		Price:        Price,
		CoverImage:   data["cover_image"],
		Tags:         tags,
		Introduction: data["introduction"],
	}, nil
}

func parseBook(data map[string]string) *models.BookCache {
	if len(data) == 0 {
		return nil
	}
	Sales, _ := strconv.ParseInt(data["sales"], 10, 64)
	Stock, _ := strconv.ParseInt(data["stock"], 10, 64)
	Price, _ := strconv.ParseInt(data["price"], 10, 64)
	Score, _ := strconv.ParseInt(data["score"], 10, 64)
	var tags []string
	_ = json.Unmarshal([]byte(data["tags"]), &tags)
	bookID, _ := strconv.ParseInt(data["book_id"], 10, 64)
	return &models.BookCache{
		BookID:       bookID,
		Title:        data["title"],
		Author:       data["author"],
		Publisher:    data["publisher"],
		Introduction: data["introduction"],
		Stock:        Stock,
		Sales:        Sales,
		Price:        Price,
		Score:        Score,
		CoverImage:   data["cover_image"],
		Tags:         tags,
	}
}

func MGetBooks(ids []int64) ([]*models.BookCache, []int64, error) {
	pipe := RDB.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, 0, len(ids))
	for _, id := range ids {
		key := "book:" + strconv.FormatInt(id, 10)
		cmds = append(cmds, pipe.HGetAll(ctx, key))
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, nil, err
	}
	var books []*models.BookCache
	var missIDs []int64
	for i, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil || len(data) == 0 {
			missIDs = append(missIDs, ids[i])
			continue
		}
		book := parseBook(data)
		if book != nil {
			books = append(books, book)
		}
	}
	return books, missIDs, nil
}

func GetRankIDsByScore(start, end int64) ([]int64, int64, error) {
	key := "book:rank:score"
	idsZ, err := RDB.ZRevRangeWithScores(ctx, key, start, end).Result()
	if err != nil {
		return nil, 0, err
	}
	var ids []int64
	for _, z := range idsZ {
		id, _ := strconv.ParseInt(z.Member.(string), 10, 64)
		ids = append(ids, id)
	}
	total, err := RDB.ZCard(ctx, key).Result()
	if err != nil {
		return nil, 0, err
	}
	return ids, total, nil
}
