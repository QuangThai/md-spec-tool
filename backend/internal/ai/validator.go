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

// ValidateColumnMapping ensures output is semantically correct
func (v *Validator) ValidateColumnMapping(result *ColumnMappingResult) error {
	if result == nil {
		return ErrAIValidationFailed
	}

	// Validate schema version
	if result.SchemaVersion != SchemaVersionColumnMapping {
		return fmt.Errorf("%w: unknown schema version: %s", ErrAIValidationFailed, result.SchemaVersion)
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
		"table":       true,
		"backlog_list": true,
		"test_cases":  true,
		"prose":       true,
		"mixed":       true,
		"unknown":     true,
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
