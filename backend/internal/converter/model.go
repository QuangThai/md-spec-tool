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
	FieldNotes        CanonicalField = "notes"
	
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
	ID           string            `json:"id,omitempty"`
	Feature      string            `json:"feature"`
	Scenario     string            `json:"scenario,omitempty"`
	Instructions string            `json:"instructions,omitempty"`
	Inputs       string            `json:"inputs,omitempty"`
	Expected     string            `json:"expected,omitempty"`
	Precondition string            `json:"precondition,omitempty"`
	Priority     string            `json:"priority,omitempty"`
	Type         string            `json:"type,omitempty"`
	Status       string            `json:"status,omitempty"`
	Endpoint     string            `json:"endpoint,omitempty"`
	Notes        string            `json:"notes,omitempty"`
	
	// Spec table fields (Phase 3)
	No                string            `json:"no,omitempty"`
	ItemName          string            `json:"item_name,omitempty"`
	ItemType          string            `json:"item_type,omitempty"`
	RequiredOptional  string            `json:"required_optional,omitempty"`
	InputRestrictions string            `json:"input_restrictions,omitempty"`
	DisplayConditions string            `json:"display_conditions,omitempty"`
	Action            string            `json:"action,omitempty"`
	NavigationDest    string            `json:"navigation_destination,omitempty"`
	
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// SpecDoc represents the complete parsed document
type SpecDoc struct {
	Title    string         `json:"title"`
	Rows     []SpecRow      `json:"rows"`
	Warnings []Warning      `json:"warnings"`
	Meta     SpecDocMeta    `json:"meta"`
	Headers  []string       `json:"headers"`
	Prose    *ProseContent  `json:"prose,omitempty"`
}

// SpecDocMeta contains metadata about the parsed document
type SpecDocMeta struct {
	SheetName       string              `json:"sheet_name,omitempty"`
	HeaderRow       int                 `json:"header_row"`
	ColumnMap       ColumnMap           `json:"column_map"`
	UnmappedColumns []string            `json:"unmapped_columns,omitempty"`
	TotalRows       int                 `json:"total_rows"`
	RowsByFeature   map[string]int      `json:"rows_by_feature,omitempty"`
	SourceURL       string              `json:"source_url,omitempty"`
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
