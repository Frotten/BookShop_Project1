package models

type Order struct {
	OrderID    int64  `gorm:"primaryKey;autoIncrement"`
	UserID     int64  `gorm:"index;not null"`
	TotalPrice int64  `gorm:"not null"`
	Status     int8   `gorm:"not null;default:0"`
	CreatedAt  string `gorm:"index;not null"`
}

type OrderItem struct {
	ID       int64 `gorm:"primaryKey;autoIncrement"`
	OrderID  int64 `gorm:"index;not null"`
	BookID   int64 `gorm:"index;not null"`
	Price    int64 `gorm:"not null"` // 下单时价格
	Quantity int64 `gorm:"not null"`
}

type OrderRequest struct {
	Items []OrderParam `json:"items" binding:"required,min=1,dive"`
}
