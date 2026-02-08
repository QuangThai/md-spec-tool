package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/ai"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

// ValidateRequest represents the request for validation with custom rules
type ValidateRequest struct {
	PasteText       string                     `json:"paste_text" binding:"required"`
	ValidationRules *converter.ValidationRules `json:"validation_rules"`
	Template        string                     `json:"template"`
}

// ValidateResponse represents the validation response with optional AI results
type ValidateResponse struct {
	Valid          bool                         `json:"valid"`
	Warnings       []converter.Warning          `json:"warnings"`
	SemanticResult *ai.SemanticValidationResult `json:"semantic_result,omitempty"`
}

// Validate handles POST /api/mdflow/validate
// Builds SpecDoc from paste_text and runs custom validation rules
// Automatically runs AI semantic validation when AI service is available
func (h *MDFlowHandler) Validate(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.cfg.MaxPasteBytes+4<<10)

	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "paste_text is required"})
		return
	}

	if int64(len(req.PasteText)) > h.cfg.MaxPasteBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: fmt.Sprintf("paste_text exceeds %s limit", humanSize(h.cfg.MaxPasteBytes))})
		return
	}

	specDoc, err := converter.BuildSpecDocFromPaste(req.PasteText)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to parse input: " + err.Error()})
		return
	}

	rules := req.ValidationRules
	if rules == nil {
		rules = &converter.ValidationRules{}
	}

	result := converter.Validate(specDoc, rules)

	resp := ValidateResponse{
		Valid:    result.Valid,
		Warnings: result.Warnings,
	}

	// Auto-run AI semantic validation when AI service is available (BYOK-aware)
	aiService := h.getAIServiceForRequest(c)
	if aiService != nil {
		semanticResult, err := aiService.ValidateSemantic(c.Request.Context(), ai.SemanticValidationRequest{
			SpecContent: req.PasteText,
			Template:    req.Template,
		})
		if err != nil {
			slog.Warn("validate AI semantic failed", "error", err)
		} else {
			resp.SemanticResult = semanticResult
			slog.Info("validate AI semantic completed", "overall", semanticResult.Overall, "score", semanticResult.Score)
		}
	}

	c.JSON(http.StatusOK, resp)
}
