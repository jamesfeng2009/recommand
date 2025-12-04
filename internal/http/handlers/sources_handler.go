package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"recommand/internal/domain"
	"recommand/internal/repository"
)

type SourceHandler struct {
	repo *repository.SourceRepo
}

func NewSourceHandler(repo *repository.SourceRepo) *SourceHandler {
	return &SourceHandler{repo: repo}
}

// ListSources GET /api/v1/crawler/sources
func (h *SourceHandler) ListSources(c *gin.Context) {
	sources, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": sources, "total": len(sources)})
}

type CreateSourceRequest struct {
	Name             string `json:"name" binding:"required"`
	Code             string `json:"code" binding:"required"`
	BaseURL          string `json:"base_url" binding:"required"`
	Language         string `json:"language" binding:"required"`
	Category         string `json:"category" binding:"required"`
	Enabled          bool   `json:"enabled"`
	CrawlIntervalMin int    `json:"crawl_interval_minutes" binding:"gte=1"`
	MaxConcurrency   int    `json:"max_concurrency" binding:"gte=1"`
}

// CreateSource POST /api/v1/crawler/sources
func (h *SourceHandler) CreateSource(c *gin.Context) {
	var req CreateSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ns := &domain.NewsSource{
		Name:             req.Name,
		Code:             req.Code,
		BaseURL:          req.BaseURL,
		Language:         req.Language,
		Category:         req.Category,
		Enabled:          req.Enabled,
		CrawlIntervalMin: req.CrawlIntervalMin,
		MaxConcurrency:   req.MaxConcurrency,
	}

	if err := h.repo.Create(c.Request.Context(), ns); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusCreated, ns)
}

type UpdateSourceRequest struct {
	Name             string `json:"name" binding:"required"`
	Code             string `json:"code" binding:"required"`
	BaseURL          string `json:"base_url" binding:"required"`
	Language         string `json:"language" binding:"required"`
	Category         string `json:"category" binding:"required"`
	Enabled          bool   `json:"enabled"`
	CrawlIntervalMin int    `json:"crawl_interval_minutes" binding:"gte=1"`
	MaxConcurrency   int    `json:"max_concurrency" binding:"gte=1"`
}

// UpdateSource PUT /api/v1/crawler/sources/:id
func (h *SourceHandler) UpdateSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	existing, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}

	var req UpdateSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing.Name = req.Name
	existing.Code = req.Code
	existing.BaseURL = req.BaseURL
	existing.Language = req.Language
	existing.Category = req.Category
	existing.Enabled = req.Enabled
	existing.CrawlIntervalMin = req.CrawlIntervalMin
	existing.MaxConcurrency = req.MaxConcurrency

	if err := h.repo.Update(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, existing)
}

type UpdateSourceStatusRequest struct {
	Enabled bool `json:"enabled"`
}

// UpdateSourceStatus PUT /api/v1/crawler/sources/:id/status
func (h *SourceHandler) UpdateSourceStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateSourceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateEnabled(c.Request.Context(), id, req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.Status(http.StatusNoContent)
}
