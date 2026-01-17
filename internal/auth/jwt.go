package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	Issuer         string
	AccessSecret   string
	RefreshSecret  string
	AccessTTLMin   int
	RefreshTTLDays int
}

type JWTManager struct {
	cfg JWTConfig
}

type Claims struct {
	UserID int64  `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTManager(cfg JWTConfig) *JWTManager {
	return &JWTManager{cfg: cfg}
}

func (m *JWTManager) AccessTTL() time.Duration {
	return time.Duration(m.cfg.AccessTTLMin) * time.Minute
}

func (m *JWTManager) RefreshTTL() time.Duration {
	return time.Duration(m.cfg.RefreshTTLDays) * 24 * time.Hour
}

func (m *JWTManager) SignAccess(userID int64, role string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(m.AccessTTL())
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString([]byte(m.cfg.AccessSecret))
	return s, exp, err
}

func (m *JWTManager) SignRefresh(userID int64, role string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(m.RefreshTTL())
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString([]byte(m.cfg.RefreshSecret))
	return s, exp, err
}

func (m *JWTManager) ParseAccess(tokenStr string) (*Claims, error) {
	return m.parse(tokenStr, []byte(m.cfg.AccessSecret))
}

func (m *JWTManager) ParseRefresh(tokenStr string) (*Claims, error) {
	return m.parse(tokenStr, []byte(m.cfg.RefreshSecret))
}

func (m *JWTManager) parse(tokenStr string, secret []byte) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
