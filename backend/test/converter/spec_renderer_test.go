package converter_test

import (
	. "github.com/yourorg/md-spec-tool/internal/converter"
	"strings"
	"testing"
)

func TestSpecRenderer_Render(t *testing.T) {
	renderer := NewSpecRenderer()

	rows := []TableRow{
		{Cells: []string{"TC-001", "Login", "User logs in with valid credentials", "test_case", "high", "active", "1. Open app\n2. Enter username\n3. Enter password\n4. Click login", "User should be logged in successfully", "User account exists", "Critical path test"}},
		{Cells: []string{"TC-002", "Login", "User logs in with invalid password", "test_case", "medium", "active", "1. Open app\n2. Enter username\n3. Enter wrong password\n4. Click login", "Error message should appear", "", ""}},
	}

	table := NewTable("Test Specification", []string{"id", "feature", "scenario", "type", "priority", "status", "instructions", "expected", "precondition", "notes"}, rows)

	output, _, err := renderer.Render(table)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify YAML front matter
	if !strings.Contains(output, "---") {
		t.Error("Output should contain YAML front matter")
	}
	if !strings.Contains(output, "name: \"Specification\"") {
		t.Error("Output should contain specification name")
	}

	// Verify summary section
	if !strings.Contains(output, "## Summary") {
		t.Error("Output should contain Summary section")
	}
	if !strings.Contains(output, "Total Items | 2") {
		t.Error("Output should show correct total items")
	}

	// Verify specifications section
	if !strings.Contains(output, "## Specifications") {
		t.Error("Output should contain Specifications section")
	}
	if !strings.Contains(output, "### Login") {
		t.Error("Output should group by feature")
	}
	if !strings.Contains(output, "TC-001") {
		t.Error("Output should contain test case ID")
	}
}

func TestSpecRenderer_EmptyRows(t *testing.T) {
	renderer := NewSpecRenderer()

	table := NewTable("Empty Spec", []string{"feature"}, []TableRow{})

	output, _, err := renderer.Render(table)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(output, "## Summary") {
		t.Error("Output should still contain summary even with empty rows")
	}
	if !strings.Contains(output, "Total Items | 0") {
		t.Error("Output should show 0 items")
	}
}

func TestSpecRenderer_SpecialCharacters(t *testing.T) {
	renderer := NewSpecRenderer()

	rows := []TableRow{
		{Cells: []string{"Feature|with|pipes", "Test\nwith\nnewlines", "Step with `code`"}},
	}

	table := NewTable("Spec with Special Chars", []string{"feature", "scenario", "instructions"}, rows)

	output, _, err := renderer.Render(table)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify output was generated
	if len(output) == 0 {
		t.Error("Output should not be empty")
	}

	// Should have some output with the spec data
	if !strings.Contains(output, "Feature|with|pipes") && !strings.Contains(output, "Test") {
		t.Error("Output should contain spec data")
	}
}

func TestGroupRowsByFeature(t *testing.T) {
	rows := []SpecRow{
		{Feature: "Login", Scenario: "valid"},
		{Feature: "Login", Scenario: "invalid"},
		{Feature: "Logout", Scenario: "success"},
		{Feature: "", Scenario: "uncategorized"},
	}

	groups := GroupRowsByFeature(rows)

	if len(groups) != 3 {
		t.Errorf("Expected 3 feature groups, got %d", len(groups))
	}

	if len(groups["Login"]) != 2 {
		t.Errorf("Expected 2 Login rows, got %d", len(groups["Login"]))
	}

	if len(groups["Logout"]) != 1 {
		t.Errorf("Expected 1 Logout row, got %d", len(groups["Logout"]))
	}

	if len(groups["Uncategorized"]) != 1 {
		t.Errorf("Expected 1 Uncategorized row, got %d", len(groups["Uncategorized"]))
	}
}

func TestFormatAsNumberedList(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Step 1\nStep 2\nStep 3",
			expected: "1. Step 1",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for i, tt := range tests {
		result := FormatAsNumberedList(tt.input)
		if tt.expected != "" && !strings.HasPrefix(result, tt.expected) {
			t.Errorf("Test %d: expected to start with %q, got %q", i, tt.expected, result)
		}
	}
}

func TestEscapeTableCell(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "simple",
			expected: "simple",
		},
		{
			input:    "with | pipe",
			expected: "with \\| pipe",
		},
	}

	for i, tt := range tests {
		result := EscapeTableCell(tt.input)
		if result != tt.expected {
			t.Errorf("Test %d: expected %q, got %q", i, tt.expected, result)
		}
	}
}
