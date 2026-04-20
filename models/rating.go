package models

type RateBook struct {
	BookID     int64 `json:"book_id" db:"book_id" gorm:"primaryKey;autoIncrement:false"`
	ScoreCount int64 `json:"score_count" db:"score_count" gorm:"not null"`
	Score      int64 `json:"score" db:"score" gorm:"not null"`
}

type UserRateBook struct {
	UserID int64 `json:"user_id" db:"user_id" gorm:"primaryKey;autoIncrement:false;not null"`
	BookID int64 `json:"book_id" db:"book_id" gorm:"primaryKey;autoIncrement:false;not null"`
	Score  int64 `json:"score" db:"score" gorm:"not null"`
	Op     int   `json:"Op" gorm:"not null"`
}

type ListBook struct {
	BookID int64    `json:"book_id" redis:"book_id"`
	Title  string   `json:"title" redis:"title"`
	Score  int64    `json:"score" redis:"score"`
	Sales  int64    `json:"sales" redis:"sales"`
	Tags   []string `json:"tags" redis:"tags"`
}

type UserRating struct {
	UserID int64  `json:"user_id"`
	BookID int64  `json:"book_id"`
	Score  int64  `json:"score"`
	Title  string `json:"book_title"`
}
