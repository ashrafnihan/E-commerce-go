package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshRepo struct {
	db *pgxpool.Pool
}

func NewRefreshRepo(db *pgxpool.Pool) *RefreshRepo {
	return &RefreshRepo{db: db}
}

func (r *RefreshRepo) Store(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1,$2,$3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (r *RefreshRepo) IsValid(ctx context.Context, userID int64, tokenHash string) (bool, error) {
	var ok bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM refresh_tokens
			WHERE user_id=$1 AND token_hash=$2
			  AND revoked_at IS NULL
			  AND expires_at > now()
		)
	`, userID, tokenHash).Scan(&ok)
	return ok, err
}

func (r *RefreshRepo) Revoke(ctx context.Context, userID int64, tokenHash string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at=now()
		WHERE user_id=$1 AND token_hash=$2 AND revoked_at IS NULL
	`, userID, tokenHash)
	return err
}
