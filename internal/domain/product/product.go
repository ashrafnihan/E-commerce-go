package product

import "time"

type ProductType struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	CategoryID int64     `json:"category_id" gorm:"not null;index"`
	Name       string    `json:"name" gorm:"type:text;not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (ProductType) TableName() string { return "product_types" }

type Product struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	CategoryID  int64     `json:"category_id" gorm:"not null;index"`
	TypeID      int64     `json:"type_id" gorm:"not null;index"`
	TypeName    string    `json:"type_name,omitempty" gorm:"-"` // computed via join
	Category    string    `json:"category,omitempty" gorm:"-"`  // computed via join
	Name        string    `json:"name" gorm:"type:text;not null"`
	Description string    `json:"description,omitempty" gorm:"type:text"`
	IsActive    bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedBy   *int64    `json:"created_by,omitempty" gorm:"index"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Variants    []Variant `json:"variants,omitempty" gorm:"foreignKey:ProductID"`
}

func (Product) TableName() string { return "products" }

type Variant struct {
	ID              int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ProductID       int64     `json:"product_id" gorm:"not null;index"`
	Size            string    `json:"size" gorm:"type:text;not null"`
	Color           string    `json:"color" gorm:"type:text;not null"`
	Price           float64   `json:"price" gorm:"type:numeric(12,2);not null"`
	DiscountPercent int       `json:"discount_percent" gorm:"not null;default:0"`
	FinalPrice      float64   `json:"final_price" gorm:"-"` // computed
	StockQty        int       `json:"stock_qty" gorm:"not null;default:0"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Variant) TableName() string { return "product_variants" }

// ComputeFinalPrice calculates the discounted price
func (v *Variant) ComputeFinalPrice() {
	v.FinalPrice = v.Price * float64(100-v.DiscountPercent) / 100.0
}
