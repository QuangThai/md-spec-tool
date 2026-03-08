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

// TestAnalyzePasteForConvert_ParseFirst_MultiColumnTable verifies parse-first strategy:
// when paste parses as multi-column table, AnalyzePasteForConvert returns Table (not DetectInputType).
func TestAnalyzePasteForConvert_ParseFirst_MultiColumnTable(t *testing.T) {
	conv := NewConverter()
	// CSV from Google Sheets - might be misclassified as markdown by DetectInputType
	// due to lack of strong table signals, but parse succeeds with 3 columns
	csv := "Endpoint,Method,Status\n/api/users,GET,200\n/api/users,POST,201"
	analysis := conv.AnalyzePasteForConvert(csv)
	if analysis.Type != InputTypeTable {
		t.Errorf("expected InputTypeTable for parseable multi-column CSV, got %s", analysis.Type)
	}
	if analysis.Confidence < 80 {
		t.Errorf("expected high confidence for parsed table, got %d", analysis.Confidence)
	}
}

// TestConvertPaste_ParseFirst_CSV_TakesTablePath verifies that CSV paste
// (e.g. from Google Sheets) always takes table path via parse-first.
func TestConvertPaste_ParseFirst_CSV_TakesTablePath(t *testing.T) {
	conv := NewConverter()
	csv := "Endpoint,Method,Status\n/api/users,GET,200\n/api/users,POST,201"
	result, err := conv.ConvertPasteWithFormatContext(
		context.Background(),
		csv,
		"spec",
		"spec",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should produce table-derived output (has structure), not raw markdown
	if result.MDFlow == "" {
		t.Error("expected non-empty output")
	}
	// Meta should indicate table path was used (e.g. has rows, not just prose)
	if result.Meta.TotalRows < 2 {
		t.Errorf("expected at least 2 data rows from CSV, got TotalRows=%d", result.Meta.TotalRows)
	}
}

// mockPartialAIMock returns low avg confidence (below threshold) but 1+ high-confidence mapping.
// Used to verify MAPPING_AI_PARTIAL_MERGE is used instead of pure fallback.
type mockPartialAI struct {
	result *ai.ColumnMappingResult
}

func (m *mockPartialAI) MapColumns(ctx context.Context, req ai.MapColumnsRequest) (*ai.ColumnMappingResult, error) {
	return m.result, nil
}
func (m *mockPartialAI) AnalyzePaste(ctx context.Context, req ai.AnalyzePasteRequest) (*ai.PasteAnalysis, error) {
	return nil, &ai.AIError{Err: ai.ErrAIUnavailable, Message: "not implemented"}
}
func (m *mockPartialAI) GetSuggestions(ctx context.Context, req ai.SuggestionsRequest) (*ai.SuggestionsResult, error) {
	return nil, &ai.AIError{Err: ai.ErrAIUnavailable, Message: "not implemented"}
}
func (m *mockPartialAI) SummarizeDiff(ctx context.Context, req ai.SummarizeDiffRequest) (*ai.DiffSummary, error) {
	return nil, &ai.AIError{Err: ai.ErrAIUnavailable, Message: "not implemented"}
}
func (m *mockPartialAI) ValidateSemantic(ctx context.Context, req ai.SemanticValidationRequest) (*ai.SemanticValidationResult, error) {
	return nil, &ai.AIError{Err: ai.ErrAIUnavailable, Message: "not implemented"}
}
func (m *mockPartialAI) GetMode() string  { return "on" }
func (m *mockPartialAI) GetModel() string { return "mock-partial" }

// TestConvertPaste_PartialMerge_UsesAIPlusFallback verifies that when AI returns
// low overall confidence but ≥1 mapping with confidence >= 0.75, we merge (not pure fallback).
func TestConvertPaste_PartialMerge_UsesAIPlusFallback(t *testing.T) {
	// AI returns: id (0.9), title (0.5) — avg ~0.7, only 1 meets 0.75
	mock := &mockPartialAI{
		result: &ai.ColumnMappingResult{
			SchemaVersion: ai.SchemaVersionColumnMapping,
			CanonicalFields: []ai.CanonicalFieldMapping{
				{CanonicalName: "id", SourceHeader: "TC ID", ColumnIndex: 0, Confidence: 0.9},
				{CanonicalName: "title", SourceHeader: "Name", ColumnIndex: 1, Confidence: 0.5},
			},
			Meta: ai.MappingMeta{AvgConfidence: 0.7, MappedColumns: 2, TotalColumns: 3},
		},
	}
	conv := NewConverter().WithAIService(mock)
	// Headers that rule-based would map differently
	input := "TC ID\tName\tStatus\n1\tTest\tPass"
	result, err := conv.ConvertPasteWithFormatContext(context.Background(), input, "spec", "spec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hasPartialMerge := false
	for _, w := range result.Warnings {
		if w.Code == "MAPPING_AI_PARTIAL_MERGE" {
			hasPartialMerge = true
			break
		}
	}
	if !hasPartialMerge {
		t.Errorf("expected MAPPING_AI_PARTIAL_MERGE warning when AI has partial high-confidence mappings, got warnings: %+v", result.Warnings)
	}
	if result.MDFlow == "" {
		t.Error("expected non-empty output")
	}
}
