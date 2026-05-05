package auth

import (
	"errors"

	"gorm.io/gorm"

	"ecommerce/internal/domain/user"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(email, passwordHash, role string) (user.User, error) {
	u := user.User{
		Email:         email,
		PasswordHash:  passwordHash,
		Role:          role,
		IsActive:      true,
		EmailVerified: false,
	}
	if err := r.db.Create(&u).Error; err != nil {
		return user.User{}, err
	}
	return u, nil
}

func (r *UserRepo) ByEmail(email string) (user.User, error) {
	var u user.User
	err := r.db.Where("email = ?", email).First(&u).Error
	if err != nil {
		return user.User{}, err
	}
	return u, nil
}

func (r *UserRepo) ByID(id int64) (user.User, error) {
	var u user.User
	err := r.db.First(&u, id).Error
	if err != nil {
		return user.User{}, err
	}
	return u, nil
}

func (r *UserRepo) UpdatePassword(userID int64, newHash string) error {
	result := r.db.Model(&user.User{}).Where("id = ?", userID).Update("password_hash", newHash)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *UserRepo) SetEmailVerified(userID int64) error {
	return r.db.Model(&user.User{}).Where("id = ?", userID).Update("email_verified", true).Error
}
