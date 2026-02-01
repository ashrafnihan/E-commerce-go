package products

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ecommerce/internal/domain/product"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

// Create type if not exists for a category (so admin can type "Formal" and next time it shows)
func (r *Repo) GetOrCreateType(ctx context.Context, categoryID int64, typeName string) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO product_types (category_id, name)
		VALUES ($1, $2)
		ON CONFLICT (category_id, name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, categoryID, typeName).Scan(&id)
	return id, err
}

type CreateProductInput struct {
	CategoryID  int64
	TypeName    string
	Name        string
	Description string
	CreatedBy   int64

	Variants []CreateVariantInput
}

type CreateVariantInput struct {
	Size            string
	Color           string
	Price           float64
	DiscountPercent int
	StockQty        int
}

func (r *Repo) CreateProductWithVariants(ctx context.Context, in CreateProductInput) (product.Product, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return product.Product{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	typeID, err := func() (int64, error) {
		var id int64
		err := tx.QueryRow(ctx, `
			INSERT INTO product_types (category_id, name)
			VALUES ($1, $2)
			ON CONFLICT (category_id, name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, in.CategoryID, in.TypeName).Scan(&id)
		return id, err
	}()
	if err != nil {
		return product.Product{}, err
	}

	var p product.Product
	err = tx.QueryRow(ctx, `
		INSERT INTO products (category_id, type_id, name, description, created_by, is_active)
		VALUES ($1,$2,$3,$4,$5,true)
		RETURNING id, category_id, type_id, name, description, is_active, created_by, created_at, updated_at
	`, in.CategoryID, typeID, in.Name, in.Description, in.CreatedBy).Scan(
		&p.ID, &p.CategoryID, &p.TypeID, &p.Name, &p.Description, &p.IsActive, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return product.Product{}, err
	}

	for _, v := range in.Variants {
		_, err := tx.Exec(ctx, `
			INSERT INTO product_variants (product_id, size, color, price, discount_percent, stock_qty)
			VALUES ($1,$2,$3,$4,$5,$6)
		`, p.ID, v.Size, v.Color, v.Price, v.DiscountPercent, v.StockQty)
		if err != nil {
			return product.Product{}, fmt.Errorf("variant insert failed: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return product.Product{}, err
	}
	return p, nil
}

func (r *Repo) ListPublic(ctx context.Context, categorySlug *string) ([]product.Product, error) {
	q := `
		SELECT
		  p.id, p.category_id, p.type_id, p.name, COALESCE(p.description,''), p.is_active, p.created_at, p.updated_at,
		  c.name as category_name,
		  pt.name as type_name
		FROM products p
		JOIN categories c ON c.id = p.category_id
		JOIN product_types pt ON pt.id = p.type_id
		WHERE p.is_active = true AND c.is_active = true
	`
	args := []any{}
	if categorySlug != nil && *categorySlug != "" {
		q += " AND c.slug = $1 "
		args = append(args, *categorySlug)
	}
	q += " ORDER BY p.created_at DESC "

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []product.Product
	for rows.Next() {
		var p product.Product
		if err := rows.Scan(
			&p.ID, &p.CategoryID, &p.TypeID, &p.Name, &p.Description, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
			&p.Category, &p.TypeName,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repo) GetProductPublic(ctx context.Context, id int64) (product.Product, error) {
	var p product.Product
	err := r.db.QueryRow(ctx, `
		SELECT
		  p.id, p.category_id, p.type_id, p.name, COALESCE(p.description,''), p.is_active, p.created_at, p.updated_at,
		  c.name as category_name, pt.name as type_name
		FROM products p
		JOIN categories c ON c.id = p.category_id
		JOIN product_types pt ON pt.id = p.type_id
		WHERE p.id = $1 AND p.is_active = true AND c.is_active = true
	`, id).Scan(
		&p.ID, &p.CategoryID, &p.TypeID, &p.Name, &p.Description, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		&p.Category, &p.TypeName,
	)
	if err != nil {
		return product.Product{}, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, product_id, size, color, price, discount_percent,
		       ROUND(price * (100 - discount_percent) / 100.0, 2) as final_price,
		       stock_qty, created_at, updated_at
		FROM product_variants
		WHERE product_id = $1
		ORDER BY id ASC
	`, p.ID)
	if err != nil {
		return product.Product{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var v product.Variant
		if err := rows.Scan(&v.ID, &v.ProductID, &v.Size, &v.Color, &v.Price, &v.DiscountPercent, &v.FinalPrice, &v.StockQty, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return product.Product{}, err
		}
		p.Variants = append(p.Variants, v)
	}
	return p, rows.Err()
}
