package converter

// TemplateConfig defines a conversion template with field mappings
type TemplateConfig struct {
	Name           string                 `yaml:"name"`
	Description    string                 `yaml:"description"`
	HeaderSynonyms map[string][]string    `yaml:"header_synonyms"`
	RequiredFields []string               `yaml:"required_fields"`
	Output         TemplateOutputConfig   `yaml:"output"`
	Metadata       map[string]interface{} `yaml:"metadata,omitempty"`
}

// TemplateOutputConfig configures output format and how unmapped columns are handled
type TemplateOutputConfig struct {
	Type              string `yaml:"type"`             // spec, table
	UnmappedColumns   string `yaml:"unmapped_columns"` // append_section, ignore
	PreserveAllFields bool   `yaml:"preserve_all_fields"`
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

	// Validate output type specific requirements.
	outputType := t.Output.Type
	switch outputType {
	case "spec", "table", "":
		// No strict field requirements - can work with any columns

	default:
		errors = append(errors, TemplateValidationError{
			"output.type", "unknown output type: " + outputType + " (supported: spec, table)"})
	}

	return errors
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
