package auth

import (
	
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ecommerce/internal/config"
	"ecommerce/internal/domain/user"
	"ecommerce/internal/mail"
	"ecommerce/internal/util"
)

type Dependencies struct {
	Cfg     config.Config
	JWT     *JWTManager
	Users   *UserRepo
	Refresh *RefreshRepo
	Resets  *ResetRepo
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
	// role is not allowed from normal registration (website-style)
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type forgotReq struct {
	Email string `json:"email" binding:"required,email"`
}

type resetReq struct {
	Token       string `json:"token" binding:"required"`
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

	// Default role for website signups
	u, err := h.deps.Users.Create(c.Request.Context(), req.Email, pwHash, "user")
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}

	access, accessExp, _ := h.deps.JWT.SignAccess(u.ID, u.Role)
	refresh, refreshExp, _ := h.deps.JWT.SignRefresh(u.ID, u.Role)

	_ = h.deps.Refresh.Store(c.Request.Context(), u.ID, HashToken(refresh), refreshExp)

	c.JSON(http.StatusCreated, gin.H{
		"user":          sanitizeUser(u),
		"access_token":  access,
		"access_exp":    accessExp,
		"refresh_token": refresh,
		"refresh_exp":   refreshExp,
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

// Rotate refresh token (very common approach)
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

	// revoke old refresh, issue new pair
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
	// website style logout = revoke refresh token (client should delete tokens too)
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

// IMPORTANT: for privacy, always return ok even if email doesn't exist (like real websites).
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

	// token valid for 30 minutes
	rawToken, err := util.RandomToken(32)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	tokenHash := HashToken(rawToken)
	expiresAt := time.Now().Add(30 * time.Minute)

	_ = h.deps.Resets.Create(c.Request.Context(), u.ID, tokenHash, expiresAt)

	resetLink := h.deps.Cfg.AppBaseURL + h.deps.Cfg.ResetPath + "?token=" + rawToken

	body := "We received a request to reset your password.\n\n" +
		"Reset link (valid for 30 minutes):\n" + resetLink + "\n\n" +
		"If you didn't request this, you can ignore this email."

	_ = h.sendMailSafe(u.Email, "Reset your password", body)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req resetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenHash := HashToken(req.Token)

	userID, ok, err := h.deps.Resets.Consume(c.Request.Context(), tokenHash)
	if err != nil || !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}

	newHash, err := HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password hash failed"})
		return
	}
	if err := h.deps.Users.UpdatePassword(c.Request.Context(), userID, newHash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password update failed"})
		return
	}

	// optional but recommended: revoke ALL refresh tokens for this user on password reset
	// (simple approach: revoke by user_id; omitted here for brevity, but easy to add)

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
		"id":         u.ID,
		"email":      u.Email,
		"role":       u.Role,
		"is_active":  u.IsActive,
		"created_at": u.CreatedAt,
		"updated_at": u.UpdatedAt,
	}
}

// func (h *Handler) sendMailSafe(to, subject, body string) error {
// 	// don’t fail the request if mail fails (website style),
// 	// but log it in real apps.
// 	return h.deps.Mailer.Send(to, subject, body)
// }
func (h *Handler) sendMailSafe(to, subject, body string) error {
    // don’t fail the request if mail fails (website style),
	// but log it in real apps.
	return h.deps.Mailer.Send(to, subject, body)
}
