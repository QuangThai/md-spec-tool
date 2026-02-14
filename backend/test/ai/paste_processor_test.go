package ai_test

import (
	"context"
	. "github.com/yourorg/md-spec-tool/internal/ai"
	"testing"
)

func TestPasteProcessor_QuickDetectTSV(t *testing.T) {
	processor := NewPasteProcessorForTest(DefaultConfig())

	content := "ID\tName\tStatus\n1\tTest1\tActive\n2\tTest2\tInactive"
	result := processor.QuickTableDetect(content)

	if result == nil {
		t.Fatal("expected TSV to be detected, got nil")
	}

	if result.DetectedFormat != "tsv" {
		t.Errorf("expected format 'tsv', got %q", result.DetectedFormat)
	}

	if result.InputType != "table" {
		t.Errorf("expected input_type 'table', got %q", result.InputType)
	}

	if result.Confidence != 0.95 {
		t.Errorf("expected confidence 0.95, got %v", result.Confidence)
	}

	if len(result.NormalizedTable) != 3 {
		t.Errorf("expected 3 rows, got %d", len(result.NormalizedTable))
	}
}

func TestPasteProcessor_QuickDetectMarkdownTable(t *testing.T) {
	processor := NewPasteProcessorForTest(DefaultConfig())

	content := `| ID | Name | Status |
| --- | --- | --- |
| 1 | Test1 | Active |
| 2 | Test2 | Inactive |`

	result := processor.QuickTableDetect(content)

	if result == nil {
		t.Fatal("expected markdown table to be detected, got nil")
	}

	if result.DetectedFormat != "markdown_table" {
		t.Errorf("expected format 'markdown_table', got %q", result.DetectedFormat)
	}

	if len(result.NormalizedTable) < 2 {
		t.Errorf("expected at least 2 rows, got %d", len(result.NormalizedTable))
	}

	headers := result.NormalizedTable[0]
	if len(headers) != 3 {
		t.Errorf("expected 3 headers, got %d", len(headers))
	}
}

func TestPasteProcessor_QuickDetectCSV(t *testing.T) {
	processor := NewPasteProcessorForTest(DefaultConfig())

	content := "ID,Name,Status\n1,Test1,Active\n2,Test2,Inactive"
	result := processor.QuickTableDetect(content)

	if result == nil {
		t.Fatal("expected CSV to be detected, got nil")
	}

	if result.DetectedFormat != "csv" {
		t.Errorf("expected format 'csv', got %q", result.DetectedFormat)
	}

	if len(result.NormalizedTable) != 3 {
		t.Errorf("expected 3 rows, got %d", len(result.NormalizedTable))
	}
}

func TestPasteProcessor_NoTableFormat(t *testing.T) {
	processor := NewPasteProcessorForTest(DefaultConfig())

	// Plain text without table delimiters
	content := "This is just a single line of text"
	result := processor.QuickTableDetect(content)

	if result != nil {
		t.Errorf("expected nil for non-table content, got %v", result)
	}
}

func TestParseMarkdownRow(t *testing.T) {
	tests := []struct {
		line     string
		expected []string
	}{
		{
			line:     "| ID | Name | Status |",
			expected: []string{"ID", "Name", "Status"},
		},
		{
			line:     "|TC-001|Test Case Name|Passed|",
			expected: []string{"TC-001", "Test Case Name", "Passed"},
		},
		{
			line:     "| --- | --- | --- |",
			expected: []string{"---", "---", "---"},
		},
	}

	for _, tt := range tests {
		got := ParseMarkdownRow(tt.line)
		if len(got) != len(tt.expected) {
			t.Errorf("ParseMarkdownRow(%q) returned %d cells, want %d", tt.line, len(got), len(tt.expected))
			continue
		}

		for i, cell := range got {
			if cell != tt.expected[i] {
				t.Errorf("ParseMarkdownRow(%q) cell %d = %q, want %q", tt.line, i, cell, tt.expected[i])
			}
		}
	}
}

func TestDetectCSVReliability(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		minConf  float64
		expected bool
	}{
		{
			name:     "valid CSV",
			lines:    []string{"ID,Name,Status", "1,Test1,Active", "2,Test2,Inactive"},
			minConf:  0.8,
			expected: true,
		},
		{
			name:     "single line",
			lines:    []string{"ID,Name,Status"},
			minConf:  0.8,
			expected: false,
		},
		{
			name:     "inconsistent columns",
			lines:    []string{"ID,Name,Status", "1,Test1", "2,Test2,Inactive,Extra"},
			minConf:  0.8,
			expected: false,
		},
		{
			name:     "no commas",
			lines:    []string{"This is plain text", "without any commas"},
			minConf:  0.8,
			expected: false,
		},
	}

	for _, tt := range tests {
		got := DetectCSVReliability(tt.lines)
		isReliable := got >= tt.minConf
		if isReliable != tt.expected {
			t.Errorf("DetectCSVReliability(%q) = %.2f (reliable=%v), want reliable=%v", tt.name, got, isReliable, tt.expected)
		}
	}
}

func TestPasteProcessor_GetNormalizedTable(t *testing.T) {
	processor := &PasteProcessorService{}

	analysis := &PasteAnalysis{
		NormalizedTable: [][]string{
			{"ID", "Name", "Status"},
			{"1", "Test1", "Active"},
		},
	}

	table := processor.GetNormalizedTable(analysis)
	if len(table) != 2 {
		t.Errorf("expected 2 rows, got %d", len(table))
	}
}

func TestPasteProcessor_GetHeaders(t *testing.T) {
	processor := &PasteProcessorService{}

	analysis := &PasteAnalysis{
		NormalizedTable: [][]string{
			{"ID", "Name", "Status"},
			{"1", "Test1", "Active"},
		},
	}

	headers := processor.GetHeaders(analysis)
	if len(headers) != 3 {
		t.Errorf("expected 3 headers, got %d", len(headers))
	}

	if headers[0] != "ID" {
		t.Errorf("expected first header 'ID', got %q", headers[0])
	}
}

func TestPasteProcessor_GetDataRows(t *testing.T) {
	processor := &PasteProcessorService{}

	analysis := &PasteAnalysis{
		NormalizedTable: [][]string{
			{"ID", "Name", "Status"},
			{"1", "Test1", "Active"},
			{"2", "Test2", "Inactive"},
		},
	}

	rows := processor.GetDataRows(analysis)
	if len(rows) != 2 {
		t.Errorf("expected 2 data rows, got %d", len(rows))
	}

	if rows[0][0] != "1" {
		t.Errorf("expected first data row ID '1', got %q", rows[0][0])
	}
}

func TestPasteProcessor_ComplexMarkdownTable(t *testing.T) {
	processor := NewPasteProcessorForTest(DefaultConfig())

	content := `| TC ID | Test Case Name | Expected Result | Status |
| --- | --- | --- | --- |
| TC-001 | Login with valid credentials | User logged in successfully | Pass |
| TC-002 | Login with invalid credentials | Error message displayed | Pass |`

	result := processor.QuickTableDetect(content)

	if result == nil {
		t.Fatal("expected markdown table to be detected")
	}

	headers := result.NormalizedTable[0]
	if len(headers) != 4 {
		t.Errorf("expected 4 headers, got %d", len(headers))
	}

	if headers[0] != "TC ID" {
		t.Errorf("expected first header 'TC ID', got %q", headers[0])
	}

	dataRows := result.NormalizedTable[1:]
	if len(dataRows) != 2 {
		t.Errorf("expected 2 data rows, got %d", len(dataRows))
	}
}

func TestPasteProcessor_AnalyzePasteWithAIService(t *testing.T) {
	mockService := &mockAIService{
		analyzePasteFunc: func(ctx context.Context, req AnalyzePasteRequest) (*PasteAnalysis, error) {
			return &PasteAnalysis{
				SchemaVersion:   SchemaVersionPasteAnalysis,
				InputType:       "test_cases",
				DetectedFormat:  "markdown_table",
				SuggestedOutput: "spec",
				Confidence:      0.9,
			}, nil
		},
	}

	cache := NewCache(100, 0)
	validator := NewValidator()
	config := DefaultConfig()

	processor := NewPasteProcessorService(mockService, cache, validator, config)

	// Use prose content (not a markdown table) to bypass quick detection
	result, err := processor.AnalyzePaste(context.Background(), AnalyzePasteRequest{
		Content: "Write me test cases for login functionality",
		MaxSize: 1000,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.InputType != "test_cases" {
		t.Errorf("expected input_type 'test_cases', got %q", result.InputType)
	}

	if result.Confidence != 0.9 {
		t.Errorf("expected confidence 0.9, got %v", result.Confidence)
	}
}

func TestPasteProcessor_TSVWithEmptyLines(t *testing.T) {
	processor := NewPasteProcessorForTest(DefaultConfig())

	content := "ID\tName\tStatus\n1\tTest1\tActive\n\n2\tTest2\tInactive"
	result := processor.QuickTableDetect(content)

	if result == nil {
		t.Fatal("expected TSV to be detected despite empty lines")
	}

	// Should skip empty lines
	if len(result.NormalizedTable) != 3 {
		t.Errorf("expected 3 rows (header + 2 data), got %d", len(result.NormalizedTable))
	}
}
