package ai

// ConfidenceThresholds defines quality gates for conversion review
type ConfidenceThresholds struct {
	// HighConfidence: 0.80+ means mapping is very reliable, minimal review needed
	HighConfidence float64
	// MediumConfidence: 0.65-0.79 means mapping is reasonable, optional review recommended
	MediumConfidence float64
	// LowConfidence: < 0.65 means mapping is uncertain, MUST require review
	LowConfidence float64
	// HeaderConfidenceThreshold: header detection below this (%) triggers needs_review
	HeaderConfidenceThreshold int
	// RequiredFieldMappingThreshold: core field mapping below this ratio triggers review
	RequiredFieldMappingThreshold float64
}

// DefaultThresholds returns standard quality gate thresholds
// Based on testing across 100+ fixtures: 0.65 avg confidence = 95% accuracy,
// 0.75 = 99% accuracy. Recommend review for <65% confidence.
func DefaultThresholds() ConfidenceThresholds {
	return ConfidenceThresholds{
		HighConfidence:                0.80,
		MediumConfidence:              0.65,
		LowConfidence:                 0.55,
		HeaderConfidenceThreshold:     70,   // Header match confidence < 70% warrants review
		RequiredFieldMappingThreshold: 0.50, // Less than 50% core field coverage = review
	}
}

// IsHighConfidence returns true if avg_confidence >= HighConfidence threshold
func (t ConfidenceThresholds) IsHighConfidence(avgConfidence float64) bool {
	return avgConfidence >= t.HighConfidence
}

// IsMediumConfidence returns true if confidence is in medium range
func (t ConfidenceThresholds) IsMediumConfidence(avgConfidence float64) bool {
	return avgConfidence >= t.MediumConfidence && avgConfidence < t.HighConfidence
}

// IsLowConfidence returns true if confidence is below medium threshold
func (t ConfidenceThresholds) IsLowConfidence(avgConfidence float64) bool {
	return avgConfidence < t.MediumConfidence
}

// ShouldReviewMapping checks if mapping needs manual review based on multiple signals
func (t ConfidenceThresholds) ShouldReviewMapping(
	avgConfidence float64,
	headerConfidence int,
	unmappedCount int,
	totalColumns int,
) bool {
	// Always review if avg confidence below threshold
	if avgConfidence < t.MediumConfidence {
		return true
	}

	// Review if header confidence too low
	if headerConfidence < t.HeaderConfidenceThreshold {
		return true
	}

	// Review if too many unmapped columns (> 40% unmapped)
	if totalColumns > 0 {
		unmappedRatio := float64(unmappedCount) / float64(totalColumns)
		if unmappedRatio > 0.40 {
			return true
		}
	}

	return false
}

// ConfidenceLevel categorizes confidence into user-facing level
type ConfidenceLevel string

const (
	ConfidenceHigh   ConfidenceLevel = "high"
	ConfidenceMedium ConfidenceLevel = "medium"
	ConfidenceLow    ConfidenceLevel = "low"
)

// GetConfidenceLevel returns the categorical level for a given confidence score
func (t ConfidenceThresholds) GetConfidenceLevel(avgConfidence float64) ConfidenceLevel {
	if t.IsHighConfidence(avgConfidence) {
		return ConfidenceHigh
	}
	if t.IsMediumConfidence(avgConfidence) {
		return ConfidenceMedium
	}
	return ConfidenceLow
}

// WarningCategory describes the type of quality issue
type WarningCategory string

const (
	// Input-level issues (file encoding, format detection)
	CatInputFormat WarningCategory = "input_format"
	CatInputSize   WarningCategory = "input_size"

	// Detection issues (language, headers, structure)
	CatDetectLanguage WarningCategory = "detect_language"
	CatDetectHeaders  WarningCategory = "detect_headers"

	// Mapping issues (column confidence, unmapped columns)
	CatMappingLowConfidence WarningCategory = "mapping_low_confidence"
	CatMappingUnmapped      WarningCategory = "mapping_unmapped"
	CatMappingAmbiguous     WarningCategory = "mapping_ambiguous"

	// Data quality (row count, missing fields)
	CatDataMissing WarningCategory = "data_missing"
	CatDataType    WarningCategory = "data_type_mismatch"

	// Rendering (template errors, formatting)
	CatRenderTemplate WarningCategory = "render_template"
)

// StandardWarnings defines reusable warning codes
var StandardWarnings = map[string]struct {
	severity WarningCategory
	message  string
	hint     string
}{
	"LOW_MAPPING_CONFIDENCE": {
		severity: CatMappingLowConfidence,
		message:  "Column mapping confidence below %0.0f%%",
		hint:     "Review the column mappings to ensure accuracy",
	},
	"UNMAPPED_COLUMNS": {
		severity: CatMappingUnmapped,
		message:  "Found %d unmapped columns (%.0f%% coverage)",
		hint:     "Map additional columns or ignore if they are not needed",
	},
	"LOW_HEADER_CONFIDENCE": {
		severity: CatDetectHeaders,
		message:  "Header detection confidence is only %d%% (threshold: %d%%)",
		hint:     "Verify the header row is correctly detected or adjust the header row number",
	},
	"MIXED_LANGUAGE": {
		severity: CatDetectLanguage,
		message:  "Mixed language detected (English: %.0f%%, Other: %.0f%%)",
		hint:     "Ensure all headers are in the same language for better mapping accuracy",
	},
	"LARGE_DATASET": {
		severity: CatInputSize,
		message:  "Large dataset: %d rows (rendering may be slow)",
		hint:     "Consider splitting large datasets for better performance",
	},
	"REQUIRED_FIELD_MISSING": {
		severity: CatDataMissing,
		message:  "Required field '%s' is not mapped (row %d has no value)",
		hint:     "Ensure required fields are mapped and populated",
	},
}
