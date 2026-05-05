package auth

import (
	"time"

	"gorm.io/gorm"
)

// PasswordReset is the GORM model for password_resets table
type PasswordReset struct {
	ID        int64      `gorm:"primaryKey;autoIncrement"`
	UserID    int64      `gorm:"not null;index"`
	TokenHash string     `gorm:"column:token_hash;not null"`
	ExpiresAt time.Time  `gorm:"not null"`
	UsedAt    *time.Time `gorm:""`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
}

func (PasswordReset) TableName() string { return "password_resets" }

type ResetRepo struct {
	db *gorm.DB
}

func NewResetRepo(db *gorm.DB) *ResetRepo {
	return &ResetRepo{db: db}
}

func (r *ResetRepo) Create(userID int64, tokenHash string, expiresAt time.Time) error {
	pr := PasswordReset{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	return r.db.Create(&pr).Error
}

func (r *ResetRepo) Consume(tokenHash string) (int64, bool, error) {
	var pr PasswordReset
	now := time.Now()

	// Find the valid reset token
	err := r.db.Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, now).
		First(&pr).Error
	if err != nil {
		return 0, false, nil // not found or expired
	}

	// Mark as used
	r.db.Model(&pr).Update("used_at", now)
	return pr.UserID, true, nil
}
