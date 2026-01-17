package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv    string
	HTTPAddr  string

	DatabaseURL string

	JWTIssuer        string
	JWTAccessSecret  string
	JWTRefreshSecret string
	AccessTokenTTLMin   int
	RefreshTokenTTLDays int

	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	AppBaseURL string
	ResetPath  string
}

func Load() Config {
	return Config{
		AppEnv:     get("APP_ENV", "dev"),
		HTTPAddr:   get("HTTP_ADDR", ":8080"),

		DatabaseURL: get("DATABASE_URL", ""),

		JWTIssuer:        get("JWT_ISSUER", "ecommerce"),
		JWTAccessSecret:  get("JWT_ACCESS_SECRET", ""),
		JWTRefreshSecret: get("JWT_REFRESH_SECRET", ""),
		AccessTokenTTLMin:   getInt("ACCESS_TOKEN_TTL_MIN", 15),
		RefreshTokenTTLDays: getInt("REFRESH_TOKEN_TTL_DAYS", 30),

		SMTPHost: get("SMTP_HOST", ""),
		SMTPPort: getInt("SMTP_PORT", 587),
		SMTPUser: get("SMTP_USER", ""),
		SMTPPass: get("SMTP_PASS", ""),
		SMTPFrom: get("SMTP_FROM", ""),

		AppBaseURL: get("APP_BASE_URL", "http://localhost:8080"),
		ResetPath:  get("RESET_PATH", "/reset-password"),
	}
}

func get(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return def
}
