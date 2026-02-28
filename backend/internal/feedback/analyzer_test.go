package feedback

import (
	"encoding/json"
	"fmt"
	"testing"
)

// newTestAnalyzer creates an Analyzer backed by an in-memory store for tests.
func newTestAnalyzer(t *testing.T) (*Analyzer, *Store) {
	t.Helper()
	s := newTestStore(t)
	return NewAnalyzer(s), s
}

// submitWithFixes is a helper that marshals a slice of ColumnCorrection into the
// column_fixes field of a new feedback entry.
func submitWithFixes(t *testing.T, s *Store, hash string, rating int, fixes []ColumnCorrection) {
	t.Helper()
	b, err := json.Marshal(fixes)
	if err != nil {
		t.Fatalf("marshal fixes: %v", err)
	}
	if err := s.Submit(&Feedback{RequestHash: hash, Rating: rating, ColumnFixes: string(b)}); err != nil {
		t.Fatalf("Submit(%s): %v", hash, err)
	}
}

// ---------------------------------------------------------------------------
// AnalyzePatterns
// ---------------------------------------------------------------------------

// TestAnalyzer_AnalyzePatterns_EmptyDB verifies that an empty store produces no patterns.
func TestAnalyzer_AnalyzePatterns_EmptyDB(t *testing.T) {
	a, _ := newTestAnalyzer(t)

	patterns, err := a.AnalyzePatterns(30)
	if err != nil {
		t.Fatalf("AnalyzePatterns: %v", err)
	}
	if len(patterns) != 0 {
		t.Errorf("expected empty patterns on empty DB, got %d pattern(s): %+v", len(patterns), patterns)
	}
}

// TestAnalyzer_AnalyzePatterns_IdentifiesLowRating verifies that ≥3 negative ratings
// for the same request hash produce a "low_rating_cluster" pattern with severity "high".
func TestAnalyzer_AnalyzePatterns_IdentifiesLowRating(t *testing.T) {
	a, s := newTestAnalyzer(t)

	// Submit 3 negative ratings for the same hash.
	for i := 0; i < 3; i++ {
		if err := s.Submit(&Feedback{RequestHash: "bad-hash", Rating: 1}); err != nil {
			t.Fatalf("Submit: %v", err)
		}
	}

	patterns, err := a.AnalyzePatterns(30)
	if err != nil {
		t.Fatalf("AnalyzePatterns: %v", err)
	}
	if len(patterns) == 0 {
		t.Fatal("expected at least one pattern, got none")
	}

	var found bool
	for _, p := range patterns {
		if p.RequestHash == "bad-hash" && p.Pattern == "low_rating_cluster" {
			found = true
			if p.Frequency < 3 {
				t.Errorf("expected Frequency >= 3, got %d", p.Frequency)
			}
			if p.Severity != "high" {
				t.Errorf("expected Severity %q, got %q", "high", p.Severity)
			}
			if p.Suggestion == "" {
				t.Error("expected non-empty Suggestion")
			}
		}
	}
	if !found {
		t.Errorf("expected low_rating_cluster pattern for hash %q, patterns: %+v", "bad-hash", patterns)
	}
}

// TestAnalyzer_AnalyzePatterns_PositiveFeedbackNoPattern verifies that purely positive
// feedback does not generate a low_rating_cluster pattern.
func TestAnalyzer_AnalyzePatterns_PositiveFeedbackNoPattern(t *testing.T) {
	a, s := newTestAnalyzer(t)

	for i := 0; i < 5; i++ {
		if err := s.Submit(&Feedback{RequestHash: "good-hash", Rating: 5}); err != nil {
			t.Fatalf("Submit: %v", err)
		}
	}

	patterns, err := a.AnalyzePatterns(30)
	if err != nil {
		t.Fatalf("AnalyzePatterns: %v", err)
	}

	for _, p := range patterns {
		if p.Pattern == "low_rating_cluster" {
			t.Errorf("unexpected low_rating_cluster for positive-only feedback: %+v", p)
		}
	}
}

// ---------------------------------------------------------------------------
// GetTopCorrections
// ---------------------------------------------------------------------------

// TestAnalyzer_GetTopCorrections verifies that corrections are aggregated and returned
// sorted by frequency (most common first).
func TestAnalyzer_GetTopCorrections(t *testing.T) {
	a, s := newTestAnalyzer(t)

	// Submit the same correction 3 times across different feedback entries.
	commonFix := ColumnCorrection{SourceHeader: "TC Name", WrongMapping: "notes", CorrectMapping: "title"}
	submitWithFixes(t, s, "h1", 1, []ColumnCorrection{commonFix})
	submitWithFixes(t, s, "h2", 1, []ColumnCorrection{commonFix})
	submitWithFixes(t, s, "h3", 1, []ColumnCorrection{commonFix})

	// Submit a different correction once.
	rareFix := ColumnCorrection{SourceHeader: "Ref", WrongMapping: "id", CorrectMapping: "reference"}
	submitWithFixes(t, s, "h4", 1, []ColumnCorrection{rareFix})

	corrections, err := a.GetTopCorrections(10)
	if err != nil {
		t.Fatalf("GetTopCorrections: %v", err)
	}

	if len(corrections) == 0 {
		t.Fatal("expected corrections, got none")
	}

	// Most frequent correction should come first.
	if corrections[0].SourceHeader != "TC Name" {
		t.Errorf("expected first correction SourceHeader %q, got %q", "TC Name", corrections[0].SourceHeader)
	}
	if corrections[0].Frequency != 3 {
		t.Errorf("expected first correction Frequency 3, got %d", corrections[0].Frequency)
	}
	if len(corrections) < 2 {
		t.Fatalf("expected at least 2 corrections, got %d", len(corrections))
	}
	if corrections[1].SourceHeader != "Ref" {
		t.Errorf("expected second correction SourceHeader %q, got %q", "Ref", corrections[1].SourceHeader)
	}
	if corrections[1].Frequency != 1 {
		t.Errorf("expected second correction Frequency 1, got %d", corrections[1].Frequency)
	}
}

// TestAnalyzer_GetTopCorrections_Limit verifies that the limit parameter is honored.
func TestAnalyzer_GetTopCorrections_Limit(t *testing.T) {
	a, s := newTestAnalyzer(t)

	// Insert 5 distinct corrections (one per entry).
	for i := 0; i < 5; i++ {
		fix := ColumnCorrection{
			SourceHeader:   fmt.Sprintf("Header%d", i),
			WrongMapping:   "wrong",
			CorrectMapping: "correct",
		}
		submitWithFixes(t, s, fmt.Sprintf("h%d", i), 1, []ColumnCorrection{fix})
	}

	corrections, err := a.GetTopCorrections(2)
	if err != nil {
		t.Fatalf("GetTopCorrections: %v", err)
	}
	if len(corrections) > 2 {
		t.Errorf("expected at most 2 corrections (limit=2), got %d", len(corrections))
	}
}

// TestAnalyzer_GetTopCorrections_EmptyDB verifies empty DB returns empty slice, not error.
func TestAnalyzer_GetTopCorrections_EmptyDB(t *testing.T) {
	a, _ := newTestAnalyzer(t)

	corrections, err := a.GetTopCorrections(10)
	if err != nil {
		t.Fatalf("unexpected error on empty DB: %v", err)
	}
	if len(corrections) != 0 {
		t.Errorf("expected empty corrections on empty DB, got %d", len(corrections))
	}
}

// ---------------------------------------------------------------------------
// GenerateExampleFromCorrections
// ---------------------------------------------------------------------------

// TestAnalyzer_GenerateExampleFromCorrections verifies the shape of generated suggestions.
func TestAnalyzer_GenerateExampleFromCorrections(t *testing.T) {
	a, _ := newTestAnalyzer(t)

	corrections := []ColumnCorrection{
		{SourceHeader: "TC Name", WrongMapping: "notes", CorrectMapping: "title", Frequency: 3},
		{SourceHeader: "Ref ID", WrongMapping: "feature", CorrectMapping: "id", Frequency: 2},
	}

	suggestions := a.GenerateExampleFromCorrections(corrections)

	if len(suggestions) == 0 {
		t.Fatal("expected at least one suggestion")
	}

	for i, sug := range suggestions {
		if sug.Operation == "" {
			t.Errorf("suggestion[%d]: missing Operation", i)
		}
		if sug.Source != "user_feedback" {
			t.Errorf("suggestion[%d]: expected Source %q, got %q", i, "user_feedback", sug.Source)
		}
		if len(sug.Headers) == 0 {
			t.Errorf("suggestion[%d]: missing Headers", i)
		}
		if len(sug.Corrections) == 0 {
			t.Errorf("suggestion[%d]: missing Corrections", i)
		}
	}
}

// TestAnalyzer_GenerateExampleFromCorrections_EmptyInput verifies nil input returns nil.
func TestAnalyzer_GenerateExampleFromCorrections_EmptyInput(t *testing.T) {
	a, _ := newTestAnalyzer(t)

	suggestions := a.GenerateExampleFromCorrections(nil)
	if suggestions != nil {
		t.Errorf("expected nil for empty input, got %v", suggestions)
	}
}
