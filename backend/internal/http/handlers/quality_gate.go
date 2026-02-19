package handlers

import (
	"github.com/yourorg/md-spec-tool/internal/ai"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

// RequiresReview signals low-confidence conversion outputs that should be reviewed in UI.
// Uses confidence thresholds to determine if mapping quality warrants review.
func RequiresReview(meta converter.SpecDocMeta, warnings []converter.Warning) bool {
	thresholds := ai.DefaultThresholds()
	qualityReport := meta.QualityReport

	// Check validation failures
	if qualityReport != nil && !qualityReport.ValidationPassed {
		return true
	}

	// Check AI confidence if available
	if meta.AIUsed && meta.AIAvgConfidence > 0 {
		if thresholds.IsLowConfidence(meta.AIAvgConfidence) {
			return true
		}
	}

	// Check header confidence
	if qualityReport != nil && qualityReport.HeaderConfidence < thresholds.HeaderConfidenceThreshold {
		return true
	}

	// Check unmapped column ratio
	if qualityReport != nil {
		totalCols := qualityReport.HeaderCount
		unmappedCols := len(meta.UnmappedColumns)
		if totalCols > 0 {
			if thresholds.ShouldReviewMapping(meta.AIAvgConfidence, qualityReport.HeaderConfidence, unmappedCols, totalCols) {
				return true
			}
		}
	}

	// Check for high-severity warnings
	for _, w := range warnings {
		if w.Category == converter.CatHeader || w.Category == converter.CatMapping {
			return true
		}
		if w.Severity == converter.SeverityError {
			return true
		}
	}

	return false
}
