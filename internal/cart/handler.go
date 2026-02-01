package cart

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ecommerce/internal/auth"
)

type Handler struct {
	repo *Repo
}

func NewHandler(repo *Repo) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) GetMyCart(c *gin.Context) {
	userIDAny, _ := c.Get(auth.CtxUserIDKey)
	userID := userIDAny.(int64)

	crt, err := h.repo.GetCart(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load cart"})
		return
	}
	c.JSON(http.StatusOK, crt)
}

type AddItemReq struct {
	VariantID int64 `json:"variant_id" binding:"required"`
	Qty       int   `json:"qty" binding:"required"`
}

func (h *Handler) AddItem(c *gin.Context) {
	userIDAny, _ := c.Get(auth.CtxUserIDKey)
	userID := userIDAny.(int64)

	var req AddItemReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Qty <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.repo.AddItem(c.Request.Context(), userID, req.VariantID, req.Qty); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to add item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type UpdateQtyReq struct {
	VariantID int64 `json:"variant_id" binding:"required"`
	Qty       int   `json:"qty" binding:"required"`
}

func (h *Handler) UpdateQty(c *gin.Context) {
	userIDAny, _ := c.Get(auth.CtxUserIDKey)
	userID := userIDAny.(int64)

	var req UpdateQtyReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Qty <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.repo.UpdateQty(c.Request.Context(), userID, req.VariantID, req.Qty); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to update qty"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type RemoveItemReq struct {
	VariantID int64 `json:"variant_id" binding:"required"`
}

func (h *Handler) RemoveItem(c *gin.Context) {
	userIDAny, _ := c.Get(auth.CtxUserIDKey)
	userID := userIDAny.(int64)

	var req RemoveItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.repo.RemoveItem(c.Request.Context(), userID, req.VariantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to remove item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
