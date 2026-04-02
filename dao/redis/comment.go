package redis

import (
	"Project1_Shop/models"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func GetCommentIDsByBookID(bookID int64) ([]int64, error) {
	key := "comment:book:" + strconv.FormatInt(bookID, 10)
	ids, err := RDB.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var res []int64
	for _, id := range ids {
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			continue
		}
		res = append(res, idInt)
	}
	return res, nil
}

func GetCommentIDsByUserID(UserID int64) ([]int64, error) {
	UserKey := "user:comments:" + strconv.FormatInt(UserID, 10)
	ids, err := RDB.SMembers(ctx, UserKey).Result()
	if err != nil {
		return nil, err
	}
	var res []int64
	for _, id := range ids {
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			continue
		}
		res = append(res, idInt)
	}
	return res, nil
}

func MGetComments(ids []int64) ([]*models.CommentBook, []int64, error) {
	pipe := RDB.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, 0, len(ids))
	for _, id := range ids {
		key := "comment:" + strconv.FormatInt(id, 10)
		cmds = append(cmds, pipe.HGetAll(ctx, key))
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, nil, err
	}
	var res []*models.CommentBook
	var miss []int64
	for i, cmd := range cmds {
		data := cmd.Val()
		if len(data) == 0 {
			miss = append(miss, ids[i])
			continue
		}
		commentID, _ := strconv.ParseInt(data["comment_id"], 10, 64)
		bookID, _ := strconv.ParseInt(data["book_id"], 10, 64)
		userID, _ := strconv.ParseInt(data["user_id"], 10, 64)
		parentID, _ := strconv.ParseInt(data["parent_id"], 10, 64)
		rootID, _ := strconv.ParseInt(data["root_id"], 10, 64)
		likeCount, _ := strconv.ParseInt(data["like_count"], 10, 64)
		replyCount, _ := strconv.ParseInt(data["reply_count"], 10, 64)
		status, _ := strconv.ParseInt(data["status"], 10, 8)
		res = append(res, &models.CommentBook{
			CommentID:   commentID,
			BookID:      bookID,
			UserID:      userID,
			ParentID:    parentID,
			RootID:      rootID,
			LikeCount:   likeCount,
			ReplyCount:  replyCount,
			Status:      int8(status),
			Comment:     data["comment"],
			CommentTime: data["comment_time"],
		})
	}
	return res, miss, nil
}

func SetCommentsToCache(c *models.CommentBook) error {
	pipe := RDB.Pipeline()
	key := "comment:" + strconv.FormatInt(c.CommentID, 10)
	UserKey := "user:comments:" + strconv.FormatInt(c.UserID, 10)
	BookKey := "comment:book:" + strconv.FormatInt(c.BookID, 10)
	pipe.HSet(ctx, key, map[string]interface{}{
		"comment_id":   c.CommentID,
		"book_id":      c.BookID,
		"user_id":      c.UserID,
		"parent_id":    c.ParentID,
		"root_id":      c.RootID,
		"like_count":   c.LikeCount,
		"reply_count":  c.ReplyCount,
		"status":       c.Status,
		"comment":      c.Comment,
		"comment_time": c.CommentTime,
	})
	pipe.Expire(ctx, key, RandTTL(CommentListTime))
	pipe.SAdd(ctx, UserKey, strconv.FormatInt(c.CommentID, 10))
	pipe.SAdd(ctx, BookKey, strconv.FormatInt(c.CommentID, 10))
	_, err := pipe.Exec(ctx)
	return err
}

func DelCommentsCache(bookID int64) error {
	key := "comment:book:" + strconv.FormatInt(bookID, 10)
	return RDB.Del(ctx, key).Err()
}
