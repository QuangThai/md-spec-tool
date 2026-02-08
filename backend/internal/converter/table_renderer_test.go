package converter

import (
	"strings"
	"testing"
)

func TestTableRenderer_Render(t *testing.T) {
	renderer := NewTableRenderer()

	rows := []TableRow{
		{Cells: []string{"TC-001", "Login", "high", "active"}},
		{Cells: []string{"TC-002", "Logout", "medium", "active"}},
	}

	table := NewTable("Simple Data Table", []string{"id", "feature", "priority", "status"}, rows)

	output, _, err := renderer.Render(table)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify title
	if !strings.Contains(output, "# Simple Data Table") {
		t.Error("Output should contain title")
	}

	// Verify table structure
	if !strings.Contains(output, "| id |") {
		t.Error("Output should contain id header")
	}
	if !strings.Contains(output, "| feature |") {
		t.Error("Output should contain feature header")
	}

	// Verify separator row
	if !strings.Contains(output, " --- |") {
		t.Error("Output should contain separator row with ---")
	}

	// Verify data
	if !strings.Contains(output, "TC-001") {
		t.Error("Output should contain first test case ID")
	}
	if !strings.Contains(output, "TC-002") {
		t.Error("Output should contain second test case ID")
	}
	if !strings.Contains(output, "Login") {
		t.Error("Output should contain feature")
	}
	if !strings.Contains(output, "high") {
		t.Error("Output should contain priority")
	}
}

func TestTableRenderer_InferHeaders(t *testing.T) {
	renderer := NewTableRenderer()

	rows := []SpecRow{
		{
			ID:           "1",
			Feature:      "F1",
			Type:         "test",
			Priority:     "high",
			Status:       "active",
			Instructions: "steps",
			Expected:     "result",
		},
		{
			Feature:  "F2",
			Type:     "bug",
			Priority: "low",
		},
	}

	headers := renderer.inferHeaders(rows)

	// Should infer common fields
	expectedFields := map[string]bool{
		"id":           true,
		"feature":      true,
		"type":         true,
		"priority":     true,
		"status":       true,
		"instructions": true,
		"expected":     true,
	}

	for _, h := range headers {
		if !expectedFields[h] {
			t.Errorf("Unexpected header: %s", h)
		}
	}
}

func TestTableRenderer_GetCellValue(t *testing.T) {
	renderer := NewTableRenderer()

	row := SpecRow{
		ID:       "TC-001",
		Feature:  "Login",
		Scenario: "Valid credentials",
		Type:     "test",
		Priority: "high",
		Status:   "active",
		Notes:    "Important test",
	}

	tests := []struct {
		header   string
		expected string
	}{
		{"id", "TC-001"},
		{"ID", "TC-001"},
		{"feature", "Login"},
		{"type", "test"},
		{"priority", "high"},
		{"status", "active"},
		{"notes", "Important test"},
	}

	for _, tt := range tests {
		result := renderer.getCellValue(row, tt.header)
		if result != tt.expected {
			t.Errorf("Header %q: expected %q, got %q", tt.header, tt.expected, result)
		}
	}
}

func TestTableRenderer_TruncateLongValues(t *testing.T) {
	renderer := NewTableRenderer()

	longValue := strings.Repeat("x", 150)

	rows := []TableRow{
		{Cells: []string{longValue}},
	}

	table := NewTable("Table", []string{"instructions"}, rows)

	output, _, err := renderer.Render(table)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify long value is preserved
	if !strings.Contains(output, longValue) {
		t.Error("Output should preserve very long values")
	}
}

func TestTableRenderer_EmptyInput(t *testing.T) {
	renderer := NewTableRenderer()

	table := NewTable("Empty Table", []string{"col1"}, []TableRow{})

	output, _, err := renderer.Render(table)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(output, "# Empty Table") {
		t.Error("Should contain title")
	}
	if !strings.Contains(output, "No data") {
		t.Error("Should show 'No data' message for empty rows")
	}
}

func TestTableRenderer_SpecialCharacters(t *testing.T) {
	renderer := NewTableRenderer()

	rows := []TableRow{
		{Cells: []string{"Feature | with | pipes", "Test\nwith\nnewlines"}},
	}

	table := NewTable("Special Chars", []string{"col | with | pipes", "col\nwith\nnewlines"}, rows)

	output, _, err := renderer.Render(table)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Pipes should be escaped
	if !strings.Contains(output, "\\|") {
		t.Error("Should escape pipe characters in table")
	}
}
