package mysql

import (
	"Project1_Shop/models"

	"gorm.io/gorm"
)

// SaveComment 保存评论，并在写入楼中楼时同步更新父评论的 reply_count
func SaveComment(p *models.CommentBook) error {
	// 以防调用方传了 0 覆盖默认值
	if p.Status == 0 {
		p.Status = 1
	}
	if err := DB.Create(p).Error; err != nil {
		return err
	}
	if p.ParentID != 0 {
		if err := DB.Model(&models.CommentBook{}).
			Where("comment_id = ?", p.ParentID).
			UpdateColumn("reply_count", gorm.Expr("reply_count + ?", 1)).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetCommentsByBookID 获取指定书籍下的一级/二级/更多级楼中楼评论（返回平铺列表，前端再组树）
// 说明：前端当前 buildCommentTree 会过滤 status===1，因此这里也只返回正常评论。
func GetCommentsByBookID(bookID int64) ([]models.CommentView, error) {
	var list []models.CommentView
	err := DB.Model(&models.CommentBook{}).
		Select("comment_books.*, users.username AS user_name, '' AS user_avatar").
		Joins("JOIN users ON users.user_id = comment_books.user_id").
		Where("comment_books.book_id = ? AND comment_books.status = ?", bookID, 1).
		Order("comment_books.comment_id DESC").
		Scan(&list).Error
	return list, err
}

// LikeComment 点赞：自增 like_count，并返回该评论所属的 book_id（用于失效 Redis 缓存）
func LikeComment(commentID int64) (int64, error) {
	var c models.CommentBook
	if err := DB.Where("comment_id = ? AND status = ?", commentID, 1).First(&c).Error; err != nil {
		return 0, err
	}
	if err := DB.Model(&models.CommentBook{}).
		Where("comment_id = ? AND status = ?", commentID, 1).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
		return 0, err
	}
	return c.BookID, nil
}
