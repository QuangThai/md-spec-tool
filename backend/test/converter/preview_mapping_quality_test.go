package converter_test

import (
	. "github.com/yourorg/md-spec-tool/internal/converter"
	"testing"
)

func TestBuildPreviewMappingQuality_RecommendsTableWhenNoCoreFields(t *testing.T) {
	// No core fields mapped, low ratio -> should recommend table
	headers := []string{"ColA", "ColB", "ColC", "ColD"}
	rows := [][]string{{"x", "y", "z", "w"}}
	columnMapping := map[string]string{"ColB": "notes"} // notes is not in core fields
	unmapped := []string{"ColA", "ColC", "ColD"}

	quality := BuildPreviewMappingQuality(30, headers, rows, columnMapping, unmapped)

	if quality.RecommendedFormat != "table" {
		t.Fatalf("expected table recommendation when no core fields mapped, got %q", quality.RecommendedFormat)
	}
}

func TestBuildPreviewMappingQuality_FlagsLowConfidenceColumns(t *testing.T) {
	headers := []string{"Case", "Result Signal", "Current State", "Notes"}
	rows := [][]string{
		{"A-1", "User should see dashboard", "open", "manual check"},
		{"A-2", "User should log out", "done", ""},
	}
	columnMapping := map[string]string{
		"Result Signal": "expected",
		"Current State": "status",
	}
	unmapped := []string{"Case", "Notes"}

	quality := BuildPreviewMappingQuality(42, headers, rows, columnMapping, unmapped)

	// With 1 core field (expected) and MappedRatio 0.5, we recommend spec (no fallback)
	if quality.RecommendedFormat != "spec" {
		t.Fatalf("expected spec recommendation (core field + reasonable ratio), got %q", quality.RecommendedFormat)
	}
	if len(quality.LowConfidenceColumns) == 0 {
		t.Fatalf("expected low confidence columns to be populated")
	}
	if quality.ColumnConfidence["Case"] != 0 {
		t.Fatalf("expected unmapped column confidence to be 0")
	}
	if len(quality.ColumnReasons["Case"]) == 0 {
		t.Fatalf("expected reasons for low-confidence unmapped column")
	}
}
