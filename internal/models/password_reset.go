package models

import "time"

type PasswordReset struct {
	ID        int64      `gorm:"primaryKey;autoIncrement"`
	UserID    int64      `gorm:"not null;index"`
	TokenHash string     `gorm:"not null;index"`
	ExpiresAt time.Time  `gorm:"not null"`
	UsedAt    *time.Time `gorm:"default:null"`
	CreatedAt time.Time  `gorm:"not null;default:now()"`
}

func (PasswordReset) TableName() string { return "password_resets" }
