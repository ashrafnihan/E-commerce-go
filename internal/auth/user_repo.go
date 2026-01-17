package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"ecommerce/internal/domain/user"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, email, passwordHash, role string) (user.User, error) {
	var u user.User
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1,$2,$3)
		RETURNING id, email, password_hash, role, is_active, created_at, updated_at
	`, email, passwordHash, role).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	return u, err
}

func (r *UserRepo) ByEmail(ctx context.Context, email string) (user.User, error) {
	var u user.User
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE email=$1
	`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (r *UserRepo) ByID(ctx context.Context, id int64) (user.User, error) {
	var u user.User
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE id=$1
	`, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID int64, newHash string) error {
	ct, err := r.db.Exec(ctx, `UPDATE users SET password_hash=$1 WHERE id=$2`, newHash, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}
