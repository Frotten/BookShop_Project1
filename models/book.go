package models

import "gorm.io/datatypes"

type Book struct {
	BookID     int64          `json:"book_id" db:"book_id" gorm:"primaryKey;autoIncrement"`
	Title      string         `json:"title" db:"title" gorm:"not null;size:100"`
	Author     string         `json:"author" db:"author" gorm:"not null;size:50"`
	Publisher  string         `json:"publisher" db:"publisher" gorm:"not null;size:50"`
	Stock      int64          `json:"stock" db:"stock" gorm:"not null"`
	Sales      int64          `json:"sales" db:"sales" gorm:"not null"`
	Price      int64          `json:"price" db:"price" gorm:"not null"` //Price和Score都用int64，前端显示时除以100,避免精度误差
	Score      int64          `json:"score" db:"score"`
	CoverImage string         `json:"cover_image" db:"cover_image" gorm:"size:255"`
	Tags       datatypes.JSON `json:"tags" db:"tags" gorm:"type:json"` //切片用法
}

type BookCache struct {
	BookID     int64    `json:"book_id"`
	Title      string   `json:"title"`
	Author     string   `json:"author"`
	Publisher  string   `json:"publisher"`
	Stock      int64    `json:"stock"`
	Sales      int64    `json:"sales"`
	Price      int64    `json:"price"`
	Score      int64    `json:"score"`
	CoverImage string   `json:"cover_image"`
	Tags       []string `json:"tags"`
}

type Page struct {
	Page  int64       `json:"page" db:"page" gorm:"not null"`
	Total int64       `json:"total" db:"total" gorm:"not null"`
	Data  interface{} `json:"data" db:"data"`
}
