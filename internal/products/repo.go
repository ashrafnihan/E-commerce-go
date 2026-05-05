package products

import (
	"fmt"
	"math"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ecommerce/internal/domain/product"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

// GetOrCreateType finds or creates a product_type for a category
func (r *Repo) GetOrCreateType(categoryID int64, typeName string) (int64, error) {
	pt := product.ProductType{
		CategoryID: categoryID,
		Name:       typeName,
	}
	err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "category_id"}, {Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"name"}),
	}).Create(&pt).Error
	return pt.ID, err
}

type CreateProductInput struct {
	CategoryID  int64
	TypeName    string
	Name        string
	Description string
	CreatedBy   int64
	Variants    []CreateVariantInput
}

type CreateVariantInput struct {
	Size            string
	Color           string
	Price           float64
	DiscountPercent int
	StockQty        int
}

func (r *Repo) CreateProductWithVariants(in CreateProductInput) (product.Product, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return product.Product{}, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get or create product type
	pt := product.ProductType{
		CategoryID: in.CategoryID,
		Name:       in.TypeName,
	}
	err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "category_id"}, {Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"name"}),
	}).Create(&pt).Error
	if err != nil {
		tx.Rollback()
		return product.Product{}, err
	}

	// Create product
	p := product.Product{
		CategoryID:  in.CategoryID,
		TypeID:      pt.ID,
		Name:        in.Name,
		Description: in.Description,
		IsActive:    true,
		CreatedBy:   &in.CreatedBy,
	}
	if err := tx.Create(&p).Error; err != nil {
		tx.Rollback()
		return product.Product{}, err
	}

	// Create variants
	for _, v := range in.Variants {
		variant := product.Variant{
			ProductID:       p.ID,
			Size:            v.Size,
			Color:           v.Color,
			Price:           v.Price,
			DiscountPercent: v.DiscountPercent,
			StockQty:        v.StockQty,
		}
		if err := tx.Create(&variant).Error; err != nil {
			tx.Rollback()
			return product.Product{}, fmt.Errorf("variant insert failed: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return product.Product{}, err
	}
	return p, nil
}

func (r *Repo) ListPublic(categorySlug *string) ([]product.Product, error) {
	query := r.db.Table("products p").
		Select(`p.id, p.category_id, p.type_id, p.name, COALESCE(p.description,'') as description,
		        p.is_active, p.created_at, p.updated_at,
		        c.name as category, pt.name as type_name`).
		Joins("JOIN categories c ON c.id = p.category_id").
		Joins("JOIN product_types pt ON pt.id = p.type_id").
		Where("p.is_active = ? AND c.is_active = ?", true, true)

	if categorySlug != nil && *categorySlug != "" {
		query = query.Where("c.slug = ?", *categorySlug)
	}

	query = query.Order("p.created_at DESC")

	rows, err := query.Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []product.Product
	for rows.Next() {
		var p product.Product
		if err := rows.Scan(
			&p.ID, &p.CategoryID, &p.TypeID, &p.Name, &p.Description,
			&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
			&p.Category, &p.TypeName,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *Repo) GetProductPublic(id int64) (product.Product, error) {
	var p product.Product

	err := r.db.Table("products p").
		Select(`p.id, p.category_id, p.type_id, p.name, COALESCE(p.description,'') as description,
		        p.is_active, p.created_at, p.updated_at,
		        c.name as category, pt.name as type_name`).
		Joins("JOIN categories c ON c.id = p.category_id").
		Joins("JOIN product_types pt ON pt.id = p.type_id").
		Where("p.id = ? AND p.is_active = ? AND c.is_active = ?", id, true, true).
		Row().Scan(
		&p.ID, &p.CategoryID, &p.TypeID, &p.Name, &p.Description,
		&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		&p.Category, &p.TypeName,
	)
	if err != nil {
		return product.Product{}, err
	}

	// Load variants
	var variants []product.Variant
	err = r.db.Where("product_id = ?", p.ID).Order("id ASC").Find(&variants).Error
	if err != nil {
		return product.Product{}, err
	}

	// Compute final price for each variant
	for i := range variants {
		variants[i].FinalPrice = math.Round(variants[i].Price*float64(100-variants[i].DiscountPercent)) / 100.0
	}
	p.Variants = variants

	return p, nil
}
