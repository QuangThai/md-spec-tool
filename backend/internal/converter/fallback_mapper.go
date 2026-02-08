package converter

import (
	"strings"
)

// FallbackMapper uses heuristic logic when AI is unavailable
type FallbackMapper struct {
	synonyms map[string]CanonicalField
}

// NewFallbackMapper creates a new fallback mapper with default synonyms
func NewFallbackMapper() *FallbackMapper {
	return &FallbackMapper{
		synonyms: HeaderSynonyms,
	}
}

// MapColumns analyzes headers and returns column mapping using heuristics
// Used when AI is unavailable or AI mapping fails/has low confidence
func (f *FallbackMapper) MapColumns(headers []string) (ColumnMap, []string) {
	colMap := make(ColumnMap)
	var unmapped []string

	for i, header := range headers {
		normalized := f.normalizeHeader(header)
		if field, ok := f.synonyms[normalized]; ok {
			// Only map if not already mapped (first occurrence wins)
			if _, exists := colMap[field]; !exists {
				colMap[field] = i
			}
		} else {
			unmapped = append(unmapped, header)
		}
	}

	return colMap, unmapped
}

// normalizeHeader converts a header to lowercase and normalizes whitespace
func (f *FallbackMapper) normalizeHeader(header string) string {
	// Convert to lowercase
	h := strings.ToLower(header)
	// Remove extra whitespace
	h = strings.TrimSpace(h)
	// Replace multiple spaces with single space
	h = strings.Join(strings.Fields(h), " ")
	return h
}

// WithCustomSynonyms allows setting custom synonyms for a mapper
func (f *FallbackMapper) WithCustomSynonyms(synonyms map[string]CanonicalField) *FallbackMapper {
	f.synonyms = synonyms
	return f
}

// GetUnmappedHeaders returns headers that couldn't be mapped
func (f *FallbackMapper) GetUnmappedHeaders(headers []string, colMap ColumnMap) []string {
	mappedIndices := make(map[int]bool)
	for _, idx := range colMap {
		mappedIndices[idx] = true
	}

	var unmapped []string
	for i, header := range headers {
		if !mappedIndices[i] {
			unmapped = append(unmapped, header)
		}
	}
	return unmapped
}
