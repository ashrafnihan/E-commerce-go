package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OTPRepo struct {
	db *pgxpool.Pool
}

func NewOTPRepo(db *pgxpool.Pool) *OTPRepo {
	return &OTPRepo{db: db}
}

// GetValid returns (otp_hash, expires_at, found, error)
func (r *OTPRepo) GetValid(ctx context.Context, userID int64, purpose string) (string, time.Time, bool, error) {
	var hash string
	var exp time.Time

	err := r.db.QueryRow(ctx, `
		SELECT otp_hash, expires_at
		FROM user_otps
		WHERE user_id=$1 AND purpose=$2 AND expires_at > now()
	`, userID, purpose).Scan(&hash, &exp)

	if err != nil {
		// Not found or any scan error -> treat as not found
		return "", time.Time{}, false, nil
	}

	return hash, exp, true, nil
}

// Upsert inserts or overwrites OTP for that (user_id, purpose)
func (r *OTPRepo) Upsert(ctx context.Context, userID int64, purpose, otpHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_otps (user_id, purpose, otp_hash, expires_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (user_id, purpose)
		DO UPDATE SET otp_hash=EXCLUDED.otp_hash, expires_at=EXCLUDED.expires_at
	`, userID, purpose, otpHash, expiresAt)
	return err
}

func (r *OTPRepo) Delete(ctx context.Context, userID int64, purpose string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM user_otps
		WHERE user_id=$1 AND purpose=$2
	`, userID, purpose)
	return err
}
