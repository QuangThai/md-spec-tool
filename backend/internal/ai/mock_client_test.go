package ai

import (
	"context"
	"errors"
	"testing"
)

func TestMockAIService_ImplementsInterface(t *testing.T) {
	// Compile-time check
	var _ Service = (*MockAIService)(nil)
}

func TestMockAIService_DefaultResponses(t *testing.T) {
	mock := NewMockAIService()
	ctx := context.Background()

	t.Run("MapColumns returns empty result", func(t *testing.T) {
		result, err := mock.MapColumns(ctx, MapColumnsRequest{Headers: []string{"ID"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.SchemaVersion != SchemaVersionColumnMapping {
			t.Errorf("expected schema version %s, got %s", SchemaVersionColumnMapping, result.SchemaVersion)
		}
	})

	t.Run("AnalyzePaste returns empty result", func(t *testing.T) {
		result, err := mock.AnalyzePaste(ctx, AnalyzePasteRequest{Content: "test"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.SchemaVersion != SchemaVersionPasteAnalysis {
			t.Errorf("expected schema version %s", SchemaVersionPasteAnalysis)
		}
	})

	t.Run("GetSuggestions returns empty result", func(t *testing.T) {
		result, err := mock.GetSuggestions(ctx, SuggestionsRequest{SpecContent: "test"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.SchemaVersion != SchemaVersionSuggestions {
			t.Errorf("expected schema version %s", SchemaVersionSuggestions)
		}
	})

	t.Run("SummarizeDiff returns empty result", func(t *testing.T) {
		result, err := mock.SummarizeDiff(ctx, SummarizeDiffRequest{Before: "a", After: "b"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected non-nil result")
		}
	})

	t.Run("ValidateSemantic returns good result", func(t *testing.T) {
		result, err := mock.ValidateSemantic(ctx, SemanticValidationRequest{SpecContent: "test"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Overall != "good" {
			t.Errorf("expected overall=good, got %s", result.Overall)
		}
	})
}

func TestMockAIServiceWithDefaults_RealisticResponses(t *testing.T) {
	mock := NewMockAIServiceWithDefaults()
	ctx := context.Background()

	t.Run("MapColumns generates mappings from headers", func(t *testing.T) {
		result, err := mock.MapColumns(ctx, MapColumnsRequest{
			Headers: []string{"TC ID", "Title", "Steps", "Expected"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.CanonicalFields) != 4 {
			t.Errorf("expected 4 fields, got %d", len(result.CanonicalFields))
		}
		// Should map "TC ID" to "id"
		if result.CanonicalFields[0].CanonicalName != "id" {
			t.Errorf("expected 'id' for 'TC ID', got %s", result.CanonicalFields[0].CanonicalName)
		}
		if result.Meta.AvgConfidence != 0.85 {
			t.Errorf("expected confidence 0.85, got %f", result.Meta.AvgConfidence)
		}
	})

	t.Run("AnalyzePaste returns table detection", func(t *testing.T) {
		result, err := mock.AnalyzePaste(ctx, AnalyzePasteRequest{Content: "ID\tTitle\n1\tTest"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.InputType != "table" {
			t.Errorf("expected input_type=table, got %s", result.InputType)
		}
	})

	t.Run("GetSuggestions returns suggestions", func(t *testing.T) {
		result, err := mock.GetSuggestions(ctx, SuggestionsRequest{SpecContent: "test"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Suggestions) == 0 {
			t.Error("expected at least 1 suggestion")
		}
	})
}

func TestMockAIService_CustomFunction(t *testing.T) {
	mock := NewMockAIService()
	ctx := context.Background()

	expectedErr := errors.New("AI is down")
	mock.MapColumnsFunc = func(_ context.Context, _ MapColumnsRequest) (*ColumnMappingResult, error) {
		return nil, expectedErr
	}

	_, err := mock.MapColumns(ctx, MapColumnsRequest{Headers: []string{"ID"}})
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected custom error, got: %v", err)
	}
}

func TestMockAIService_CallTracking(t *testing.T) {
	mock := NewMockAIServiceWithDefaults()
	ctx := context.Background()

	if mock.CallCount() != 0 {
		t.Error("expected 0 calls initially")
	}

	mock.MapColumns(ctx, MapColumnsRequest{Headers: []string{"ID"}})
	mock.MapColumns(ctx, MapColumnsRequest{Headers: []string{"Title"}})
	mock.AnalyzePaste(ctx, AnalyzePasteRequest{Content: "test"})

	if mock.CallCount() != 3 {
		t.Errorf("expected 3 total calls, got %d", mock.CallCount())
	}

	if mock.CallCountFor("MapColumns") != 2 {
		t.Errorf("expected 2 MapColumns calls, got %d", mock.CallCountFor("MapColumns"))
	}

	if mock.CallCountFor("AnalyzePaste") != 1 {
		t.Errorf("expected 1 AnalyzePaste call, got %d", mock.CallCountFor("AnalyzePaste"))
	}

	last, ok := mock.LastCall()
	if !ok {
		t.Fatal("expected last call")
	}
	if last.Method != "AnalyzePaste" {
		t.Errorf("expected last call to be AnalyzePaste, got %s", last.Method)
	}
}

func TestMockAIService_Reset(t *testing.T) {
	mock := NewMockAIServiceWithDefaults()
	ctx := context.Background()

	mock.MapColumns(ctx, MapColumnsRequest{Headers: []string{"ID"}})
	mock.Reset()

	if mock.CallCount() != 0 {
		t.Error("expected 0 calls after reset")
	}
}

func TestMockAIService_GetModeAndModel(t *testing.T) {
	mock := NewMockAIService()

	if mock.GetMode() != "on" {
		t.Errorf("expected mode=on, got %s", mock.GetMode())
	}
	if mock.GetModel() != "gpt-4o-mini-mock" {
		t.Errorf("expected model=gpt-4o-mini-mock, got %s", mock.GetModel())
	}

	mock.Mode = "off"
	mock.Model = "custom-model"

	if mock.GetMode() != "off" {
		t.Errorf("expected mode=off, got %s", mock.GetMode())
	}
	if mock.GetModel() != "custom-model" {
		t.Errorf("expected model=custom-model, got %s", mock.GetModel())
	}
}

func TestMockAIService_ConcurrentSafety(t *testing.T) {
	mock := NewMockAIServiceWithDefaults()
	ctx := context.Background()

	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			mock.MapColumns(ctx, MapColumnsRequest{Headers: []string{"ID"}})
			mock.CallCount()
			mock.CallCountFor("MapColumns")
			mock.LastCall()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
	// No panic = pass
}

func TestGuessCanonical(t *testing.T) {
	tests := []struct {
		header   string
		expected string
	}{
		{"TC ID", "id"},
		{"Issue #", "id"},
		{"Title", "title"},
		{"Test Case Name", "title"},
		{"Description", "description"},
		{"Steps", "instructions"},
		{"Expected Result", "expected"},
		{"Status", "status"},
		{"Priority", "priority"},
		{"Type", "type"},
		{"Notes", "notes"},
		{"Remarks", "notes"},
		{"Unknown Column", "notes"}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := guessCanonical(tt.header)
			if got != tt.expected {
				t.Errorf("guessCanonical(%q) = %q, want %q", tt.header, got, tt.expected)
			}
		})
	}
}
