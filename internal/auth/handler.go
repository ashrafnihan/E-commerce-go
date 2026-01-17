package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ecommerce/internal/config"
	"ecommerce/internal/domain/user"
	"ecommerce/internal/mail"
	"ecommerce/internal/util"
)

const (
	OTPPurposeVerifyEmail   = "verify_email"
	OTPPurposeResetPassword = "reset_password"
)

type Dependencies struct {
	Cfg     config.Config
	JWT     *JWTManager
	Users   *UserRepo
	Refresh *RefreshRepo
	Resets  *ResetRepo // (kept for compatibility; not used in OTP flow)
	OTP     *OTPRepo
	Mailer  mail.Mailer
}

type Handler struct {
	deps Dependencies
}

func NewHandler(d Dependencies) *Handler {
	return &Handler{deps: d}
}

type registerReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type verifyOTPReq struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

type forgotReq struct {
	Email string `json:"email" binding:"required,email"`
}

type resetWithOTPReq struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *Handler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	pwHash, err := HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password hash failed"})
		return
	}

	// Create user (default role = user). EmailVerified should be false by default.
	u, err := h.deps.Users.Create(c.Request.Context(), req.Email, pwHash, "user")
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}

	otp, exp, err := h.issueOTP(c, u.ID, OTPPurposeVerifyEmail)
	if err == nil {
		_ = h.sendOTPEmail(u.Email, otp, exp, "Verify your email")
	}

	// Do NOT login user until verified (common website behavior)
	c.JSON(http.StatusCreated, gin.H{
		"ok":      true,
		"message": "Account created. OTP sent to your email for verification.",
	})
}

func (h *Handler) ResendVerifyOTP(c *gin.Context) {
	var req forgotReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	u, err := h.deps.Users.ByEmail(c.Request.Context(), req.Email)
	if err != nil || !u.IsActive {
		// privacy: always ok
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	if u.EmailVerified {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	otp, exp, err := h.issueOTP(c, u.ID, OTPPurposeVerifyEmail)
	if err == nil {
		_ = h.sendOTPEmail(u.Email, otp, exp, "Verify your email")
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) VerifyEmailOTP(c *gin.Context) {
	var req verifyOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	u, err := h.deps.Users.ByEmail(c.Request.Context(), req.Email)
	if err != nil || !u.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if u.EmailVerified {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	ok, err := h.verifyOTP(c, u.ID, OTPPurposeVerifyEmail, req.OTP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired otp"})
		return
	}

	_ = h.deps.Users.SetEmailVerified(c.Request.Context(), u.ID)
	_ = h.deps.OTP.Delete(c.Request.Context(), u.ID, OTPPurposeVerifyEmail)

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "Email verified successfully. Now you can login.",
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	u, err := h.deps.Users.ByEmail(c.Request.Context(), req.Email)
	if err != nil || !u.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// block login if not verified
	if !u.EmailVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "email not verified"})
		return
	}

	if !CheckPassword(u.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	access, accessExp, _ := h.deps.JWT.SignAccess(u.ID, u.Role)
	refresh, refreshExp, _ := h.deps.JWT.SignRefresh(u.ID, u.Role)
	_ = h.deps.Refresh.Store(c.Request.Context(), u.ID, HashToken(refresh), refreshExp)

	c.JSON(http.StatusOK, gin.H{
		"user":          sanitizeUser(u),
		"access_token":  access,
		"access_exp":    accessExp,
		"refresh_token": refresh,
		"refresh_exp":   refreshExp,
	})
}

// Rotate refresh token
func (h *Handler) Refresh(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := h.deps.JWT.ParseRefresh(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	ok, err := h.deps.Refresh.IsValid(c.Request.Context(), claims.UserID, HashToken(req.RefreshToken))
	if err != nil || !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token expired or revoked"})
		return
	}

	_ = h.deps.Refresh.Revoke(c.Request.Context(), claims.UserID, HashToken(req.RefreshToken))

	access, accessExp, _ := h.deps.JWT.SignAccess(claims.UserID, claims.Role)
	newRefresh, refreshExp, _ := h.deps.JWT.SignRefresh(claims.UserID, claims.Role)
	_ = h.deps.Refresh.Store(c.Request.Context(), claims.UserID, HashToken(newRefresh), refreshExp)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"access_exp":    accessExp,
		"refresh_token": newRefresh,
		"refresh_exp":   refreshExp,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claims, err := h.deps.JWT.ParseRefresh(req.RefreshToken)
	if err == nil {
		_ = h.deps.Refresh.Revoke(c.Request.Context(), claims.UserID, HashToken(req.RefreshToken))
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Forgot password -> send OTP (privacy-safe)
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req forgotReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	u, err := h.deps.Users.ByEmail(c.Request.Context(), req.Email)
	if err != nil || !u.IsActive {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	otp, exp, err := h.issueOTP(c, u.ID, OTPPurposeResetPassword)
	if err == nil {
		_ = h.sendOTPEmail(u.Email, otp, exp, "Reset password OTP")
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Reset password using email + otp + new_password
func (h *Handler) ResetPassword(c *gin.Context) {
	var req resetWithOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	u, err := h.deps.Users.ByEmail(c.Request.Context(), req.Email)
	if err != nil || !u.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	ok, err := h.verifyOTP(c, u.ID, OTPPurposeResetPassword, req.OTP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired otp"})
		return
	}

	newHash, err := HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password hash failed"})
		return
	}
	if err := h.deps.Users.UpdatePassword(c.Request.Context(), u.ID, newHash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password update failed"})
		return
	}

	_ = h.deps.OTP.Delete(c.Request.Context(), u.ID, OTPPurposeResetPassword)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) Me(c *gin.Context) {
	uidAny, _ := c.Get(CtxUserIDKey)
	uid, _ := uidAny.(int64)

	u, err := h.deps.Users.ByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, sanitizeUser(u))
}

func sanitizeUser(u user.User) gin.H {
	return gin.H{
		"id":             u.ID,
		"email":          u.Email,
		"role":           u.Role,
		"is_active":      u.IsActive,
		"email_verified": u.EmailVerified,
		"created_at":     u.CreatedAt,
		"updated_at":     u.UpdatedAt,
	}
}

func (h *Handler) issueOTP(c *gin.Context, userID int64, purpose string) (string, time.Time, error) {
	otp, err := util.GenerateOTP6()
	if err != nil {
		return "", time.Time{}, err
	}

	ttl := h.deps.Cfg.OTPTTLMin
	if ttl <= 0 {
		ttl = 10
	}
	expiresAt := time.Now().Add(time.Duration(ttl) * time.Minute)

	if h.deps.OTP == nil {
		return "", time.Time{}, errors.New("otp repo not configured")
	}
	if err := h.deps.OTP.Upsert(c.Request.Context(), userID, purpose, HashToken(otp), expiresAt); err != nil {
		return "", time.Time{}, err
	}
	return otp, expiresAt, nil
}

func (h *Handler) verifyOTP(c *gin.Context, userID int64, purpose string, otp string) (bool, error) {
	if h.deps.OTP == nil {
		return false, errors.New("otp repo not configured")
	}
	hash, exp, ok, _ := h.deps.OTP.GetValid(c.Request.Context(), userID, purpose)
	if !ok || time.Now().After(exp) {
		return false, nil
	}
	return HashToken(otp) == hash, nil
}

func (h *Handler) sendMailSafe(to, subject, body string) error {
	// website style: do not fail request if mail fails
	return h.deps.Mailer.Send(to, subject, body)
}

func (h *Handler) sendOTPEmail(to, otp string, exp time.Time, subject string) error {
	body := "Your OTP code is: " + otp + "\n\n" +
		"It expires at: " + exp.Format(time.RFC1123) + "\n\n" +
		"If you didn't request this, ignore this email."
	return h.sendMailSafe(to, subject, body)
}
