package cart

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"ecommerce/internal/domain/cart"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) getOrCreateCartID(ctx context.Context, userID int64) (int64, error) {
	var cartID int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO carts (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO UPDATE SET updated_at = now()
		RETURNING id
	`, userID).Scan(&cartID)
	return cartID, err
}

func (r *Repo) AddItem(ctx context.Context, userID, variantID int64, qty int) error {
	cartID, err := r.getOrCreateCartID(ctx, userID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO cart_items (cart_id, variant_id, qty)
		VALUES ($1,$2,$3)
		ON CONFLICT (cart_id, variant_id)
		DO UPDATE SET qty = cart_items.qty + EXCLUDED.qty
	`, cartID, variantID, qty)
	return err
}

func (r *Repo) UpdateQty(ctx context.Context, userID, variantID int64, qty int) error {
	cartID, err := r.getOrCreateCartID(ctx, userID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, `
		UPDATE cart_items
		SET qty = $3
		WHERE cart_id = $1 AND variant_id = $2
	`, cartID, variantID, qty)
	return err
}

func (r *Repo) RemoveItem(ctx context.Context, userID, variantID int64) error {
	cartID, err := r.getOrCreateCartID(ctx, userID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, `
		DELETE FROM cart_items
		WHERE cart_id = $1 AND variant_id = $2
	`, cartID, variantID)
	return err
}

func (r *Repo) GetCart(ctx context.Context, userID int64) (cart.Cart, error) {
	cartID, err := r.getOrCreateCartID(ctx, userID)
	if err != nil {
		return cart.Cart{}, err
	}

	out := cart.Cart{ID: cartID, UserID: userID}

	rows, err := r.db.Query(ctx, `
		SELECT
		  ci.id, ci.variant_id, ci.qty,
		  p.id as product_id,
		  p.name as product_name,
		  c.name as category_name,
		  pt.name as type_name,
		  v.size, v.color,
		  v.price, v.discount_percent,
		  ROUND(v.price * (100 - v.discount_percent) / 100.0, 2) as final_price,
		  v.stock_qty
		FROM cart_items ci
		JOIN product_variants v ON v.id = ci.variant_id
		JOIN products p ON p.id = v.product_id
		JOIN categories c ON c.id = p.category_id
		JOIN product_types pt ON pt.id = p.type_id
		WHERE ci.cart_id = $1
		ORDER BY ci.created_at DESC
	`, cartID)
	if err != nil {
		return cart.Cart{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var it cart.CartItem
		if err := rows.Scan(
			&it.ID, &it.VariantID, &it.Qty,
			&it.ProductID, &it.Product, &it.Category, &it.TypeName,
			&it.Size, &it.Color,
			&it.Price, &it.Discount, &it.FinalPrice, &it.StockQty,
		); err != nil {
			return cart.Cart{}, err
		}
		out.Items = append(out.Items, it)
	}
	return out, rows.Err()
}
