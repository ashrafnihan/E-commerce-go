package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"ecommerce/internal/auth"
	"ecommerce/internal/cart"
	"ecommerce/internal/categories"
	"ecommerce/internal/config"
	"ecommerce/internal/db"
	"ecommerce/internal/mail"
	"ecommerce/internal/products"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	gormDB, err := db.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	sqlDB, _ := gormDB.DB()
	defer sqlDB.Close()

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

	// Repos (all using GORM now)
	userRepo := auth.NewUserRepo(gormDB)
	refreshRepo := auth.NewRefreshRepo(gormDB)
	resetRepo := auth.NewResetRepo(gormDB)
	otpRepo := auth.NewOTPRepo(gormDB)

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

	// Catalog repos/handlers (GORM)
	catRepo := categories.NewRepo(gormDB)
	catHandler := categories.NewHandler(catRepo)

	prodRepo := products.NewRepo(gormDB)
	prodHandler := products.NewHandler(prodRepo)

	cartRepo := cart.NewRepo(gormDB)
	cartHandler := cart.NewHandler(cartRepo)

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

	// Public catalog routes (no login required)
	api.GET("/categories", catHandler.ListPublic)
	api.GET("/products", prodHandler.ListPublic)
	api.GET("/products/:id", prodHandler.GetPublic)

	// Protected routes
	protected := api.Group("/")
	protected.Use(auth.AuthMiddleware(jwtMgr))
	{
		protected.GET("/me", h.Me)

		// Cart (user must login)
		protected.GET("/cart", cartHandler.GetMyCart)
		protected.POST("/cart/items", cartHandler.AddItem)
		protected.PATCH("/cart/items", cartHandler.UpdateQty)
		protected.DELETE("/cart/items", cartHandler.RemoveItem)

		adminOnly := protected.Group("/admin")
		adminOnly.Use(auth.RequireRole("admin"))

		adminOnly.GET("/dashboard", func(c *gin.Context) {
			c.JSON(200, gin.H{"ok": true, "message": "admin access granted"})
		})

		// Admin category CRUD
		adminOnly.GET("/categories", catHandler.AdminList)
		adminOnly.POST("/categories", catHandler.AdminCreate)
		adminOnly.PATCH("/categories/:id", catHandler.AdminUpdate)

		// Admin add product
		adminOnly.POST("/products", prodHandler.AdminCreate)
	}

	log.Printf("listening on %s", cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
