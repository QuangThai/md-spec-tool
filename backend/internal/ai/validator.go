package ai

import "fmt"

// Validator checks AI output for semantic correctness
type Validator struct {
	canonicalFields map[string]struct{}
}

// NewValidator creates a new validator with canonical fields
func NewValidator() *Validator {
	fields := make(map[string]struct{})
	for name := range CanonicalFields {
		fields[name] = struct{}{}
	}
	return &Validator{canonicalFields: fields}
}

// ValidateColumnMapping ensures output is semantically correct.
// Use ValidateColumnMappingWithHeaders when header count is available for range validation.
func (v *Validator) ValidateColumnMapping(result *ColumnMappingResult) error {
	return v.ValidateColumnMappingWithHeaders(result, -1)
}

// ValidateColumnMappingWithHeaders ensures output is semantically correct and column_index is in [0, headerCount).
// If headerCount < 0, skips column index range validation.
func (v *Validator) ValidateColumnMappingWithHeaders(result *ColumnMappingResult, headerCount int) error {
	if result == nil {
		return ErrAIValidationFailed
	}

	// Validate schema version
	if result.SchemaVersion != SchemaVersionColumnMapping {
		return fmt.Errorf("%w: unknown schema version: %s", ErrAIValidationFailed, result.SchemaVersion)
	}

	// Dedupe by canonical_name: keep first occurrence
	seen := make(map[string]bool)
	var validMappings []CanonicalFieldMapping
	for _, m := range result.CanonicalFields {
		if seen[m.CanonicalName] {
			continue
		}
		seen[m.CanonicalName] = true
		validMappings = append(validMappings, m)
	}
	result.CanonicalFields = validMappings

	// Recalculate metadata after dedupe
	result.Meta.MappedColumns = len(result.CanonicalFields)
	if result.Meta.TotalColumns > 0 {
		unmapped := result.Meta.TotalColumns - result.Meta.MappedColumns
		if unmapped < 0 {
			unmapped = 0
		}
		result.Meta.UnmappedColumns = unmapped
	} else if len(result.ExtraColumns) > 0 {
		result.Meta.UnmappedColumns = len(result.ExtraColumns)
	}
	if len(result.CanonicalFields) > 0 {
		sum := 0.0
		for _, m := range result.CanonicalFields {
			sum += m.Confidence
		}
		result.Meta.AvgConfidence = sum / float64(len(result.CanonicalFields))
	} else {
		result.Meta.AvgConfidence = 0
	}

	// Validate canonical field mappings
	for i, m := range result.CanonicalFields {
		// Validate canonical name is known
		if _, ok := v.canonicalFields[m.CanonicalName]; !ok {
			return fmt.Errorf("%w: unknown canonical field: %s", ErrAIValidationFailed, m.CanonicalName)
		}

		// Validate confidence bounds
		if m.Confidence < 0 || m.Confidence > 1 {
			result.CanonicalFields[i].Confidence = clamp(m.Confidence, 0, 1)
		}

		// Validate column index
		if m.ColumnIndex < 0 {
			return fmt.Errorf("%w: negative column index", ErrAIValidationFailed)
		}
		if headerCount >= 0 && m.ColumnIndex >= headerCount {
			return fmt.Errorf("%w: column_index %d out of range [0, %d)", ErrAIValidationFailed, m.ColumnIndex, headerCount)
		}

		// Cap reasoning length
		if len(m.Reasoning) > 256 {
			result.CanonicalFields[i].Reasoning = m.Reasoning[:256] + "..."
		}

		// Validate alternatives
		for j, alt := range m.Alternatives {
			if alt.Confidence < 0 || alt.Confidence > 1 {
				result.CanonicalFields[i].Alternatives[j].Confidence = clamp(alt.Confidence, 0, 1)
			}
			if alt.ColumnIndex < 0 {
				return fmt.Errorf("%w: negative column index in alternative", ErrAIValidationFailed)
			}
		}
	}

	// Validate extra column mappings
	for i, extra := range result.ExtraColumns {
		if extra.Confidence < 0 || extra.Confidence > 1 {
			result.ExtraColumns[i].Confidence = clamp(extra.Confidence, 0, 1)
		}
		if extra.ColumnIndex < 0 {
			return fmt.Errorf("%w: negative column index in extra", ErrAIValidationFailed)
		}
		if headerCount >= 0 && extra.ColumnIndex >= headerCount {
			return fmt.Errorf("%w: extra column_index %d out of range [0, %d)", ErrAIValidationFailed, extra.ColumnIndex, headerCount)
		}
	}

	// Validate metadata
	if result.Meta.TotalColumns < 0 || result.Meta.MappedColumns < 0 || result.Meta.UnmappedColumns < 0 {
		return fmt.Errorf("%w: negative column counts", ErrAIValidationFailed)
	}

	if result.Meta.AvgConfidence < 0 || result.Meta.AvgConfidence > 1 {
		result.Meta.AvgConfidence = clamp(result.Meta.AvgConfidence, 0, 1)
	}

	return nil
}

// ValidatePasteAnalysis ensures paste analysis output is valid
func (v *Validator) ValidatePasteAnalysis(result *PasteAnalysis) error {
	if result == nil {
		return ErrAIValidationFailed
	}

	// Validate schema version
	if result.SchemaVersion != SchemaVersionPasteAnalysis {
		return fmt.Errorf("%w: unknown schema version: %s", ErrAIValidationFailed, result.SchemaVersion)
	}

	// Validate input type
	validInputTypes := map[string]bool{
		"table":           true,
		"backlog_list":    true,
		"test_cases":      true,
		"product_backlog": true,
		"issue_tracker":   true,
		"api_spec":        true,
		"ui_spec":         true,
		"prose":           true,
		"mixed":           true,
		"unknown":         true,
	}
	if !validInputTypes[result.InputType] {
		return fmt.Errorf("%w: invalid input type: %s", ErrAIValidationFailed, result.InputType)
	}

	// Validate detected format
	validFormats := map[string]bool{
		"csv":            true,
		"tsv":            true,
		"markdown_table": true,
		"free_text":      true,
		"mixed":          true,
	}
	if !validFormats[result.DetectedFormat] {
		return fmt.Errorf("%w: invalid detected format: %s", ErrAIValidationFailed, result.DetectedFormat)
	}

	// Validate suggested output
	validOutputs := map[string]bool{
		"spec":  true,
		"table": true,
	}
	if !validOutputs[result.SuggestedOutput] {
		return fmt.Errorf("%w: invalid suggested output: %s", ErrAIValidationFailed, result.SuggestedOutput)
	}

	// Validate confidence
	if result.Confidence < 0 || result.Confidence > 1 {
		result.Confidence = clamp(result.Confidence, 0, 1)
	}

	// Validate normalized table structure
	if len(result.NormalizedTable) > 0 {
		headerCount := len(result.NormalizedTable[0])
		for i, row := range result.NormalizedTable[1:] {
			if len(row) != headerCount {
				return fmt.Errorf("%w: inconsistent row length at row %d", ErrAIValidationFailed, i+1)
			}
		}
	}

	return nil
}

// ValidateMappingSemantics performs semantic validation on column mappings
// to detect conflicting mappings, suspicious confidence scores, and missing required fields
func (v *Validator) ValidateMappingSemantics(result *ColumnMappingResult, detectedSchema string) *SemanticValidationResult {
	issues := []SemanticIssue{}
	avgConfidence := result.Meta.AvgConfidence

	// Check for suspicious low average confidence
	if avgConfidence < 0.5 {
		issues = append(issues, SemanticIssue{
			Type:       "ambiguous",
			Severity:   "warn",
			Message:    "Low average mapping confidence suggests ambiguous headers",
			Suggestion: "Review mappings and consider refinement or interactive disambiguation",
		})
	}

	// Check for conflicting column indices
	usedIndices := make(map[int]string)
	for _, m := range result.CanonicalFields {
		if existing, exists := usedIndices[m.ColumnIndex]; exists && existing != m.CanonicalName {
			issues = append(issues, SemanticIssue{
				Type:       "inconsistent",
				Severity:   "error",
				Message:    fmt.Sprintf("Column index %d mapped to multiple fields: %s and %s", m.ColumnIndex, existing, m.CanonicalName),
				Suggestion: "Resolve duplicate column mappings",
			})
		}
		usedIndices[m.ColumnIndex] = m.CanonicalName
	}

	// Check for suspicious low confidence mappings
	for i, m := range result.CanonicalFields {
		if m.Confidence < 0.4 && m.Confidence > 0 {
			issues = append(issues, SemanticIssue{
				Type:       "ambiguous",
				Severity:   "warn",
				Message:    fmt.Sprintf("Low confidence mapping (%.1f%%) for '%s' -> '%s'", m.Confidence*100, m.SourceHeader, m.CanonicalName),
				Suggestion: fmt.Sprintf("Consider moving '%s' to extra_columns or refining with more context", m.SourceHeader),
			})
		}

		// Check for alternatives with higher confidence than selected mapping
		if len(m.Alternatives) > 0 {
			for _, alt := range m.Alternatives {
				if alt.Confidence > m.Confidence {
					issues = append(issues, SemanticIssue{
						Type:       "inconsistent",
						Severity:   "warn",
						Message:    fmt.Sprintf("Alternative mapping has higher confidence (%.1f%% vs %.1f%%) for column %d", alt.Confidence*100, m.Confidence*100, i),
						Suggestion: "Review and consider swapping selected mapping with higher-confidence alternative",
					})
					break // Only report once per field
				}
			}
		}
	}

	// Check for schema-specific required fields
	requiredBySchema := getRequiredFieldsBySchema(detectedSchema)
	mappedFields := make(map[string]bool)
	for _, m := range result.CanonicalFields {
		mappedFields[m.CanonicalName] = true
	}

	for _, required := range requiredBySchema {
		if !mappedFields[required] {
			issues = append(issues, SemanticIssue{
				Type:       "incomplete",
				Severity:   "warn",
				Message:    fmt.Sprintf("Missing required field '%s' for detected schema '%s'", required, detectedSchema),
				Suggestion: fmt.Sprintf("Add mapping for '%s' or reconsider schema detection", required),
			})
		}
	}

	// Determine overall status
	overall := "good"
	if len(issues) > 0 {
		overall = "needs_improvement"
		for _, issue := range issues {
			if issue.Severity == "error" {
				overall = "poor"
				break
			}
		}
	}

	// Calculate confidence as inverse of issues
	confidence := 1.0 - (float64(len(issues)) * 0.1) // Each issue reduces by 0.1
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return &SemanticValidationResult{
		Issues:     issues,
		Overall:    overall,
		Score:      avgConfidence,
		Confidence: confidence,
	}
}

// getRequiredFieldsBySchema returns required canonical fields for a detected schema type
func getRequiredFieldsBySchema(schema string) []string {
	requirements := map[string][]string{
		"test_case":       {"id", "scenario", "instructions", "expected"},
		"product_backlog": {"id", "title", "description", "acceptance_criteria"},
		"issue_tracker":   {"id", "feature", "priority", "status"},
		"api_spec":        {"endpoint", "method", "parameters", "response"},
		"ui_spec":         {"item_name", "item_type", "action"},
		"generic":         {"id", "feature"},
	}
	if reqs, ok := requirements[schema]; ok {
		return reqs
	}
	return []string{"id", "feature"} // minimal default
}

// clamp constrains a value between min and max
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
