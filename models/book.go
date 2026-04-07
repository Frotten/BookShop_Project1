package models

const AvgAllBooks = 5.0
const AvgAllRateUser = 20
const PageSize = 8

type Book struct {
	BookID       int64  `json:"book_id" db:"book_id" gorm:"primaryKey;autoIncrement"`
	Title        string `json:"title" db:"title" gorm:"not null;size:100"`
	Author       string `json:"author" db:"author" gorm:"not null;size:50"`
	Publisher    string `json:"publisher" db:"publisher" gorm:"not null;size:50"`
	Introduction string `json:"introduction" db:"introduction" gorm:"not null;size:500"`
	Stock        int64  `json:"stock" db:"stock" gorm:"not null"`
	Sales        int64  `json:"sales" db:"sales" gorm:"not null"`
	Price        int64  `json:"price" db:"price" gorm:"not null"` //Price和Score都用int64，前端显示时除以100,避免精度误差
	Score        int64  `json:"score" db:"score" gorm:"default:0"`
	CoverImage   string `json:"cover_image" db:"cover_image" gorm:"size:255"`
	Tags         []Tag  `json:"tags" db:"tags" gorm:"-"` //切片用法
}

type BookCache struct {
	BookID       int64    `json:"book_id" redis:"book_id"`
	Title        string   `json:"title" redis:"title"`
	Author       string   `json:"author" redis:"author"`
	Publisher    string   `json:"publisher" redis:"publisher"`
	Introduction string   `json:"introduction" redis:"introduction"`
	Stock        int64    `json:"stock" redis:"stock"`
	Sales        int64    `json:"sales" redis:"sales"`
	Price        int64    `json:"price" redis:"price"`
	Score        int64    `json:"score" redis:"score"`
	CoverImage   string   `json:"cover_image" redis:"cover_image"`
	Tags         []string `json:"tags" redis:"tags"`
}

type Page struct {
	Page  int64       `json:"page" db:"page" gorm:"not null"`
	Total int64       `json:"total" db:"total" gorm:"not null"`
	Data  interface{} `json:"data" db:"data"`
}

type Tag struct {
	ID   int64  `json:"id" gorm:"not null;primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"not null;unique;size:255"`
}

type BookTag struct {
	BookID int64 `gorm:"column:book_id"`
	TagID  int64 `gorm:"column:tag_id"`
}

func WeightedCalculation(AllScore, Count int64) float64 {
	if Count == 0 {
		return 0
	}
	return (float64(AllScore)/float64(Count)*float64(Count) + AvgAllBooks*AvgAllRateUser) / (float64(Count) + AvgAllRateUser)
}
