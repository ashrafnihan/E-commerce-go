package user

import "time"

type User struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Email         string    `json:"email" gorm:"type:citext;uniqueIndex;not null"`
	PasswordHash  string    `json:"-" gorm:"column:password_hash;not null"`
	Role          string    `json:"role" gorm:"type:text;not null;default:'user'"`
	IsActive      bool      `json:"is_active" gorm:"not null;default:true"`
	EmailVerified bool      `json:"email_verified" gorm:"not null;default:false"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (User) TableName() string { return "users" }
