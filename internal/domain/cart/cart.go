package cart

import "time"

type Cart struct {
	ID        int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int64      `json:"user_id" gorm:"uniqueIndex;not null"`
	CreatedAt time.Time  `json:"-" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"-" gorm:"autoUpdateTime"`
	Items     []CartItem `json:"items" gorm:"foreignKey:CartID"`
}

func (Cart) TableName() string { return "carts" }

type CartItem struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	CartID    int64     `json:"-" gorm:"not null;index;uniqueIndex:idx_cart_variant"`
	VariantID int64     `json:"variant_id" gorm:"not null;uniqueIndex:idx_cart_variant"`
	Qty       int       `json:"qty" gorm:"not null"`
	CreatedAt time.Time `json:"-" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"-" gorm:"autoUpdateTime"`

	// Computed fields (populated via joins, not stored)
	ProductID  int64   `json:"product_id" gorm:"-"`
	Product    string  `json:"product" gorm:"-"`
	Category   string  `json:"category" gorm:"-"`
	TypeName   string  `json:"type" gorm:"-"`
	Size       string  `json:"size" gorm:"-"`
	Color      string  `json:"color" gorm:"-"`
	Price      float64 `json:"price" gorm:"-"`
	Discount   int     `json:"discount_percent" gorm:"-"`
	FinalPrice float64 `json:"final_price" gorm:"-"`
	StockQty   int     `json:"stock_qty" gorm:"-"`
}

func (CartItem) TableName() string { return "cart_items" }
