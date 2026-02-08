package handlers

import (
	"log/slog"
	"net/http"
	"strings"

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
	aiService ai.Service
	cfg       *config.Config
}

func NewDiffHandler(aiService ai.Service, cfg *config.Config) *DiffHandler {
	return &DiffHandler{aiService: aiService, cfg: cfg}
}

// getAIServiceForRequest returns an AI service for the current request (BYOK-aware).
func (h *DiffHandler) getAIServiceForRequest(c *gin.Context) ai.Service {
	userKey := strings.TrimSpace(c.GetHeader(BYOKHeader))
	if userKey == "" {
		return h.aiService
	}

	aiCfg := ai.DefaultConfig()
	aiCfg.APIKey = userKey
	if h.cfg != nil {
		aiCfg.Model = h.cfg.OpenAIModel
		aiCfg.RequestTimeout = h.cfg.AIRequestTimeout
		aiCfg.MaxRetries = h.cfg.AIMaxRetries
		aiCfg.RetryBaseDelay = h.cfg.AIRetryBaseDelay
	}

	service, err := ai.NewService(aiCfg)
	if err != nil {
		slog.Warn("BYOK: diff handler failed to create AI service", "error", err)
		return h.aiService
	}
	return service
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
	aiService := h.getAIServiceForRequest(c)
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
	handler := &DiffHandler{aiService: nil}
	return handler.DiffMDFlow
}
