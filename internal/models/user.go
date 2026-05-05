package models

import "time"

type User struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	Email        string    `gorm:"type:citext;uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	Role         string    `gorm:"not null"` // user/admin (migration enforces check)
	IsActive     bool      `gorm:"not null;default:true"`
	EmailVerified bool     `gorm:"not null;default:false"` // added by 002_otp.sql
	CreatedAt    time.Time `gorm:"not null;default:now()"`
	UpdatedAt    time.Time `gorm:"not null;default:now()"`
}

func (User) TableName() string { return "users" }
