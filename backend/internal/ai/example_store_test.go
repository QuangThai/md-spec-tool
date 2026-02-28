package ai

import (
	"testing"
)

func TestExampleStore_RegisterAndGet(t *testing.T) {
	store := NewExampleStore()
	store.Register(Example{
		Operation:  "column_mapping",
		SchemaType: "test_case",
		Language:   "en",
		Headers:    []string{"TC ID", "Test Case Name", "Expected"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "TC ID", ColumnIndex: 0, Confidence: 1.0},
		},
	})

	examples := store.GetExamples("column_mapping", ExampleFilter{})
	if len(examples) != 1 {
		t.Errorf("expected 1 example, got %d", len(examples))
	}
}

func TestExampleStore_FilterBySchemaType(t *testing.T) {
	store := NewExampleStore()
	store.Register(Example{Operation: "column_mapping", SchemaType: "test_case", Language: "en", Headers: []string{"TC ID"}})
	store.Register(Example{Operation: "column_mapping", SchemaType: "ui_spec", Language: "en", Headers: []string{"No"}})
	store.Register(Example{Operation: "column_mapping", SchemaType: "api_spec", Language: "en", Headers: []string{"Endpoint"}})

	examples := store.GetExamples("column_mapping", ExampleFilter{SchemaType: "test_case"})
	if len(examples) != 1 {
		t.Errorf("expected 1 test_case example, got %d", len(examples))
	}
	if examples[0].SchemaType != "test_case" {
		t.Errorf("expected test_case, got %s", examples[0].SchemaType)
	}
}

func TestExampleStore_FilterByLanguage(t *testing.T) {
	store := NewExampleStore()
	store.Register(Example{Operation: "column_mapping", SchemaType: "test_case", Language: "en", Headers: []string{"TC ID"}})
	store.Register(Example{Operation: "column_mapping", SchemaType: "issue_tracker", Language: "ja", Headers: []string{"Issue #"}})

	examples := store.GetExamples("column_mapping", ExampleFilter{Language: "ja"})
	if len(examples) != 1 {
		t.Errorf("expected 1 ja example, got %d", len(examples))
	}
}

func TestExampleStore_FilterCombined(t *testing.T) {
	store := NewExampleStore()
	store.Register(Example{Operation: "column_mapping", SchemaType: "test_case", Language: "en", Headers: []string{"TC ID"}})
	store.Register(Example{Operation: "column_mapping", SchemaType: "test_case", Language: "ja", Headers: []string{"テストID"}})
	store.Register(Example{Operation: "column_mapping", SchemaType: "ui_spec", Language: "en", Headers: []string{"No"}})

	examples := store.GetExamples("column_mapping", ExampleFilter{SchemaType: "test_case", Language: "en"})
	if len(examples) != 1 {
		t.Errorf("expected 1 example, got %d", len(examples))
	}
}

func TestExampleStore_LimitResults(t *testing.T) {
	store := NewExampleStore()
	for i := 0; i < 10; i++ {
		store.Register(Example{Operation: "column_mapping", SchemaType: "test_case", Language: "en", Headers: []string{"H"}})
	}

	examples := store.GetExamples("column_mapping", ExampleFilter{MaxResults: 3})
	if len(examples) != 3 {
		t.Errorf("expected 3 examples, got %d", len(examples))
	}
}

func TestExampleStore_EmptyFilter(t *testing.T) {
	store := NewExampleStore()
	store.Register(Example{Operation: "column_mapping", SchemaType: "test_case", Language: "en", Headers: []string{"TC ID"}})
	store.Register(Example{Operation: "column_mapping", SchemaType: "ui_spec", Language: "en", Headers: []string{"No"}})

	examples := store.GetExamples("column_mapping", ExampleFilter{})
	if len(examples) != 2 {
		t.Errorf("expected 2 examples with empty filter, got %d", len(examples))
	}
}

func TestExampleStore_WrongOperation(t *testing.T) {
	store := NewExampleStore()
	store.Register(Example{Operation: "column_mapping", SchemaType: "test_case", Language: "en", Headers: []string{"TC ID"}})

	examples := store.GetExamples("paste_analysis", ExampleFilter{})
	if len(examples) != 0 {
		t.Errorf("expected 0 examples for wrong operation, got %d", len(examples))
	}
}

func TestExampleStore_FormatForPrompt(t *testing.T) {
	store := NewExampleStore()
	store.Register(Example{
		Operation:  "column_mapping",
		SchemaType: "test_case",
		Language:   "en",
		Headers:    []string{"TC ID", "Test Case Name", "Expected"},
		Mappings: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "TC ID", ColumnIndex: 0, Confidence: 1.0, Reasoning: "Test case identifier"},
		},
	})

	examples := store.GetExamples("column_mapping", ExampleFilter{SchemaType: "test_case"})
	formatted := FormatExamplesForPrompt(examples)
	if formatted == "" {
		t.Error("expected non-empty formatted output")
	}
	// Should contain headers and mapping info
	if !containsString(formatted, "TC ID") {
		t.Error("formatted output should contain header 'TC ID'")
	}
}

func TestExampleStore_ConcurrentSafety(t *testing.T) {
	store := NewExampleStore()
	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			store.Register(Example{Operation: "test", SchemaType: "t", Language: "en", Headers: []string{"H"}})
			store.GetExamples("test", ExampleFilter{})
			done <- struct{}{}
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
}

func TestDefaultExampleStore_HasAllSchemaTypes(t *testing.T) {
	store := DefaultExampleStore()

	requiredTypes := []string{"test_case", "issue_tracker", "ui_spec", "product_backlog", "api_spec", "generic"}
	for _, schemaType := range requiredTypes {
		examples := store.GetExamples("column_mapping", ExampleFilter{SchemaType: schemaType})
		if len(examples) == 0 {
			t.Errorf("expected at least 1 example for schema type %q", schemaType)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
