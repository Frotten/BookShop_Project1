package models

// CommentView 用于前端展示评论：除了 CommentBook 字段外，还需要用户信息与统计字段。
// 前端在构建楼中楼树时使用 parent_id/root_id，并依赖 user_name/user_avatar、like_count、reply_count 等字段。
type CommentView struct {
	CommentBook
	UserName   string `json:"user_name"`
	UserAvatar string `json:"user_avatar"`
}

