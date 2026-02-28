package converter

import (
	"context"
	"testing"

	"github.com/yourorg/md-spec-tool/internal/ai"
)

// mockUnavailableAI simulates circuit breaker open (ErrAIUnavailable)
type mockUnavailableAI struct{}

func (m *mockUnavailableAI) MapColumns(ctx context.Context, req ai.MapColumnsRequest) (*ai.ColumnMappingResult, error) {
	return nil, &ai.AIError{Err: ai.ErrAIUnavailable, Message: "circuit breaker open"}
}
func (m *mockUnavailableAI) AnalyzePaste(ctx context.Context, req ai.AnalyzePasteRequest) (*ai.PasteAnalysis, error) {
	return nil, &ai.AIError{Err: ai.ErrAIUnavailable, Message: "circuit breaker open"}
}
func (m *mockUnavailableAI) GetSuggestions(ctx context.Context, req ai.SuggestionsRequest) (*ai.SuggestionsResult, error) {
	return nil, &ai.AIError{Err: ai.ErrAIUnavailable, Message: "circuit breaker open"}
}
func (m *mockUnavailableAI) SummarizeDiff(ctx context.Context, req ai.SummarizeDiffRequest) (*ai.DiffSummary, error) {
	return nil, &ai.AIError{Err: ai.ErrAIUnavailable, Message: "circuit breaker open"}
}
func (m *mockUnavailableAI) ValidateSemantic(ctx context.Context, req ai.SemanticValidationRequest) (*ai.SemanticValidationResult, error) {
	return nil, &ai.AIError{Err: ai.ErrAIUnavailable, Message: "circuit breaker open"}
}
func (m *mockUnavailableAI) GetMode() string  { return "on" }
func (m *mockUnavailableAI) GetModel() string { return "mock" }

func TestConvertPaste_FallbackWhenAIUnavailable(t *testing.T) {
	conv := NewConverter().WithAIService(&mockUnavailableAI{})

	result, err := conv.ConvertPasteWithFormatContext(
		context.Background(),
		"ID\tTitle\tDescription\n1\tTest\tA test row",
		"spec",
		"spec",
	)
	if err != nil {
		t.Fatalf("should not error on AI unavailable, got: %v", err)
	}
	if result.MDFlow == "" {
		t.Error("should produce heuristic output even without AI")
	}
	// Check warnings contain AI unavailable notice
	hasWarning := false
	for _, w := range result.Warnings {
		if w.Code == "AI_UNAVAILABLE" {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Errorf("expected AI_UNAVAILABLE warning in response, got warnings: %+v", result.Warnings)
	}
}

func TestConvertPaste_MetaShowsDegradedWhenFallback(t *testing.T) {
	conv := NewConverter().WithAIService(&mockUnavailableAI{})
	// Use multi-column tabular input so DetectInputType returns table (not markdown)
	result, err := conv.ConvertPasteWithFormatContext(
		context.Background(),
		"ID\tTitle\tDescription\n1\tTest\tA test row",
		"spec",
		"spec",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Meta.AIDegraded {
		t.Error("meta.AIDegraded should be true when fallback")
	}
	if result.Meta.AIFallbackReason != "ai_unavailable" {
		t.Errorf("expected AIFallbackReason 'ai_unavailable', got %q", result.Meta.AIFallbackReason)
	}
}

func TestConvertPaste_NoAIService_StillWorks(t *testing.T) {
	conv := NewConverter() // no AI service
	result, err := conv.ConvertPasteWithFormatContext(
		context.Background(),
		"ID\tTitle\tDescription\n1\tTest\tA test row",
		"spec",
		"spec",
	)
	if err != nil {
		t.Fatalf("should work without AI: %v", err)
	}
	if result.MDFlow == "" {
		t.Error("should produce output without AI")
	}
}
