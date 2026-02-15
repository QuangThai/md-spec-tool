package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/ai"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

type fakeAIService struct{}

func (fakeAIService) MapColumns(context.Context, ai.MapColumnsRequest) (*ai.ColumnMappingResult, error) {
	return nil, nil
}

func (fakeAIService) AnalyzePaste(context.Context, ai.AnalyzePasteRequest) (*ai.PasteAnalysis, error) {
	return nil, nil
}

func (fakeAIService) GetSuggestions(context.Context, ai.SuggestionsRequest) (*ai.SuggestionsResult, error) {
	return nil, nil
}

func (fakeAIService) SummarizeDiff(context.Context, ai.SummarizeDiffRequest) (*ai.DiffSummary, error) {
	return nil, nil
}

func (fakeAIService) ValidateSemantic(context.Context, ai.SemanticValidationRequest) (*ai.SemanticValidationResult, error) {
	return nil, nil
}

func (fakeAIService) GetMode() string { return "on" }

func TestGetConverterForRequest_BYOKServiceUnavailable_DoesNotFallbackToBaseAI(t *testing.T) {
	cfg := config.LoadConfig()
	provider := NewAIServiceProvider(cfg)

	baseConverter := converter.NewConverter().WithAIService(fakeAIService{})
	if !baseConverter.HasAIService() {
		t.Fatal("expected base converter to have AI service")
	}

	provider.byokCache = nil

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/test", nil)
	c.Request.Header.Set(BYOKHeader, "user-byok-key")

	conv := provider.GetConverterForRequest(c, baseConverter)
	if conv == nil {
		t.Fatal("expected converter")
	}
	if conv == baseConverter {
		t.Fatal("expected BYOK path not to reuse base converter when BYOK service is unavailable")
	}
	if conv.HasAIService() {
		t.Fatal("expected converter without AI when BYOK service is unavailable")
	}
}
