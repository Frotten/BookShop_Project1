package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/dao/redis"
	"Project1_Shop/models"
)

// GetCommentsByBookID 获取评论列表：优先 Redis，未命中回落 MySQL，并写入 Redis
func GetCommentsByBookID(bookID int64) ([]models.CommentView, models.ResCode) {
	if bookID <= 0 {
		return nil, models.CodeInvalidParam
	}
	// 1) Redis 优先
	if cached, ok, err := redis.GetCommentsFromCache(bookID); err == nil && ok {
		return cached, models.CodeSuccess
	}
	// 2) MySQL 回落
	list, err := mysql.GetCommentsByBookID(bookID)
	if err != nil {
		return nil, models.CodeServerBusy
	}
	// 3) 写入 Redis（忽略写入失败，避免影响主链路）
	_ = redis.SetCommentsToCache(bookID, list)
	return list, models.CodeSuccess
}

// LikeComment 点赞：更新 MySQL，然后失效 Redis
func LikeComment(commentID int64) models.ResCode {
	bookID, err := mysql.LikeComment(commentID)
	if err != nil {
		return models.CodeServerBusy
	}
	_ = redis.DelCommentsCache(bookID)
	return models.CodeSuccess
}

func CommentBook(p *models.CommentBook) models.ResCode {
	if err := mysql.SaveComment(p); err != nil {
		return models.CodeServerBusy
	}
	_ = redis.DelCommentsCache(p.BookID)
	return models.CodeSuccess
}
