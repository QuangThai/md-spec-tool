package converter

import (
	"strconv"
)

// HeaderResolver resolves headers to canonical fields using a template configuration
// This replaces the hardcoded ColumnMapper approach from column_map.go
type HeaderResolver struct {
	template     *TemplateConfig
	headerMap    map[string]CanonicalField // normalized header -> canonical field
}

// NewHeaderResolver creates a new HeaderResolver with a given template
func NewHeaderResolver(template *TemplateConfig) *HeaderResolver {
	resolver := &HeaderResolver{
		template:  template,
		headerMap: make(map[string]CanonicalField),
	}

	// Build the normalized header map from template synonyms
	for fieldName, synonyms := range template.HeaderSynonyms {
		canonicalField := CanonicalField(fieldName)

		for _, synonym := range synonyms {
			normalized := normalizeHeader(synonym)
			// First match wins (to maintain consistent behavior with ColumnMapper)
			if _, exists := resolver.headerMap[normalized]; !exists {
				resolver.headerMap[normalized] = canonicalField
			}
		}
	}

	return resolver
}

// ResolveHeaders analyzes headers and returns column mapping
// Returns (ColumnMap, unmappedHeaders, warnings)
// Matches the signature of ColumnMapper.MapColumns
func (r *HeaderResolver) ResolveHeaders(headers []string) (ColumnMap, []string, []string) {
	colMap := make(ColumnMap)
	var unmapped []string
	var warnings []string

	seenFields := make(map[CanonicalField]int) // Track which fields we've already mapped

	for i, header := range headers {
		normalized := normalizeHeader(header)

		if canonicalField, ok := r.headerMap[normalized]; ok {
			// Only map if not already mapped (first occurrence wins)
			if _, exists := seenFields[canonicalField]; !exists {
				colMap[canonicalField] = i
				seenFields[canonicalField] = i
			} else {
				// Duplicate field mapping
				warnings = append(warnings, "Duplicate header '"+header+"' ignored (field '"+string(canonicalField)+"' already mapped at column "+strconv.Itoa(seenFields[canonicalField])+")")
				unmapped = append(unmapped, header)
			}
		} else {
			unmapped = append(unmapped, header)
		}
	}

	return colMap, unmapped, warnings
}

// GetFieldValue extracts a field value from a row using the resolved column map
// Matches the signature and behavior of GetFieldValue in column_map.go
func (r *HeaderResolver) GetFieldValue(row []string, colMap ColumnMap, field CanonicalField) string {
	if idx, ok := colMap[field]; ok && idx < len(row) {
		return row[idx]
	}
	return ""
}

// Note: normalizeHeader is defined in renderer.go and used here.
// The function is defined once in the package to avoid duplication.

// GetTemplate returns the underlying template configuration
func (r *HeaderResolver) GetTemplate() *TemplateConfig {
	return r.template
}
