package ai

import (
	"testing"
)

func TestValidateColumnMappingBusinessRules_RejectsDuplicateCanonicalNames(t *testing.T) {
	result := &ColumnMappingResult{
		CanonicalFields: []CanonicalFieldMapping{
			{CanonicalName: "id", ColumnIndex: 0, Confidence: 0.9},
			{CanonicalName: "id", ColumnIndex: 1, Confidence: 0.8},
		},
	}
	err := ValidateColumnMappingBusinessRules(result)
	if err == nil {
		t.Error("expected error for duplicate canonical names")
	}
}

func TestValidateColumnMappingBusinessRules_RejectsNegativeConfidence(t *testing.T) {
	result := &ColumnMappingResult{
		CanonicalFields: []CanonicalFieldMapping{
			{CanonicalName: "id", ColumnIndex: 0, Confidence: -0.1},
		},
	}
	err := ValidateColumnMappingBusinessRules(result)
	if err == nil {
		t.Error("expected error for negative confidence")
	}
}

func TestValidateColumnMappingBusinessRules_RejectsNegativeColumnIndex(t *testing.T) {
	result := &ColumnMappingResult{
		CanonicalFields: []CanonicalFieldMapping{
			{CanonicalName: "id", ColumnIndex: -1, Confidence: 0.9},
		},
	}
	err := ValidateColumnMappingBusinessRules(result)
	if err == nil {
		t.Error("expected error for negative column index")
	}
}

func TestValidateColumnMappingBusinessRules_RejectsUnknownCanonicalName(t *testing.T) {
	result := &ColumnMappingResult{
		CanonicalFields: []CanonicalFieldMapping{
			{CanonicalName: "not_a_real_field", ColumnIndex: 0, Confidence: 0.9},
		},
	}
	err := ValidateColumnMappingBusinessRules(result)
	if err == nil {
		t.Error("expected error for unknown canonical name")
	}
}

func TestValidateColumnMappingBusinessRules_AcceptsValidResult(t *testing.T) {
	result := &ColumnMappingResult{
		CanonicalFields: []CanonicalFieldMapping{
			{CanonicalName: "id", ColumnIndex: 0, Confidence: 0.9},
			{CanonicalName: "title", ColumnIndex: 1, Confidence: 0.85},
		},
	}
	err := ValidateColumnMappingBusinessRules(result)
	if err != nil {
		t.Errorf("expected no error for valid result, got: %v", err)
	}
}

func TestValidateColumnMappingBusinessRules_NilResult(t *testing.T) {
	err := ValidateColumnMappingBusinessRules(nil)
	if err == nil {
		t.Error("expected error for nil result")
	}
}
