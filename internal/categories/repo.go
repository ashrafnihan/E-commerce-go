package categories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"ecommerce/internal/domain/category"
	"ecommerce/internal/util"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) ListActive(ctx context.Context) ([]category.Category, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, slug, is_active, sort_order, created_at, updated_at
		FROM categories
		WHERE is_active = true
		ORDER BY sort_order ASC, name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []category.Category
	for rows.Next() {
		var c category.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.IsActive, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repo) AdminListAll(ctx context.Context) ([]category.Category, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, slug, is_active, sort_order, created_at, updated_at
		FROM categories
		ORDER BY sort_order ASC, name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []category.Category
	for rows.Next() {
		var c category.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.IsActive, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repo) Create(ctx context.Context, name string, sortOrder int) (category.Category, error) {
	slug := util.Slugify(name)

	var c category.Category
	err := r.db.QueryRow(ctx, `
		INSERT INTO categories (name, slug, sort_order, is_active)
		VALUES ($1, $2, $3, true)
		RETURNING id, name, slug, is_active, sort_order, created_at, updated_at
	`, name, slug, sortOrder).Scan(
		&c.ID, &c.Name, &c.Slug, &c.IsActive, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (r *Repo) Update(ctx context.Context, id int64, name *string, sortOrder *int, isActive *bool) (category.Category, error) {
	// Keep slug synced with name if name updated (simple approach)
	var c category.Category
	err := r.db.QueryRow(ctx, `
		UPDATE categories
		SET
		  name = COALESCE($2, name),
		  slug = CASE WHEN $2 IS NULL THEN slug ELSE $5 END,
		  sort_order = COALESCE($3, sort_order),
		  is_active = COALESCE($4, is_active)
		WHERE id = $1
		RETURNING id, name, slug, is_active, sort_order, created_at, updated_at
	`, id, name, sortOrder, isActive, func() any {
		if name == nil {
			return nil
		}
		s := util.Slugify(*name)
		return s
	}()).Scan(&c.ID, &c.Name, &c.Slug, &c.IsActive, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}
