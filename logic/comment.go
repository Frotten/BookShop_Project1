package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
	"strconv"
)

func GetCommentsByBookID(bookID int64) ([]*models.CommentBook, models.ResCode) {
	if bookID <= 0 {
		return nil, models.CodeInvalidParam
	}
	z, err, _ := redis.G.Do(strconv.FormatInt(bookID, 10), func() (interface{}, error) {
		ids, err := redis.GetCommentIDsByBookID(bookID)
		if err != nil || len(ids) == 0 {
			dbList, err := mysql.GetCommentsByBookID(bookID)
			if err != nil {
				return nil, err
			}
			for _, c := range dbList {
				_ = redis.SetCommentsToCache(c)
				UserView, _ := GetUserInfo(c.UserID)
				c.Username = UserView.Username
			}
			return dbList, err
		}
		list, missIDs, err := redis.MGetComments(ids)
		if err != nil {
			return nil, err
		}
		if len(missIDs) > 0 {
			dbList, err := mysql.GetCommentsByIDs(missIDs)
			if err != nil {
				return nil, err
			}
			list = append(list, dbList...)
			for _, c := range dbList {
				_ = redis.SetCommentsToCache(c)
			}
		}
		for _, c := range list {
			UserView, _ := GetUserInfo(c.UserID)
			c.Username = UserView.Username
		}
		return list, err
	})
	if err != nil {
		return nil, models.CodeServerBusy
	}
	return z.([]*models.CommentBook), models.CodeSuccess
}

func LikeComment(commentID int64) models.ResCode {
	bookID, err := mysql.LikeComment(commentID)
	if err != nil {
		return models.CodeMySQLError
	}
	_ = redis.DelCommentsCache(bookID)
	return models.CodeSuccess
}

func CommentBook(p *models.CommentBook) models.ResCode {
	if err := mysql.SaveComment(p); err != nil {
		return models.CodeMySQLError
	}
	_ = redis.DelCommentsCache(p.BookID)
	return models.CodeSuccess
}

func GetCommentsByUser(UserID int64, UserName string) ([]*models.CommentBook, models.ResCode) {
	z, err, _ := redis.G.Do(strconv.FormatInt(UserID, 10), func() (interface{}, error) {
		ids, err := redis.GetCommentIDsByUserID(UserID)
		if err != nil {
			res, err := mysql.GetCommentsByUserID(UserID)
			if err != nil {
				return nil, err
			}
			for _, c := range res {
				_ = redis.SetCommentsToCache(c)
				c.Username = UserName
			}
			return res, err
		}
		if len(ids) == 0 {
			return nil, err
		}
		list, missIDs, err := redis.MGetComments(ids)
		if err != nil {
			return nil, err
		}
		if len(missIDs) > 0 {
			dbList, err := mysql.GetCommentsByIDs(missIDs)
			if err != nil {
				return nil, err
			}
			for _, c := range dbList {
				_ = redis.SetCommentsToCache(c)
				list = append(list, c)
			}
		}
		var res []*models.CommentBook
		for _, c := range list {
			res = append(res, c)
		}
		for _, c := range res {
			c.Username = UserName
		}
		return res, err
	})
	if err != nil {
		return nil, models.CodeServerBusy
	}
	return z.([]*models.CommentBook), models.CodeSuccess
}
