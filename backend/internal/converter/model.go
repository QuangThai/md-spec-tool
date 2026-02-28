package converter

// WarningSeverity represents the importance level of a warning
type WarningSeverity string

const (
	SeverityInfo  WarningSeverity = "info"
	SeverityWarn  WarningSeverity = "warn"
	SeverityError WarningSeverity = "error" // non-fatal error-level warning
)

// WarningCategory represents the type of warning
type WarningCategory string

const (
	CatInput   WarningCategory = "input"
	CatDetect  WarningCategory = "detect"
	CatHeader  WarningCategory = "header"
	CatMapping WarningCategory = "mapping"
	CatRows    WarningCategory = "rows"
	CatRender  WarningCategory = "render"
)

// Warning represents a structured warning from the conversion pipeline
type Warning struct {
	Code     string          `json:"code"`
	Message  string          `json:"message"`
	Severity WarningSeverity `json:"severity"`
	Category WarningCategory `json:"category"`
	Hint     string          `json:"hint,omitempty"`    // user-facing suggestion
	Details  map[string]any  `json:"details,omitempty"` // structured metadata
}

// CanonicalField represents mapped column types
type CanonicalField string

const (
	FieldID           CanonicalField = "id"
	FieldTitle        CanonicalField = "title"
	FieldDescription  CanonicalField = "description"
	FieldAcceptance   CanonicalField = "acceptance_criteria"
	FieldFeature      CanonicalField = "feature"
	FieldScenario     CanonicalField = "scenario"
	FieldInstructions CanonicalField = "instructions"
	FieldInputs       CanonicalField = "inputs"
	FieldExpected     CanonicalField = "expected"
	FieldPrecondition CanonicalField = "precondition"
	FieldPriority     CanonicalField = "priority"
	FieldType         CanonicalField = "type"
	FieldStatus       CanonicalField = "status"
	FieldEndpoint     CanonicalField = "endpoint"
	FieldMethod       CanonicalField = "method"
	FieldParameters   CanonicalField = "parameters"
	FieldResponse     CanonicalField = "response"
	FieldStatusCode   CanonicalField = "status_code"
	FieldNotes        CanonicalField = "notes"
	FieldComponent    CanonicalField = "component"
	FieldAssignee     CanonicalField = "assignee"
	FieldCategory     CanonicalField = "category"

	// Spec table fields (Phase 3)
	FieldNo                CanonicalField = "no"
	FieldItemName          CanonicalField = "item_name"
	FieldItemType          CanonicalField = "item_type"
	FieldRequiredOptional  CanonicalField = "required_optional"
	FieldInputRestrictions CanonicalField = "input_restrictions"
	FieldDisplayConditions CanonicalField = "display_conditions"
	FieldAction            CanonicalField = "action"
	FieldNavigationDest    CanonicalField = "navigation_destination"
)

// CellMatrix represents normalized spreadsheet data
type CellMatrix [][]string

// ColumnMap maps canonical fields to column indices
type ColumnMap map[CanonicalField]int

// SpecRow represents a single spec requirement
type SpecRow struct {
	ID           string `json:"id,omitempty"`
	Title        string `json:"title,omitempty"`
	Description  string `json:"description,omitempty"`
	Acceptance   string `json:"acceptance_criteria,omitempty"`
	Feature      string `json:"feature"`
	Scenario     string `json:"scenario,omitempty"`
	Instructions string `json:"instructions,omitempty"`
	Inputs       string `json:"inputs,omitempty"`
	Expected     string `json:"expected,omitempty"`
	Precondition string `json:"precondition,omitempty"`
	Priority     string `json:"priority,omitempty"`
	Type         string `json:"type,omitempty"`
	Status       string `json:"status,omitempty"`
	Endpoint     string `json:"endpoint,omitempty"`
	Method       string `json:"method,omitempty"`
	Parameters   string `json:"parameters,omitempty"`
	Response     string `json:"response,omitempty"`
	StatusCode   string `json:"status_code,omitempty"`
	Notes        string `json:"notes,omitempty"`
	Component    string `json:"component,omitempty"`
	Assignee     string `json:"assignee,omitempty"`
	Category     string `json:"category,omitempty"`

	// Spec table fields (Phase 3)
	No                string `json:"no,omitempty"`
	ItemName          string `json:"item_name,omitempty"`
	ItemType          string `json:"item_type,omitempty"`
	RequiredOptional  string `json:"required_optional,omitempty"`
	InputRestrictions string `json:"input_restrictions,omitempty"`
	DisplayConditions string `json:"display_conditions,omitempty"`
	Action            string `json:"action,omitempty"`
	NavigationDest    string `json:"navigation_destination,omitempty"`

	Metadata map[string]string `json:"metadata,omitempty"`
}

// SpecDoc represents the complete parsed document
type SpecDoc struct {
	Title    string        `json:"title"`
	Rows     []SpecRow     `json:"rows"`
	Warnings []Warning     `json:"warnings"`
	Meta     SpecDocMeta   `json:"meta"`
	Headers  []string      `json:"headers"`
	Prose    *ProseContent `json:"prose,omitempty"`
}

// SpecDocMeta contains metadata about the parsed document
type SpecDocMeta struct {
	SheetName               string         `json:"sheet_name,omitempty"`
	HeaderRow               int            `json:"header_row"`
	ColumnMap               ColumnMap      `json:"column_map"`
	UnmappedColumns         []string       `json:"unmapped_columns,omitempty"`
	TotalRows               int            `json:"total_rows"`
	RowsByFeature           map[string]int `json:"rows_by_feature,omitempty"`
	SourceURL               string         `json:"source_url,omitempty"`
	AIMode                  string         `json:"ai_mode,omitempty"`
	AIUsed                  bool           `json:"ai_used,omitempty"`
	AIDegraded              bool           `json:"ai_degraded,omitempty"`
	AIFallbackReason        string         `json:"ai_fallback_reason,omitempty"`
	AIModel                 string         `json:"ai_model,omitempty"`
	AIPromptVersion         string         `json:"ai_prompt_version,omitempty"`
	AIAvgConfidence         float64        `json:"ai_avg_confidence,omitempty"`
	AIMappedColumns         int            `json:"ai_mapped_columns,omitempty"`
	AIUnmappedColumns       int            `json:"ai_unmapped_columns,omitempty"`
	AIEstimatedInputTokens  int            `json:"ai_estimated_input_tokens,omitempty"`
	AIEstimatedOutputTokens int            `json:"ai_estimated_output_tokens,omitempty"`
	AIEstimatedCostUSD      float64        `json:"ai_estimated_cost_usd,omitempty"`
	OutputFormat            string         `json:"output_format,omitempty"`
	QualityReport           *QualityReport `json:"quality_report,omitempty"`
}

type QualityReport struct {
	StrictMode          bool            `json:"strict_mode"`
	ValidationPassed    bool            `json:"validation_passed"`
	ValidationReason    string          `json:"validation_reason,omitempty"`
	HeaderConfidence    int             `json:"header_confidence"`
	MinHeaderConfidence int             `json:"min_header_confidence"`
	SourceRows          int             `json:"source_rows"`
	ConvertedRows       int             `json:"converted_rows"`
	RowLossRatio        float64         `json:"row_loss_ratio"`
	MaxRowLossRatio     float64         `json:"max_row_loss_ratio"`
	HeaderCount         int             `json:"header_count"`
	MappedColumns       int             `json:"mapped_columns"`
	MappedRatio         float64         `json:"mapped_ratio"`
	CoreFieldCoverage   map[string]bool `json:"core_field_coverage,omitempty"`
}

// ConvertRequest represents the API request for conversion
type ConvertRequest struct {
	PasteText string `json:"paste_text,omitempty"`
	SheetName string `json:"sheet_name,omitempty"`
	Template  string `json:"template,omitempty"`
}

// ConvertResponse represents the API response
type ConvertResponse struct {
	MDFlow   string      `json:"mdflow"`
	Warnings []Warning   `json:"warnings"`
	Meta     SpecDocMeta `json:"meta"`
}
