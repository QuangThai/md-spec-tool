package converter

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// ---------------------------------------------------------------------------
// SanitizeHeaders — edge cases
// ---------------------------------------------------------------------------

func TestSanitizeHeaders_EmptySlice(t *testing.T) {
	result := SanitizeHeaders([]string{})
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(result))
	}
}

func TestSanitizeHeaders_AllWhitespace(t *testing.T) {
	input := []string{"   ", "\t\t", "\n", "  \t  "}
	result := SanitizeHeaders(input)
	for i, h := range result {
		if h != "" {
			t.Errorf("index %d: expected empty string after trimming all-whitespace, got %q", i, h)
		}
	}
}

func TestSanitizeHeaders_MixedWhitespaceAndContent(t *testing.T) {
	input := []string{"\t  Name  \n", "  ", "   Value"}
	result := SanitizeHeaders(input)
	expected := []string{"Name", "", "Value"}
	for i, want := range expected {
		if result[i] != want {
			t.Errorf("index %d: expected %q, got %q", i, want, result[i])
		}
	}
}

func TestSanitizeHeaders_ExactlyAtColumnLimit(t *testing.T) {
	input := make([]string, MaxColumnCount)
	for i := range input {
		input[i] = "col"
	}
	result := SanitizeHeaders(input)
	if len(result) != MaxColumnCount {
		t.Errorf("expected exactly %d columns, got %d", MaxColumnCount, len(result))
	}
}

// ---------------------------------------------------------------------------
// SanitizeCellContent — edge cases
// ---------------------------------------------------------------------------

func TestSanitizeCellContent_EmptyString(t *testing.T) {
	if got := SanitizeCellContent(""); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestSanitizeCellContent_AllWhitespace(t *testing.T) {
	for _, input := range []string{"   ", "\t\t", "\n\r\n", "  \t  "} {
		if got := SanitizeCellContent(input); got != "" {
			t.Errorf("SanitizeCellContent(%q): expected empty after trim, got %q", input, got)
		}
	}
}

func TestSanitizeCellContent_ControlCharactersOnly(t *testing.T) {
	// ASCII control chars (\x01-\x1f) after NFKC + TrimSpace should collapse to ""
	input := "\x01\x02\x03\x04\x1f"
	got := SanitizeCellContent(input)
	// After NFKC normalization and TrimSpace, the result may still contain
	// low-ASCII controls (they are valid UTF-8). Verify it doesn't panic and
	// does not exceed MaxCellLength.
	if utf8.RuneCountInString(got) > MaxCellLength {
		t.Errorf("result exceeds MaxCellLength after control-char input")
	}
}

func TestSanitizeCellContent_LeadingAndTrailingWhitespaceStripped(t *testing.T) {
	input := "  hello world  "
	if got := SanitizeCellContent(input); got != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", got)
	}
}

func TestSanitizeCellContent_VeryLongUnicodeString(t *testing.T) {
	// Build a string longer than MaxCellLength using multi-byte CJK characters
	var sb strings.Builder
	// U+4E00 = '一' (CJK, 3 bytes in UTF-8 but 1 rune)
	for i := 0; i < MaxCellLength+500; i++ {
		sb.WriteRune('一')
	}
	result := SanitizeCellContent(sb.String())
	runeCount := utf8.RuneCountInString(result)
	// Must be exactly MaxCellLength runes + 3 for "..."
	if runeCount != MaxCellLength+3 {
		t.Errorf("expected %d runes (MaxCellLength+3), got %d", MaxCellLength+3, runeCount)
	}
	if !strings.HasSuffix(result, "...") {
		t.Error("expected truncated string to end with '...'")
	}
}

func TestSanitizeCellContent_VeryLongASCIIString(t *testing.T) {
	longString := strings.Repeat("x", MaxCellLength*2)
	result := SanitizeCellContent(longString)
	runeCount := utf8.RuneCountInString(result)
	if runeCount != MaxCellLength+3 {
		t.Errorf("expected %d chars after truncation, got %d", MaxCellLength+3, runeCount)
	}
}

func TestSanitizeCellContent_ExactlyAtMaxLength(t *testing.T) {
	// A string exactly MaxCellLength runes long should NOT be truncated
	exact := strings.Repeat("a", MaxCellLength)
	result := SanitizeCellContent(exact)
	if result != exact {
		t.Errorf("expected string of exactly MaxCellLength to be preserved, got length %d", utf8.RuneCountInString(result))
	}
	if strings.HasSuffix(result, "...") {
		t.Error("expected no truncation ellipsis for string at exact max length")
	}
}

func TestSanitizeCellContent_UnicodeNormalization(t *testing.T) {
	// Fullwidth digit '１' (U+FF11) should be normalized to '1' (ASCII) under NFKC
	input := "Test\uff11\uff12\uff13"
	result := SanitizeCellContent(input)
	if strings.Contains(result, "\uff11") {
		t.Error("expected fullwidth digits to be normalized to ASCII under NFKC")
	}
	if !strings.Contains(result, "123") {
		t.Errorf("expected '123' after NFKC normalization, got %q", result)
	}
}

func TestSanitizeCellContent_SpecialCharactersPreserved(t *testing.T) {
	// Special printable characters should survive unchanged
	input := "Hello, <World> & \"Friends\" | 100%"
	result := SanitizeCellContent(input)
	if result != input {
		t.Errorf("expected special chars preserved, got %q", result)
	}
}

// ---------------------------------------------------------------------------
// SanitizeSampleRows — edge cases
// ---------------------------------------------------------------------------

func TestSanitizeSampleRows_SingleRow(t *testing.T) {
	input := [][]string{{"alpha", "beta"}}
	result := SanitizeSampleRows(input)
	if len(result) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result))
	}
	if result[0][0] != "alpha" || result[0][1] != "beta" {
		t.Errorf("expected [alpha beta], got %v", result[0])
	}
}

func TestSanitizeSampleRows_TwoRows(t *testing.T) {
	input := [][]string{{"r0"}, {"r1"}}
	result := SanitizeSampleRows(input)
	if len(result) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result))
	}
}

func TestSanitizeSampleRows_ExactlyMaxRows(t *testing.T) {
	input := make([][]string, MaxSampleRows)
	for i := range input {
		input[i] = []string{"cell"}
	}
	result := SanitizeSampleRows(input)
	if len(result) != MaxSampleRows {
		t.Errorf("expected %d rows, got %d", MaxSampleRows, len(result))
	}
}

func TestSanitizeSampleRows_SanitizesCellContents(t *testing.T) {
	// Whitespace in cells must be trimmed
	input := [][]string{{"  padded  "}}
	result := SanitizeSampleRows(input)
	if result[0][0] != "padded" {
		t.Errorf("expected cell content trimmed, got %q", result[0][0])
	}
}

func TestSanitizeSampleRows_EmptyRows(t *testing.T) {
	// Should not panic on empty outer slice
	result := SanitizeSampleRows([][]string{})
	if len(result) != 0 {
		t.Errorf("expected empty result for empty input, got %d", len(result))
	}
}

func TestSanitizeSampleRows_LargeLongCells(t *testing.T) {
	// Each cell is 2000 chars; must be truncated by SanitizeCellContent
	longCell := strings.Repeat("z", 2000)
	input := [][]string{{longCell, longCell}}
	result := SanitizeSampleRows(input)
	for _, cell := range result[0] {
		if utf8.RuneCountInString(cell) > MaxCellLength+3 {
			t.Errorf("expected cell truncated to MaxCellLength+3, got %d runes", utf8.RuneCountInString(cell))
		}
	}
}

// ---------------------------------------------------------------------------
// NormalizeUnicode — idempotency and roundtrip
// ---------------------------------------------------------------------------

func TestNormalizeUnicode_Idempotent(t *testing.T) {
	// Applying NFKC twice must yield the same result as once
	input := "Héllo\uff01 Wörld"
	once := NormalizeUnicode(input)
	twice := NormalizeUnicode(once)
	if once != twice {
		t.Errorf("NormalizeUnicode is not idempotent: first=%q, second=%q", once, twice)
	}
}

func TestNormalizeUnicode_EmptyString(t *testing.T) {
	if got := NormalizeUnicode(""); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestNormalizeUnicode_PureASCII(t *testing.T) {
	input := "Hello, World! 123"
	if got := NormalizeUnicode(input); got != input {
		t.Errorf("pure ASCII should be unchanged, got %q", got)
	}
}
