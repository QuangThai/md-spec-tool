package converter

import (
	"testing"
)

func TestFallbackMapper_MapColumns_BasicHeaders(t *testing.T) {
	mapper := NewFallbackMapper()

	headers := []string{"TC ID", "Test Name", "Expected Result", "Status"}
	colMap, unmapped := mapper.MapColumns(headers)

	// Verify mappings
	expectedMappings := map[CanonicalField]int{
		FieldID:       0, // TC ID
		FieldScenario: 1, // Test Name (test_name synonym)
		FieldExpected: 2, // Expected Result
		FieldStatus:   3, // Status
	}

	for field, expectedIdx := range expectedMappings {
		if idx, ok := colMap[field]; !ok {
			t.Errorf("expected field %q to be mapped, but it wasn't", field)
		} else if idx != expectedIdx {
			t.Errorf("field %q mapped to index %d, expected %d", field, idx, expectedIdx)
		}
	}

	if len(unmapped) != 0 {
		t.Errorf("expected no unmapped headers, got %d: %v", len(unmapped), unmapped)
	}
}

func TestFallbackMapper_MapColumns_UnmappedHeaders(t *testing.T) {
	mapper := NewFallbackMapper()

	headers := []string{"ID", "Title", "Unknown Field 1", "Unknown Field 2", "Status"}
	colMap, unmapped := mapper.MapColumns(headers)

	if len(unmapped) != 2 {
		t.Errorf("expected 2 unmapped headers, got %d", len(unmapped))
	}

	if len(colMap) != 3 {
		t.Errorf("expected 3 mapped fields, got %d", len(colMap))
	}

	// Verify unmapped headers
	if len(unmapped) > 0 && unmapped[0] != "Unknown Field 1" {
		t.Errorf("expected first unmapped to be 'Unknown Field 1', got %q", unmapped[0])
	}
}

func TestFallbackMapper_MapColumns_CaseInsensitivity(t *testing.T) {
	mapper := NewFallbackMapper()

	headers := []string{"tc id", "test name", "expected result", "status"}
	colMap, _ := mapper.MapColumns(headers)

	// All headers should map correctly despite case differences
	if len(colMap) != 4 {
		t.Errorf("expected 4 mapped fields, got %d", len(colMap))
	}
}

func TestFallbackMapper_MapColumns_MultiLanguageHeaders(t *testing.T) {
	mapper := NewFallbackMapper()

	// Test with Japanese headers that are in the HeaderSynonyms
	headers := []string{"番号", "項目名", "種別", "remarks"}
	colMap, _ := mapper.MapColumns(headers)

	expectedMappings := map[CanonicalField]bool{
		FieldNo:        true, // 番号
		FieldItemName:  true, // 項目名
		FieldItemType:  true, // 種別
		FieldNotes:     true, // remarks (maps to notes)
	}

	for field, shouldExist := range expectedMappings {
		_, exists := colMap[field]
		if exists != shouldExist {
			t.Errorf("field %q existence: got %v, want %v", field, exists, shouldExist)
		}
	}
}

func TestFallbackMapper_MapColumns_DuplicateHeaders(t *testing.T) {
	mapper := NewFallbackMapper()

	// When same canonical field appears multiple times, first occurrence wins
	headers := []string{"ID", "Reference", "Test Case Name", "Status", "State"}
	colMap, _ := mapper.MapColumns(headers)

	// FieldID should map to first occurrence (index 0)
	if idx, ok := colMap[FieldID]; !ok {
		t.Error("expected FieldID to be mapped")
	} else if idx != 0 {
		t.Errorf("FieldID mapped to index %d, expected 0", idx)
	}

	// FieldStatus should map to first occurrence (index 3, not 4 which is "State")
	if idx, ok := colMap[FieldStatus]; !ok {
		t.Error("expected FieldStatus to be mapped")
	} else if idx != 3 {
		t.Errorf("FieldStatus mapped to index %d, expected 3", idx)
	}
}

func TestFallbackMapper_GetUnmappedHeaders(t *testing.T) {
	mapper := NewFallbackMapper()

	headers := []string{"ID", "Unknown1", "Title", "Unknown2"}
	colMap, _ := mapper.MapColumns(headers)

	unmapped := mapper.GetUnmappedHeaders(headers, colMap)

	if len(unmapped) != 2 {
		t.Errorf("expected 2 unmapped headers, got %d", len(unmapped))
	}

	expectedUnmapped := map[string]bool{
		"Unknown1": true,
		"Unknown2": true,
	}

	for _, h := range unmapped {
		if !expectedUnmapped[h] {
			t.Errorf("unexpected unmapped header: %q", h)
		}
	}
}

func TestFallbackMapper_WithCustomSynonyms(t *testing.T) {
	mapper := NewFallbackMapper()

	customSynonyms := map[string]CanonicalField{
		"custom_field": FieldNotes,
	}

	mapper.WithCustomSynonyms(customSynonyms)

	headers := []string{"custom_field"}
	colMap, unmapped := mapper.MapColumns(headers)

	if len(unmapped) != 0 {
		t.Errorf("expected custom synonym to be recognized, but got unmapped: %v", unmapped)
	}

	if idx, ok := colMap[FieldNotes]; !ok {
		t.Error("expected custom field to map to FieldNotes")
	} else if idx != 0 {
		t.Errorf("custom field mapped to index %d, expected 0", idx)
	}
}

func TestFallbackMapper_Whitespace_Normalization(t *testing.T) {
	mapper := NewFallbackMapper()

	headers := []string{"  TC  ID  ", "test  name", "expected  result"}
	colMap, unmapped := mapper.MapColumns(headers)

	// All headers should be normalized and mapped correctly
	if len(unmapped) != 0 {
		t.Errorf("expected no unmapped headers after whitespace normalization, got: %v", unmapped)
	}

	if len(colMap) != 3 {
		t.Errorf("expected 3 mapped fields, got %d", len(colMap))
	}
}

func TestFallbackMapper_SpecTableFields(t *testing.T) {
	mapper := NewFallbackMapper()

	// Test spec table specific fields (Phase 3)
	headers := []string{"No", "Item Name", "Item Type", "Required/Optional", "Input Restrictions"}
	colMap, unmapped := mapper.MapColumns(headers)

	expectedFields := []CanonicalField{
		FieldNo,
		FieldItemName,
		FieldItemType,
		FieldRequiredOptional,
		FieldInputRestrictions,
	}

	for _, field := range expectedFields {
		if _, ok := colMap[field]; !ok {
			t.Errorf("expected field %q to be mapped", field)
		}
	}

	if len(unmapped) != 0 {
		t.Errorf("expected no unmapped headers, got: %v", unmapped)
	}
}

func TestFallbackMapper_EmptyHeaders(t *testing.T) {
	mapper := NewFallbackMapper()

	headers := []string{}
	colMap, unmapped := mapper.MapColumns(headers)

	if len(colMap) != 0 {
		t.Errorf("expected empty colMap for empty headers, got %d fields", len(colMap))
	}

	if len(unmapped) != 0 {
		t.Errorf("expected no unmapped for empty headers, got %d", len(unmapped))
	}
}
