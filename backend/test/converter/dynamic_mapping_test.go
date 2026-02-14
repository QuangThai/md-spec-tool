package converter_test

import (
	. "github.com/yourorg/md-spec-tool/internal/converter"
	"testing"
)

func TestEnhanceColumnMapping_InferNonStandardHeaders(t *testing.T) {
	headers := []string{"Case Ref", "Flow Name", "Success Signal", "Current State"}
	rows := [][]string{
		{"TC-1", "Login", "User sees dashboard", "open"},
		{"TC-2", "Logout", "User returns to home", "done"},
	}

	colMap, unmapped, warnings := EnhanceColumnMapping(headers, rows, ColumnMap{})

	if _, ok := colMap[FieldExpected]; !ok {
		t.Fatalf("expected dynamic mapping to infer %q", FieldExpected)
	}
	if _, ok := colMap[FieldStatus]; !ok {
		t.Fatalf("expected dynamic mapping to infer %q", FieldStatus)
	}
	if len(unmapped) >= len(headers) {
		t.Fatalf("expected fewer unmapped headers, got %v", unmapped)
	}
	if len(warnings) == 0 {
		t.Fatalf("expected inference warning metadata")
	}
}

func TestShouldFallbackToTable_WhenQualityLow(t *testing.T) {
	headers := []string{"A", "B", "C", "D"}
	quality := EvaluateMappingQuality(20, headers, ColumnMap{FieldEndpoint: 0})

	if !ShouldFallbackToTable("spec", quality) {
		t.Fatalf("expected fallback for low mapping quality: %+v", quality)
	}
}

func TestShouldFallbackToTable_WhenQualityGood(t *testing.T) {
	headers := []string{"Feature", "Scenario", "Steps", "Expected"}
	colMap := ColumnMap{
		FieldFeature:      0,
		FieldScenario:     1,
		FieldInstructions: 2,
		FieldExpected:     3,
	}
	quality := EvaluateMappingQuality(90, headers, colMap)

	if ShouldFallbackToTable("spec", quality) {
		t.Fatalf("did not expect fallback for strong mapping quality: %+v", quality)
	}
}
