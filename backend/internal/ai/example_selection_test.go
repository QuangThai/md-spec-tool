package ai

import (
	"testing"
)

func TestSelectExamples_ExactSchemaMatch(t *testing.T) {
	store := DefaultExampleStore()
	selected := store.SelectExamples("column_mapping", SelectionContext{
		SchemaHint:  "test_case",
		Language:    "en",
		ColumnCount: 6,
	})
	if len(selected) == 0 {
		t.Fatal("expected at least 1 example")
	}
	// First example should be exact schema match
	if selected[0].SchemaType != "test_case" {
		t.Errorf("expected first example to be test_case, got %s", selected[0].SchemaType)
	}
}

func TestSelectExamples_LanguagePreference(t *testing.T) {
	store := DefaultExampleStore()
	selected := store.SelectExamples("column_mapping", SelectionContext{
		SchemaHint:  "",
		Language:    "ja",
		ColumnCount: 5,
	})
	if len(selected) == 0 {
		t.Fatal("expected at least 1 example")
	}
	// Japanese example should be prioritized
	hasJa := false
	for _, ex := range selected {
		if ex.Language == "ja" {
			hasJa = true
			break
		}
	}
	if !hasJa {
		t.Error("expected at least one Japanese example when language=ja")
	}
}

func TestSelectExamples_ColumnCountSimilarity(t *testing.T) {
	store := DefaultExampleStore()
	// UI spec has 8 headers, most others have 5-6
	selected := store.SelectExamples("column_mapping", SelectionContext{
		SchemaHint:  "",
		Language:    "en",
		ColumnCount: 8,
	})
	if len(selected) == 0 {
		t.Fatal("expected at least 1 example")
	}
	// UI spec (8 columns) should be preferred for 8-column input
	hasUISpec := false
	for _, ex := range selected {
		if ex.SchemaType == "ui_spec" {
			hasUISpec = true
			break
		}
	}
	if !hasUISpec {
		t.Error("expected ui_spec example for 8-column input (column count similarity)")
	}
}

func TestSelectExamples_DefaultMaxResults(t *testing.T) {
	store := DefaultExampleStore()
	selected := store.SelectExamples("column_mapping", SelectionContext{
		ColumnCount: 5,
	})
	if len(selected) > 3 {
		t.Errorf("expected max 3 default results, got %d", len(selected))
	}
	if len(selected) == 0 {
		t.Error("expected at least 1 result")
	}
}

func TestSelectExamples_CustomMaxResults(t *testing.T) {
	store := DefaultExampleStore()
	selected := store.SelectExamples("column_mapping", SelectionContext{
		ColumnCount: 5,
		MaxResults:  1,
	})
	if len(selected) != 1 {
		t.Errorf("expected exactly 1 result, got %d", len(selected))
	}
}

func TestSelectExamples_AlwaysReturnsAtLeastOne(t *testing.T) {
	store := DefaultExampleStore()
	// Even with weird context, should fallback to generic
	selected := store.SelectExamples("column_mapping", SelectionContext{
		SchemaHint:  "nonexistent_schema",
		Language:    "xx",
		ColumnCount: 100,
	})
	if len(selected) == 0 {
		t.Error("should always return at least 1 example (generic fallback)")
	}
}

func TestSelectExamples_EmptyStore(t *testing.T) {
	store := NewExampleStore()
	selected := store.SelectExamples("column_mapping", SelectionContext{
		ColumnCount: 5,
	})
	if len(selected) != 0 {
		t.Errorf("expected 0 examples from empty store, got %d", len(selected))
	}
}

func TestSelectExamples_ScoreCalculation(t *testing.T) {
	score := calculateExampleScore(
		Example{SchemaType: "test_case", Language: "en", Headers: []string{"A", "B", "C", "D", "E", "F"}},
		SelectionContext{SchemaHint: "test_case", Language: "en", ColumnCount: 6},
	)
	// Should get: schema=100 + language=50 + column_count=30 (exact match) = 180
	if score < 150 {
		t.Errorf("expected high score for exact match, got %d", score)
	}
}

func TestSelectExamples_ScoreDifferentSchema(t *testing.T) {
	score := calculateExampleScore(
		Example{SchemaType: "api_spec", Language: "en", Headers: []string{"A", "B", "C", "D", "E", "F"}},
		SelectionContext{SchemaHint: "test_case", Language: "en", ColumnCount: 6},
	)
	// No schema match, but language and column count match
	if score >= 150 {
		t.Errorf("expected lower score for non-matching schema, got %d", score)
	}
}
