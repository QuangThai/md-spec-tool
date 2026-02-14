package converter_test

import (
	. "github.com/yourorg/md-spec-tool/internal/converter"
	"testing"
)

func TestEvaluateMappingQuality_SpecTableSchema(t *testing.T) {
	// Spec-table headers: No, Item Name, Item Type, Display Conditions, Action, Navigation Dest
	headers := []string{"No", "Item Name", "Item Type", "Required/Optional", "Display Conditions", "Action", "Navigation Destination"}
	colMap := ColumnMap{
		FieldNo:                0,
		FieldItemName:          1,
		FieldItemType:          2,
		FieldRequiredOptional:  3,
		FieldDisplayConditions: 4,
		FieldAction:            5,
		FieldNavigationDest:    6,
	}

	quality := EvaluateMappingQuality(80, headers, colMap)

	// Should have high core coverage from spec-table schema (6 of 6 spec-table fields)
	if quality.CoreMapped != 6 {
		t.Errorf("expected CoreMapped=6 for spec-table, got %d", quality.CoreMapped)
	}
	if quality.CoreCoverage < 0.9 {
		t.Errorf("expected CoreCoverage >= 0.9, got %f", quality.CoreCoverage)
	}
}

func TestShouldFallbackToTable_SpecTableNoFallback(t *testing.T) {
	// Spec-table with good mapping: should NOT fallback
	quality := MappingQuality{
		Score:        0.55,
		MappedRatio:  0.8,
		CoreMapped:   6,
		CoreCoverage: 1.0,
	}

	if ShouldFallbackToTable("spec", quality) {
		t.Error("expected no fallback for spec-table with good mapping")
	}
}

func TestShouldFallbackToTable_NoCoreFieldsFallback(t *testing.T) {
	// No core fields: should fallback
	quality := MappingQuality{
		Score:        0.3,
		MappedRatio:  0.2,
		CoreMapped:   0,
		CoreCoverage: 0,
	}

	if !ShouldFallbackToTable("spec", quality) {
		t.Error("expected fallback when no core fields mapped")
	}
}
