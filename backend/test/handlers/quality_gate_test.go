package handlers

import (
	"testing"

	"github.com/yourorg/md-spec-tool/internal/converter"
	"github.com/yourorg/md-spec-tool/internal/http/handlers"
)

func TestRequiresReview_HighConfidence(t *testing.T) {
	// High confidence mapping should not require review
	meta := converter.SpecDocMeta{
		AIUsed:          true,
		AIAvgConfidence: 0.92,
		UnmappedColumns: []string{},
		QualityReport: &converter.QualityReport{
			ValidationPassed: true,
			HeaderConfidence: 90,
			HeaderCount:      5,
			MappedColumns:    5,
			MappedRatio:      1.0,
		},
	}
	warnings := []converter.Warning{}

	result := handlers.RequiresReview(meta, warnings)
	if result {
		t.Errorf("expected needs_review=false for high confidence (0.92), got true")
	}
}

func TestRequiresReview_LowConfidence(t *testing.T) {
	// Low confidence mapping should require review
	meta := converter.SpecDocMeta{
		AIUsed:          true,
		AIAvgConfidence: 0.55,
		UnmappedColumns: []string{},
		QualityReport: &converter.QualityReport{
			ValidationPassed: true,
			HeaderConfidence: 85,
			HeaderCount:      5,
			MappedColumns:    3,
			MappedRatio:      0.60,
		},
	}
	warnings := []converter.Warning{}

	result := handlers.RequiresReview(meta, warnings)
	if !result {
		t.Errorf("expected needs_review=true for low confidence (0.55), got false")
	}
}

func TestRequiresReview_LowHeaderConfidence(t *testing.T) {
	// Low header confidence should require review
	meta := converter.SpecDocMeta{
		AIUsed:          true,
		AIAvgConfidence: 0.80,
		UnmappedColumns: []string{},
		QualityReport: &converter.QualityReport{
			ValidationPassed: true,
			HeaderConfidence: 60, // Below 70% threshold
			HeaderCount:      5,
			MappedColumns:    4,
			MappedRatio:      0.80,
		},
	}
	warnings := []converter.Warning{}

	result := handlers.RequiresReview(meta, warnings)
	if !result {
		t.Errorf("expected needs_review=true for low header confidence (60%%), got false")
	}
}

func TestRequiresReview_HighUnmappedRatio(t *testing.T) {
	// High unmapped ratio should require review
	meta := converter.SpecDocMeta{
		AIUsed:          true,
		AIAvgConfidence: 0.72,
		UnmappedColumns: []string{"Col4", "Col5", "Col6"}, // 3 unmapped = 60% ratio
		QualityReport: &converter.QualityReport{
			ValidationPassed: true,
			HeaderConfidence: 80,
			HeaderCount:      5,
			MappedColumns:    2,
			MappedRatio:      0.40,
		},
	}
	warnings := []converter.Warning{}

	result := handlers.RequiresReview(meta, warnings)
	if !result {
		t.Errorf("expected needs_review=true for high unmapped ratio (60%%), got false")
	}
}

func TestRequiresReview_ValidationFailed(t *testing.T) {
	// Validation failure should require review
	meta := converter.SpecDocMeta{
		AIUsed:          true,
		AIAvgConfidence: 0.80,
		UnmappedColumns: []string{},
		QualityReport: &converter.QualityReport{
			ValidationPassed: false,
			HeaderConfidence: 85,
			HeaderCount:      5,
			MappedColumns:    5,
			MappedRatio:      1.0,
		},
	}
	warnings := []converter.Warning{}

	result := handlers.RequiresReview(meta, warnings)
	if !result {
		t.Errorf("expected needs_review=true for validation failure, got false")
	}
}

func TestRequiresReview_HeaderWarning(t *testing.T) {
	// Header category warnings should require review
	meta := converter.SpecDocMeta{
		AIUsed:          true,
		AIAvgConfidence: 0.80,
		UnmappedColumns: []string{},
		QualityReport: &converter.QualityReport{
			ValidationPassed: true,
			HeaderConfidence: 85,
			HeaderCount:      5,
			MappedColumns:    5,
			MappedRatio:      1.0,
		},
	}
	warnings := []converter.Warning{
		{
			Code:     "LOW_HEADER_CONFIDENCE",
			Message:  "Header confidence is low",
			Severity: converter.SeverityWarn,
			Category: converter.CatHeader,
		},
	}

	result := handlers.RequiresReview(meta, warnings)
	if !result {
		t.Errorf("expected needs_review=true for header warning, got false")
	}
}

func TestRequiresReview_ErrorSeverityWarning(t *testing.T) {
	// Error severity warnings should require review
	meta := converter.SpecDocMeta{
		AIUsed:          true,
		AIAvgConfidence: 0.80,
		UnmappedColumns: []string{},
		QualityReport: &converter.QualityReport{
			ValidationPassed: true,
			HeaderConfidence: 85,
			HeaderCount:      5,
			MappedColumns:    5,
			MappedRatio:      1.0,
		},
	}
	warnings := []converter.Warning{
		{
			Code:     "SOME_ERROR",
			Message:  "Non-fatal error",
			Severity: converter.SeverityError,
			Category: converter.CatRows,
		},
	}

	result := handlers.RequiresReview(meta, warnings)
	if !result {
		t.Errorf("expected needs_review=true for error severity warning, got false")
	}
}

func TestRequiresReview_MediumConfidenceWithoutWarnings(t *testing.T) {
	// Medium confidence without warnings should not require review
	meta := converter.SpecDocMeta{
		AIUsed:          true,
		AIAvgConfidence: 0.72, // In medium range [0.65, 0.80)
		UnmappedColumns: []string{},
		QualityReport: &converter.QualityReport{
			ValidationPassed: true,
			HeaderConfidence: 85,
			HeaderCount:      5,
			MappedColumns:    4,
			MappedRatio:      0.80,
		},
	}
	warnings := []converter.Warning{}

	result := handlers.RequiresReview(meta, warnings)
	if result {
		t.Errorf("expected needs_review=false for medium confidence without warnings, got true")
	}
}

func TestRequiresReview_NilQualityReportDoesNotPanic(t *testing.T) {
	meta := converter.SpecDocMeta{
		AIUsed:          true,
		AIAvgConfidence: 0.72,
		UnmappedColumns: []string{"Notes"},
		QualityReport:   nil,
	}
	warnings := []converter.Warning{}

	result := handlers.RequiresReview(meta, warnings)
	if result {
		t.Errorf("expected needs_review=false when quality report is nil and no warning thresholds are violated, got true")
	}
}
