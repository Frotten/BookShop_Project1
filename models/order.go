package models

type Order struct {
	OrderID    int64  `gorm:"primaryKey;autoIncrement"`
	UserID     int64  `gorm:"index;not null"`
	TotalPrice int64  `gorm:"not null"`
	Status     int8   `gorm:"not null;default:0"`
	CreatedAt  string `gorm:"index;not null"`
}

type OrderItem struct {
	ID       int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID  int64  `gorm:"index;not null" json:"order_id"`
	BookID   int64  `gorm:"index;not null" json:"book_id"`
	Price    int64  `gorm:"not null" json:"price"` // 下单时价格
	Quantity int64  `gorm:"not null" json:"quantity"`
	Title    string `gorm:"-" json:"title"`
}

type OrderRequest struct {
	Items []OrderParam `json:"items" binding:"required,min=1,dive"`
}

type OrderView struct {
	OrderID    int64  `json:"order_id"`
	UserID     int64  `json:"user_id"`
	TotalPrice int64  `json:"total_price"`
	Status     int8   `json:"status"`
	CreatedAt  string `json:"created_at"`
	Items      []*OrderItem
}

type OrderConfirmParam struct {
	OrderID int64 `json:"order_id" binding:"required"`
}
