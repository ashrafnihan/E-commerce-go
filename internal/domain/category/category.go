package category

import "time"

type Category struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"type:text;not null"`
	Slug      string    `json:"slug" gorm:"type:citext;uniqueIndex;not null"`
	IsActive  bool      `json:"is_active" gorm:"not null;default:true"`
	SortOrder int       `json:"sort_order" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Category) TableName() string { return "categories" }
