package handlers

import "github.com/yourorg/md-spec-tool/internal/converter"

// Request types for conversion endpoints
type PasteConvertRequest struct {
	PasteText       string                     `json:"paste_text" binding:"required"`
	Template        string                     `json:"template"`
	Format          string                     `json:"format"`
	ColumnOverrides map[string]string           `json:"column_overrides,omitempty"`
	ValidationRules *converter.ValidationRules `json:"validation_rules,omitempty"`
	// Phase 3: Convert options
	IncludeMetadata *bool `json:"include_metadata,omitempty"` // default true when nil
	NumberRows      *bool `json:"number_rows,omitempty"`     // default false when nil
}

// Response types for conversion endpoints
type MDFlowConvertResponse struct {
	MDFlow   string                `json:"mdflow"`
	Warnings []converter.Warning   `json:"warnings"`
	Meta     converter.SpecDocMeta `json:"meta"`
	Format   string                `json:"format"`
	Template string                `json:"template"`
}

type InputAnalysisResponse struct {
	Type       string  `json:"type"` // 'markdown' | 'table' | 'unknown'
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason,omitempty"`
}

type SheetsResponse struct {
	Sheets      []string `json:"sheets"`
	ActiveSheet string   `json:"active_sheet"`
}

// Response types for preview endpoints
type PreviewResponse struct {
	Headers            []string                         `json:"headers"`
	Rows               [][]string                       `json:"rows"`
	TotalRows          int                              `json:"total_rows"`
	PreviewRows        int                              `json:"preview_rows"`
	HeaderRow          int                              `json:"header_row"`
	Confidence         int                              `json:"confidence"`
	ColumnMapping      map[string]string                `json:"column_mapping"`
	UnmappedCols       []string                         `json:"unmapped_columns"`
	MappingQuality     *converter.PreviewMappingQuality `json:"mapping_quality,omitempty"`
	Blocks             []PreviewBlock                   `json:"blocks,omitempty"`
	SelectedBlockID    string                           `json:"selected_block_id,omitempty"`
	SelectedBlockRange string                           `json:"selected_block_range,omitempty"`
	InputType          string                           `json:"input_type"`
	AIAvailable        bool                             `json:"ai_available"`
}

type PreviewBlock struct {
	ID             string                           `json:"id"`
	Range          string                           `json:"range"`
	TotalRows      int                              `json:"total_rows"`
	TotalColumns   int                              `json:"total_columns"`
	LanguageHint   string                           `json:"language_hint"`
	EnglishScore   float64                          `json:"english_score"`
	HeaderRow      int                              `json:"header_row"`
	Confidence     int                              `json:"confidence"`
	MappingQuality *converter.PreviewMappingQuality `json:"mapping_quality,omitempty"`
}
