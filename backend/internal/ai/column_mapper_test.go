package ai

import (
	"context"
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// TestColumnMapper_TableFormatBypass
// Table format must NOT call AI; all headers go to ExtraColumns.
// ---------------------------------------------------------------------------

func TestColumnMapper_TableFormatBypass(t *testing.T) {
	mock := NewMockAIService()
	svc := NewColumnMapperService(mock)

	req := MapColumnsRequest{
		Headers:    []string{"ID", "Title", "Description"},
		SampleRows: [][]string{{"1", "Login test", "A basic login test"}},
		Format:     "table",
	}

	result, err := svc.MapColumns(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No AI call should have been made
	if got := mock.CallCountFor("MapColumns"); got != 0 {
		t.Errorf("expected 0 AI calls for table format, got %d", got)
	}

	// Every header lands in ExtraColumns
	if got := len(result.ExtraColumns); got != len(req.Headers) {
		t.Errorf("expected %d extra columns, got %d", len(req.Headers), got)
	}

	// CanonicalFields must be empty
	if got := len(result.CanonicalFields); got != 0 {
		t.Errorf("expected 0 canonical fields, got %d", got)
	}

	// Meta reflects table-bypass behaviour
	if result.Meta.DetectedType != "table" {
		t.Errorf("expected Meta.DetectedType=%q, got %q", "table", result.Meta.DetectedType)
	}
	if result.Meta.TotalColumns != len(req.Headers) {
		t.Errorf("expected Meta.TotalColumns=%d, got %d", len(req.Headers), result.Meta.TotalColumns)
	}
	if result.Meta.MappedColumns != 0 {
		t.Errorf("expected Meta.MappedColumns=0, got %d", result.Meta.MappedColumns)
	}
	if result.Meta.UnmappedColumns != len(req.Headers) {
		t.Errorf("expected Meta.UnmappedColumns=%d, got %d", len(req.Headers), result.Meta.UnmappedColumns)
	}
}

func TestColumnMapper_TableFormat_SemanticRoleAndConfidence(t *testing.T) {
	mock := NewMockAIService()
	svc := NewColumnMapperService(mock)

	req := MapColumnsRequest{
		Headers: []string{"Alpha", "Beta"},
		Format:  "table",
	}

	result, err := svc.MapColumns(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, col := range result.ExtraColumns {
		if col.SemanticRole != "table_column" {
			t.Errorf("col %q: expected SemanticRole=%q, got %q", col.Name, "table_column", col.SemanticRole)
		}
		if col.Confidence != 1.0 {
			t.Errorf("col %q: expected Confidence=1.0, got %f", col.Name, col.Confidence)
		}
	}
}

func TestColumnMapper_TableFormat_PreservesColumnOrder(t *testing.T) {
	mock := NewMockAIService()
	svc := NewColumnMapperService(mock)

	headers := []string{"Alpha", "Beta", "Gamma"}
	req := MapColumnsRequest{Headers: headers, Format: "table"}

	result, err := svc.MapColumns(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, col := range result.ExtraColumns {
		if col.Name != headers[i] {
			t.Errorf("index %d: expected name %q, got %q", i, headers[i], col.Name)
		}
		if col.ColumnIndex != i {
			t.Errorf("index %d: expected ColumnIndex=%d, got %d", i, i, col.ColumnIndex)
		}
	}
}

// ---------------------------------------------------------------------------
// TestColumnMapper_CallsAIForSpecFormat
// Spec format must delegate to the AI service exactly once.
// ---------------------------------------------------------------------------

func TestColumnMapper_CallsAIForSpecFormat(t *testing.T) {
	mock := NewMockAIServiceWithDefaults()
	svc := NewColumnMapperService(mock)

	req := MapColumnsRequest{
		Headers:    []string{"ID", "Title"},
		SampleRows: [][]string{{"TC-001", "Login Test"}},
		Format:     "spec",
		FileType:   "xlsx",
	}

	result, err := svc.MapColumns(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// AI must have been called exactly once
	if got := mock.CallCountFor("MapColumns"); got != 1 {
		t.Errorf("expected 1 AI call for spec format, got %d", got)
	}

	// Mock default returns canonical fields
	if len(result.CanonicalFields) == 0 {
		t.Error("expected non-empty CanonicalFields from spec mapping")
	}
}

func TestColumnMapper_EmptyFormatCallsAI(t *testing.T) {
	// An empty format string is NOT "table", so it falls through to AI.
	mock := NewMockAIServiceWithDefaults()
	svc := NewColumnMapperService(mock)

	req := MapColumnsRequest{Headers: []string{"ID"}, Format: ""}

	_, err := svc.MapColumns(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := mock.CallCountFor("MapColumns"); got != 1 {
		t.Errorf("expected 1 AI call for empty format, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// TestColumnMapper_HandlesAIError
// Errors from the AI service must propagate unchanged.
// ---------------------------------------------------------------------------

func TestColumnMapper_HandlesAIError(t *testing.T) {
	sentinelErr := errors.New("ai_unavailable: API timeout")

	mock := NewMockAIService()
	mock.MapColumnsFunc = func(_ context.Context, _ MapColumnsRequest) (*ColumnMappingResult, error) {
		return nil, sentinelErr
	}
	svc := NewColumnMapperService(mock)

	req := MapColumnsRequest{Headers: []string{"ID", "Title"}, Format: "spec"}

	result, err := svc.MapColumns(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
	if !errors.Is(err, sentinelErr) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestColumnMapper_NilAIService_ReturnsErrAIUnavailable(t *testing.T) {
	svc := NewColumnMapperService(nil)

	req := MapColumnsRequest{Headers: []string{"ID"}, Format: "spec"}

	_, err := svc.MapColumns(context.Background(), req)
	if err == nil {
		t.Fatal("expected ErrAIUnavailable for nil service, got nil")
	}
	if !errors.Is(err, ErrAIUnavailable) {
		t.Errorf("expected ErrAIUnavailable, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// GetCanonicalName / GetExtraColumnName helpers
// ---------------------------------------------------------------------------

func TestGetCanonicalName_ReturnsName(t *testing.T) {
	result := &ColumnMappingResult{
		CanonicalFields: []CanonicalFieldMapping{
			{CanonicalName: "title", ColumnIndex: 2},
		},
	}
	if got := GetCanonicalName(result, 2); got != "title" {
		t.Errorf("expected %q, got %q", "title", got)
	}
}

func TestGetCanonicalName_ReturnsEmptyForMissing(t *testing.T) {
	result := &ColumnMappingResult{
		CanonicalFields: []CanonicalFieldMapping{
			{CanonicalName: "title", ColumnIndex: 0},
		},
	}
	if got := GetCanonicalName(result, 99); got != "" {
		t.Errorf("expected empty string for missing index, got %q", got)
	}
}

func TestGetExtraColumnName_ReturnsName(t *testing.T) {
	result := &ColumnMappingResult{
		ExtraColumns: []ExtraColumnMapping{
			{Name: "risk", ColumnIndex: 3},
		},
	}
	if got := GetExtraColumnName(result, 3); got != "risk" {
		t.Errorf("expected %q, got %q", "risk", got)
	}
}

func TestGetExtraColumnName_ReturnsEmptyForMissing(t *testing.T) {
	result := &ColumnMappingResult{
		ExtraColumns: []ExtraColumnMapping{
			{Name: "risk", ColumnIndex: 0},
		},
	}
	if got := GetExtraColumnName(result, 9); got != "" {
		t.Errorf("expected empty string for missing index, got %q", got)
	}
}
