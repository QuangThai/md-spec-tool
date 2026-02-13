package ai

// Schema versions - bump when changing structure
const (
	SchemaVersionColumnMapping = "v2"
	SchemaVersionPasteAnalysis = "v1"
	SchemaVersionSuggestions   = "v1"
)

// ==================== Column Mapping ====================

// ColumnMappingResult is the structured output for column mapping
type ColumnMappingResult struct {
	SchemaVersion   string                  `json:"schema_version"`
	CanonicalFields []CanonicalFieldMapping `json:"canonical_fields"`
	ExtraColumns    []ExtraColumnMapping    `json:"extra_columns,omitempty"`
	Meta            MappingMeta             `json:"meta"`
}

// CanonicalFieldMapping maps a source header to a known canonical field
type CanonicalFieldMapping struct {
	CanonicalName string              `json:"canonical_name"`         // e.g., "id", "title", "description"
	SourceHeader  string              `json:"source_header"`          // Original header text
	ColumnIndex   int                 `json:"column_index"`           // 0-based index
	Confidence    float64             `json:"confidence"`             // 0.0-1.0
	Reasoning     string              `json:"reasoning,omitempty"`    // Brief explanation (max 256 chars)
	Alternatives  []AlternativeColumn `json:"alternatives,omitempty"` // For ambiguous cases
}

// AlternativeColumn represents an alternative mapping candidate
type AlternativeColumn struct {
	SourceHeader string  `json:"source_header"`
	ColumnIndex  int     `json:"column_index"`
	Confidence   float64 `json:"confidence"`
}

// ExtraColumnMapping preserves columns that don't match canonical fields
type ExtraColumnMapping struct {
	Name         string  `json:"name"`          // Original header
	SemanticRole string  `json:"semantic_role"` // Free text: "risk", "component", "milestone"
	ColumnIndex  int     `json:"column_index"`
	Confidence   float64 `json:"confidence"`
}

// MappingMeta contains metadata about the mapping operation
type MappingMeta struct {
	DetectedType    string  `json:"detected_type"`   // "test_case", "backlog", "requirements", "generic"
	SourceLanguage  string  `json:"source_language"` // "en", "ja", "vi", etc.
	TotalColumns    int     `json:"total_columns"`
	MappedColumns   int     `json:"mapped_columns"`
	UnmappedColumns int     `json:"unmapped_columns"`
	AvgConfidence   float64 `json:"avg_confidence"`
}

// ==================== Paste Analysis ====================

// PasteAnalysis is the structured output for paste processing
type PasteAnalysis struct {
	SchemaVersion   string     `json:"schema_version"`
	InputType       string     `json:"input_type"`                 // "table", "test_cases", "product_backlog", "issue_tracker", "api_spec", "ui_spec", "prose", "mixed", "unknown"
	DetectedFormat  string     `json:"detected_format"`            // "csv", "tsv", "markdown_table", "free_text", "mixed"
	DetectedSchema  string     `json:"detected_schema,omitempty"`  // "test_case", "product_backlog", "issue_tracker", "api_spec", "ui_spec", "generic"
	NormalizedTable [][]string `json:"normalized_table,omitempty"` // [headers, ...rows]
	DetectedColumns []string   `json:"detected_columns,omitempty"`
	SuggestedOutput string     `json:"suggested_output"` // "spec" or "table"
	Confidence      float64    `json:"confidence"`
	Notes           string     `json:"notes,omitempty"`
}

// ==================== AI Suggestions ====================

// SuggestionType represents the type of AI suggestion
type SuggestionType string

const (
	SuggestionMissingField     SuggestionType = "missing_field"
	SuggestionVagueDescription SuggestionType = "vague_description"
	SuggestionIncompleteSteps  SuggestionType = "incomplete_steps"
	SuggestionFormatting       SuggestionType = "formatting"
	SuggestionCoverage         SuggestionType = "coverage"
)

// SuggestionsRequest is the input for AI suggestions
type SuggestionsRequest struct {
	SpecContent string `json:"spec_content"` // Formatted spec rows
	Template    string `json:"template"`     // Template name
	RowCount    int    `json:"row_count"`    // Number of rows for context
}

// SuggestionsResult contains AI-generated improvement suggestions
type SuggestionsResult struct {
	SchemaVersion string       `json:"schema_version"`
	Suggestions   []Suggestion `json:"suggestions"`
}

// Suggestion represents a single AI-generated improvement suggestion
type Suggestion struct {
	Type       SuggestionType `json:"type"`              // missing_field, vague_description, etc.
	Severity   string         `json:"severity"`          // info, warn, error
	Message    string         `json:"message"`           // Brief description of the issue
	RowRef     *int           `json:"row_ref,omitempty"` // Row number (1-based) if applicable
	Field      string         `json:"field,omitempty"`   // Field name if applicable
	Suggestion string         `json:"suggestion"`        // Specific actionable improvement
}

// ValidSuggestionTypes returns all valid suggestion types for schema building
func ValidSuggestionTypes() []string {
	return []string{
		string(SuggestionMissingField),
		string(SuggestionVagueDescription),
		string(SuggestionIncompleteSteps),
		string(SuggestionFormatting),
		string(SuggestionCoverage),
	}
}

// ==================== Canonical Fields Definition ====================

// CanonicalFields defines the known vocabulary (keep small: 10-20 max)
var CanonicalFields = map[string]string{
	"id":                     "Unique identifier (TC ID, Issue #, etc.)",
	"title":                  "Title or short summary",
	"feature":                "Feature, module, or epic name",
	"scenario":               "Scenario or test case name",
	"instructions":           "Step-by-step instructions",
	"inputs":                 "Inputs or test data",
	"expected":               "Expected outcome or acceptance criteria",
	"precondition":           "Prerequisites or setup requirements",
	"priority":               "Priority level (high, medium, low, P0-P3)",
	"type":                   "Type classification (bug, feature, task)",
	"status":                 "Current status (active, done, pending)",
	"endpoint":               "API endpoint or URL",
	"method":                 "HTTP method (GET, POST, PUT, DELETE)",
	"parameters":             "API parameters or request body fields",
	"response":               "API response structure or output format",
	"status_code":            "HTTP status code (200, 400, 404, etc.)",
	"notes":                  "Additional notes or comments",
	"description":            "Detailed description of feature/item/issue",
	"acceptance_criteria":    "Acceptance criteria or done definition",
	"component":              "System component or module",
	"assignee":               "Person responsible or team",
	"category":               "Category or classification tag",
	"no":                     "Row number or sequence",
	"item_name":              "UI/UX item name",
	"item_type":              "UI/UX item type",
	"required_optional":      "Required/optional indicator",
	"input_restrictions":     "Input constraints or validation",
	"display_conditions":     "Display conditions",
	"action":                 "User action",
	"navigation_destination": "Navigation destination",
}
