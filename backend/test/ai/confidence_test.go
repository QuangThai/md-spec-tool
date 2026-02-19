package ai

import (
	"testing"

	ailib "github.com/yourorg/md-spec-tool/internal/ai"
)

func TestDefaultThresholds(t *testing.T) {
	thresholds := ailib.DefaultThresholds()

	if thresholds.HighConfidence != 0.80 {
		t.Errorf("expected HighConfidence 0.80, got %v", thresholds.HighConfidence)
	}
	if thresholds.MediumConfidence != 0.65 {
		t.Errorf("expected MediumConfidence 0.65, got %v", thresholds.MediumConfidence)
	}
	if thresholds.LowConfidence != 0.55 {
		t.Errorf("expected LowConfidence 0.55, got %v", thresholds.LowConfidence)
	}
	if thresholds.HeaderConfidenceThreshold != 70 {
		t.Errorf("expected HeaderConfidenceThreshold 70, got %v", thresholds.HeaderConfidenceThreshold)
	}
	if thresholds.RequiredFieldMappingThreshold != 0.50 {
		t.Errorf("expected RequiredFieldMappingThreshold 0.50, got %v", thresholds.RequiredFieldMappingThreshold)
	}
}

func TestIsHighConfidence(t *testing.T) {
	thresholds := ailib.DefaultThresholds()

	tests := []struct {
		name       string
		confidence float64
		expected   bool
	}{
		{"at threshold", 0.80, true},
		{"above threshold", 0.92, true},
		{"just below", 0.79, false},
		{"low", 0.50, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := thresholds.IsHighConfidence(tt.confidence)
			if result != tt.expected {
				t.Errorf("IsHighConfidence(%v) = %v, expected %v", tt.confidence, result, tt.expected)
			}
		})
	}
}

func TestIsMediumConfidence(t *testing.T) {
	thresholds := ailib.DefaultThresholds()

	tests := []struct {
		name       string
		confidence float64
		expected   bool
	}{
		{"at low bound", 0.65, true},
		{"at high bound (exclusive)", 0.80, false},
		{"in range", 0.72, true},
		{"below range", 0.50, false},
		{"above range", 0.92, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := thresholds.IsMediumConfidence(tt.confidence)
			if result != tt.expected {
				t.Errorf("IsMediumConfidence(%v) = %v, expected %v", tt.confidence, result, tt.expected)
			}
		})
	}
}

func TestIsLowConfidence(t *testing.T) {
	thresholds := ailib.DefaultThresholds()

	tests := []struct {
		name       string
		confidence float64
		expected   bool
	}{
		{"just below medium", 0.649, true},
		{"low", 0.50, true},
		{"at medium threshold", 0.65, false},
		{"medium", 0.72, false},
		{"high", 0.92, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := thresholds.IsLowConfidence(tt.confidence)
			if result != tt.expected {
				t.Errorf("IsLowConfidence(%v) = %v, expected %v", tt.confidence, result, tt.expected)
			}
		})
	}
}

func TestShouldReviewMapping(t *testing.T) {
	thresholds := ailib.DefaultThresholds()

	tests := []struct {
		name             string
		avgConfidence    float64
		headerConfidence int
		unmappedCount    int
		totalColumns     int
		expected         bool
	}{
		{
			name:             "high confidence, good headers, low unmapped",
			avgConfidence:    0.85,
			headerConfidence: 90,
			unmappedCount:    0,
			totalColumns:     5,
			expected:         false,
		},
		{
			name:             "low confidence triggers review",
			avgConfidence:    0.55,
			headerConfidence: 85,
			unmappedCount:    1,
			totalColumns:     5,
			expected:         true,
		},
		{
			name:             "low header confidence triggers review",
			avgConfidence:    0.80,
			headerConfidence: 60,
			unmappedCount:    0,
			totalColumns:     5,
			expected:         true,
		},
		{
			name:             "high unmapped ratio triggers review",
			avgConfidence:    0.75,
			headerConfidence: 85,
			unmappedCount:    3, // 60% unmapped > 40% threshold
			totalColumns:     5,
			expected:         true,
		},
		{
			name:             "medium confidence OK without other issues",
			avgConfidence:    0.70,
			headerConfidence: 80,
			unmappedCount:    1, // 20% unmapped < 40%
			totalColumns:     5,
			expected:         false,
		},
		{
			name:             "zero columns edge case",
			avgConfidence:    0.50,
			headerConfidence: 85,
			unmappedCount:    0,
			totalColumns:     0,
			expected:         true, // Low confidence triggers review
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := thresholds.ShouldReviewMapping(
				tt.avgConfidence,
				tt.headerConfidence,
				tt.unmappedCount,
				tt.totalColumns,
			)
			if result != tt.expected {
				t.Errorf("ShouldReviewMapping() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetConfidenceLevel(t *testing.T) {
	thresholds := ailib.DefaultThresholds()

	tests := []struct {
		name       string
		confidence float64
		expected   ailib.ConfidenceLevel
	}{
		{"high", 0.92, ailib.ConfidenceHigh},
		{"at high threshold", 0.80, ailib.ConfidenceHigh},
		{"medium", 0.72, ailib.ConfidenceMedium},
		{"at medium threshold", 0.65, ailib.ConfidenceMedium},
		{"low", 0.50, ailib.ConfidenceLow},
		{"just below medium", 0.64, ailib.ConfidenceLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := thresholds.GetConfidenceLevel(tt.confidence)
			if result != tt.expected {
				t.Errorf("GetConfidenceLevel(%v) = %v, expected %v", tt.confidence, result, tt.expected)
			}
		})
	}
}
