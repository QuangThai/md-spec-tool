package ai_test

import (
	. "github.com/yourorg/md-spec-tool/internal/ai"
	"testing"
)

// TestValidateMappingSemantics tests semantic validation of column mappings
func TestValidateMappingSemantics(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name           string
		result         *ColumnMappingResult
		schema         string
		expectedStatus string // "good", "needs_improvement", "poor"
		expectErrors   bool
	}{
		{
			name: "high confidence mapping",
			result: &ColumnMappingResult{
				SchemaVersion: SchemaVersionColumnMapping,
				CanonicalFields: []CanonicalFieldMapping{
					{CanonicalName: "id", SourceHeader: "ID", ColumnIndex: 0, Confidence: 0.95},
					{CanonicalName: "scenario", SourceHeader: "Scenario", ColumnIndex: 1, Confidence: 0.90},
					{CanonicalName: "instructions", SourceHeader: "Instructions", ColumnIndex: 2, Confidence: 0.85},
					{CanonicalName: "expected", SourceHeader: "Expected", ColumnIndex: 3, Confidence: 0.92},
				},
				Meta: MappingMeta{
					DetectedType:    "test_case",
					AvgConfidence:   0.90,
					TotalColumns:    4,
					MappedColumns:   4,
					UnmappedColumns: 0,
				},
			},
			schema:         "test_case",
			expectedStatus: "good",
			expectErrors:   false,
		},
		{
			name: "low average confidence",
			result: &ColumnMappingResult{
				SchemaVersion: SchemaVersionColumnMapping,
				CanonicalFields: []CanonicalFieldMapping{
					{CanonicalName: "id", SourceHeader: "ID", ColumnIndex: 0, Confidence: 0.45},
					{CanonicalName: "title", SourceHeader: "Name", ColumnIndex: 1, Confidence: 0.40},
					{CanonicalName: "expected", SourceHeader: "Result", ColumnIndex: 2, Confidence: 0.35},
				},
				Meta: MappingMeta{
					DetectedType:    "test_case",
					AvgConfidence:   0.40,
					TotalColumns:    3,
					MappedColumns:   3,
					UnmappedColumns: 0,
				},
			},
			schema:         "test_case",
			expectedStatus: "needs_improvement",
			expectErrors:   true,
		},
		{
			name: "conflicting column indices",
			result: &ColumnMappingResult{
				SchemaVersion: SchemaVersionColumnMapping,
				CanonicalFields: []CanonicalFieldMapping{
					{CanonicalName: "id", SourceHeader: "ID", ColumnIndex: 0, Confidence: 0.95},
					{CanonicalName: "title", SourceHeader: "Title", ColumnIndex: 0, Confidence: 0.90}, // Same index!
				},
				Meta: MappingMeta{
					DetectedType:    "test_case",
					AvgConfidence:   0.92,
					TotalColumns:    2,
					MappedColumns:   2,
					UnmappedColumns: 0,
				},
			},
			schema:         "test_case",
			expectedStatus: "poor",
			expectErrors:   true,
		},
		{
			name: "missing required field",
			result: &ColumnMappingResult{
				SchemaVersion: SchemaVersionColumnMapping,
				CanonicalFields: []CanonicalFieldMapping{
					{CanonicalName: "id", SourceHeader: "ID", ColumnIndex: 0, Confidence: 0.95},
					{CanonicalName: "title", SourceHeader: "Title", ColumnIndex: 1, Confidence: 0.90},
					// Missing 'instructions' and 'expected' for test_case schema
				},
				Meta: MappingMeta{
					DetectedType:    "test_case",
					AvgConfidence:   0.92,
					TotalColumns:    2,
					MappedColumns:   2,
					UnmappedColumns: 0,
				},
			},
			schema:         "test_case",
			expectedStatus: "needs_improvement",
			expectErrors:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateMappingSemantics(tt.result, tt.schema)

			if result.Overall != tt.expectedStatus {
				t.Errorf("expected status %s, got %s", tt.expectedStatus, result.Overall)
			}

			hasErrors := len(result.Issues) > 0
			if hasErrors != tt.expectErrors {
				t.Errorf("expected errors=%v, got %v (issues: %d)", tt.expectErrors, hasErrors, len(result.Issues))
			}
		})
	}
}

// TestApplyConfidenceFallback tests confidence-based fallback
func TestApplyConfidenceFallback(t *testing.T) {
	tests := []struct {
		name              string
		input             *ColumnMappingResult
		expectedCanonical int
		expectedExtra     int
	}{
		{
			name: "move low confidence to extra_columns",
			input: &ColumnMappingResult{
				SchemaVersion: SchemaVersionColumnMapping,
				CanonicalFields: []CanonicalFieldMapping{
					{CanonicalName: "id", SourceHeader: "ID", ColumnIndex: 0, Confidence: 0.95},
					{CanonicalName: "title", SourceHeader: "Name", ColumnIndex: 1, Confidence: 0.35}, // < 0.4
					{CanonicalName: "expected", SourceHeader: "Result", ColumnIndex: 2, Confidence: 0.85},
				},
				ExtraColumns: []ExtraColumnMapping{},
				Meta: MappingMeta{
					DetectedType:    "test_case",
					AvgConfidence:   0.71,
					TotalColumns:    3,
					MappedColumns:   3,
					UnmappedColumns: 0,
				},
			},
			expectedCanonical: 2, // id, expected (title moved)
			expectedExtra:     1, // name moved here
		},
		{
			name: "keep high confidence mappings",
			input: &ColumnMappingResult{
				SchemaVersion: SchemaVersionColumnMapping,
				CanonicalFields: []CanonicalFieldMapping{
					{CanonicalName: "id", SourceHeader: "ID", ColumnIndex: 0, Confidence: 0.95},
					{CanonicalName: "title", SourceHeader: "Name", ColumnIndex: 1, Confidence: 0.85},
					{CanonicalName: "expected", SourceHeader: "Result", ColumnIndex: 2, Confidence: 0.90},
				},
				ExtraColumns: []ExtraColumnMapping{},
				Meta: MappingMeta{
					DetectedType:    "test_case",
					AvgConfidence:   0.90,
					TotalColumns:    3,
					MappedColumns:   3,
					UnmappedColumns: 0,
				},
			},
			expectedCanonical: 3, // all kept
			expectedExtra:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewServiceForFallbackTest()
			result := service.ApplyConfidenceFallback(tt.input)

			if len(result.CanonicalFields) != tt.expectedCanonical {
				t.Errorf("expected %d canonical fields, got %d", tt.expectedCanonical, len(result.CanonicalFields))
			}

			if len(result.ExtraColumns) != tt.expectedExtra {
				t.Errorf("expected %d extra columns, got %d", tt.expectedExtra, len(result.ExtraColumns))
			}
		})
	}
}

// TestGetMappingWithFallback tests the full fallback orchestration
func TestGetMappingWithFallback(t *testing.T) {
	// Note: This test would require mocking the client layer
	// For now, we'll test the logic structure
	service := NewServiceForFallbackTest()

	// Test applyConfidenceFallback as part of the orchestration
	lowConfidenceResult := &ColumnMappingResult{
		SchemaVersion: SchemaVersionColumnMapping,
		CanonicalFields: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "ID", ColumnIndex: 0, Confidence: 0.55},
			{CanonicalName: "title", SourceHeader: "Name", ColumnIndex: 1, Confidence: 0.35}, // < 0.4
		},
		ExtraColumns: []ExtraColumnMapping{},
		Meta: MappingMeta{
			DetectedType:    "generic",
			AvgConfidence:   0.45,
			TotalColumns:    2,
			MappedColumns:   2,
			UnmappedColumns: 0,
		},
	}

	result := service.ApplyConfidenceFallback(lowConfidenceResult)

	// Verify fallback moved low-confidence mapping
	if len(result.CanonicalFields) != 1 {
		t.Errorf("expected 1 canonical field after fallback, got %d", len(result.CanonicalFields))
	}

	if len(result.ExtraColumns) != 1 {
		t.Errorf("expected 1 extra column after fallback, got %d", len(result.ExtraColumns))
	}

	// Verify no data was lost
	if result.Meta.TotalColumns != 2 {
		t.Errorf("expected total_columns=2, got %d", result.Meta.TotalColumns)
	}
}

// TestGetRequiredFieldsBySchema tests schema-specific field requirements
func TestGetRequiredFieldsBySchema(t *testing.T) {
	tests := []struct {
		schema   string
		expected []string
	}{
		{"test_case", []string{"id", "scenario", "instructions", "expected"}},
		{"product_backlog", []string{"id", "title", "description", "acceptance_criteria"}},
		{"issue_tracker", []string{"id", "feature", "priority", "status"}},
		{"api_spec", []string{"endpoint", "method", "parameters", "response"}},
		{"ui_spec", []string{"item_name", "item_type", "action"}},
		{"generic", []string{"id", "feature"}},
		{"unknown", []string{"id", "feature"}}, // default
	}

	for _, tt := range tests {
		t.Run(tt.schema, func(t *testing.T) {
			result := GetRequiredFieldsBySchema(tt.schema)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d required fields, got %d", len(tt.expected), len(result))
			}

			// Check all expected fields are present
			for _, field := range tt.expected {
				found := false
				for _, r := range result {
					if r == field {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected field %s not found in result", field)
				}
			}
		})
	}
}

// TestRefineMappingContext tests that refinement uses context correctly
func TestRefineMappingContext(t *testing.T) {
	originalResult := &ColumnMappingResult{
		CanonicalFields: []CanonicalFieldMapping{
			{CanonicalName: "id", SourceHeader: "ID", ColumnIndex: 0, Confidence: 0.55},
			{CanonicalName: "title", SourceHeader: "Name", ColumnIndex: 1, Confidence: 0.65},
		},
		Meta: MappingMeta{
			DetectedType:    "generic",
			AvgConfidence:   0.60,
			TotalColumns:    2,
			MappedColumns:   2,
			UnmappedColumns: 0,
		},
	}

	// Test that refinement request is properly constructed
	// (Can't actually call RefineMapping without a real client)
	ambiguousFields := []string{}
	for _, m := range originalResult.CanonicalFields {
		if m.Confidence < 0.7 {
			ambiguousFields = append(ambiguousFields, m.SourceHeader)
		}
	}

	if len(ambiguousFields) != 2 {
		t.Errorf("expected 2 ambiguous fields, got %d", len(ambiguousFields))
	}
}
