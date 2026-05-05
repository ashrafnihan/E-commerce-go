package auth

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserOTP is the GORM model for user_otps table
type UserOTP struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	UserID    int64     `gorm:"not null;uniqueIndex:idx_user_purpose"`
	Purpose   string    `gorm:"type:text;not null;uniqueIndex:idx_user_purpose"`
	OTPHash   string    `gorm:"column:otp_hash;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (UserOTP) TableName() string { return "user_otps" }

type OTPRepo struct {
	db *gorm.DB
}

func NewOTPRepo(db *gorm.DB) *OTPRepo {
	return &OTPRepo{db: db}
}

// GetValid returns (otp_hash, expires_at, found, error)
func (r *OTPRepo) GetValid(userID int64, purpose string) (string, time.Time, bool, error) {
	var otp UserOTP
	err := r.db.Where("user_id = ? AND purpose = ? AND expires_at > ?",
		userID, purpose, time.Now()).
		First(&otp).Error
	if err != nil {
		return "", time.Time{}, false, nil
	}
	return otp.OTPHash, otp.ExpiresAt, true, nil
}

// Upsert inserts or overwrites OTP for that (user_id, purpose)
func (r *OTPRepo) Upsert(userID int64, purpose, otpHash string, expiresAt time.Time) error {
	otp := UserOTP{
		UserID:    userID,
		Purpose:   purpose,
		OTPHash:   otpHash,
		ExpiresAt: expiresAt,
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "purpose"}},
		DoUpdates: clause.AssignmentColumns([]string{"otp_hash", "expires_at"}),
	}).Create(&otp).Error
}

func (r *OTPRepo) Delete(userID int64, purpose string) error {
	return r.db.Where("user_id = ? AND purpose = ?", userID, purpose).
		Delete(&UserOTP{}).Error
}
