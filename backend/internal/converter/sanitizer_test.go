package converter

import (
	"fmt"
	"testing"
)

func TestSanitizeHeaders_TrimsWhitespace(t *testing.T) {
	input := []string{"  ID  ", "\tTitle\n", "  Description  "}
	result := SanitizeHeaders(input)
	expected := []string{"ID", "Title", "Description"}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("index %d: expected %q, got %q", i, expected[i], v)
		}
	}
}

func TestSanitizeHeaders_LimitsColumnCount(t *testing.T) {
	input := make([]string, 60)
	for i := range input {
		input[i] = fmt.Sprintf("col_%d", i)
	}
	result := SanitizeHeaders(input)
	if len(result) > MaxColumnCount {
		t.Errorf("expected max %d columns, got %d", MaxColumnCount, len(result))
	}
}

func TestSanitizeHeaders_EmptyInput(t *testing.T) {
	result := SanitizeHeaders(nil)
	if len(result) != 0 {
		t.Errorf("expected empty result for nil input, got %d", len(result))
	}
}

func TestSanitizeSampleRows_LimitsRowCount(t *testing.T) {
	input := make([][]string, 200)
	for i := range input {
		input[i] = []string{"data"}
	}
	result := SanitizeSampleRows(input)
	if len(result) > MaxSampleRows {
		t.Errorf("expected max %d rows, got %d", MaxSampleRows, len(result))
	}
}

func TestSanitizeSampleRows_SelectsRepresentativeRows(t *testing.T) {
	input := make([][]string, 20)
	for i := range input {
		input[i] = []string{fmt.Sprintf("row_%d", i)}
	}
	result := SanitizeSampleRows(input)
	// Verify first row included
	if result[0][0] != "row_0" {
		t.Errorf("expected first row, got %s", result[0][0])
	}
	// Verify last row included
	last := result[len(result)-1]
	if last[0] != "row_19" {
		t.Errorf("expected last row, got %s", last[0])
	}
}

func TestSanitizeSampleRows_SmallInput(t *testing.T) {
	input := [][]string{{"a"}, {"b"}}
	result := SanitizeSampleRows(input)
	if len(result) != 2 {
		t.Errorf("expected 2 rows for small input, got %d", len(result))
	}
}

func TestNormalizeUnicode(t *testing.T) {
	// NFKC normalization: fullwidth chars → ASCII equivalents
	input := "\uff0a\uff0a\uff0a fullwidth asterisks"
	result := NormalizeUnicode(input)
	if result == input {
		t.Error("expected unicode normalization to change fullwidth chars")
	}
}

func TestSanitizeCellContent_TruncatesLongCells(t *testing.T) {
	longString := make([]byte, 2000)
	for i := range longString {
		longString[i] = 'a'
	}
	result := SanitizeCellContent(string(longString))
	if len(result) > MaxCellLength+3 { // +3 for "..."
		t.Errorf("expected max %d chars, got %d", MaxCellLength+3, len(result))
	}
}

func TestSanitizeCellContent_PreservesShortStrings(t *testing.T) {
	input := "Hello World"
	result := SanitizeCellContent(input)
	if result != input {
		t.Errorf("expected %q, got %q", input, result)
	}
}
