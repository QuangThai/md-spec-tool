package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/models"
	"github.com/yourorg/md-spec-tool/internal/services"
)

type ShareHandler struct {
	shareService *services.ShareService
}

func NewShareHandler(shareService *services.ShareService) *ShareHandler {
	return &ShareHandler{shareService: shareService}
}

// ShareSpec shares a spec with another user
// POST /spec/:id/share
func (h *ShareHandler) ShareSpec(c *gin.Context) {
	specID := c.Param("id")
	userID := c.GetString("user_id")

	var req models.ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	share, err := h.shareService.ShareSpec(c.Request.Context(), specID, userID, req.SharedWithUserID, req.PermissionLevel)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.ShareResponse{
		Message: "spec shared successfully",
		Share:   share,
	})
}

// UnshareSpec removes a share
// DELETE /spec/:id/share/:user_id
func (h *ShareHandler) UnshareSpec(c *gin.Context) {
	specID := c.Param("id")
	sharedWithUserID := c.Param("user_id")
	ownerID := c.GetString("user_id")

	err := h.shareService.UnshareSpec(c.Request.Context(), specID, ownerID, sharedWithUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "share removed successfully"})
}

// GetSpecShares returns all shares for a spec
// GET /spec/:id/shares
func (h *ShareHandler) GetSpecShares(c *gin.Context) {
	specID := c.Param("id")
	ownerID := c.GetString("user_id")

	shares, err := h.shareService.GetSpecShares(c.Request.Context(), specID, ownerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if shares == nil {
		shares = []*models.Share{}
	}

	c.JSON(http.StatusOK, models.ShareResponse{
		Shares: shares,
	})
}

// UpdateSharePermission updates share permission level
// PUT /spec/:id/share/:user_id
func (h *ShareHandler) UpdateSharePermission(c *gin.Context) {
	specID := c.Param("id")
	sharedWithUserID := c.Param("user_id")
	ownerID := c.GetString("user_id")

	var req struct {
		PermissionLevel string `json:"permission_level" binding:"required,oneof=view edit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	err := h.shareService.UpdateSharePermission(c.Request.Context(), specID, ownerID, sharedWithUserID, req.PermissionLevel)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission updated successfully"})
}

// GetSharedSpecs returns all specs shared with the user
// GET /spec/shared/mine
func (h *ShareHandler) GetSharedSpecs(c *gin.Context) {
	userID := c.GetString("user_id")

	specs, err := h.shareService.GetSharedSpecs(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if specs == nil {
		specs = []*models.Spec{}
	}

	c.JSON(http.StatusOK, gin.H{
		"specs": specs,
		"count": len(specs),
	})
}
