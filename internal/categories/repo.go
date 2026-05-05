package categories

import (
	"gorm.io/gorm"

	"ecommerce/internal/domain/category"
	"ecommerce/internal/util"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) ListActive() ([]category.Category, error) {
	var out []category.Category
	err := r.db.Where("is_active = ?", true).
		Order("sort_order ASC, name ASC").
		Find(&out).Error
	return out, err
}

func (r *Repo) AdminListAll() ([]category.Category, error) {
	var out []category.Category
	err := r.db.Order("sort_order ASC, name ASC").Find(&out).Error
	return out, err
}

func (r *Repo) Create(name string, sortOrder int) (category.Category, error) {
	c := category.Category{
		Name:      name,
		Slug:      util.Slugify(name),
		SortOrder: sortOrder,
		IsActive:  true,
	}
	if err := r.db.Create(&c).Error; err != nil {
		return category.Category{}, err
	}
	return c, nil
}

func (r *Repo) Update(id int64, name *string, sortOrder *int, isActive *bool) (category.Category, error) {
	var c category.Category
	if err := r.db.First(&c, id).Error; err != nil {
		return category.Category{}, err
	}

	updates := map[string]interface{}{}
	if name != nil {
		updates["name"] = *name
		updates["slug"] = util.Slugify(*name)
	}
	if sortOrder != nil {
		updates["sort_order"] = *sortOrder
	}
	if isActive != nil {
		updates["is_active"] = *isActive
	}

	if len(updates) > 0 {
		if err := r.db.Model(&c).Updates(updates).Error; err != nil {
			return category.Category{}, err
		}
	}

	// Reload to return fresh data
	r.db.First(&c, id)
	return c, nil
}
