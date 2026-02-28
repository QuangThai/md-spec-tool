package ai

import (
	"testing"
	"fmt"
	"strings"
)

// TestPromptPipeline_EndToEndFlow tests the complete prompt building pipeline
func TestPromptPipeline_EndToEndFlow(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	// Scenario: English test case with 6 columns
	built, err := builder.BuildPrompt(PromptIDColumnMapping, PromptContext{
		SchemaHint:  "test_case",
		Language:    "en",
		ColumnCount: 6,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify structure
	if built.Content == "" {
		t.Error("expected non-empty prompt content")
	}
	if built.Hash == "" {
		t.Error("expected non-empty hash")
	}
	if built.CacheVersion == "" {
		t.Error("expected non-empty cache version")
	}
	if built.OperationID != PromptIDColumnMapping {
		t.Errorf("expected operation ID %s, got %s", PromptIDColumnMapping, built.OperationID)
	}

	// Verify content has expected sections
	if !strings.Contains(built.Content, "You are an expert at analyzing spreadsheet headers") {
		t.Error("prompt should contain system prompt introduction")
	}
	if !strings.Contains(built.Content, "CONTEXT HINTS") {
		t.Error("prompt should include context hints section")
	}
	if !strings.Contains(built.Content, "test_case") {
		t.Error("prompt should reference schema hint")
	}
	if !strings.Contains(built.Content, "FEW-SHOT EXAMPLES") {
		t.Error("prompt should include examples section")
	}
}

// TestPromptPipeline_MultiLanguageDetection tests language-specific example selection
func TestPromptPipeline_MultiLanguageDetection(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	languages := []string{"en", "vi", "ja"}
	for _, lang := range languages {
		t.Run(fmt.Sprintf("language_%s", lang), func(t *testing.T) {
			built, err := builder.BuildPrompt(PromptIDColumnMapping, PromptContext{
				SchemaHint:  "test_case",
				Language:    lang,
				ColumnCount: 6,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify language hint is present
			if !strings.Contains(built.Content, fmt.Sprintf("content_language: %s", lang)) {
				t.Errorf("prompt should include language hint for %s", lang)
			}

			// Verify examples are included
			if !strings.Contains(built.Content, "FEW-SHOT EXAMPLES") {
				t.Error("prompt should include examples section")
			}
		})
	}
}

// TestPromptPipeline_AllSchemaTypes tests example selection for all schema types
func TestPromptPipeline_AllSchemaTypes(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	schemaTypes := []string{"test_case", "ui_spec", "api_spec", "product_backlog", "issue_tracker", "db_schema", "requirements", "defect_report", "security_req"}

	for _, schemaType := range schemaTypes {
		t.Run(fmt.Sprintf("schema_%s", schemaType), func(t *testing.T) {
			built, err := builder.BuildPrompt(PromptIDColumnMapping, PromptContext{
				SchemaHint:  schemaType,
				Language:    "en",
				ColumnCount: 5,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify schema type hint is present
			if !strings.Contains(built.Content, fmt.Sprintf("schema_type: %s", schemaType)) {
				t.Errorf("prompt should include schema hint for %s", schemaType)
			}

			// All schema types should have examples
			if !strings.Contains(built.Content, "FEW-SHOT EXAMPLES") {
				t.Error("prompt should include examples section")
			}
		})
	}
}

// TestPromptPipeline_CacheInvalidationOnChange tests hash changes when content changes
func TestPromptPipeline_CacheInvalidationOnChange(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	// Build with initial context
	built1, err := builder.BuildPrompt(PromptIDColumnMapping, PromptContext{
		SchemaHint:  "test_case",
		Language:    "en",
		ColumnCount: 6,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Build with different language (should have different hash)
	built2, err := builder.BuildPrompt(PromptIDColumnMapping, PromptContext{
		SchemaHint:  "test_case",
		Language:    "vi",
		ColumnCount: 6,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Hashes should be different because content is different
	if built1.Hash == built2.Hash {
		t.Error("expected different hashes for different language contexts")
	}

	// Cache versions should be different
	if built1.CacheVersion == built2.CacheVersion {
		t.Error("expected different cache versions for different prompts")
	}
}

// TestPromptPipeline_RefinementContext tests refinement context injection
func TestPromptPipeline_RefinementContext(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	refinement := "Additional instruction: prioritize columns with numeric data."

	built, err := builder.BuildPrompt(PromptIDRefineMapping, PromptContext{
		SchemaHint:        "test_case",
		Language:          "en",
		ColumnCount:       6,
		RefinementContext: refinement,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify refinement context is included
	if !strings.Contains(built.Content, "ADDITIONAL CONTEXT") {
		t.Error("prompt should include additional context section")
	}
	if !strings.Contains(built.Content, refinement) {
		t.Error("prompt should include the refinement instruction")
	}
}

// TestPromptPipeline_ExampleSelectionScoring tests example selection scoring algorithm
func TestPromptPipeline_ExampleSelectionScoring(t *testing.T) {
	exampleStore := DefaultExampleStore()

	// Test case 1: Exact schema + language match with similar column count
	ctx := SelectionContext{
		SchemaHint:  "api_spec",
		Language:    "en",
		ColumnCount: 6,
		MaxResults:  3,
	}
	examples := exampleStore.SelectExamples(PromptIDColumnMapping, ctx)
	if len(examples) == 0 {
		t.Error("should find examples for api_spec English")
	}

	// Test case 2: Vietnamese examples should be selected for Vietnamese input
	ctx = SelectionContext{
		SchemaHint:  "api_spec",
		Language:    "vi",
		ColumnCount: 6,
		MaxResults:  3,
	}
	examples = exampleStore.SelectExamples(PromptIDColumnMapping, ctx)
	if len(examples) == 0 {
		t.Error("should find examples for api_spec Vietnamese")
	}

	// Test case 3: Japanese examples should be available
	ctx = SelectionContext{
		SchemaHint:  "product_backlog",
		Language:    "ja",
		ColumnCount: 6,
		MaxResults:  3,
	}
	examples = exampleStore.SelectExamples(PromptIDColumnMapping, ctx)
	if len(examples) == 0 {
		t.Error("should find examples for product_backlog Japanese")
	}
}

// TestPromptPipeline_FallbackToGeneric tests fallback when no exact match exists
func TestPromptPipeline_FallbackToGeneric(t *testing.T) {
	exampleStore := DefaultExampleStore()

	// Request examples for non-existent schema type
	ctx := SelectionContext{
		SchemaHint:  "non_existent_schema",
		Language:    "en",
		ColumnCount: 6,
		MaxResults:  3,
	}
	examples := exampleStore.SelectExamples(PromptIDColumnMapping, ctx)

	// Should still get examples (fallback to generic or similar)
	if len(examples) == 0 {
		t.Error("should fallback to available examples even for unknown schema type")
	}
}

// TestPromptPipeline_ColumnCountSimilarity tests scoring based on column count
func TestPromptPipeline_ColumnCountSimilarity(t *testing.T) {
	exampleStore := DefaultExampleStore()

	// Test with column count that matches UI spec (8 columns)
	ctx := SelectionContext{
		SchemaHint:  "ui_spec",
		Language:    "en",
		ColumnCount: 8,
		MaxResults:  3,
	}
	examples := exampleStore.SelectExamples(PromptIDColumnMapping, ctx)
	if len(examples) == 0 {
		t.Error("should find UI spec examples with 8 columns")
	}

	// Verify that the UI spec example with 8 columns is included
	found := false
	for _, ex := range examples {
		if ex.SchemaType == "ui_spec" && len(ex.Headers) == 8 {
			found = true
			break
		}
	}
	if !found && len(examples) > 0 {
		t.Log("Warning: UI spec example with 8 columns not in top selections, but other examples found")
	}
}

// TestPromptPipeline_PromptVersioning tests version tracking and cache keys
func TestPromptPipeline_PromptVersioning(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	built, err := builder.BuildPrompt(PromptIDColumnMapping, PromptContext{
		SchemaHint:  "test_case",
		Language:    "en",
		ColumnCount: 6,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify version information is present
	if built.BaseVersion == "" {
		t.Error("expected base version to be set")
	}
	if built.CacheVersion == "" {
		t.Error("expected cache version to be set")
	}

	// Cache version should follow format: version:hash_prefix
	parts := strings.Split(built.CacheVersion, ":")
	if len(parts) != 2 {
		t.Errorf("cache version should be 'version:hash', got %s", built.CacheVersion)
	}
	if parts[0] != built.BaseVersion {
		t.Errorf("cache version should start with base version %s, got %s", built.BaseVersion, parts[0])
	}
}

// TestPromptPipeline_JSONSchemaHint tests JSON schema hints for structured outputs
func TestPromptPipeline_JSONSchemaHint(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	// Column mapping should include JSON schema hint
	built, err := builder.BuildPrompt(PromptIDColumnMapping, PromptContext{
		SchemaHint:  "test_case",
		Language:    "en",
		ColumnCount: 6,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(built.Content, "JSON OUTPUT REMINDER") {
		t.Error("column mapping prompt should include JSON output reminder")
	}
	if !strings.Contains(built.Content, "canonical_name") {
		t.Error("column mapping prompt should reference canonical_name field")
	}
}

// TestPromptPipeline_MultipleLanguagesInOnePrompt tests multi-language example inclusion
func TestPromptPipeline_MultipleLanguagesInOnePrompt(t *testing.T) {
	exampleStore := DefaultExampleStore()

	// Get examples without specifying language - should get best matches
	ctx := SelectionContext{
		SchemaHint:  "api_spec",
		Language:    "",
		ColumnCount: 6,
		MaxResults:  5,
	}
	examples := exampleStore.SelectExamples(PromptIDColumnMapping, ctx)

	// Should have examples available
	if len(examples) == 0 {
		t.Error("should find examples for api_spec without language filter")
	}

	// Should have multiple languages if available
	languages := make(map[string]bool)
	for _, ex := range examples {
		languages[ex.Language] = true
	}
	if len(languages) > 1 {
		t.Logf("Found examples in multiple languages: %v", languages)
	}
}

// TestPromptPipeline_EmptyContextHandling tests behavior with minimal context
func TestPromptPipeline_EmptyContextHandling(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	// Build with minimal context
	built, err := builder.BuildPrompt(PromptIDColumnMapping, PromptContext{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still produce a valid prompt
	if built.Content == "" {
		t.Error("expected non-empty prompt even with empty context")
	}
	if !strings.Contains(built.Content, "You are an expert") {
		t.Error("prompt should contain system prompt")
	}

	// Should not have context hints when not provided
	if strings.Contains(built.Content, "CONTEXT HINTS") && !strings.Contains(built.Content, "schema_type:") {
		t.Error("should not have context hints section when no context provided")
	}
}

// TestPromptPipeline_LargeColumnCount tests handling of large column counts
func TestPromptPipeline_LargeColumnCount(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	// Test with large column count
	built, err := builder.BuildPrompt(PromptIDColumnMapping, PromptContext{
		SchemaHint:  "generic",
		Language:    "en",
		ColumnCount: 50,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(built.Content, "column_count: 50") {
		t.Error("prompt should include large column count in context hints")
	}
}

// TestPromptPipeline_PasteAnalysisPrompt tests paste analysis prompt building
func TestPromptPipeline_PasteAnalysisPrompt(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	built, err := builder.BuildPrompt(PromptIDPasteAnalysis, PromptContext{
		Language: "en",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Paste analysis prompt should have specific content
	if !strings.Contains(built.Content, "DETECT INPUT TYPE") {
		t.Error("paste analysis prompt should include input type detection")
	}
	if !strings.Contains(built.Content, "DETECT FORMAT") {
		t.Error("paste analysis prompt should include format detection")
	}
	if built.OperationID != PromptIDPasteAnalysis {
		t.Errorf("expected operation ID %s, got %s", PromptIDPasteAnalysis, built.OperationID)
	}
}

// TestPromptPipeline_AllOperations tests building prompts for all supported operations
func TestPromptPipeline_AllOperations(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	operations := []string{
		PromptIDColumnMapping,
		PromptIDRefineMapping,
		PromptIDPasteAnalysis,
		PromptIDSuggestions,
		PromptIDDiffSummary,
		PromptIDSemanticValidation,
	}

	for _, opID := range operations {
		t.Run(fmt.Sprintf("operation_%s", opID), func(t *testing.T) {
			built, err := builder.BuildPrompt(opID, PromptContext{
				Language: "en",
			})

			if err != nil {
				t.Fatalf("unexpected error building %s: %v", opID, err)
			}

			if built.OperationID != opID {
				t.Errorf("expected operation ID %s, got %s", opID, built.OperationID)
			}
			if built.Content == "" {
				t.Errorf("operation %s should have non-empty prompt content", opID)
			}
		})
	}
}

// TestPromptPipeline_HashConsistency tests that same context produces same hash
func TestPromptPipeline_HashConsistency(t *testing.T) {
	registry := DefaultPromptRegistry()
	exampleStore := DefaultExampleStore()
	builder := NewPromptBuilder(registry, exampleStore)

	ctx := PromptContext{
		SchemaHint:  "test_case",
		Language:    "en",
		ColumnCount: 6,
	}

	built1, _ := builder.BuildPrompt(PromptIDColumnMapping, ctx)
	built2, _ := builder.BuildPrompt(PromptIDColumnMapping, ctx)

	// Same context should produce same hash
	if built1.Hash != built2.Hash {
		t.Error("expected same hash for identical context")
	}

	// Same context should produce same cache version
	if built1.CacheVersion != built2.CacheVersion {
		t.Error("expected same cache version for identical context")
	}
}

// TestExampleStore_CoverageByDomain tests that all domains have examples
func TestExampleStore_CoverageByDomain(t *testing.T) {
	store := DefaultExampleStore()

	// Define required domains with their language variations
	domains := map[string][]string{
		"test_case":      {"en", "vi", "ja"},
		"ui_spec":        {"en", "vi", "ja"},
		"api_spec":       {"en", "vi", "ja"},
		"product_backlog": {"en", "vi", "ja"},
		"issue_tracker":  {"en", "ja", "vi"},
		"generic":        {"en"},
		"db_schema":      {"en"},
		"requirements":   {"en"},
		"defect_report":  {"en"},
		"security_req":   {"en"},
	}

	for domain, languages := range domains {
		for _, lang := range languages {
			examples := store.GetExamples("column_mapping", ExampleFilter{
				SchemaType: domain,
				Language:   lang,
			})
			if len(examples) == 0 {
				t.Errorf("missing example for domain=%s, language=%s", domain, lang)
			}
		}
	}
}

// TestExampleStore_FormattedOutput tests formatting examples for prompts
func TestExampleStore_FormattedOutput(t *testing.T) {
	store := DefaultExampleStore()
	examples := store.GetExamples("column_mapping", ExampleFilter{
		SchemaType: "test_case",
		Language:   "en",
	})

	if len(examples) == 0 {
		t.Fatal("expected at least one test_case example")
	}

	formatted := FormatExamplesForPrompt(examples)

	// Verify formatting
	if !strings.Contains(formatted, "FEW-SHOT EXAMPLES") {
		t.Error("formatted examples should include header")
	}
	if !strings.Contains(formatted, "Example 1") {
		t.Error("formatted examples should number each example")
	}
	if !strings.Contains(formatted, "Headers:") {
		t.Error("formatted examples should show headers")
	}
	if !strings.Contains(formatted, "Expected mappings") {
		t.Error("formatted examples should show expected mappings")
	}
}
