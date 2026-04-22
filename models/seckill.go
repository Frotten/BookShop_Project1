package models

import "time"

type SeckillProduct struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	BookID       int64     `json:"book_id" gorm:"index;not null"`
	Title        string    `json:"title" gorm:"size:100;not null"`
	SeckillPrice int64     `json:"seckill_price" gorm:"not null"`
	OrigPrice    int64     `json:"orig_price" gorm:"not null"`
	Stock        int64     `json:"stock" gorm:"not null"`
	StartTime    time.Time `json:"start_time" gorm:"not null"`
	EndTime      time.Time `json:"end_time" gorm:"not null"`
	Status       int8      `json:"status" gorm:"not null;default:1"` // 1=进行中 0=已下架
}

type SeckillOrder struct {
	ID        int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int64  `json:"user_id" gorm:"index;not null"`
	ProductID int64  `json:"product_id" gorm:"index;not null"` // SeckillProduct.ID
	BookID    int64  `json:"book_id" gorm:"not null"`
	Price     int64  `json:"price" gorm:"not null"`            // 实际支付价
	OrderID   int64  `json:"order_id" gorm:"index;default:0"`  // 关联的 Order.OrderID（异步生成后回填）
	Status    int8   `json:"status" gorm:"not null;default:0"` // 0=待下单 1=下单成功 2=已失效
	CreatedAt string `json:"created_at" gorm:"not null"`
}

type SeckillProductView struct {
	ID           int64  `json:"id"`
	BookID       int64  `json:"book_id"`
	Title        string `json:"title"`
	SeckillPrice int64  `json:"seckill_price"`
	OrigPrice    int64  `json:"orig_price"`
	Stock        int64  `json:"stock"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Status       int8   `json:"status"`
}

type SeckillParam struct {
	BookID       int64  `json:"book_id" binding:"required"`
	SeckillPrice int64  `json:"seckill_price" binding:"required"`
	Stock        int64  `json:"stock" binding:"required,min=1"`
	StartTime    string `json:"start_time" binding:"required"`
	EndTime      string `json:"end_time" binding:"required"`
}

type SeckillRequest struct {
	ProductID int64 `json:"product_id" binding:"required"`
}

type SeckillMsg struct {
	UserID    int64 `json:"user_id"`
	ProductID int64 `json:"product_id"`
}
