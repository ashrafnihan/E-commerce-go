package models

import "time"

type ProductType struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	CategoryID int64     `gorm:"not null;uniqueIndex:uniq_cat_type"`
	Name       string    `gorm:"not null;uniqueIndex:uniq_cat_type"`
	CreatedAt  time.Time `gorm:"not null;default:now()"`
}

func (ProductType) TableName() string { return "product_types" }

type Product struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	CategoryID  int64     `gorm:"not null;index"`
	TypeID      int64     `gorm:"not null;index"`
	Name        string    `gorm:"not null"`
	Description *string
	IsActive    bool      `gorm:"not null;default:true"`
	CreatedBy   *int64
	CreatedAt   time.Time `gorm:"not null;default:now()"`
	UpdatedAt   time.Time `gorm:"not null;default:now()"`
	PopularityScore int     `gorm:"not null;default:0"`
    AvgRating       float64 `gorm:"not null;default:0"`
    RatingCount     int     `gorm:"not null;default:0"`
}

func (Product) TableName() string { return "products" }

type ProductVariant struct {
	ID              int64     `gorm:"primaryKey;autoIncrement"`
	ProductID       int64     `gorm:"not null;uniqueIndex:uniq_prod_size_color"`
	Size            string    `gorm:"not null;uniqueIndex:uniq_prod_size_color"`
	Color           string    `gorm:"not null;uniqueIndex:uniq_prod_size_color"`
	Price           float64   `gorm:"not null"`
	DiscountPercent int       `gorm:"not null;default:0"`
	StockQty        int       `gorm:"not null;default:0"`
	CreatedAt       time.Time `gorm:"not null;default:now()"`
	UpdatedAt       time.Time `gorm:"not null;default:now()"`
}

func (ProductVariant) TableName() string { return "product_variants" }
