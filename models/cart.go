package models

type Cart struct {
	CartID   int64 `json:"cart_id" gorm:"primaryKey;autoIncrement"`
	UserID   int64 `json:"user_id" gorm:"not null;index:idx_user_book,unique"` //创建用户ID和书籍ID的唯一索引
	BookID   int64 `json:"book_id" gorm:"not null;index:idx_user_book,unique"`
	Quantity int64 `json:"quantity" gorm:"not null"`
}

type CartView struct {
	CartID     int64  `json:"cart_id"`
	UserID     int64  `json:"user_id"`
	BookID     int64  `json:"book_id"`
	Quantity   int64  `json:"quantity"`
	Stock      int64  `json:"stock"`
	Price      int64  `json:"price"`
	Title      string `json:"title"`
	Author     string `json:"author"`
	CoverImage string `json:"cover_image"`
}
