package converter

import (
	"os"
	"path/filepath"
	"testing"
)

// useCasesDir is relative to package dir (backend/internal/converter) when go test runs
const useCasesDir = "../../../use-cases"

func readUseCase(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(useCasesDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("use-cases file not found (run from backend/): %v", err)
		return ""
	}
	return string(data)
}

// Preset rules aligned with frontend DEFAULT_PRESETS
var (
	specTableRules = &ValidationRules{
		RequiredFields: []string{"item_name", "item_type"},
		FormatRules:    nil,
		CrossField: []CrossFieldRule{
			{IfField: "action", ThenField: "navigation_destination", Message: "When Action is set, Navigation destination is required for navigation actions"},
		},
	}
	defaultRules = &ValidationRules{
		RequiredFields: []string{"feature", "scenario", "expected"},
		FormatRules:    &FormatRules{IDPattern: `^[A-Z]{2,}-\d+$`},
		CrossField: []CrossFieldRule{
			{IfField: "id", ThenField: "feature"},
			{IfField: "instructions", ThenField: "expected"},
		},
	}
	testPlanRules = &ValidationRules{
		RequiredFields: []string{"id", "feature", "scenario", "expected"},
		FormatRules:    &FormatRules{IDPattern: `^[A-Z]{2,}-\d+$`},
		CrossField:     []CrossFieldRule{{IfField: "id", ThenField: "feature"}},
	}
)

func TestValidate_UseCases_Markdown_NoRows(t *testing.T) {
	// example-1.md is markdown/prose → BuildSpecDocFromPaste returns SpecDoc with Rows: []
	content := readUseCase(t, "example-1.md")
	if content == "" {
		return
	}
	doc, err := BuildSpecDocFromPaste(content)
	if err != nil {
		t.Fatalf("BuildSpecDocFromPaste: %v", err)
	}
	if doc == nil {
		t.Fatal("doc is nil")
	}
	if len(doc.Rows) != 0 {
		t.Errorf("markdown input should produce 0 rows, got %d", len(doc.Rows))
	}

	result := Validate(doc, defaultRules)
	if !result.Valid {
		t.Errorf("expected valid (no rows to validate), got valid=%v", result.Valid)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("expected 0 warnings for empty rows, got %d: %+v", len(result.Warnings), result.Warnings)
	}
}

func TestValidate_UseCases_SpecTable_Example2(t *testing.T) {
	// example-2.md is TSV: No, Item Name, Item Type, ... (Spec Table format)
	content := readUseCase(t, "example-2.md")
	if content == "" {
		return
	}
	doc, err := BuildSpecDocFromPaste(content)
	if err != nil {
		t.Fatalf("BuildSpecDocFromPaste: %v", err)
	}
	if doc == nil {
		t.Fatal("doc is nil")
	}
	if len(doc.Rows) == 0 {
		t.Fatal("expected at least one row from example-2 table")
	}

	result := Validate(doc, specTableRules)
	// Spec Table preset: required item_name, item_type. example-2 has a row "other" with empty Item Type → expect at least 1 VALIDATION_REQUIRED
	requiredWarnings := 0
	for _, w := range result.Warnings {
		if w.Code == "VALIDATION_REQUIRED" {
			requiredWarnings++
		}
	}
	if requiredWarnings < 1 {
		t.Errorf("example-2: expected at least 1 VALIDATION_REQUIRED (row 'other' has empty Item Type), got %d; total warnings=%d", requiredWarnings, len(result.Warnings))
	}
	if result.Valid {
		t.Errorf("example-2 with Spec Table rules (required item_name, item_type) should be invalid when a row has empty item_type")
	}
}

func TestValidate_UseCases_SpecTable_Example3(t *testing.T) {
	content := readUseCase(t, "example-3.md")
	if content == "" {
		return
	}
	doc, err := BuildSpecDocFromPaste(content)
	if err != nil {
		t.Fatalf("BuildSpecDocFromPaste: %v", err)
	}
	if doc == nil {
		t.Fatal("doc is nil")
	}

	result := Validate(doc, specTableRules)
	// example-3 has Item Name and Item Type for each row; some rows have Action but Destination empty → possible cross-field warnings
	t.Logf("example-3: rows=%d, valid=%v, warnings=%d", len(doc.Rows), result.Valid, len(result.Warnings))
	for i, w := range result.Warnings {
		t.Logf("  warning[%d]: %s %s", i, w.Code, w.Message)
	}
}

func TestValidate_UseCases_SpecTable_TableTSV(t *testing.T) {
	content := readUseCase(t, "table.tsv")
	if content == "" {
		return
	}
	doc, err := BuildSpecDocFromPaste(content)
	if err != nil {
		t.Fatalf("BuildSpecDocFromPaste: %v", err)
	}
	if doc == nil {
		t.Fatal("doc is nil")
	}
	// table.tsv uses "Display Condition" and "Destination" (column_map has "display condition", "destination")
	if len(doc.Rows) == 0 {
		t.Fatal("expected rows from table.tsv")
	}

	result := Validate(doc, specTableRules)
	t.Logf("table.tsv: rows=%d, valid=%v, warnings=%d", len(doc.Rows), result.Valid, len(result.Warnings))
	// Rows with Action but no Destination should get VALIDATION_CROSS_FIELD
	crossCount := 0
	for _, w := range result.Warnings {
		if w.Code == "VALIDATION_CROSS_FIELD" {
			crossCount++
		}
	}
	t.Logf("  VALIDATION_CROSS_FIELD count: %d", crossCount)
}

func TestValidate_NilDocOrRules(t *testing.T) {
	doc, _ := BuildSpecDocFromPaste("id\tfeature\tscenario\nTC-01\tLogin\tValid login")
	if doc == nil {
		t.Fatal("doc nil")
	}

	r1 := Validate(nil, specTableRules)
	if !r1.Valid || len(r1.Warnings) != 0 {
		t.Errorf("Validate(nil, rules) should return valid, no warnings; got valid=%v warnings=%d", r1.Valid, len(r1.Warnings))
	}

	r2 := Validate(doc, nil)
	if !r2.Valid || len(r2.Warnings) != 0 {
		t.Errorf("Validate(doc, nil) should return valid, no warnings; got valid=%v warnings=%d", r2.Valid, len(r2.Warnings))
	}
}

func TestValidate_RequiredAndCrossField(t *testing.T) {
	// Minimal TSV: id, feature, scenario, expected. One row missing expected.
	content := "id\tfeature\tscenario\texpected\nTC-01\tLogin\tValid\tOK\nTC-02\tLogin\tInvalid\t"
	doc, err := BuildSpecDocFromPaste(content)
	if err != nil {
		t.Fatalf("BuildSpecDocFromPaste: %v", err)
	}
	rules := &ValidationRules{
		RequiredFields: []string{"expected"},
		CrossField:     []CrossFieldRule{{IfField: "id", ThenField: "feature"}},
	}
	result := Validate(doc, rules)
	// Row 2 has empty expected → VALIDATION_REQUIRED
	var required int
	for _, w := range result.Warnings {
		if w.Code == "VALIDATION_REQUIRED" {
			required++
		}
	}
	if required < 1 {
		t.Errorf("expected at least 1 VALIDATION_REQUIRED (empty expected), got %d", required)
	}
	if result.Valid {
		t.Errorf("expected valid=false when required field missing")
	}
}
