package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/models"
	"github.com/yourorg/md-spec-tool/internal/services"
)

type ActivityHandler struct {
	activityService *services.ActivityService
}

func NewActivityHandler(activityService *services.ActivityService) *ActivityHandler {
	return &ActivityHandler{activityService: activityService}
}

// GetSpecActivity retrieves activity log for a spec
// GET /spec/:id/activity
func (h *ActivityHandler) GetSpecActivity(c *gin.Context) {
	specID := c.Param("id")

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	logs, total, err := h.activityService.GetSpecActivity(c.Request.Context(), specID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if logs == nil {
		logs = []*models.ActivityLog{}
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetUserActivity retrieves activity for current user
// GET /activity/mine
func (h *ActivityHandler) GetUserActivity(c *gin.Context) {
	userID := c.GetString("user_id")

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	logs, total, err := h.activityService.GetUserActivity(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if logs == nil {
		logs = []*models.ActivityLog{}
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// PHASE 7: Activity filtering and export endpoints

// FilterActivities retrieves filtered activity logs
// POST /spec/:id/activities/filter
func (h *ActivityHandler) FilterActivities(c *gin.Context) {
	specID := c.Param("id")

	var req models.ActivityFilterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Ensure spec_id matches path parameter
	req.SpecID = specID

	// Validate limit
	if req.Limit == 0 {
		req.Limit = 50
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	logs, total, err := h.activityService.GetFilteredActivities(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if logs == nil {
		logs = []*models.ActivityLog{}
	}

	c.JSON(http.StatusOK, models.ActivityFilterResponse{
		Data: logs,
		Pagination: models.PaginationInfo{
			Limit:  req.Limit,
			Offset: req.Offset,
			Total:  total,
		},
	})
}

// GetActivityStats retrieves statistics for activities on a spec
// GET /spec/:id/activities/stats
func (h *ActivityHandler) GetActivityStats(c *gin.Context) {
	specID := c.Param("id")

	stats, err := h.activityService.GetActivityStats(c.Request.Context(), specID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ExportActivities exports activities in specified format
// POST /spec/:id/activities/export
func (h *ActivityHandler) ExportActivities(c *gin.Context) {
	specID := c.Param("id")

	var req models.ActivityExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Ensure spec_id matches
	req.SpecID = specID
	req.Filters.SpecID = specID

	// Validate format
	if req.Format != "json" && req.Format != "csv" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format must be 'json' or 'csv'"})
		return
	}

	// Set default limit if not specified
	if req.Filters.Limit == 0 {
		req.Filters.Limit = 1000
	}

	data, mimeType, err := h.activityService.ExportActivities(c.Request.Context(), &req.Filters, req.Format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set response headers for file download
	c.Header("Content-Type", mimeType)
	c.Header("Content-Disposition", "attachment; filename=\"activities."+req.Format+"\"")
	c.Data(http.StatusOK, mimeType, data)
}
