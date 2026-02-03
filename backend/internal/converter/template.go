package converter

import "fmt"

// TemplateConfig defines a conversion template with field mappings
type TemplateConfig struct {
	Name            string                       `yaml:"name"`
	Description     string                       `yaml:"description"`
	HeaderSynonyms  map[string][]string          `yaml:"header_synonyms"`
	RequiredFields  []string                     `yaml:"required_fields"`
	Output          TemplateOutputConfig         `yaml:"output"`
	Metadata        map[string]interface{}       `yaml:"metadata,omitempty"`
}

// HeaderReference allows selecting a header by exact name or synonym list (first match wins)
type HeaderReference struct {
	Exact  string   `yaml:"exact,omitempty"`
	AnyOf  []string `yaml:"any_of,omitempty"`
}

// RowCardsConfig configures per-row card/section output format
type RowCardsConfig struct {
	TitleFrom HeaderReference `yaml:"title_from"`
	Sections  []RowCardSection `yaml:"sections"`
	Extras    ExtrasConfig    `yaml:"extras,omitempty"`
}

// RowCardSection defines a section within a row card
type RowCardSection struct {
	Label string           `yaml:"label"`
	From  HeaderReference `yaml:"from"`
}

// ExtrasConfig configures how unmapped columns are handled in cards
type ExtrasConfig struct {
	Mode  string `yaml:"mode"`  // append_section, ignore
	Label string `yaml:"label"` // Section label if mode=append_section
}

// NarrativeConfig configures narrative/template-based output (future)
type NarrativeConfig struct {
	Template string `yaml:"template"`
}

// MultiSheetConfig handles multi-sheet document conversion
type MultiSheetConfig struct {
	Enabled           bool              `yaml:"enabled"`
	PerSheetTemplate  map[string]string `yaml:"per_sheet_template,omitempty"`
	DefaultTemplate   string            `yaml:"default_template,omitempty"`
	SheetHeadingLevel int               `yaml:"sheet_heading_level"` // 1 for #, 2 for ##, etc
}

// TemplateOutputConfig configures output format and how unmapped columns are handled
type TemplateOutputConfig struct {
	Type              string               `yaml:"type"` // generic_table, test_spec_markdown, row_cards, narrative
	UnmappedColumns   string               `yaml:"unmapped_columns"` // append_section, ignore, generic_table
	PreserveAllFields bool                 `yaml:"preserve_all_fields"`
	RowCards          *RowCardsConfig      `yaml:"row_cards,omitempty"`
	Narrative         *NarrativeConfig     `yaml:"narrative,omitempty"`
	MultiSheet        *MultiSheetConfig    `yaml:"multi_sheet,omitempty"`
}

// TemplateValidationError represents a template validation error
type TemplateValidationError struct {
	Field   string
	Message string
}

// Validate checks if the template configuration is valid
func (t *TemplateConfig) Validate() []TemplateValidationError {
	var errors []TemplateValidationError

	if t.Name == "" {
		errors = append(errors, TemplateValidationError{"name", "template name is required"})
	}

	if len(t.HeaderSynonyms) == 0 {
		errors = append(errors, TemplateValidationError{"header_synonyms", "at least one field mapping is required"})
	}

	// Validate output type specific requirements (Phase 4)
	outputType := t.Output.Type
	switch outputType {
	case "generic_table":
		// No special requirements
		
	case "test_spec_markdown":
		// No strict field requirements - can work with any columns
		
	case "row_cards":
		if t.Output.RowCards == nil {
			errors = append(errors, TemplateValidationError{
				"output.row_cards", "RowCardsConfig required for row_cards output type"})
		} else {
			// Validate row_cards config references valid headers
			if !isValidHeaderReference(t.Output.RowCards.TitleFrom, t.HeaderSynonyms) {
				errors = append(errors, TemplateValidationError{
					"output.row_cards.title_from", "title_from must reference a field in header_synonyms"})
			}
			for i, section := range t.Output.RowCards.Sections {
				if !isValidHeaderReference(section.From, t.HeaderSynonyms) {
					errors = append(errors, TemplateValidationError{
						"output.row_cards.sections", 
						fmt.Sprintf("Section %d 'from' must reference a field in header_synonyms", i)})
				}
			}
		}
		
	case "narrative":
		if t.Output.Narrative == nil {
			errors = append(errors, TemplateValidationError{
				"output.narrative", "NarrativeConfig required for narrative output type"})
		}
		
	default:
		if outputType != "" {
			errors = append(errors, TemplateValidationError{
				"output.type", "unknown output type: "+outputType})
		}
	}

	return errors
}

// isValidHeaderReference checks if a HeaderReference can resolve to a field in header_synonyms
func isValidHeaderReference(ref HeaderReference, headerSynonyms map[string][]string) bool {
	// If exact match, check if field exists in synonyms
	if ref.Exact != "" {
		normalized := normalizeCanonicalField(ref.Exact)
		for field := range headerSynonyms {
			if field == normalized {
				return true
			}
		}
		return false
	}

	// If any_of, at least one must match a field in synonyms
	if len(ref.AnyOf) > 0 {
		for _, synonym := range ref.AnyOf {
			normalized := normalizeCanonicalField(synonym)
			for field := range headerSynonyms {
				if field == normalized {
					return true
				}
			}
		}
		return false
	}

	return false
}

// GetCanonicalFieldName extracts the canonical field name from the YAML key
// Converts "FieldID" -> "id", "FieldScenario" -> "scenario", etc.
func (t *TemplateConfig) GetCanonicalFieldName(yamlKey string) string {
	// Remove "Field" prefix if present
	if len(yamlKey) > 5 && yamlKey[:5] == "Field" {
		yamlKey = yamlKey[5:]
	}
	// Convert to lowercase and use as canonical field
	return normalizeCanonicalField(yamlKey)
}

// BuildHeaderMap creates a mapping from normalized headers to canonical fields
// This replaces the hardcoded HeaderSynonyms from column_map.go
func (t *TemplateConfig) BuildHeaderMap() map[string]string {
	headerMap := make(map[string]string)

	// For each canonical field and its synonyms
	for fieldName, synonyms := range t.HeaderSynonyms {
		for _, synonym := range synonyms {
			// Normalize the synonym (lowercase, trim, collapse spaces)
			normalized := normalizeHeaderForMatching(synonym)
			headerMap[normalized] = fieldName
		}
	}

	return headerMap
}

// normalizeHeaderForMatching converts a header to normalized form for matching
// Matches the logic in column_map.go's normalizeHeader method
func normalizeHeaderForMatching(header string) string {
	// Convert to lowercase
	h := header
	// Don't use strings.ToLower as it won't work with Japanese characters as expected
	// but for now keep the same logic as column_map.go
	normalized := normalizeHeader(h)
	return normalized
}

// normalizeCanonicalField normalizes field names (used internally)
func normalizeCanonicalField(field string) string {
	// Convert CamelCase to snake_case for consistency
	// For now, just lowercase
	switch field {
	case "ID":
		return "id"
	case "Feature":
		return "feature"
	case "Scenario":
		return "scenario"
	case "Instructions":
		return "instructions"
	case "Inputs":
		return "inputs"
	case "Expected":
		return "expected"
	case "Precondition":
		return "precondition"
	case "Priority":
		return "priority"
	case "Type":
		return "type"
	case "Status":
		return "status"
	case "Endpoint":
		return "endpoint"
	case "Notes":
		return "notes"
	case "No":
		return "no"
	case "ItemName":
		return "item_name"
	case "ItemType":
		return "item_type"
	case "RequiredOptional":
		return "required_optional"
	case "InputRestrictions":
		return "input_restrictions"
	case "DisplayConditions":
		return "display_conditions"
	case "Action":
		return "action"
	case "NavigationDest":
		return "navigation_destination"
	default:
		return field
	}
}

// GetFieldValue extracts the canonical field name for a header from template synonyms
// Returns the CanonicalField and whether it was found
func (t *TemplateConfig) GetFieldValue(normalizedHeader string) (CanonicalField, bool) {
	headerMap := t.BuildHeaderMap()
	if fieldName, ok := headerMap[normalizedHeader]; ok {
		return CanonicalField(fieldName), true
	}
	return "", false
}

// Resolve finds the index of a header matching this HeaderReference
// Returns -1 if no match found
// Searches through actual table headers and uses template header_synonyms for matching
func (ref HeaderReference) Resolve(headers []string, headerMap map[string]string) int {
	// 1. Try exact match first
	if ref.Exact != "" {
		for i, h := range headers {
			if h == ref.Exact {
				return i
			}
		}
		// Try to find by normalized synonym match
		normalized := normalizeHeaderForMatching(ref.Exact)
		if fieldName, ok := headerMap[normalized]; ok {
			// Find first column that maps to this field
			for i, h := range headers {
				if normalizeHeaderForMatching(h) != "" {
					checkNorm := normalizeHeaderForMatching(h)
					if checkNorm == normalized || headerMap[checkNorm] == fieldName {
						return i
					}
				}
			}
		}
		return -1
	}

	// 2. Try any_of list (first match wins)
	if len(ref.AnyOf) > 0 {
		for _, synonym := range ref.AnyOf {
			normalized := normalizeHeaderForMatching(synonym)
			if fieldName, ok := headerMap[normalized]; ok {
				// Find first column that maps to this field
				for i, h := range headers {
					checkNorm := normalizeHeaderForMatching(h)
					if checkNorm == normalized || headerMap[checkNorm] == fieldName {
						return i
					}
				}
			}
		}
		return -1
	}

	return -1
}
