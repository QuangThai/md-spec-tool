package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/ai"
	"github.com/yourorg/md-spec-tool/internal/converter"
	"github.com/yourorg/md-spec-tool/internal/suggest"
)

// AISuggestRequest represents the request for AI suggestions
type AISuggestRequest struct {
	PasteText string `json:"paste_text" binding:"required"`
	Template  string `json:"template"`
}

// AISuggestResponse represents the AI suggestions response
type AISuggestResponse struct {
	Suggestions     []suggest.AISuggestion `json:"suggestions"`
	Error           string                 `json:"error,omitempty"`
	Configured      bool                   `json:"configured"`
	AIModel         string                 `json:"ai_model,omitempty"`
	AIPromptVersion string                 `json:"ai_prompt_version,omitempty"`
}

// GetAISuggestions handles POST /api/mdflow/ai/suggest
// Analyzes spec content and returns AI-powered improvement suggestions
// Supports BYOK: users can provide their own OpenAI key via X-OpenAI-API-Key header
func (h *MDFlowHandler) GetAISuggestions(c *gin.Context) {
	// Get suggester for this request (BYOK-aware)
	suggester := h.getSuggesterForRequest(c)
	aiService := h.getAIServiceForRequest(c)
	aiModel := ""
	if aiService != nil {
		aiModel = aiService.GetModel()
	}

	if suggester == nil || !suggester.IsConfigured() {
		c.JSON(http.StatusOK, AISuggestResponse{
			Suggestions:     []suggest.AISuggestion{},
			Error:           "OpenAI API key not configured. Add your key in Studio settings.",
			Configured:      false,
			AIModel:         aiModel,
			AIPromptVersion: ai.PromptVersionSuggestions,
		})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.cfg.MaxPasteBytes+4<<10)

	var req AISuggestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "request body exceeds limit"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "paste_text is required"})
		return
	}

	// Validate input size
	if int64(len(req.PasteText)) > h.cfg.MaxPasteBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: fmt.Sprintf("paste_text exceeds %s limit", humanSize(h.cfg.MaxPasteBytes))})
		return
	}

	template, err := normalizeTemplate(req.Template)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if template == "" {
		template = "spec"
	}

	// Parse the paste text into a SpecDoc
	specDoc, err := converter.BuildSpecDocFromPaste(req.PasteText)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to parse input: " + err.Error()})
		return
	}

	// Call AI suggester
	suggestReq := &suggest.SuggestionRequest{
		SpecDoc:  specDoc,
		Template: template,
	}

	resp, err := suggester.GetSuggestions(c.Request.Context(), suggestReq)
	if err != nil {
		slog.Error("AI suggestion error", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get AI suggestions"})
		return
	}

	if resp.Error != "" {
		c.JSON(http.StatusOK, AISuggestResponse{
			Suggestions:     []suggest.AISuggestion{},
			Error:           resp.Error,
			Configured:      true,
			AIModel:         aiModel,
			AIPromptVersion: ai.PromptVersionSuggestions,
		})
		return
	}

	c.JSON(http.StatusOK, AISuggestResponse{
		Suggestions:     resp.Suggestions,
		Configured:      true,
		AIModel:         aiModel,
		AIPromptVersion: ai.PromptVersionSuggestions,
	})
}
