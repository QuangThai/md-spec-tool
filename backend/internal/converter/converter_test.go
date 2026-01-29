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

// Phase 5: Integration tests for markdown and table paths
func TestConvertPaste_MarkdownInput(t *testing.T) {
	input := readTestFile(t, "example-1.md")
	converter := NewConverter()

	result, err := converter.ConvertPaste(input, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No header detection warnings
	if len(result.Warnings) > 0 {
		t.Errorf("expected no warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}

	// Check for section content
	if !strings.Contains(result.MDFlow, "Background") {
		t.Error("expected 'Background' section in output")
	}
	if !strings.Contains(result.MDFlow, "Scope") {
		t.Error("expected 'Scope' section in output")
	}
	if !strings.Contains(result.MDFlow, "Requirements") {
		t.Error("expected 'Requirements' section in output")
	}

	// Original markdown should be preserved
	if !strings.Contains(result.MDFlow, "WebinarStock") {
		t.Error("expected original content to be preserved")
	}

	// Should use markdown template (not default table template)
	if strings.Contains(result.MDFlow, "_inputs") {
		t.Error("expected markdown output without _inputs")
	}
}

func TestConvertPaste_Markdown_DefaultTemplateParam(t *testing.T) {
	input := readTestFile(t, "example-1.md")
	converter := NewConverter()

	result, err := converter.ConvertPaste(input, "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result.MDFlow, "_inputs") {
		t.Error("expected markdown output without _inputs when template=default")
	}
	if !strings.Contains(result.MDFlow, "type: \"specification\"") {
		t.Error("expected markdown template output when template=default")
	}
}

func TestConvertPaste_Markdown_IgnoresTemplateParam(t *testing.T) {
	input := readTestFile(t, "example-1.md")
	converter := NewConverter()

	result, err := converter.ConvertPaste(input, "feature-spec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.MDFlow, "## Background") {
		t.Error("expected markdown section heading in output")
	}
}

func TestConvertPaste_Markdown_WithTemplates(t *testing.T) {
	input := readTestFile(t, "example-1.md")
	converter := NewConverter()

	resultFeature, err := converter.ConvertPaste(input, "feature-spec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resultFeature.MDFlow, "## Background") {
		t.Error("expected feature-spec to include section heading")
	}
	if !strings.Contains(resultFeature.MDFlow, "type: \"feature-spec\"") {
		t.Error("expected feature-spec marker in output")
	}

	resultTestPlan, err := converter.ConvertPaste(input, "test-plan")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resultTestPlan.MDFlow, "Test Plan") {
		t.Error("expected test-plan output")
	}

	resultAPI, err := converter.ConvertPaste(input, "api-endpoint")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resultAPI.MDFlow, "API Specification") {
		t.Error("expected api-endpoint output")
	}
}

func TestConvertPaste_Markdown_BoldHeadings(t *testing.T) {
	input := readTestFile(t, "example-5.md")
	converter := NewConverter()

	result, err := converter.ConvertPaste(input, "feature-spec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.MDFlow, "Summary") {
		t.Error("expected bold heading Summary to be parsed")
	}
	if !strings.Contains(result.MDFlow, "Details") {
		t.Error("expected bold heading Details to be parsed")
	}
}

func TestConvertPaste_TableInput(t *testing.T) {
	input := readTestFile(t, "example-2.md")
	converter := NewConverter()

	result, err := converter.ConvertPaste(input, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No "Low confidence" warning
	for _, w := range result.Warnings {
		if strings.Contains(w, "Low confidence") {
			t.Errorf("unexpected 'Low confidence' warning: %s", w)
		}
	}

	// Should have parsed rows
	if result.Meta.TotalRows == 0 {
		t.Error("expected rows to be parsed from table input")
	}

	// Check column mappings for Phase 3 fields
	if idx, ok := result.Meta.ColumnMap[FieldItemName]; !ok || idx != 1 {
		t.Errorf("expected ItemName field at column 1, got %d", idx)
	}
	if idx, ok := result.Meta.ColumnMap[FieldItemType]; !ok || idx != 2 {
		t.Errorf("expected ItemType field at column 2, got %d", idx)
	}
	if idx, ok := result.Meta.ColumnMap[FieldRequiredOptional]; !ok || idx != 3 {
		t.Errorf("expected RequiredOptional field at column 3, got %d", idx)
	}

	// Spec-table content should map into Feature/Scenario for default templates
	if !strings.Contains(result.MDFlow, "Popular Video Ranking") {
		t.Error("expected table output to include ItemName content")
	}
	if !strings.Contains(result.MDFlow, "Input Restrictions") {
		t.Error("expected table output to include Input Restrictions label")
	}
}

func TestConvertPaste_WithSpecTableTemplate(t *testing.T) {
	input := readTestFile(t, "example-2.md")
	converter := NewConverter()

	result, err := converter.ConvertPaste(input, "spec-table")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain spec-table type marker
	if !strings.Contains(result.MDFlow, "type: \"spec-table\"") {
		t.Error("expected 'spec-table' type in output")
	}

	// Should contain summary table
	if !strings.Contains(result.MDFlow, "Summary Table") {
		t.Error("expected 'Summary Table' in spec-table output")
	}

	// Should contain item details
	if !strings.Contains(result.MDFlow, "Item Details") {
		t.Error("expected 'Item Details' in spec-table output")
	}
}

func TestInputTypeDetection_Markdown(t *testing.T) {
	input := readTestFile(t, "example-1.md")
	analysis := DetectInputType(input)

	if analysis.Type != InputTypeMarkdown {
		t.Errorf("expected InputTypeMarkdown, got %s", analysis.Type)
	}
	if analysis.Confidence < 70 {
		t.Errorf("expected confidence >= 70, got %d", analysis.Confidence)
	}
}

func TestInputTypeDetection_Table(t *testing.T) {
	input := readTestFile(t, "example-2.md")
	analysis := DetectInputType(input)

	if analysis.Type != InputTypeTable {
		t.Errorf("expected InputTypeTable, got %s", analysis.Type)
	}
	if analysis.Confidence < 80 {
		t.Errorf("expected confidence >= 80, got %d", analysis.Confidence)
	}
}
