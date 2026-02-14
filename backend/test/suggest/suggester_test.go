package suggest_test

import (
	"context"
	. "github.com/yourorg/md-spec-tool/internal/suggest"
	"testing"

	"github.com/yourorg/md-spec-tool/internal/ai"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

// mockAIService implements ai.Service for testing
type mockAIService struct {
	getSuggestionsFunc func(ctx context.Context, req ai.SuggestionsRequest) (*ai.SuggestionsResult, error)
	mode               string
}

func (m *mockAIService) MapColumns(ctx context.Context, req ai.MapColumnsRequest) (*ai.ColumnMappingResult, error) {
	return nil, nil
}

func (m *mockAIService) AnalyzePaste(ctx context.Context, req ai.AnalyzePasteRequest) (*ai.PasteAnalysis, error) {
	return nil, nil
}

func (m *mockAIService) GetSuggestions(ctx context.Context, req ai.SuggestionsRequest) (*ai.SuggestionsResult, error) {
	if m.getSuggestionsFunc != nil {
		return m.getSuggestionsFunc(ctx, req)
	}
	return &ai.SuggestionsResult{
		SchemaVersion: ai.SchemaVersionSuggestions,
		Suggestions:   []ai.Suggestion{},
	}, nil
}

func (m *mockAIService) GetMode() string {
	if m.mode != "" {
		return m.mode
	}
	return "on"
}

func (m *mockAIService) SummarizeDiff(ctx context.Context, req ai.SummarizeDiffRequest) (*ai.DiffSummary, error) {
	return &ai.DiffSummary{
		Summary:    "Mock summary",
		KeyChanges: []string{},
		Confidence: 1.0,
	}, nil
}

func (m *mockAIService) ValidateSemantic(ctx context.Context, req ai.SemanticValidationRequest) (*ai.SemanticValidationResult, error) {
	return &ai.SemanticValidationResult{
		Issues:     []ai.SemanticIssue{},
		Overall:    "good",
		Score:      1.0,
		Confidence: 1.0,
	}, nil
}

func TestNewSuggester(t *testing.T) {
	mockService := &mockAIService{}
	suggester := NewSuggester(mockService)

	if suggester == nil {
		t.Fatal("expected suggester to not be nil")
	}

	if suggester.AIService() != mockService {
		t.Error("expected aiService to be set")
	}
}

func TestSuggester_IsConfigured(t *testing.T) {
	tests := []struct {
		name     string
		service  ai.Service
		mode     string
		expected bool
	}{
		{
			name:     "nil service",
			service:  nil,
			expected: false,
		},
		{
			name:     "service with mode on",
			service:  &mockAIService{mode: "on"},
			expected: true,
		},
		{
			name:     "service with mode off",
			service:  &mockAIService{mode: "off"},
			expected: false,
		},
		{
			name:     "service with mode mock",
			service:  &mockAIService{mode: "mock"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggester := NewSuggester(tt.service)
			got := suggester.IsConfigured()
			if got != tt.expected {
				t.Errorf("IsConfigured() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSuggester_GetSuggestions_NotConfigured(t *testing.T) {
	suggester := NewSuggester(nil)

	resp, err := suggester.GetSuggestions(context.Background(), &SuggestionRequest{
		SpecDoc:  &converter.SpecDoc{},
		Template: "spec",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Error == "" {
		t.Error("expected error message when not configured")
	}
}

func TestSuggester_GetSuggestions_EmptySpecDoc(t *testing.T) {
	mockService := &mockAIService{mode: "on"}
	suggester := NewSuggester(mockService)

	// Test with nil SpecDoc
	resp, err := suggester.GetSuggestions(context.Background(), &SuggestionRequest{
		SpecDoc:  nil,
		Template: "spec",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Suggestions) != 0 {
		t.Errorf("expected empty suggestions for nil SpecDoc, got %d", len(resp.Suggestions))
	}

	// Test with empty rows
	resp, err = suggester.GetSuggestions(context.Background(), &SuggestionRequest{
		SpecDoc:  &converter.SpecDoc{Rows: []converter.SpecRow{}},
		Template: "spec",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Suggestions) != 0 {
		t.Errorf("expected empty suggestions for empty rows, got %d", len(resp.Suggestions))
	}
}

func TestSuggester_GetSuggestions_Success(t *testing.T) {
	rowRef := 1
	mockService := &mockAIService{
		mode: "on",
		getSuggestionsFunc: func(ctx context.Context, req ai.SuggestionsRequest) (*ai.SuggestionsResult, error) {
			// Verify request is properly built
			if req.Template != "spec" {
				t.Errorf("expected template 'spec', got %q", req.Template)
			}
			if req.RowCount != 2 {
				t.Errorf("expected row count 2, got %d", req.RowCount)
			}
			if req.SpecContent == "" {
				t.Error("expected spec content to not be empty")
			}

			return &ai.SuggestionsResult{
				SchemaVersion: ai.SchemaVersionSuggestions,
				Suggestions: []ai.Suggestion{
					{
						Type:       ai.SuggestionVagueDescription,
						Severity:   "warn",
						Message:    "Expected result is too vague",
						RowRef:     &rowRef,
						Field:      "expected",
						Suggestion: "Be more specific",
					},
				},
			}, nil
		},
	}

	suggester := NewSuggester(mockService)

	resp, err := suggester.GetSuggestions(context.Background(), &SuggestionRequest{
		SpecDoc: &converter.SpecDoc{
			Rows: []converter.SpecRow{
				{ID: "TC001", Scenario: "Login", Expected: "check result"},
				{ID: "TC002", Scenario: "Logout", Expected: "verify"},
			},
		},
		Template: "spec",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Error != "" {
		t.Errorf("unexpected error in response: %s", resp.Error)
	}

	if len(resp.Suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(resp.Suggestions))
	}

	suggestion := resp.Suggestions[0]
	if suggestion.Type != ai.SuggestionVagueDescription {
		t.Errorf("expected type 'vague_description', got %q", suggestion.Type)
	}
	if suggestion.Severity != "warn" {
		t.Errorf("expected severity 'warn', got %q", suggestion.Severity)
	}
	if suggestion.RowRef == nil || *suggestion.RowRef != 1 {
		t.Errorf("expected row_ref 1, got %v", suggestion.RowRef)
	}
}

func TestBuildSpecContent(t *testing.T) {
	doc := &converter.SpecDoc{
		Rows: []converter.SpecRow{
			{
				ID:           "TC001",
				Feature:      "Login",
				Scenario:     "Valid credentials",
				Instructions: "1. Enter username\n2. Enter password",
				Expected:     "User logged in",
				Precondition: "User exists",
				Priority:     "High",
				Notes:        "Critical path",
			},
			{
				ID:       "TC002",
				Scenario: "Invalid credentials",
				Expected: "Error message",
			},
		},
	}

	content := BuildSpecContent(doc)

	// Verify content includes all fields
	if content == "" {
		t.Fatal("expected content to not be empty")
	}

	expectedParts := []string{
		"Row 1",
		"Row 2",
		"ID: TC001",
		"Feature: Login",
		"Scenario: Valid credentials",
		"Instructions: 1. Enter username",
		"Expected: User logged in",
		"Precondition: User exists",
		"Priority: High",
		"Notes: Critical path",
		"ID: TC002",
		"Scenario: Invalid credentials",
	}

	for _, part := range expectedParts {
		if !contains(content, part) {
			t.Errorf("expected content to contain %q", part)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
