package feedback

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/yourorg/md-spec-tool/internal/ai"
)

// setupLearner creates a Learner, its backing Store, and ExampleStore for tests.
func setupLearner(t *testing.T) (*Learner, *Store, *ai.ExampleStore) {
	t.Helper()
	s := newTestStore(t)
	a := NewAnalyzer(s)
	es := ai.NewExampleStore()
	return NewLearner(a, es), s, es
}

// ---------------------------------------------------------------------------
// LearnFromFeedback — no data
// ---------------------------------------------------------------------------

// TestLearner_LearnFromFeedback_NoData verifies that an empty store produces a
// zero-value LearningReport without error.
func TestLearner_LearnFromFeedback_NoData(t *testing.T) {
	l, _, _ := setupLearner(t)

	report, err := l.LearnFromFeedback(30)
	if err != nil {
		t.Fatalf("LearnFromFeedback: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.PatternsFound != 0 {
		t.Errorf("expected PatternsFound 0, got %d", report.PatternsFound)
	}
	if report.CorrectionsFound != 0 {
		t.Errorf("expected CorrectionsFound 0, got %d", report.CorrectionsFound)
	}
	if report.ExamplesGenerated != 0 {
		t.Errorf("expected ExamplesGenerated 0, got %d", report.ExamplesGenerated)
	}
}

// ---------------------------------------------------------------------------
// LearnFromFeedback — with corrections
// ---------------------------------------------------------------------------

// TestLearner_LearnFromFeedback_WithCorrections verifies that column corrections in
// feedback lead to new examples being registered in the ExampleStore.
func TestLearner_LearnFromFeedback_WithCorrections(t *testing.T) {
	l, s, es := setupLearner(t)

	// Submit 3 feedback entries, each with the same column correction.
	fix := ColumnCorrection{SourceHeader: "TC Name", WrongMapping: "notes", CorrectMapping: "title"}
	for i := 0; i < 3; i++ {
		b, _ := json.Marshal([]ColumnCorrection{fix})
		if err := s.Submit(&Feedback{
			RequestHash: fmt.Sprintf("hash-%d", i),
			Rating:      1,
			ColumnFixes: string(b),
		}); err != nil {
			t.Fatalf("Submit: %v", err)
		}
	}

	report, err := l.LearnFromFeedback(30)
	if err != nil {
		t.Fatalf("LearnFromFeedback: %v", err)
	}

	if report.CorrectionsFound == 0 {
		t.Error("expected CorrectionsFound > 0")
	}
	if report.ExamplesGenerated == 0 {
		t.Error("expected ExamplesGenerated > 0")
	}

	// Verify examples were actually registered in the ExampleStore.
	examples := es.GetExamples("column_mapping", ai.ExampleFilter{})
	if len(examples) == 0 {
		t.Error("expected at least one example to be registered in the ExampleStore")
	}
}

// ---------------------------------------------------------------------------
// LearningReport
// ---------------------------------------------------------------------------

// TestLearner_LearningReport verifies counts and improvement messages.
func TestLearner_LearningReport(t *testing.T) {
	l, s, _ := setupLearner(t)

	// Submit 3 negative ratings for the same hash to trigger low_rating_cluster.
	for i := 0; i < 3; i++ {
		if err := s.Submit(&Feedback{RequestHash: "bad", Rating: 1}); err != nil {
			t.Fatalf("Submit: %v", err)
		}
	}

	report, err := l.LearnFromFeedback(30)
	if err != nil {
		t.Fatalf("LearnFromFeedback: %v", err)
	}

	if report.PatternsFound == 0 {
		t.Error("expected PatternsFound > 0")
	}
	if len(report.Improvements) == 0 {
		t.Error("expected Improvements to have at least one entry")
	}
}

// TestLearner_LearningReport_ExamplesRegisteredCorrectly verifies that registered
// examples contain the corrected mapping (not the wrong one).
func TestLearner_LearningReport_ExamplesRegisteredCorrectly(t *testing.T) {
	l, s, es := setupLearner(t)

	fix := ColumnCorrection{SourceHeader: "My Header", WrongMapping: "bad_field", CorrectMapping: "good_field"}
	b, _ := json.Marshal([]ColumnCorrection{fix})
	_ = s.Submit(&Feedback{RequestHash: "h1", Rating: 1, ColumnFixes: string(b)})
	_ = s.Submit(&Feedback{RequestHash: "h2", Rating: 1, ColumnFixes: string(b)})
	_ = s.Submit(&Feedback{RequestHash: "h3", Rating: 1, ColumnFixes: string(b)})

	if _, err := l.LearnFromFeedback(30); err != nil {
		t.Fatalf("LearnFromFeedback: %v", err)
	}

	examples := es.GetExamples("column_mapping", ai.ExampleFilter{SchemaType: "user_correction"})
	if len(examples) == 0 {
		t.Fatal("expected user_correction examples to be registered")
	}

	// The mapping should use the corrected canonical name.
	found := false
	for _, ex := range examples {
		for _, m := range ex.Mappings {
			if m.SourceHeader == "My Header" && m.CanonicalName == "good_field" {
				found = true
			}
			if m.SourceHeader == "My Header" && m.CanonicalName == "bad_field" {
				t.Error("registered example contains the wrong mapping 'bad_field'")
			}
		}
	}
	if !found {
		t.Error("expected registered example with SourceHeader='My Header' → CanonicalName='good_field'")
	}
}
