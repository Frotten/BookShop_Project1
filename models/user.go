package models

type User struct {
	UserID   int64  `json:"user_id" db:"user_id" gorm:"primaryKey;autoIncrement"`
	Username string `json:"username" db:"username" gorm:"unique;not null;size:30"`
	Password string `json:"password" db:"password" gorm:"not null;size:255"`
	Email    string `json:"email" db:"email" gorm:"not null;size:50"`
	Gender   int8   `json:"gender" db:"gender" gorm:"not null;default:0"`
	Token    string `json:"token"`
}
