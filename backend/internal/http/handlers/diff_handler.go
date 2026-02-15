package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/ai"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/diff"
)

type DiffRequest struct {
	Before string `json:"before" binding:"required"`
	After  string `json:"after" binding:"required"`
}

type DiffResponse struct {
	Format  string          `json:"format"`
	Hunks   []diff.DiffHunk `json:"hunks"`
	Added   int             `json:"added_lines"`
	Removed int             `json:"removed_lines"`
	Text    string          `json:"text"`
	Summary *ai.DiffSummary `json:"summary,omitempty"` // AI-generated summary
}

type DiffHandler struct {
	provider *AIServiceProvider
	cfg      *config.Config
}

func NewDiffHandler(provider *AIServiceProvider, cfg *config.Config) *DiffHandler {
	if provider == nil {
		provider = NewAIServiceProvider(cfg)
	}
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	return &DiffHandler{
		provider: provider,
		cfg:      cfg,
	}
}

func (h *DiffHandler) DiffMDFlow(c *gin.Context) {
	var req DiffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("diff request binding error", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
		return
	}

	// Compute diff
	d := diff.Diff(req.Before, req.After)
	diffText := diff.FormatUnified(d)

	resp := DiffResponse{
		Format:  "json",
		Hunks:   d.Hunks,
		Added:   d.Added,
		Removed: d.Removed,
		Text:    diffText,
	}

	// Auto-generate AI summary when AI service is available (BYOK-aware)
	aiService := h.provider.GetAIServiceForRequest(c)
	if aiService != nil {
		summary, err := aiService.SummarizeDiff(c.Request.Context(), ai.SummarizeDiffRequest{
			Before:   req.Before,
			After:    req.After,
			DiffText: diffText,
		})
		if err != nil {
			slog.Warn("diff AI summary failed", "error", err)
		} else {
			resp.Summary = summary
			slog.Info("diff AI summary generated", "confidence", summary.Confidence)
		}
	}

	c.JSON(http.StatusOK, resp)
}

// DiffMDFlow returns a handler function for backwards compatibility
func DiffMDFlow() gin.HandlerFunc {
	handler := NewDiffHandler(nil, nil)
	return handler.DiffMDFlow
}
