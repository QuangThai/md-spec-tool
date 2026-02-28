package ai

import (
	"testing"
)

// ---------------------------------------------------------------------------
// ConfidenceThresholds — unit tests for every public helper
// ---------------------------------------------------------------------------

func TestDefaultThresholds_Values(t *testing.T) {
	th := DefaultThresholds()
	if th.HighConfidence != 0.80 {
		t.Errorf("HighConfidence: expected 0.80, got %f", th.HighConfidence)
	}
	if th.MediumConfidence != 0.65 {
		t.Errorf("MediumConfidence: expected 0.65, got %f", th.MediumConfidence)
	}
	if th.LowConfidence != 0.55 {
		t.Errorf("LowConfidence: expected 0.55, got %f", th.LowConfidence)
	}
	if th.HeaderConfidenceThreshold != 70 {
		t.Errorf("HeaderConfidenceThreshold: expected 70, got %d", th.HeaderConfidenceThreshold)
	}
	if th.RequiredFieldMappingThreshold != 0.50 {
		t.Errorf("RequiredFieldMappingThreshold: expected 0.50, got %f", th.RequiredFieldMappingThreshold)
	}
}

func TestIsHighConfidence(t *testing.T) {
	th := DefaultThresholds()
	cases := []struct {
		v    float64
		want bool
	}{
		{0.80, true},
		{0.95, true},
		{1.00, true},
		{0.79, false},
		{0.65, false},
		{0.00, false},
	}
	for _, tc := range cases {
		if got := th.IsHighConfidence(tc.v); got != tc.want {
			t.Errorf("IsHighConfidence(%v) = %v, want %v", tc.v, got, tc.want)
		}
	}
}

func TestIsMediumConfidence(t *testing.T) {
	th := DefaultThresholds()
	cases := []struct {
		v    float64
		want bool
	}{
		{0.65, true},
		{0.75, true},
		{0.79, true},
		{0.80, false}, // High boundary
		{0.64, false},
		{0.00, false},
	}
	for _, tc := range cases {
		if got := th.IsMediumConfidence(tc.v); got != tc.want {
			t.Errorf("IsMediumConfidence(%v) = %v, want %v", tc.v, got, tc.want)
		}
	}
}

func TestIsLowConfidence(t *testing.T) {
	th := DefaultThresholds()
	cases := []struct {
		v    float64
		want bool
	}{
		{0.00, true},
		{0.50, true},
		{0.64, true},
		{0.65, false}, // Medium boundary
		{0.80, false},
	}
	for _, tc := range cases {
		if got := th.IsLowConfidence(tc.v); got != tc.want {
			t.Errorf("IsLowConfidence(%v) = %v, want %v", tc.v, got, tc.want)
		}
	}
}

func TestGetConfidenceLevel(t *testing.T) {
	th := DefaultThresholds()
	cases := []struct {
		v    float64
		want ConfidenceLevel
	}{
		{0.90, ConfidenceHigh},
		{0.80, ConfidenceHigh},
		{0.75, ConfidenceMedium},
		{0.65, ConfidenceMedium},
		{0.64, ConfidenceLow},
		{0.00, ConfidenceLow},
	}
	for _, tc := range cases {
		if got := th.GetConfidenceLevel(tc.v); got != tc.want {
			t.Errorf("GetConfidenceLevel(%v) = %v, want %v", tc.v, got, tc.want)
		}
	}
}

func TestShouldReviewMapping_LowAvgConfidence(t *testing.T) {
	th := DefaultThresholds()
	// avgConfidence below MediumConfidence (0.65) → always review
	if !th.ShouldReviewMapping(0.50, 80, 1, 5) {
		t.Error("expected ShouldReviewMapping=true for low avg confidence")
	}
}

func TestShouldReviewMapping_LowHeaderConfidence(t *testing.T) {
	th := DefaultThresholds()
	// avgConfidence OK, but headerConfidence below 70 → review
	if !th.ShouldReviewMapping(0.75, 60, 1, 5) {
		t.Error("expected ShouldReviewMapping=true for low header confidence")
	}
}

func TestShouldReviewMapping_TooManyUnmapped(t *testing.T) {
	th := DefaultThresholds()
	// 3 unmapped out of 5 = 60% → review (> 40% threshold)
	if !th.ShouldReviewMapping(0.75, 80, 3, 5) {
		t.Error("expected ShouldReviewMapping=true for >40% unmapped")
	}
}

func TestShouldReviewMapping_ExactlyAtThreshold(t *testing.T) {
	th := DefaultThresholds()
	// 2 unmapped out of 5 = 40% → NOT above 40%, so no review needed from unmapped signal
	// all other signals also pass
	if th.ShouldReviewMapping(0.75, 80, 2, 5) {
		t.Error("expected ShouldReviewMapping=false when all signals pass")
	}
}

func TestShouldReviewMapping_NoColumns(t *testing.T) {
	th := DefaultThresholds()
	// totalColumns=0: unmapped ratio skip, rely on other signals
	// avgConfidence=0.70 (medium-high), headerConfidence=80 → no review
	if th.ShouldReviewMapping(0.70, 80, 0, 0) {
		t.Error("expected ShouldReviewMapping=false when totalColumns=0 and all other signals pass")
	}
}

// ---------------------------------------------------------------------------
// ApplyConfidenceFallback — test via test_exports.go
// Threshold: confidence < 0.4 moves to extra_columns.
// ---------------------------------------------------------------------------

func TestApplyConfidenceFallback_NoChangesAboveThreshold(t *testing.T) {
	svc := NewServiceForFallbackTest()

	result := &ColumnMappingResult{
		SchemaVersion: SchemaVersionColumnMapping,
		CanonicalFields: []CanonicalFieldMapping{
			{CanonicalName: "title", SourceHeader: "Title", ColumnIndex: 0, Confidence: 0.90},
			{CanonicalName: "description", SourceHeader: "Desc", ColumnIndex: 1, Confidence: 0.80},
			{CanonicalName: "status", SourceHeader: "Status", ColumnIndex: 2, Confidence: 0.70},
		},
		ExtraColumns: nil,
	}

	out := svc.ApplyConfidenceFallback(result)

	// All fields are above 0.4 → should remain in CanonicalFields
	if len(out.CanonicalFields) != 3 {
		t.Errorf("expected 3 canonical fields unchanged, got %d", len(out.CanonicalFields))
	}
	if len(out.ExtraColumns) != 0 {
		t.Errorf("expected 0 extra columns, got %d", len(out.ExtraColumns))
	}
}

func TestApplyConfidenceFallback_MovesLowConfidenceToExtra(t *testing.T) {
	svc := NewServiceForFallbackTest()

	result := &ColumnMappingResult{
		SchemaVersion: SchemaVersionColumnMapping,
		CanonicalFields: []CanonicalFieldMapping{
			{CanonicalName: "title", SourceHeader: "Heading", ColumnIndex: 0, Confidence: 0.20},
			{CanonicalName: "notes", SourceHeader: "Misc", ColumnIndex: 1, Confidence: 0.10},
		},
		ExtraColumns: nil,
	}

	out := svc.ApplyConfidenceFallback(result)

	// Both are below 0.4 → moved to ExtraColumns
	if len(out.CanonicalFields) != 0 {
		t.Errorf("expected 0 canonical fields after fallback, got %d", len(out.CanonicalFields))
	}
	if len(out.ExtraColumns) != 2 {
		t.Errorf("expected 2 extra columns after fallback, got %d", len(out.ExtraColumns))
	}

	// Meta should reflect the move
	if out.Meta.MappedColumns != 0 {
		t.Errorf("expected Meta.MappedColumns=0, got %d", out.Meta.MappedColumns)
	}
	if out.Meta.AvgConfidence != 0 {
		t.Errorf("expected Meta.AvgConfidence=0 when all fields moved, got %f", out.Meta.AvgConfidence)
	}
}

func TestApplyConfidenceFallback_MixedConfidence(t *testing.T) {
	svc := NewServiceForFallbackTest()

	result := &ColumnMappingResult{
		SchemaVersion: SchemaVersionColumnMapping,
		CanonicalFields: []CanonicalFieldMapping{
			{CanonicalName: "title", SourceHeader: "Title", ColumnIndex: 0, Confidence: 0.85},   // keep
			{CanonicalName: "notes", SourceHeader: "Misc", ColumnIndex: 1, Confidence: 0.30},    // move
			{CanonicalName: "status", SourceHeader: "Status", ColumnIndex: 2, Confidence: 0.70}, // keep
			{CanonicalName: "type", SourceHeader: "Kind", ColumnIndex: 3, Confidence: 0.15},     // move
		},
		ExtraColumns: nil,
	}

	out := svc.ApplyConfidenceFallback(result)

	if len(out.CanonicalFields) != 2 {
		t.Errorf("expected 2 canonical fields remaining, got %d", len(out.CanonicalFields))
	}
	if len(out.ExtraColumns) != 2 {
		t.Errorf("expected 2 extra columns from fallback, got %d", len(out.ExtraColumns))
	}

	// Verify the correct fields stayed
	keptNames := map[string]bool{}
	for _, f := range out.CanonicalFields {
		keptNames[f.SourceHeader] = true
	}
	if !keptNames["Title"] {
		t.Error("expected 'Title' to remain in CanonicalFields")
	}
	if !keptNames["Status"] {
		t.Error("expected 'Status' to remain in CanonicalFields")
	}

	// Verify moved fields ended up in ExtraColumns
	movedNames := map[string]bool{}
	for _, e := range out.ExtraColumns {
		movedNames[e.Name] = true
	}
	if !movedNames["Misc"] {
		t.Error("expected 'Misc' to be moved to ExtraColumns")
	}
	if !movedNames["Kind"] {
		t.Error("expected 'Kind' to be moved to ExtraColumns")
	}
}

func TestApplyConfidenceFallback_PreservesExistingExtraColumns(t *testing.T) {
	svc := NewServiceForFallbackTest()

	result := &ColumnMappingResult{
		SchemaVersion: SchemaVersionColumnMapping,
		CanonicalFields: []CanonicalFieldMapping{
			{CanonicalName: "notes", SourceHeader: "Ambiguous", ColumnIndex: 1, Confidence: 0.25},
		},
		ExtraColumns: []ExtraColumnMapping{
			{Name: "AlreadyExtra", SemanticRole: "custom", ColumnIndex: 0, Confidence: 1.0},
		},
	}

	out := svc.ApplyConfidenceFallback(result)

	// Should have: 1 original extra + 1 moved = 2 total extra
	if len(out.ExtraColumns) != 2 {
		t.Errorf("expected 2 extra columns (1 pre-existing + 1 moved), got %d", len(out.ExtraColumns))
	}
}

func TestApplyConfidenceFallback_AvgConfidenceRecalculated(t *testing.T) {
	svc := NewServiceForFallbackTest()

	result := &ColumnMappingResult{
		SchemaVersion: SchemaVersionColumnMapping,
		CanonicalFields: []CanonicalFieldMapping{
			{CanonicalName: "title", SourceHeader: "T1", ColumnIndex: 0, Confidence: 0.80},
			{CanonicalName: "notes", SourceHeader: "T2", ColumnIndex: 1, Confidence: 0.60},
			{CanonicalName: "type", SourceHeader: "T3", ColumnIndex: 2, Confidence: 0.20}, // moved
		},
	}

	out := svc.ApplyConfidenceFallback(result)

	// AvgConfidence should now only reflect the 2 kept fields: (0.80+0.60)/2 = 0.70
	wantAvg := (0.80 + 0.60) / 2.0
	if out.Meta.AvgConfidence != wantAvg {
		t.Errorf("expected AvgConfidence=%f, got %f", wantAvg, out.Meta.AvgConfidence)
	}
}

// ---------------------------------------------------------------------------
// NormalizePromptProfile — tests for all recognized aliases
// ---------------------------------------------------------------------------

func TestNormalizePromptProfile_StaticVariants(t *testing.T) {
	cases := []string{"static", "static_v3", "v3", "STATIC", "Static_V3", "V3"}
	for _, input := range cases {
		if got := NormalizePromptProfile(input); got != PromptProfileStaticV3 {
			t.Errorf("NormalizePromptProfile(%q) = %q, want %q", input, got, PromptProfileStaticV3)
		}
	}
}

func TestNormalizePromptProfile_LegacyVariants(t *testing.T) {
	cases := []string{"legacy", "legacy_v2", "v2", "LEGACY", "Legacy_V2", "V2"}
	for _, input := range cases {
		if got := NormalizePromptProfile(input); got != PromptProfileLegacyV2 {
			t.Errorf("NormalizePromptProfile(%q) = %q, want %q", input, got, PromptProfileLegacyV2)
		}
	}
}

func TestNormalizePromptProfile_DefaultsToStaticV3(t *testing.T) {
	cases := []string{"", "unknown", "v4", "   "}
	for _, input := range cases {
		if got := NormalizePromptProfile(input); got != PromptProfileStaticV3 {
			t.Errorf("NormalizePromptProfile(%q) = %q, want %q (default)", input, got, PromptProfileStaticV3)
		}
	}
}

func TestNormalizePromptProfile_TrimsWhitespace(t *testing.T) {
	if got := NormalizePromptProfile("  v3  "); got != PromptProfileStaticV3 {
		t.Errorf("expected whitespace to be trimmed, got %q", got)
	}
	if got := NormalizePromptProfile("  legacy  "); got != PromptProfileLegacyV2 {
		t.Errorf("expected whitespace to be trimmed, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// ColumnMappingPromptVersion — routes to correct version constant
// ---------------------------------------------------------------------------

func TestColumnMappingPromptVersion_StaticProfile(t *testing.T) {
	for _, profile := range []string{"static_v3", "v3", "", "anything"} {
		got := ColumnMappingPromptVersion(profile)
		if got != PromptVersionColumnMapping {
			t.Errorf("ColumnMappingPromptVersion(%q) = %q, want %q", profile, got, PromptVersionColumnMapping)
		}
	}
}

func TestColumnMappingPromptVersion_LegacyProfile(t *testing.T) {
	for _, profile := range []string{"legacy", "legacy_v2", "v2"} {
		got := ColumnMappingPromptVersion(profile)
		if got != PromptVersionColumnMappingLegacy {
			t.Errorf("ColumnMappingPromptVersion(%q) = %q, want %q", profile, got, PromptVersionColumnMappingLegacy)
		}
	}
}

func TestSuggestionsPromptVersion_StaticProfile(t *testing.T) {
	for _, profile := range []string{"static_v3", "v3", ""} {
		got := SuggestionsPromptVersion(profile)
		if got != PromptVersionSuggestions {
			t.Errorf("SuggestionsPromptVersion(%q) = %q, want %q", profile, got, PromptVersionSuggestions)
		}
	}
}

func TestSuggestionsPromptVersion_LegacyProfile(t *testing.T) {
	for _, profile := range []string{"legacy", "legacy_v2", "v2"} {
		got := SuggestionsPromptVersion(profile)
		if got != PromptVersionSuggestionsLegacy {
			t.Errorf("SuggestionsPromptVersion(%q) = %q, want %q", profile, got, PromptVersionSuggestionsLegacy)
		}
	}
}
