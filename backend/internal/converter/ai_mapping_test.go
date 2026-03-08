package converter

import (
	"testing"

	"github.com/yourorg/md-spec-tool/internal/ai"
)

func TestCountHighConfidenceAIMappings(t *testing.T) {
	headers := []string{"ID", "Title", "Status", "Notes"}
	tests := []struct {
		name     string
		result   *ai.ColumnMappingResult
		threshold float64
		want     int
	}{
		{
			name:     "nil result",
			result:   nil,
			threshold: 0.75,
			want:     0,
		},
		{
			name: "all below threshold",
			result: &ai.ColumnMappingResult{
				CanonicalFields: []ai.CanonicalFieldMapping{
					{CanonicalName: "id", ColumnIndex: 0, Confidence: 0.5},
					{CanonicalName: "title", ColumnIndex: 1, Confidence: 0.6},
				},
			},
			threshold: 0.75,
			want:      0,
		},
		{
			name: "two meet threshold",
			result: &ai.ColumnMappingResult{
				CanonicalFields: []ai.CanonicalFieldMapping{
					{CanonicalName: "id", ColumnIndex: 0, Confidence: 0.8},
					{CanonicalName: "title", ColumnIndex: 1, Confidence: 0.9},
					{CanonicalName: "status", ColumnIndex: 2, Confidence: 0.6},
				},
			},
			threshold: 0.75,
			want:      2,
		},
		{
			name: "invalid column index excluded",
			result: &ai.ColumnMappingResult{
				CanonicalFields: []ai.CanonicalFieldMapping{
					{CanonicalName: "id", ColumnIndex: 0, Confidence: 0.9},
					{CanonicalName: "title", ColumnIndex: 10, Confidence: 0.9}, // out of range
				},
			},
			threshold: 0.75,
			want:      1,
		},
		{
			name: "unknown canonical excluded",
			result: &ai.ColumnMappingResult{
				CanonicalFields: []ai.CanonicalFieldMapping{
					{CanonicalName: "id", ColumnIndex: 0, Confidence: 0.9},
					{CanonicalName: "unknown_field", ColumnIndex: 1, Confidence: 0.9},
				},
			},
			threshold: 0.75,
			want:      1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countHighConfidenceAIMappings(tt.result, headers, tt.threshold)
			if got != tt.want {
				t.Errorf("countHighConfidenceAIMappings() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMergeAIMappingWithFallback(t *testing.T) {
	headers := []string{"ID", "Title", "Status", "Notes"}
	fallbackMap := ColumnMap{
		FieldID:    0,
		FieldTitle: 1,
		FieldStatus: 2,
	}
	tests := []struct {
		name        string
		result      *ai.ColumnMappingResult
		threshold   float64
		wantUsed    bool
		wantIDCol   int // expected column for FieldID after merge
		wantTitleCol int
		wantLen     int // expected len(merged)
	}{
		{
			name:        "nil result returns fallback",
			result:     nil,
			threshold:  0.75,
			wantUsed:   false,
			wantIDCol:  0,
			wantTitleCol: 1,
			wantLen:    3,
		},
		{
			name: "AI overrides ID and Title, adds none for Status",
			result: &ai.ColumnMappingResult{
				CanonicalFields: []ai.CanonicalFieldMapping{
					{CanonicalName: "id", ColumnIndex: 0, Confidence: 0.9},
					{CanonicalName: "title", ColumnIndex: 1, Confidence: 0.85},
					{CanonicalName: "status", ColumnIndex: 2, Confidence: 0.5}, // below threshold
				},
			},
			threshold:   0.75,
			wantUsed:    true,
			wantIDCol:   0,
			wantTitleCol: 1,
			wantLen:     3,
		},
		{
			name: "AI overrides Title to different column",
			result: &ai.ColumnMappingResult{
				CanonicalFields: []ai.CanonicalFieldMapping{
					{CanonicalName: "title", ColumnIndex: 3, Confidence: 0.9}, // Notes column
				},
			},
			threshold:   0.75,
			wantUsed:    true,
			wantIDCol:   0, // from fallback
			wantTitleCol: 3, // from AI
			wantLen:     3,
		},
		{
			name: "no AI mappings meet threshold",
			result: &ai.ColumnMappingResult{
				CanonicalFields: []ai.CanonicalFieldMapping{
					{CanonicalName: "id", ColumnIndex: 0, Confidence: 0.5},
					{CanonicalName: "title", ColumnIndex: 1, Confidence: 0.6},
				},
			},
			threshold:   0.75,
			wantUsed:    false,
			wantIDCol:   0,
			wantTitleCol: 1,
			wantLen:     3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged, unmapped, used := mergeAIMappingWithFallback(tt.result, fallbackMap, headers, tt.threshold)
			if used != tt.wantUsed {
				t.Errorf("mergeAIMappingWithFallback() used = %v, want %v", used, tt.wantUsed)
			}
			if len(merged) != tt.wantLen {
				t.Errorf("mergeAIMappingWithFallback() len(merged) = %d, want %d", len(merged), tt.wantLen)
			}
			if idx, ok := merged[FieldID]; ok && idx != tt.wantIDCol {
				t.Errorf("FieldID column = %d, want %d", idx, tt.wantIDCol)
			}
			if idx, ok := merged[FieldTitle]; ok && idx != tt.wantTitleCol {
				t.Errorf("FieldTitle column = %d, want %d", idx, tt.wantTitleCol)
			}
			// Unmapped should include columns not in merged
			_ = unmapped // avoid unused
		})
	}
}
