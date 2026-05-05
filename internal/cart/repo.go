package cart

import (
	"math"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	cartDomain "ecommerce/internal/domain/cart"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) getOrCreateCartID(userID int64) (int64, error) {
	c := cartDomain.Cart{UserID: userID}
	err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
	}).Create(&c).Error
	if err != nil {
		// If the upsert returned an existing ID, query it
		var existing cartDomain.Cart
		if findErr := r.db.Where("user_id = ?", userID).First(&existing).Error; findErr != nil {
			return 0, err
		}
		return existing.ID, nil
	}
	return c.ID, nil
}

func (r *Repo) AddItem(userID, variantID int64, qty int) error {
	cartID, err := r.getOrCreateCartID(userID)
	if err != nil {
		return err
	}

	item := cartDomain.CartItem{
		CartID:    cartID,
		VariantID: variantID,
		Qty:       qty,
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "cart_id"}, {Name: "variant_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"qty": gorm.Expr("cart_items.qty + ?", qty),
		}),
	}).Create(&item).Error
}

func (r *Repo) UpdateQty(userID, variantID int64, qty int) error {
	cartID, err := r.getOrCreateCartID(userID)
	if err != nil {
		return err
	}

	return r.db.Model(&cartDomain.CartItem{}).
		Where("cart_id = ? AND variant_id = ?", cartID, variantID).
		Update("qty", qty).Error
}

func (r *Repo) RemoveItem(userID, variantID int64) error {
	cartID, err := r.getOrCreateCartID(userID)
	if err != nil {
		return err
	}

	return r.db.Where("cart_id = ? AND variant_id = ?", cartID, variantID).
		Delete(&cartDomain.CartItem{}).Error
}

func (r *Repo) GetCart(userID int64) (cartDomain.Cart, error) {
	cartID, err := r.getOrCreateCartID(userID)
	if err != nil {
		return cartDomain.Cart{}, err
	}

	out := cartDomain.Cart{ID: cartID, UserID: userID}

	rows, err := r.db.Table("cart_items ci").
		Select(`ci.id, ci.variant_id, ci.qty,
		        p.id as product_id,
		        p.name as product_name,
		        c.name as category_name,
		        pt.name as type_name,
		        v.size, v.color,
		        v.price, v.discount_percent,
		        ROUND(v.price * (100 - v.discount_percent) / 100.0, 2) as final_price,
		        v.stock_qty`).
		Joins("JOIN product_variants v ON v.id = ci.variant_id").
		Joins("JOIN products p ON p.id = v.product_id").
		Joins("JOIN categories c ON c.id = p.category_id").
		Joins("JOIN product_types pt ON pt.id = p.type_id").
		Where("ci.cart_id = ?", cartID).
		Order("ci.created_at DESC").
		Rows()
	if err != nil {
		return cartDomain.Cart{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var it cartDomain.CartItem
		if err := rows.Scan(
			&it.ID, &it.VariantID, &it.Qty,
			&it.ProductID, &it.Product, &it.Category, &it.TypeName,
			&it.Size, &it.Color,
			&it.Price, &it.Discount, &it.FinalPrice, &it.StockQty,
		); err != nil {
			return cartDomain.Cart{}, err
		}
		_ = math.Round(it.FinalPrice*100) / 100 // ensure precision
		out.Items = append(out.Items, it)
	}
	return out, nil
}
