package ai

import (
	"testing"
)

func TestModelRouter_SimpleInput(t *testing.T) {
	router := NewModelRouter(ModelRouterConfig{
		SimpleModel:  "gpt-4o-mini",
		ComplexModel: "gpt-4o",
	})

	model := router.SelectModel(RoutingContext{
		ColumnCount: 5,
		Headers:     []string{"ID", "Title", "Description", "Status", "Priority"},
		Language:    "en",
	})

	if model != "gpt-4o-mini" {
		t.Errorf("expected gpt-4o-mini for simple input, got %s", model)
	}
}

func TestModelRouter_ComplexHighColumnCount(t *testing.T) {
	router := NewModelRouter(ModelRouterConfig{
		SimpleModel:  "gpt-4o-mini",
		ComplexModel: "gpt-4o",
	})

	headers := make([]string, 25)
	for i := range headers {
		headers[i] = "Column"
	}

	model := router.SelectModel(RoutingContext{
		ColumnCount: 25,
		Headers:     headers,
		Language:    "en",
	})

	if model != "gpt-4o" {
		t.Errorf("expected gpt-4o for 25 columns, got %s", model)
	}
}

func TestModelRouter_ComplexNonEnglish(t *testing.T) {
	router := NewModelRouter(ModelRouterConfig{
		SimpleModel:  "gpt-4o-mini",
		ComplexModel: "gpt-4o",
	})

	model := router.SelectModel(RoutingContext{
		ColumnCount: 5,
		Headers:     []string{"テストID", "テスト名", "期待結果"},
		Language:    "ja",
	})

	if model != "gpt-4o" {
		t.Errorf("expected gpt-4o for Japanese headers, got %s", model)
	}
}

func TestModelRouter_DefaultModel(t *testing.T) {
	router := NewModelRouter(ModelRouterConfig{
		SimpleModel:  "gpt-4o-mini",
		ComplexModel: "gpt-4o",
	})

	model := router.SelectModel(RoutingContext{
		ColumnCount: 10,
		Headers:     []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"},
		Language:    "en",
	})

	// 10 columns, English → still simple (threshold is >20)
	if model != "gpt-4o-mini" {
		t.Errorf("expected gpt-4o-mini for borderline, got %s", model)
	}
}

func TestModelRouter_MixedLanguageHeaders(t *testing.T) {
	router := NewModelRouter(ModelRouterConfig{
		SimpleModel:  "gpt-4o-mini",
		ComplexModel: "gpt-4o",
	})

	model := router.SelectModel(RoutingContext{
		ColumnCount: 6,
		Headers:     []string{"ID", "Name", "説明", "Status", "優先度", "Notes"},
		Language:    "en", // even if language is "en", mixed headers → complex
	})

	if model != "gpt-4o" {
		t.Errorf("expected gpt-4o for mixed language headers, got %s", model)
	}
}

func TestModelRouter_EmptyConfig(t *testing.T) {
	router := NewModelRouter(ModelRouterConfig{})

	model := router.SelectModel(RoutingContext{ColumnCount: 5})
	if model == "" {
		t.Error("expected non-empty default model")
	}
}

func TestDetectNonASCIIHeaders(t *testing.T) {
	tests := []struct {
		headers  []string
		expected bool
	}{
		{[]string{"ID", "Title"}, false},
		{[]string{"テストID", "Name"}, true},
		{[]string{"ID", "Título"}, true},
		{[]string{}, false},
	}
	for _, tt := range tests {
		got := hasNonASCIIHeaders(tt.headers)
		if got != tt.expected {
			t.Errorf("hasNonASCIIHeaders(%v) = %v, want %v", tt.headers, got, tt.expected)
		}
	}
}
