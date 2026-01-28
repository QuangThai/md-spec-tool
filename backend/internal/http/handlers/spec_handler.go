package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/models"
	"github.com/yourorg/md-spec-tool/internal/services"
)

type SpecHandler struct {
	specService *services.SpecService
}

func NewSpecHandler(specService *services.SpecService) *SpecHandler {
	return &SpecHandler{specService: specService}
}

type SaveSpecRequest struct {
	Title      string  `json:"title" binding:"required,min=1"`
	Content    string  `json:"content" binding:"required"`
	TemplateID *string `json:"template_id"`
}

type UpdateSpecRequest struct {
	Title   string `json:"title" binding:"required,min=1"`
	Content string `json:"content" binding:"required"`
}

type SpecResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Version   int    `json:"version"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type SearchRequest struct {
	Query string `json:"query" binding:"required,min=1"`
}

func (h *SpecHandler) SaveSpec(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user_id not found in context"})
		return
	}

	var req SaveSpecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	spec := &models.Spec{
		UserID:     userID,
		Title:      req.Title,
		Content:    req.Content,
		Version:    1,
		TemplateID: nil,
	}

	spec.TemplateID = req.TemplateID

	if err := h.specService.SaveSpec(c.Request.Context(), spec); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, SpecResponse{
		ID:        spec.ID,
		Title:     spec.Title,
		Content:   spec.Content,
		Version:   spec.Version,
		CreatedAt: spec.CreatedAt.String(),
		UpdatedAt: spec.UpdatedAt.String(),
	})
}

func (h *SpecHandler) GetSpec(c *gin.Context) {
	specID := c.Param("id")
	if specID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "spec id required"})
		return
	}

	spec, err := h.specService.GetSpec(c.Request.Context(), specID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	// Verify ownership
	userID := c.GetString("user_id")
	if spec.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "not authorized to view this spec"})
		return
	}

	c.JSON(http.StatusOK, SpecResponse{
		ID:        spec.ID,
		Title:     spec.Title,
		Content:   spec.Content,
		Version:   spec.Version,
		CreatedAt: spec.CreatedAt.String(),
		UpdatedAt: spec.UpdatedAt.String(),
	})
}

func (h *SpecHandler) ListSpecs(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user_id not found in context"})
		return
	}

	specs, err := h.specService.ListSpecs(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	var responses []SpecResponse
	for _, spec := range specs {
		responses = append(responses, SpecResponse{
			ID:        spec.ID,
			Title:     spec.Title,
			Content:   spec.Content,
			Version:   spec.Version,
			CreatedAt: spec.CreatedAt.String(),
			UpdatedAt: spec.UpdatedAt.String(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"specs": responses,
		"count": len(responses),
	})
}

func (h *SpecHandler) GetVersions(c *gin.Context) {
	specID := c.Param("id")
	if specID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "spec id required"})
		return
	}

	// Verify ownership first
	spec, err := h.specService.GetSpec(c.Request.Context(), specID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "spec not found"})
		return
	}

	userID := c.GetString("user_id")
	if spec.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "not authorized"})
		return
	}

	versions, err := h.specService.GetVersions(c.Request.Context(), specID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	var responses []SpecResponse
	for _, v := range versions {
		responses = append(responses, SpecResponse{
			ID:        v.ID,
			Title:     v.Title,
			Content:   v.Content,
			Version:   v.Version,
			CreatedAt: v.CreatedAt.String(),
			UpdatedAt: v.UpdatedAt.String(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"versions": responses,
		"count":    len(responses),
	})
}

func (h *SpecHandler) UpdateSpec(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user_id not found in context"})
		return
	}

	specID := c.Param("id")
	if specID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "spec id required"})
		return
	}

	// Verify ownership
	spec, err := h.specService.GetSpec(c.Request.Context(), specID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "spec not found"})
		return
	}

	if spec.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "not authorized to update this spec"})
		return
	}

	var req UpdateSpecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	spec.Title = req.Title
	spec.Content = req.Content

	if err := h.specService.UpdateSpec(c.Request.Context(), spec); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SpecResponse{
		ID:        spec.ID,
		Title:     spec.Title,
		Content:   spec.Content,
		Version:   spec.Version,
		CreatedAt: spec.CreatedAt.String(),
		UpdatedAt: spec.UpdatedAt.String(),
	})
}

func (h *SpecHandler) DeleteSpec(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user_id not found in context"})
		return
	}

	specID := c.Param("id")
	if specID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "spec id required"})
		return
	}

	// Verify ownership
	spec, err := h.specService.GetSpec(c.Request.Context(), specID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "spec not found"})
		return
	}

	if spec.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "not authorized to delete this spec"})
		return
	}

	if err := h.specService.DeleteSpec(c.Request.Context(), specID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "spec deleted successfully"})
}

func (h *SpecHandler) SearchSpecs(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user_id not found in context"})
		return
	}

	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	specs, err := h.specService.SearchSpecs(c.Request.Context(), userID, req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	var responses []SpecResponse
	for _, spec := range specs {
		responses = append(responses, SpecResponse{
			ID:        spec.ID,
			Title:     spec.Title,
			Content:   spec.Content,
			Version:   spec.Version,
			CreatedAt: spec.CreatedAt.String(),
			UpdatedAt: spec.UpdatedAt.String(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"specs": responses,
		"count": len(responses),
	})
}

func (h *SpecHandler) DownloadSpec(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user_id not found in context"})
		return
	}

	specID := c.Param("id")
	if specID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "spec id required"})
		return
	}

	spec, err := h.specService.GetSpec(c.Request.Context(), specID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "spec not found"})
		return
	}

	if spec.UserID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "not authorized to download this spec"})
		return
	}

	// Set headers for file download
	c.Header("Content-Disposition", "attachment; filename="+spec.Title+".md")
	c.Header("Content-Type", "text/markdown")
	c.String(http.StatusOK, spec.Content)
}
