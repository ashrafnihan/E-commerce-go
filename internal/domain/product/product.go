package product

import "time"

type ProductType struct {
	ID         int64     `json:"id"`
	CategoryID int64     `json:"category_id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
}

type Product struct {
	ID          int64     `json:"id"`
	CategoryID  int64     `json:"category_id"`
	TypeID      int64     `json:"type_id"`
	TypeName    string    `json:"type_name,omitempty"`
	Category    string    `json:"category,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   *int64    `json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Variants    []Variant `json:"variants,omitempty"`
}

type Variant struct {
	ID              int64     `json:"id"`
	ProductID       int64     `json:"product_id"`
	Size            string    `json:"size"`
	Color           string    `json:"color"`
	Price           float64   `json:"price"`
	DiscountPercent int       `json:"discount_percent"`
	FinalPrice      float64   `json:"final_price"`
	StockQty        int       `json:"stock_qty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
