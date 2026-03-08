package models

type Book struct {
	BookID     int64  `json:"book_id" db:"book_id" gorm:"primaryKey;autoIncrement"`
	Title      string `json:"title" db:"title" gorm:"not null;size:100"`
	Author     string `json:"author" db:"author" gorm:"not null;size:50"`
	Publisher  string `json:"publisher" db:"publisher" gorm:"not null;size:50"`
	Stock      int64  `json:"stock" db:"stock" gorm:"not null"`
	Sales      int64  `json:"sales" db:"sales" gorm:"not null"`
	Price      int64  `json:"price" db:"price" gorm:"not null"`
	Score      int64  `json:"score" db:"score"`
	CoverImage string `json:"cover_image" db:"cover_image" gorm:"size:255"`
}

type BookTag struct {
	BookID int64    `json:"book_id" db:"book_id" gorm:"primaryKey"`
	Tags   []string `json:"tag" db:"tag" gorm:"serializer:json"` //切片用法
}
