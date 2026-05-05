package models

import "time"

type Cart struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	UserID    int64     `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

func (Cart) TableName() string { return "carts" }

type CartItem struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	CartID    int64     `gorm:"not null;uniqueIndex:uniq_cart_variant"`
	VariantID int64     `gorm:"not null;uniqueIndex:uniq_cart_variant"`
	Qty       int       `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

func (CartItem) TableName() string { return "cart_items" }
