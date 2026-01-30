package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/diff"
)

type DiffRequest struct {
	Before string `json:"before" binding:"required"`
	After  string `json:"after" binding:"required"`
}

type DiffResponse struct {
	Format  string      `json:"format"`
	Hunks   []diff.DiffHunk `json:"hunks"`
	Added   int         `json:"added_lines"`
	Removed int         `json:"removed_lines"`
	Text    string      `json:"text"`
}

func DiffMDFlow() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DiffRequest
		if err := c.BindJSON(&req); err != nil {
			slog.Error("diff request binding error", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Compute diff
		d := diff.Diff(req.Before, req.After)

		// Return both JSON + text formats
		c.JSON(http.StatusOK, DiffResponse{
			Format:  "json",
			Hunks:   d.Hunks,
			Added:   d.Added,
			Removed: d.Removed,
			Text:    diff.FormatUnified(d),
		})
	}
}
