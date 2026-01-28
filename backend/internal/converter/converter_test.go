package converter

import (
	"strings"
	"testing"
)

func TestPasteParser_TSV(t *testing.T) {
	parser := NewPasteParser()

	input := "Feature\tDescription\tExpected\nLogin\tEnter creds\tSuccess\nLogout\tClick btn\tRedirect"

	matrix, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matrix) != 3 {
		t.Errorf("expected 3 rows, got %d", len(matrix))
	}

	if matrix[0][0] != "Feature" {
		t.Errorf("expected 'Feature', got '%s'", matrix[0][0])
	}
}

func TestPasteParser_CSV(t *testing.T) {
	parser := NewPasteParser()

	input := "Feature,Description,Expected\nLogin,Enter creds,Success"

	matrix, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matrix) != 2 {
		t.Errorf("expected 2 rows, got %d", len(matrix))
	}
}

func TestColumnMapper_MapColumns(t *testing.T) {
	mapper := NewColumnMapper()

	headers := []string{"TC_ID", "Feature", "Steps", "Expected Result", "Priority", "Custom"}

	colMap, unmapped := mapper.MapColumns(headers)

	if colMap[FieldID] != 0 {
		t.Errorf("expected FieldID at index 0")
	}

	if colMap[FieldFeature] != 1 {
		t.Errorf("expected FieldFeature at index 1")
	}

	if colMap[FieldInstructions] != 2 {
		t.Errorf("expected FieldInstructions at index 2")
	}

	if colMap[FieldExpected] != 3 {
		t.Errorf("expected FieldExpected at index 3")
	}

	if len(unmapped) != 1 || unmapped[0] != "Custom" {
		t.Errorf("expected 'Custom' as unmapped, got %v", unmapped)
	}
}

func TestHeaderDetector_DetectHeaderRow(t *testing.T) {
	detector := NewHeaderDetector()

	// Matrix with header in first row
	matrix := CellMatrix{
		{"Feature", "Description", "Expected"},
		{"Login", "Enter credentials", "Success"},
		{"Logout", "Click logout", "Redirect"},
	}

	row, confidence := detector.DetectHeaderRow(matrix)

	if row != 0 {
		t.Errorf("expected header row 0, got %d", row)
	}

	if confidence < 50 {
		t.Errorf("expected confidence >= 50, got %d", confidence)
	}
}

func TestHeaderDetector_DetectHeaderRow_WithTitleRow(t *testing.T) {
	detector := NewHeaderDetector()

	// Matrix with title row before header
	matrix := CellMatrix{
		{"Authentication Test Cases", "", ""},
		{"Feature", "Description", "Expected"},
		{"Login", "Enter credentials", "Success"},
	}

	row, _ := detector.DetectHeaderRow(matrix)

	if row != 1 {
		t.Errorf("expected header row 1, got %d", row)
	}
}

func TestCellMatrix_Normalize(t *testing.T) {
	matrix := CellMatrix{
		{"  Feature  ", "Description", ""},
		{"Login", "  test  ", "Success"},
		{"", "", ""}, // Empty row should be removed
		{"Logout", "click", "Redirect"},
	}

	normalized := matrix.Normalize()

	if len(normalized) != 3 {
		t.Errorf("expected 3 rows after normalization, got %d", len(normalized))
	}

	if normalized[0][0] != "Feature" {
		t.Errorf("expected trimmed 'Feature', got '%s'", normalized[0][0])
	}
}

func TestConverter_ConvertPaste(t *testing.T) {
	conv := NewConverter()

	input := `TC_ID	Feature	Scenario	Steps	Expected Result	Priority
AUTH-001	Register	Register with valid email	1. Open Register page	Account created	High
AUTH-002	Login	Login with valid credentials	1. Open Login page	User logged in	High`

	result, err := conv.ConvertPaste(input, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.MDFlow == "" {
		t.Error("expected non-empty MDFlow output")
	}

	if !strings.Contains(result.MDFlow, "AUTH-001") {
		t.Error("expected MDFlow to contain AUTH-001")
	}

	if !strings.Contains(result.MDFlow, "Register") {
		t.Error("expected MDFlow to contain 'Register'")
	}

	if result.Meta.TotalRows != 2 {
		t.Errorf("expected 2 rows, got %d", result.Meta.TotalRows)
	}
}

func TestConverter_ConvertPaste_WithTemplate(t *testing.T) {
	conv := NewConverter()

	input := `Feature	Scenario	Steps	Expected Result
Login	Valid login	1. Enter credentials	Success`

	result, err := conv.ConvertPaste(input, "test-plan")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.MDFlow, "Test Plan") {
		t.Error("expected MDFlow to contain 'Test Plan' for test-plan template")
	}
}

func TestMDFlowRenderer_Render(t *testing.T) {
	renderer := NewMDFlowRenderer()

	doc := &SpecDoc{
		Title: "Test Spec",
		Rows: []SpecRow{
			{
				ID:           "TC-001",
				Feature:      "Authentication",
				Scenario:     "Login with valid credentials",
				Instructions: "1. Enter email\n2. Enter password\n3. Click login",
				Inputs:       "email=test@test.com",
				Expected:     "User logged in successfully",
				Priority:     "High",
				Type:         "Positive",
			},
		},
		Meta: SpecDocMeta{
			TotalRows: 1,
		},
	}

	output, err := renderer.Render(doc, "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check frontmatter
	if !strings.Contains(output, "---") {
		t.Error("expected YAML frontmatter")
	}

	// Check content
	if !strings.Contains(output, "TC-001") {
		t.Error("expected TC-001 in output")
	}

	if !strings.Contains(output, "Authentication") {
		t.Error("expected 'Authentication' in output")
	}

	if !strings.Contains(output, "High") {
		t.Error("expected 'High' priority in output")
	}
}

func TestMDFlowRenderer_GetTemplateNames(t *testing.T) {
	renderer := NewMDFlowRenderer()
	names := renderer.GetTemplateNames()

	if len(names) < 3 {
		t.Errorf("expected at least 3 templates, got %d", len(names))
	}

	hasDefault := false
	for _, name := range names {
		if name == "default" {
			hasDefault = true
			break
		}
	}

	if !hasDefault {
		t.Error("expected 'default' template")
	}
}

func TestGetFieldValue(t *testing.T) {
	row := []string{"TC-001", "Login", "Valid login", "Success"}
	colMap := ColumnMap{
		FieldID:       0,
		FieldFeature:  1,
		FieldScenario: 2,
		FieldExpected: 3,
	}

	if GetFieldValue(row, colMap, FieldID) != "TC-001" {
		t.Error("expected TC-001")
	}

	if GetFieldValue(row, colMap, FieldFeature) != "Login" {
		t.Error("expected Login")
	}

	// Test missing field
	if GetFieldValue(row, colMap, FieldNotes) != "" {
		t.Error("expected empty string for missing field")
	}
}

func TestXLSXParser_ParseFile(t *testing.T) {
	parser := NewXLSXParser()
	
	// Test with the fixture file - try multiple paths
	paths := []string{
		"../../../../docs/fixtures/input/authentication_requirements_testcases.xlsx",
		"../../../docs/fixtures/input/authentication_requirements_testcases.xlsx",
		"../../docs/fixtures/input/authentication_requirements_testcases.xlsx",
	}
	
	var result *XLSXResult
	var err error
	for _, path := range paths {
		result, err = parser.ParseFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Skipf("fixture file not found, skipping: %v", err)
	}

	if len(result.Sheets) == 0 {
		t.Error("expected at least one sheet")
	}

	if result.ActiveSheet == "" {
		t.Error("expected active sheet to be set")
	}

	matrix := result.GetMatrix(result.ActiveSheet)
	if len(matrix) == 0 {
		t.Error("expected non-empty matrix")
	}

	// Check that first row contains headers
	headers := matrix.GetRow(0)
	if len(headers) == 0 {
		t.Error("expected headers")
	}

	// Verify some expected headers
	hasFeature := false
	for _, h := range headers {
		if h == "Feature" {
			hasFeature = true
			break
		}
	}
	if !hasFeature {
		t.Errorf("expected 'Feature' header, got: %v", headers)
	}
}

func TestConverter_ConvertXLSX(t *testing.T) {
	conv := NewConverter()

	paths := []string{
		"../../../../docs/fixtures/input/authentication_requirements_testcases.xlsx",
		"../../../docs/fixtures/input/authentication_requirements_testcases.xlsx",
		"../../docs/fixtures/input/authentication_requirements_testcases.xlsx",
	}
	
	var result *ConvertResponse
	var err error
	for _, path := range paths {
		result, err = conv.ConvertXLSX(path, "", "")
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Skipf("fixture file not found, skipping: %v", err)
	}

	if result.MDFlow == "" {
		t.Error("expected non-empty MDFlow output")
	}

	if !strings.Contains(result.MDFlow, "AUTH-001") {
		t.Error("expected MDFlow to contain AUTH-001")
	}

	if !strings.Contains(result.MDFlow, "Register") {
		t.Error("expected MDFlow to contain 'Register'")
	}

	if result.Meta.TotalRows < 10 {
		t.Errorf("expected at least 10 rows, got %d", result.Meta.TotalRows)
	}
}
