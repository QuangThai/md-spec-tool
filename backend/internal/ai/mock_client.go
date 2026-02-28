package ai

import (
	"context"
	"sync"
)

// MockAIService is a configurable mock implementing the Service interface.
// Use it for unit/integration tests to control AI responses without calling OpenAI.
//
// Usage:
//
//	mock := NewMockAIService()
//	mock.MapColumnsFunc = func(ctx context.Context, req MapColumnsRequest) (*ColumnMappingResult, error) {
//	    return &ColumnMappingResult{...}, nil
//	}
//	result, err := mock.MapColumns(ctx, req)
type MockAIService struct {
	mu sync.RWMutex

	// Override functions for each operation
	MapColumnsFunc       func(ctx context.Context, req MapColumnsRequest) (*ColumnMappingResult, error)
	AnalyzePasteFunc     func(ctx context.Context, req AnalyzePasteRequest) (*PasteAnalysis, error)
	GetSuggestionsFunc   func(ctx context.Context, req SuggestionsRequest) (*SuggestionsResult, error)
	SummarizeDiffFunc    func(ctx context.Context, req SummarizeDiffRequest) (*DiffSummary, error)
	ValidateSemanticFunc func(ctx context.Context, req SemanticValidationRequest) (*SemanticValidationResult, error)

	// Call tracking
	Calls []MockCall

	// Control
	Mode  string
	Model string
}

// MockCall records a single method invocation for assertions
type MockCall struct {
	Method string
	Args   interface{}
}

// NewMockAIService creates a mock with default "happy path" responses
func NewMockAIService() *MockAIService {
	return &MockAIService{
		Mode:  "on",
		Model: "gpt-4o-mini-mock",
	}
}

// NewMockAIServiceWithDefaults creates a mock pre-configured with realistic responses
func NewMockAIServiceWithDefaults() *MockAIService {
	m := NewMockAIService()

	m.MapColumnsFunc = func(_ context.Context, req MapColumnsRequest) (*ColumnMappingResult, error) {
		fields := make([]CanonicalFieldMapping, 0, len(req.Headers))
		for i, h := range req.Headers {
			fields = append(fields, CanonicalFieldMapping{
				CanonicalName: guessCanonical(h),
				SourceHeader:  h,
				ColumnIndex:   i,
				Confidence:    0.85,
				Reasoning:     "mock mapping",
			})
		}
		return &ColumnMappingResult{
			SchemaVersion:   SchemaVersionColumnMapping,
			CanonicalFields: fields,
			Meta: MappingMeta{
				DetectedType:   "generic",
				SourceLanguage: "en",
				TotalColumns:   len(req.Headers),
				MappedColumns:  len(req.Headers),
				AvgConfidence:  0.85,
			},
		}, nil
	}

	m.AnalyzePasteFunc = func(_ context.Context, req AnalyzePasteRequest) (*PasteAnalysis, error) {
		return &PasteAnalysis{
			SchemaVersion:   SchemaVersionPasteAnalysis,
			InputType:       "table",
			DetectedFormat:  "tsv",
			DetectedSchema:  "generic",
			SuggestedOutput: "spec",
			Confidence:      0.9,
		}, nil
	}

	m.GetSuggestionsFunc = func(_ context.Context, req SuggestionsRequest) (*SuggestionsResult, error) {
		return &SuggestionsResult{
			SchemaVersion: SchemaVersionSuggestions,
			Suggestions: []Suggestion{
				{
					Type:       SuggestionMissingField,
					Severity:   "info",
					Message:    "Consider adding expected results",
					Suggestion: "Add an 'Expected' column for test cases",
				},
			},
		}, nil
	}

	m.SummarizeDiffFunc = func(_ context.Context, req SummarizeDiffRequest) (*DiffSummary, error) {
		return &DiffSummary{
			Summary:    "Mock diff summary",
			KeyChanges: []string{"change 1", "change 2"},
			Confidence: 0.9,
		}, nil
	}

	m.ValidateSemanticFunc = func(_ context.Context, req SemanticValidationRequest) (*SemanticValidationResult, error) {
		return &SemanticValidationResult{
			Issues:     nil,
			Overall:    "good",
			Score:      0.9,
			Confidence: 0.85,
		}, nil
	}

	return m
}

// ---- Service interface implementation ----

func (m *MockAIService) MapColumns(ctx context.Context, req MapColumnsRequest) (*ColumnMappingResult, error) {
	m.recordCall("MapColumns", req)
	if m.MapColumnsFunc != nil {
		return m.MapColumnsFunc(ctx, req)
	}
	return &ColumnMappingResult{SchemaVersion: SchemaVersionColumnMapping}, nil
}

func (m *MockAIService) AnalyzePaste(ctx context.Context, req AnalyzePasteRequest) (*PasteAnalysis, error) {
	m.recordCall("AnalyzePaste", req)
	if m.AnalyzePasteFunc != nil {
		return m.AnalyzePasteFunc(ctx, req)
	}
	return &PasteAnalysis{SchemaVersion: SchemaVersionPasteAnalysis}, nil
}

func (m *MockAIService) GetSuggestions(ctx context.Context, req SuggestionsRequest) (*SuggestionsResult, error) {
	m.recordCall("GetSuggestions", req)
	if m.GetSuggestionsFunc != nil {
		return m.GetSuggestionsFunc(ctx, req)
	}
	return &SuggestionsResult{SchemaVersion: SchemaVersionSuggestions}, nil
}

func (m *MockAIService) SummarizeDiff(ctx context.Context, req SummarizeDiffRequest) (*DiffSummary, error) {
	m.recordCall("SummarizeDiff", req)
	if m.SummarizeDiffFunc != nil {
		return m.SummarizeDiffFunc(ctx, req)
	}
	return &DiffSummary{}, nil
}

func (m *MockAIService) ValidateSemantic(ctx context.Context, req SemanticValidationRequest) (*SemanticValidationResult, error) {
	m.recordCall("ValidateSemantic", req)
	if m.ValidateSemanticFunc != nil {
		return m.ValidateSemanticFunc(ctx, req)
	}
	return &SemanticValidationResult{Overall: "good", Score: 1.0, Confidence: 1.0}, nil
}

func (m *MockAIService) GetMode() string {
	if m.Mode != "" {
		return m.Mode
	}
	return "on"
}

func (m *MockAIService) GetModel() string {
	if m.Model != "" {
		return m.Model
	}
	return "gpt-4o-mini-mock"
}

// ---- Call tracking helpers ----

func (m *MockAIService) recordCall(method string, args interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockCall{Method: method, Args: args})
}

// CallCount returns total number of calls
func (m *MockAIService) CallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Calls)
}

// CallCountFor returns calls for a specific method
func (m *MockAIService) CallCountFor(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, c := range m.Calls {
		if c.Method == method {
			count++
		}
	}
	return count
}

// LastCall returns the most recent call
func (m *MockAIService) LastCall() (MockCall, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.Calls) == 0 {
		return MockCall{}, false
	}
	return m.Calls[len(m.Calls)-1], true
}

// Reset clears all call history
func (m *MockAIService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = nil
}

// ---- Helpers ----

// guessCanonical provides a simple heuristic mapping for mock responses
func guessCanonical(header string) string {
	lower := toLower(header)
	switch {
	case contains(lower, "id") || contains(lower, "#"):
		return "id"
	case contains(lower, "title") || contains(lower, "name"):
		return "title"
	case contains(lower, "desc"):
		return "description"
	case contains(lower, "step") || contains(lower, "instruct"):
		return "instructions"
	case contains(lower, "expect"):
		return "expected"
	case contains(lower, "status"):
		return "status"
	case contains(lower, "priority"):
		return "priority"
	case contains(lower, "type"):
		return "type"
	case contains(lower, "note") || contains(lower, "remark"):
		return "notes"
	default:
		return "notes"
	}
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
