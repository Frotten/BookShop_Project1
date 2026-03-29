package redis

import (
	"Project1_Shop/models"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const CommentListTime = time.Minute * 15

func CommentListKey(bookID int64) string {
	return "comment:book:" + strconv.FormatInt(bookID, 10) + ":list"
}

// GetCommentsFromCache 从 Redis 获取评论列表（优先）
func GetCommentsFromCache(bookID int64) ([]models.CommentView, bool, error) {
	key := CommentListKey(bookID)
	raw, err := RDB.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}
	var list []models.CommentView
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		// 反序列化失败直接视为未命中，避免页面一直拿到脏数据
		return nil, false, nil
	}
	return list, true, nil
}

func SetCommentsToCache(bookID int64, list []models.CommentView) error {
	key := CommentListKey(bookID)
	bs, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return RDB.Set(ctx, key, string(bs), RandTTL(CommentListTime)).Err()
}

func DelCommentsCache(bookID int64) error {
	key := CommentListKey(bookID)
	return RDB.Del(ctx, key).Err()
}
