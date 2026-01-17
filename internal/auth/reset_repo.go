package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ResetRepo struct {
	db *pgxpool.Pool
}

func NewResetRepo(db *pgxpool.Pool) *ResetRepo {
	return &ResetRepo{db: db}
}

func (r *ResetRepo) Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO password_resets (user_id, token_hash, expires_at)
		VALUES ($1,$2,$3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (r *ResetRepo) Consume(ctx context.Context, tokenHash string) (int64, bool, error) {
	// Mark as used if valid; return user_id
	var userID int64
	err := r.db.QueryRow(ctx, `
		UPDATE password_resets
		SET used_at=now()
		WHERE token_hash=$1
		  AND used_at IS NULL
		  AND expires_at > now()
		RETURNING user_id
	`, tokenHash).Scan(&userID)
	if err != nil {
		// no rows => invalid/expired/used
		return 0, false, nil
	}
	return userID, true, nil
}
