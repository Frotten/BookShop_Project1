package mysql

import (
	"Project1_Shop/models"

	"gorm.io/gorm"
)

func SaveComment(p *models.CommentBook) error {
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

func GetCommentsByBookID(bookID int64) ([]*models.CommentBook, error) {
	var list []*models.CommentBook
	err := DB.Where("book_id = ? AND status = ?", bookID, 1).Find(list).Error
	return list, err
}

func GetCommentsByUserID(UserID int64) ([]*models.CommentBook, error) {
	var list []*models.CommentBook
	err := DB.Where("user_id = ? AND status = ?", UserID, 1).Find(list).Error
	return list, err
}

func GetCommentsByIDs(ids []int64) ([]*models.CommentBook, error) {
	var list []*models.CommentBook
	err := DB.Where("comment_id IN ? AND status = ?", ids, 1).Find(list).Error
	return list, err
}

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
