package categories

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

func (h *Handler) ListPublic(c *gin.Context) {
	items, err := h.repo.ListActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list categories"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) AdminList(c *gin.Context) {
	items, err := h.repo.AdminListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list categories"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

type CreateCategoryReq struct {
	Name      string `json:"name" binding:"required"`
	SortOrder int    `json:"sort_order"`
}

func (h *Handler) AdminCreate(c *gin.Context) {
	// admin is already enforced by middleware
	_ = c.GetString(auth.CtxRoleKey)

	var req CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	created, err := h.repo.Create(c.Request.Context(), req.Name, req.SortOrder)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create (slug may be duplicate)"})
		return
	}
	c.JSON(http.StatusCreated, created)
}

type UpdateCategoryReq struct {
	Name      *string `json:"name"`
	SortOrder *int    `json:"sort_order"`
	IsActive  *bool   `json:"is_active"`
}

func (h *Handler) AdminUpdate(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req UpdateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	updated, err := h.repo.Update(c.Request.Context(), id, req.Name, req.SortOrder, req.IsActive)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to update"})
		return
	}
	c.JSON(http.StatusOK, updated)
}
