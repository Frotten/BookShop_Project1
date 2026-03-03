package models

import "time"

type User struct {
	UserID   int64  `json:"user_id" db:"user_id" gorm:"primaryKey;autoIncrement"`
	Username string `json:"username" db:"username" gorm:"unique;not null;size:30"`
	Password string `json:"password" db:"password" gorm:"not null;size:255"`
	Email    string `json:"email" db:"email" gorm:"not null;size:50"`
	Gender   int8   `json:"gender" db:"gender" gorm:"not null;default:0"`
}

type RefreshToken struct { //刷新令牌模型,可以用Redis进行快速缓存
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at"`
}
