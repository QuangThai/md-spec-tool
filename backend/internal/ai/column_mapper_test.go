package ai

import (
	"context"
	"testing"
)

func TestColumnMapperService_MapColumnsTableFormat(t *testing.T) {
	// Create a mock AI service that returns nil to force table format behavior
	mockService := &mockAIService{}
	mapper := NewColumnMapperService(mockService)

	req := MapColumnsRequest{
		Headers: []string{"ID", "Name", "Value"},
		Format:  "table",
	}

	result, err := mapper.MapColumns(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result to not be nil")
	}

	// For table format, all columns should be in extra_columns
	if len(result.ExtraColumns) != 3 {
		t.Errorf("expected 3 extra columns, got %d", len(result.ExtraColumns))
	}

	if len(result.CanonicalFields) != 0 {
		t.Errorf("expected 0 canonical fields, got %d", len(result.CanonicalFields))
	}

	if result.Meta.TotalColumns != 3 {
		t.Errorf("expected 3 total columns, got %d", result.Meta.TotalColumns)
	}

	if result.Meta.UnmappedColumns != 3 {
		t.Errorf("expected 3 unmapped columns, got %d", result.Meta.UnmappedColumns)
	}

	// Verify extra columns preserve original names
	for i, header := range req.Headers {
		if result.ExtraColumns[i].Name != header {
			t.Errorf("expected column %d name %q, got %q", i, header, result.ExtraColumns[i].Name)
		}
	}
}

func TestGetCanonicalName(t *testing.T) {
	result := &ColumnMappingResult{
		CanonicalFields: []CanonicalFieldMapping{
			{CanonicalName: "id", ColumnIndex: 0},
			{CanonicalName: "feature", ColumnIndex: 1},
		},
	}

	tests := []struct {
		index    int
		expected string
	}{
		{0, "id"},
		{1, "feature"},
		{2, ""},
	}

	for _, tt := range tests {
		got := GetCanonicalName(result, tt.index)
		if got != tt.expected {
			t.Errorf("GetCanonicalName(%d) = %q, want %q", tt.index, got, tt.expected)
		}
	}
}

func TestGetExtraColumnName(t *testing.T) {
	result := &ColumnMappingResult{
		ExtraColumns: []ExtraColumnMapping{
			{Name: "custom_field", ColumnIndex: 0},
			{Name: "metadata", ColumnIndex: 1},
		},
	}

	tests := []struct {
		index    int
		expected string
	}{
		{0, "custom_field"},
		{1, "metadata"},
		{2, ""},
	}

	for _, tt := range tests {
		got := GetExtraColumnName(result, tt.index)
		if got != tt.expected {
			t.Errorf("GetExtraColumnName(%d) = %q, want %q", tt.index, got, tt.expected)
		}
	}
}

func TestColumnMapperService_MixedFormatHeaders(t *testing.T) {
	mockService := &mockAIService{}
	mapper := NewColumnMapperService(mockService)

	// Test with various header formats
	req := MapColumnsRequest{
		Headers: []string{"TEST_ID", "Test Case Name", "Expected Result", "Unused Column"},
		Format:  "table",
	}

	result, err := mapper.MapColumns(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.SchemaVersion != SchemaVersionColumnMapping {
		t.Errorf("expected schema version %s, got %s", SchemaVersionColumnMapping, result.SchemaVersion)
	}

	if result.Meta.DetectedType != "table" {
		t.Errorf("expected detected type 'table', got %q", result.Meta.DetectedType)
	}
}

// mockAIService implements the Service interface for testing
type mockAIService struct {
	mapColumnsFunc     func(ctx context.Context, req MapColumnsRequest) (*ColumnMappingResult, error)
	analyzePasteFunc   func(ctx context.Context, req AnalyzePasteRequest) (*PasteAnalysis, error)
	getSuggestionsFunc func(ctx context.Context, req SuggestionsRequest) (*SuggestionsResult, error)
	modeFunc           func() string
}

func (m *mockAIService) MapColumns(ctx context.Context, req MapColumnsRequest) (*ColumnMappingResult, error) {
	if m.mapColumnsFunc != nil {
		return m.mapColumnsFunc(ctx, req)
	}
	return nil, ErrAIUnavailable
}

func (m *mockAIService) AnalyzePaste(ctx context.Context, req AnalyzePasteRequest) (*PasteAnalysis, error) {
	if m.analyzePasteFunc != nil {
		return m.analyzePasteFunc(ctx, req)
	}
	return nil, ErrAIUnavailable
}

func (m *mockAIService) GetSuggestions(ctx context.Context, req SuggestionsRequest) (*SuggestionsResult, error) {
	if m.getSuggestionsFunc != nil {
		return m.getSuggestionsFunc(ctx, req)
	}
	return &SuggestionsResult{
		SchemaVersion: SchemaVersionSuggestions,
		Suggestions:   []Suggestion{},
	}, nil
}

func (m *mockAIService) GetMode() string {
	if m.modeFunc != nil {
		return m.modeFunc()
	}
	return "mock"
}

func (m *mockAIService) SummarizeDiff(ctx context.Context, req SummarizeDiffRequest) (*DiffSummary, error) {
	return &DiffSummary{
		Summary:    "Mock summary",
		KeyChanges: []string{},
		Confidence: 1.0,
	}, nil
}

func (m *mockAIService) ValidateSemantic(ctx context.Context, req SemanticValidationRequest) (*SemanticValidationResult, error) {
	return &SemanticValidationResult{
		Issues:     []SemanticIssue{},
		Overall:    "good",
		Score:      1.0,
		Confidence: 1.0,
	}, nil
}
