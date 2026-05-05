package auth

import (
	"time"

	"gorm.io/gorm"
)

// RefreshToken is the GORM model for refresh_tokens table
type RefreshToken struct {
	ID        int64      `gorm:"primaryKey;autoIncrement"`
	UserID    int64      `gorm:"not null;index"`
	TokenHash string     `gorm:"column:token_hash;not null"`
	ExpiresAt time.Time  `gorm:"not null"`
	RevokedAt *time.Time `gorm:""`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

type RefreshRepo struct {
	db *gorm.DB
}

func NewRefreshRepo(db *gorm.DB) *RefreshRepo {
	return &RefreshRepo{db: db}
}

func (r *RefreshRepo) Store(userID int64, tokenHash string, expiresAt time.Time) error {
	rt := RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	return r.db.Create(&rt).Error
}

func (r *RefreshRepo) IsValid(userID int64, tokenHash string) (bool, error) {
	var count int64
	err := r.db.Model(&RefreshToken{}).
		Where("user_id = ? AND token_hash = ? AND revoked_at IS NULL AND expires_at > ?",
			userID, tokenHash, time.Now()).
		Count(&count).Error
	return count > 0, err
}

func (r *RefreshRepo) Revoke(userID int64, tokenHash string) error {
	now := time.Now()
	return r.db.Model(&RefreshToken{}).
		Where("user_id = ? AND token_hash = ? AND revoked_at IS NULL", userID, tokenHash).
		Update("revoked_at", now).Error
}
