package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"ecommerce/internal/auth"
	"ecommerce/internal/config"
	"ecommerce/internal/db"
	"ecommerce/internal/mail"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	pool, err := db.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	mailer := mail.NewSMTPMailer(mail.SMTPConfig{
		Host: cfg.SMTPHost,
		Port: cfg.SMTPPort,
		User: cfg.SMTPUser,
		Pass: cfg.SMTPPass,
		From: cfg.SMTPFrom,
	})

	jwtMgr := auth.NewJWTManager(auth.JWTConfig{
		Issuer:         cfg.JWTIssuer,
		AccessSecret:   cfg.JWTAccessSecret,
		RefreshSecret:  cfg.JWTRefreshSecret,
		AccessTTLMin:   cfg.AccessTokenTTLMin,
		RefreshTTLDays: cfg.RefreshTokenTTLDays,
	})

	// Repos
	userRepo := auth.NewUserRepo(pool)
	refreshRepo := auth.NewRefreshRepo(pool)
	resetRepo := auth.NewResetRepo(pool) // kept (not used in OTP reset flow, but fine)
	otpRepo := auth.NewOTPRepo(pool)

	// Handler with OTP dependency
	h := auth.NewHandler(auth.Dependencies{
		Cfg:     cfg,
		JWT:     jwtMgr,
		Users:   userRepo,
		Refresh: refreshRepo,
		Resets:  resetRepo,
		OTP:     otpRepo,
		Mailer:  mailer,
	})

	r := gin.Default()

	// Auth routes
	api := r.Group("/api")
	authGroup := api.Group("/auth")
	{
		// Register + Email Verification OTP
		authGroup.POST("/register", h.Register)
		authGroup.POST("/verify-email", h.VerifyEmailOTP)
		authGroup.POST("/resend-verify", h.ResendVerifyOTP)

		// Login / Refresh / Logout
		authGroup.POST("/login", h.Login)
		authGroup.POST("/refresh", h.Refresh)
		authGroup.POST("/logout", h.Logout)

		// Forgot/Reset password using OTP
		authGroup.POST("/forgot-password", h.ForgotPassword)
		authGroup.POST("/reset-password", h.ResetPassword)
	}

	// Protected example routes
	protected := api.Group("/")
	protected.Use(auth.AuthMiddleware(jwtMgr))
	{
		protected.GET("/me", h.Me)

		adminOnly := protected.Group("/admin")
		adminOnly.Use(auth.RequireRole("admin"))
		adminOnly.GET("/dashboard", func(c *gin.Context) {
			c.JSON(200, gin.H{"ok": true, "message": "admin access granted"})
		})
	}

	log.Printf("listening on %s", cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
