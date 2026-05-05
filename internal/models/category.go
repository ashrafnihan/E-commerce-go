package models

import "time"

type Category struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"not null"`
	Slug      string    `gorm:"type:citext;uniqueIndex;not null"`
	IsActive  bool      `gorm:"not null;default:true"`
	SortOrder int       `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

func (Category) TableName() string { return "categories" }
