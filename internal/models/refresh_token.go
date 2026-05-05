package models

import "time"

type RefreshToken struct {
	ID        int64      `gorm:"primaryKey;autoIncrement"`
	UserID    int64      `gorm:"not null;index"`
	TokenHash string     `gorm:"not null;index"`
	ExpiresAt time.Time  `gorm:"not null"`
	RevokedAt *time.Time `gorm:"default:null"`
	CreatedAt time.Time  `gorm:"not null;default:now()"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
