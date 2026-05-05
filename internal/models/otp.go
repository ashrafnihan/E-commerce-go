package models

import "time"

type UserOTP struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	UserID    int64     `gorm:"not null;index;uniqueIndex:uniq_user_purpose"`
	Purpose   string    `gorm:"not null;uniqueIndex:uniq_user_purpose"` // verify_email/reset_password
	OTPHash   string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

func (UserOTP) TableName() string { return "user_otps" }
