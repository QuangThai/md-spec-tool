package feedback

import (
	"fmt"

	"github.com/yourorg/md-spec-tool/internal/ai"
)

// ExampleSuggestion is a candidate example to add to the ai.ExampleStore.
// It carries enough information to construct a full ai.Example that encodes
// a user-verified correct column mapping.
type ExampleSuggestion struct {
	Operation   string             // e.g., "column_mapping"
	SchemaType  string             // e.g., "user_correction"
	Language    string             // Content language; "" if unknown
	Headers     []string           // Source headers covered by this example
	Corrections []ColumnCorrection // The corrections that informed this example
	Source      string             // "user_feedback"
}

// Learner applies feedback patterns to improve the AI pipeline by registering
// new few-shot examples derived from user corrections.
type Learner struct {
	analyzer     *Analyzer
	exampleStore *ai.ExampleStore
}

// NewLearner creates a Learner that reads from analyzer and writes to exampleStore.
func NewLearner(analyzer *Analyzer, exampleStore *ai.ExampleStore) *Learner {
	return &Learner{
		analyzer:     analyzer,
		exampleStore: exampleStore,
	}
}

// LearningReport summarises what LearnFromFeedback discovered and applied.
type LearningReport struct {
	PatternsFound     int      // Number of feedback patterns identified
	CorrectionsFound  int      // Number of unique column corrections found
	ExamplesGenerated int      // Number of new examples registered in ExampleStore
	Improvements      []string // Human-readable description of each improvement
}

// LearnFromFeedback analyses the last [days] days of feedback, converts the top
// column corrections into new ai.Example entries, registers them in the ExampleStore,
// and returns a summary LearningReport.
//
// A non-positive days value is treated as 30.
func (l *Learner) LearnFromFeedback(days int) (*LearningReport, error) {
	// 1. Identify patterns.
	patterns, err := l.analyzer.AnalyzePatterns(days)
	if err != nil {
		return nil, fmt.Errorf("feedback: learn: analyze patterns: %w", err)
	}

	// 2. Collect the top column corrections.
	const maxCorrections = 20
	corrections, err := l.analyzer.GetTopCorrections(maxCorrections)
	if err != nil {
		return nil, fmt.Errorf("feedback: learn: get corrections: %w", err)
	}

	// 3. Convert corrections into example suggestions.
	suggestions := l.analyzer.GenerateExampleFromCorrections(corrections)

	// 4. Register each suggestion as an ai.Example.
	for _, sug := range suggestions {
		example := ai.Example{
			Operation:  sug.Operation,
			SchemaType: sug.SchemaType,
			Language:   sug.Language,
			Headers:    sug.Headers,
		}
		for i, c := range sug.Corrections {
			example.Mappings = append(example.Mappings, ai.CanonicalFieldMapping{
				CanonicalName: c.CorrectMapping,
				SourceHeader:  c.SourceHeader,
				ColumnIndex:   i,
				Confidence:    1.0,
				Reasoning: fmt.Sprintf(
					"User correction: was %q, corrected to %q (seen %d time(s))",
					c.WrongMapping, c.CorrectMapping, c.Frequency,
				),
			})
		}
		l.exampleStore.Register(example)
	}

	// 5. Build the report.
	report := &LearningReport{
		PatternsFound:     len(patterns),
		CorrectionsFound:  len(corrections),
		ExamplesGenerated: len(suggestions),
		Improvements:      make([]string, 0, len(patterns)+len(suggestions)),
	}

	for _, p := range patterns {
		report.Improvements = append(report.Improvements, p.Suggestion)
	}
	for _, sug := range suggestions {
		if len(sug.Corrections) > 0 {
			c := sug.Corrections[0]
			report.Improvements = append(report.Improvements, fmt.Sprintf(
				"Example added: header %q now maps to %q (was incorrectly %q, corrected %d time(s))",
				c.SourceHeader, c.CorrectMapping, c.WrongMapping, c.Frequency,
			))
		}
	}

	return report, nil
}
