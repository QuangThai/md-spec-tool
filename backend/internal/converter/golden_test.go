package converter

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/md-spec-tool/internal/ai"
)

// update controls whether golden files are regenerated instead of compared.
// Run with: go test -run TestGolden ./internal/converter/ -update
var update = flag.Bool("update", false, "update golden files with actual output")

// goldenAIMappings holds deterministic per-test-case column mappings for the mock AI.
// Keys match the subdirectory names under testdata/golden/.
var goldenAIMappings = map[string][]ai.CanonicalFieldMapping{
	"simple_test_case": {
		{CanonicalName: "id", SourceHeader: "TC ID", ColumnIndex: 0, Confidence: 0.95, Reasoning: "TC ID is a test case identifier"},
		{CanonicalName: "scenario", SourceHeader: "Test Case Name", ColumnIndex: 1, Confidence: 0.95, Reasoning: "Test Case Name maps to scenario"},
		{CanonicalName: "precondition", SourceHeader: "Precondition", ColumnIndex: 2, Confidence: 0.95, Reasoning: "Precondition is the test precondition"},
		{CanonicalName: "instructions", SourceHeader: "Steps", ColumnIndex: 3, Confidence: 0.95, Reasoning: "Steps maps to instructions"},
		{CanonicalName: "expected", SourceHeader: "Expected Result", ColumnIndex: 4, Confidence: 0.95, Reasoning: "Expected Result maps to expected"},
		{CanonicalName: "status", SourceHeader: "Status", ColumnIndex: 5, Confidence: 0.95, Reasoning: "Status maps to status"},
	},
	"ui_spec": {
		{CanonicalName: "no", SourceHeader: "No", ColumnIndex: 0, Confidence: 0.95, Reasoning: "No is the row number"},
		{CanonicalName: "item_name", SourceHeader: "Item Name", ColumnIndex: 1, Confidence: 0.95, Reasoning: "Item Name is the UI element name"},
		{CanonicalName: "item_type", SourceHeader: "Item Type", ColumnIndex: 2, Confidence: 0.95, Reasoning: "Item Type is the UI element type"},
		{CanonicalName: "required_optional", SourceHeader: "Required/Optional", ColumnIndex: 3, Confidence: 0.95, Reasoning: "Required/Optional maps to required_optional"},
		{CanonicalName: "input_restrictions", SourceHeader: "Input Restrictions", ColumnIndex: 4, Confidence: 0.95, Reasoning: "Input Restrictions maps to input_restrictions"},
		{CanonicalName: "display_conditions", SourceHeader: "Display Conditions", ColumnIndex: 5, Confidence: 0.95, Reasoning: "Display Conditions maps to display_conditions"},
		{CanonicalName: "action", SourceHeader: "Action", ColumnIndex: 6, Confidence: 0.95, Reasoning: "Action maps to action"},
		{CanonicalName: "navigation_destination", SourceHeader: "Navigation Destination", ColumnIndex: 7, Confidence: 0.95, Reasoning: "Navigation Destination maps to navigation_destination"},
	},
	"api_spec": {
		{CanonicalName: "endpoint", SourceHeader: "Endpoint", ColumnIndex: 0, Confidence: 0.95, Reasoning: "Endpoint is the API path"},
		{CanonicalName: "method", SourceHeader: "Method", ColumnIndex: 1, Confidence: 0.95, Reasoning: "Method is the HTTP verb"},
		{CanonicalName: "description", SourceHeader: "Description", ColumnIndex: 2, Confidence: 0.95, Reasoning: "Description is the endpoint description"},
		{CanonicalName: "parameters", SourceHeader: "Parameters", ColumnIndex: 3, Confidence: 0.95, Reasoning: "Parameters maps to parameters"},
		{CanonicalName: "response", SourceHeader: "Response", ColumnIndex: 4, Confidence: 0.95, Reasoning: "Response maps to response"},
		{CanonicalName: "status_code", SourceHeader: "Status Code", ColumnIndex: 5, Confidence: 0.95, Reasoning: "Status Code maps to status_code"},
	},
	"japanese_headers": {
		{CanonicalName: "id", SourceHeader: "テストID", ColumnIndex: 0, Confidence: 0.95, Reasoning: "テストID is a test case identifier"},
		{CanonicalName: "scenario", SourceHeader: "テストケース名", ColumnIndex: 1, Confidence: 0.95, Reasoning: "テストケース名 is the test case name"},
		{CanonicalName: "precondition", SourceHeader: "前提条件", ColumnIndex: 2, Confidence: 0.95, Reasoning: "前提条件 means precondition"},
		{CanonicalName: "instructions", SourceHeader: "手順", ColumnIndex: 3, Confidence: 0.95, Reasoning: "手順 means steps/instructions"},
		{CanonicalName: "expected", SourceHeader: "期待結果", ColumnIndex: 4, Confidence: 0.95, Reasoning: "期待結果 means expected result"},
		{CanonicalName: "status", SourceHeader: "ステータス", ColumnIndex: 5, Confidence: 0.95, Reasoning: "ステータス means status"},
	},
}

// TestGolden runs end-to-end converter tests against golden output files.
//
// Each subdirectory of testdata/golden/ is one test case:
//   - input.tsv:    tab-separated input data fed into ConvertPasteWithFormatContext
//   - expected.md:  the golden output to compare against
//
// To regenerate golden files after an intentional output change:
//
//	go test -run TestGolden ./internal/converter/ -update -v
func TestGolden(t *testing.T) {
	goldenRoot := filepath.Join("testdata", "golden")

	entries, err := os.ReadDir(goldenRoot)
	if err != nil {
		t.Fatalf("failed to read testdata/golden: %v", err)
	}

	for _, entry := range entries {
		entry := entry
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			inputPath := filepath.Join(goldenRoot, name, "input.tsv")
			goldenPath := filepath.Join(goldenRoot, name, "expected.md")

			// Read input
			inputBytes, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("failed to read input file %s: %v", inputPath, err)
			}

			// Build converter with deterministic mock AI
			mock := buildGoldenMock(t, name)
			conv := NewConverter().WithAIService(mock)

			// Run conversion (spec template, spec output format)
			resp, err := conv.ConvertPasteWithFormatContext(
				context.Background(),
				string(inputBytes),
				"spec",
				"spec",
			)
			if err != nil {
				t.Fatalf("conversion failed for %s: %v", name, err)
			}

			actual := resp.MDFlow

			if *update {
				// Regenerate golden file
				if err := os.WriteFile(goldenPath, []byte(actual), 0644); err != nil {
					t.Fatalf("failed to write golden file %s: %v", goldenPath, err)
				}
				t.Logf("updated golden file: %s", goldenPath)
				return
			}

			// Read golden file for comparison
			goldenBytes, err := os.ReadFile(goldenPath)
			if err != nil {
				if os.IsNotExist(err) {
					t.Fatalf("golden file not found: %s\n\nRun with -update flag to generate it:\n  go test -run TestGolden ./internal/converter/ -update -v", goldenPath)
				}
				t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
			}

			if actual != string(goldenBytes) {
				t.Errorf("output mismatch for test case %q\n\n--- expected (golden) ---\n%s\n--- actual ---\n%s",
					name, string(goldenBytes), actual)
			}
		})
	}
}

// buildGoldenMock creates a MockAIService with deterministic column mappings
// for the given test case name. Falls back to default heuristic mock if no
// explicit mapping is registered.
func buildGoldenMock(t *testing.T, caseName string) *ai.MockAIService {
	t.Helper()

	mappings, ok := goldenAIMappings[caseName]
	if !ok {
		t.Logf("no explicit mapping for %q, falling back to heuristic defaults", caseName)
		return ai.NewMockAIServiceWithDefaults()
	}

	nCols := len(mappings)
	mock := ai.NewMockAIService()
	mock.MapColumnsFunc = func(_ context.Context, req ai.MapColumnsRequest) (*ai.ColumnMappingResult, error) {
		return &ai.ColumnMappingResult{
			SchemaVersion:   ai.SchemaVersionColumnMapping,
			CanonicalFields: mappings,
			Meta: ai.MappingMeta{
				DetectedType:   "generic",
				SourceLanguage: "en",
				TotalColumns:   nCols,
				MappedColumns:  nCols,
				AvgConfidence:  0.95,
			},
		}, nil
	}
	return mock
}
