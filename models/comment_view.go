package models

type CommentBook struct {
	CommentID   int64  `json:"comment_id" gorm:"primaryKey;autoIncrement"`
	BookID      int64  `json:"book_id" gorm:"index;not null"`
	UserID      int64  `json:"user_id" gorm:"index;not null"`
	BookTitle   string `json:"book_title" gorm:"-"`
	Username    string `json:"username" gorm:"-"`          // 仅用于返回给前端显示，不存储在数据库中
	ParentID    int64  `json:"parent_id" gorm:"default:0"` // 0表示一级评论
	RootID      int64  `json:"root_id" gorm:"index"`       // 根评论ID（用于楼中楼）,一级评论的RootID为0，二级评论的RootID为一级评论的CommentID
	LikeCount   int64  `json:"like_count" gorm:"default:0"`
	ReplyCount  int64  `json:"reply_count" gorm:"default:0"`
	Status      int8   `json:"status" gorm:"default:1"` // 1=正常 2=删除 3=审核中
	Comment     string `json:"comment" gorm:"type:text;not null"`
	CommentTime string `json:"comment_time" gorm:"index"`
}
