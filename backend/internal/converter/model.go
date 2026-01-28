package converter

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
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// SpecDoc represents the complete parsed document
type SpecDoc struct {
	Title    string         `json:"title"`
	Rows     []SpecRow      `json:"rows"`
	Warnings []string       `json:"warnings"`
	Meta     SpecDocMeta    `json:"meta"`
	Headers  []string       `json:"headers"`
}

// SpecDocMeta contains metadata about the parsed document
type SpecDocMeta struct {
	SheetName       string              `json:"sheet_name,omitempty"`
	HeaderRow       int                 `json:"header_row"`
	ColumnMap       ColumnMap           `json:"column_map"`
	UnmappedColumns []string            `json:"unmapped_columns,omitempty"`
	TotalRows       int                 `json:"total_rows"`
	RowsByFeature   map[string]int      `json:"rows_by_feature,omitempty"`
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
	Warnings []string    `json:"warnings"`
	Meta     SpecDocMeta `json:"meta"`
}
