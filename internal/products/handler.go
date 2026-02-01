package products

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ecommerce/internal/auth"
)

type Handler struct {
	repo *Repo
}

func NewHandler(repo *Repo) *Handler {
	return &Handler{repo: repo}
}

// Public: list products (optional category=slug)
func (h *Handler) ListPublic(c *gin.Context) {
	var cat *string
	if v := c.Query("category"); v != "" {
		cat = &v
	}

	items, err := h.repo.ListPublic(c.Request.Context(), cat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list products"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// Public: product details with variants
func (h *Handler) GetPublic(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	p, err := h.repo.GetProductPublic(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

type CreateProductReq struct {
	CategoryID  int64  `json:"category_id" binding:"required"`
	TypeName    string `json:"type_name" binding:"required"` // e.g. "Formal Shirt"
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`

	Variants []CreateVariantReq `json:"variants" binding:"required"`
}

type CreateVariantReq struct {
	Size            string  `json:"size" binding:"required"` // e.g. M, L, XL
	Color           string  `json:"color" binding:"required"` // e.g. Red, Black
	Price           float64 `json:"price" binding:"required"`
	DiscountPercent int     `json:"discount_percent"` // 0-100
	StockQty        int     `json:"stock_qty" binding:"required"`
}

// Admin: create product + variants
func (h *Handler) AdminCreate(c *gin.Context) {
	var req CreateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userIDAny, _ := c.Get(auth.CtxUserIDKey)
	userID := userIDAny.(int64)

	var vars []CreateVariantInput
	for _, v := range req.Variants {
		vars = append(vars, CreateVariantInput{
			Size:            v.Size,
			Color:           v.Color,
			Price:           v.Price,
			DiscountPercent: v.DiscountPercent,
			StockQty:        v.StockQty,
		})
	}

	p, err := h.repo.CreateProductWithVariants(c.Request.Context(), CreateProductInput{
		CategoryID:  req.CategoryID,
		TypeName:    req.TypeName,
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
		Variants:    vars,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, p)
}
