package cart

type Cart struct {
	ID     int64      `json:"id"`
	UserID int64      `json:"user_id"`
	Items  []CartItem `json:"items"`
}

type CartItem struct {
	ID         int64   `json:"id"`
	VariantID  int64   `json:"variant_id"`
	Qty        int     `json:"qty"`
	ProductID  int64   `json:"product_id"`
	Product    string  `json:"product"`
	Category   string  `json:"category"`
	TypeName   string  `json:"type"`
	Size       string  `json:"size"`
	Color      string  `json:"color"`
	Price      float64 `json:"price"`
	Discount   int     `json:"discount_percent"`
	FinalPrice float64 `json:"final_price"`
	StockQty   int     `json:"stock_qty"`
}
