package handlers_test

import (
	"context"
	. "github.com/yourorg/md-spec-tool/internal/http/handlers"
	"testing"

	"github.com/yourorg/md-spec-tool/internal/converter"
)

func newValidationTestHandler(strict bool) *MDFlowHandler {
	return NewValidationTestHandler(strict, 50, 0.5)
}

func TestNormalizeSheetRange(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		override string
		expected string
	}{
		{
			name:     "empty range uses whole sheet",
			title:    "Sheet1",
			override: "",
			expected: "Sheet1",
		},
		{
			name:     "a1 range is scoped to selected sheet",
			title:    "Popular Videos",
			override: "A1:M7",
			expected: "'Popular Videos'!A1:M7",
		},
		{
			name:     "fully qualified range is preserved",
			title:    "Sheet1",
			override: "Other!B2:D10",
			expected: "Other!B2:D10",
		},
		{
			name:     "bang-prefixed range is scoped to selected sheet",
			title:    "Data",
			override: "!A1:C5",
			expected: "Data!A1:C5",
		},
		{
			name:     "sheet title with apostrophe is escaped",
			title:    "Jon's Data",
			override: "A1:B2",
			expected: "'Jon''s Data'!A1:B2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeSheetRange(tt.title, tt.override)
			if got != tt.expected {
				t.Fatalf("NormalizeSheetRange() = %q, expected %q", got, tt.expected)
			}
		})
	}
}

func TestSelectPreferredBlockMatrix_PrefersMultiRowBlock(t *testing.T) {
	conv := converter.NewConverter()
	matrix := converter.NewCellMatrix([][]string{
		{"画面名", "概要", "", "Item Name", "Overview"},
		{"人気動画一覧", "説明", "", "Video List", "Shows videos"},
		{"", "", "", "Search", "Search videos"},
	}).Normalize()

	selected := SelectPreferredBlockMatrix(context.Background(), conv, matrix, "spec")
	if len(selected) == 0 || len(selected[0]) == 0 {
		t.Fatalf("selected matrix is empty")
	}

	if got := selected[0][0]; got != "Item Name" {
		t.Fatalf("expected multi-row block to be selected, got first header %q", got)
	}
}

func TestParseA1Start(t *testing.T) {
	tests := []struct {
		name     string
		rangeStr string
		wantCol  int
		wantRow  int
	}{
		{name: "quoted sheet range", rangeStr: "'登録チャンネル_一覧A004'!A26:I31", wantCol: 0, wantRow: 25},
		{name: "plain sheet range", rangeStr: "Sheet1!C10:K99", wantCol: 2, wantRow: 9},
		{name: "single cell", rangeStr: "B7", wantCol: 1, wantRow: 6},
		{name: "invalid", rangeStr: "Sheet1", wantCol: 0, wantRow: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCol, gotRow := ParseA1Start(tt.rangeStr)
			if gotCol != tt.wantCol || gotRow != tt.wantRow {
				t.Fatalf("ParseA1Start(%q) = (%d,%d), want (%d,%d)", tt.rangeStr, gotCol, gotRow, tt.wantCol, tt.wantRow)
			}
		})
	}
}

func TestSelectMatrixForConvert_UsesRangeMatrixWhenBlockIDMissing(t *testing.T) {
	conv := converter.NewConverter()
	matrix := converter.NewCellMatrix([][]string{
		{"A", "B", "", "X", "Y"},
		{"1", "2", "", "x1", "y1"},
		{"3", "4", "", "x2", "y2"},
	}).Normalize()

	selected := SelectMatrixForConvert(context.Background(), conv, matrix, "spec", "block_99", "Sheet1!A1:E3")
	if len(selected) != len(matrix) || len(selected[0]) != len(matrix[0]) {
		t.Fatalf("expected original matrix to be used when range is explicitly provided")
	}
}

func TestBuildConvertValidationError_LowHeaderConfidence(t *testing.T) {
	h := newValidationTestHandler(true)
	stats := ConvertValidationStats{SourceRows: 4, HeaderRow: 0, HeaderConfidence: 12}
	result := &converter.ConvertResponse{Meta: converter.SpecDocMeta{TotalRows: 4}}
	err := BuildConvertValidationError(h, string(converter.OutputFormatSpec), stats, result)
	if err == nil {
		t.Fatalf("expected validation error for low header confidence")
	}
	if err.ValidationReason != "low_header_confidence" {
		t.Fatalf("unexpected validation reason: %s", err.ValidationReason)
	}
}

func TestBuildConvertValidationError_RowLoss(t *testing.T) {
	h := newValidationTestHandler(true)
	stats := ConvertValidationStats{SourceRows: 6, HeaderRow: 0, HeaderConfidence: 90}
	result := &converter.ConvertResponse{Meta: converter.SpecDocMeta{TotalRows: 2}}
	err := BuildConvertValidationError(h, string(converter.OutputFormatSpec), stats, result)
	if err == nil {
		t.Fatalf("expected validation error for severe row loss")
	}
	if err.ValidationReason != "row_loss" {
		t.Fatalf("unexpected validation reason: %s", err.ValidationReason)
	}
}

func TestBuildConvertValidationError_TableFormatSkipsValidation(t *testing.T) {
	h := newValidationTestHandler(true)
	stats := ConvertValidationStats{SourceRows: 6, HeaderRow: 0, HeaderConfidence: 10}
	result := &converter.ConvertResponse{Meta: converter.SpecDocMeta{TotalRows: 1}}
	err := BuildConvertValidationError(h, string(converter.OutputFormatTable), stats, result)
	if err != nil {
		t.Fatalf("expected no validation error for table format")
	}
}

func TestBuildConvertValidationError_StrictModeDisabled(t *testing.T) {
	h := newValidationTestHandler(false)
	stats := ConvertValidationStats{SourceRows: 6, HeaderRow: 0, HeaderConfidence: 10}
	result := &converter.ConvertResponse{Meta: converter.SpecDocMeta{TotalRows: 1}}
	err := BuildConvertValidationError(h, string(converter.OutputFormatSpec), stats, result)
	if err != nil {
		t.Fatalf("expected no validation error when strict mode is disabled")
	}
}
