package feedback

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// FeedbackPattern represents a discovered pattern from feedback data.
type FeedbackPattern struct {
	Pattern     string // "low_rating_cluster", "widespread_negative", "column_correction_needed"
	Frequency   int    // How many times this pattern occurred
	RequestHash string // Most relevant request hash (empty for global patterns)
	Suggestion  string // Human-readable action to take
	Severity    string // "high", "medium", "low"
}

// ColumnCorrection represents a user's correction to a column mapping.
type ColumnCorrection struct {
	SourceHeader   string `json:"source_header"`
	WrongMapping   string `json:"wrong_mapping"`   // What the AI produced
	CorrectMapping string `json:"correct_mapping"` // What the user corrected it to
	Frequency      int    `json:"frequency"`       // Computed at query time; not stored per-row
}

// Analyzer processes feedback to identify improvement opportunities.
type Analyzer struct {
	store *Store
}

// NewAnalyzer creates an Analyzer backed by the given feedback store.
func NewAnalyzer(store *Store) *Analyzer {
	return &Analyzer{store: store}
}

// AnalyzePatterns examines feedback from the last [days] days and returns
// actionable patterns. A non-positive days value defaults to 30.
//
// Detected patterns:
//   - "low_rating_cluster"        — ≥3 negative ratings for the same request hash
//   - "widespread_negative"       — overall negative rate >50% with ≥5 entries
//   - "column_correction_needed"  — ≥3 entries contain column_fixes JSON
func (a *Analyzer) AnalyzePatterns(days int) ([]FeedbackPattern, error) {
	if days <= 0 {
		days = 30
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")

	rows, err := a.store.db.Query(`
		SELECT request_hash,
		       COUNT(*)                                                          AS total,
		       SUM(CASE WHEN rating = 1 THEN 1 ELSE 0 END)                      AS negative,
		       SUM(CASE WHEN column_fixes != '' THEN 1 ELSE 0 END)              AS with_fixes
		FROM   feedback
		WHERE  created_at >= ?
		GROUP  BY request_hash
	`, cutoff)
	if err != nil {
		return nil, fmt.Errorf("feedback: analyze patterns: %w", err)
	}
	defer rows.Close()

	type hashStat struct {
		hash      string
		total     int
		negative  int
		withFixes int
	}

	var allStats []hashStat
	var totalEntries, totalNegative, totalWithFixes int

	for rows.Next() {
		var st hashStat
		if err := rows.Scan(&st.hash, &st.total, &st.negative, &st.withFixes); err != nil {
			return nil, fmt.Errorf("feedback: scan pattern row: %w", err)
		}
		allStats = append(allStats, st)
		totalEntries += st.total
		totalNegative += st.negative
		totalWithFixes += st.withFixes
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("feedback: pattern rows: %w", err)
	}

	var patterns []FeedbackPattern

	// Pattern 1: Low rating clusters (≥3 negatives on the same request hash).
	const lowRatingThreshold = 3
	for _, st := range allStats {
		if st.negative >= lowRatingThreshold {
			patterns = append(patterns, FeedbackPattern{
				Pattern:     "low_rating_cluster",
				Frequency:   st.negative,
				RequestHash: st.hash,
				Suggestion: fmt.Sprintf(
					"Request %q has %d negative ratings — review the AI output for this request type",
					st.hash, st.negative,
				),
				Severity: "high",
			})
		}
	}

	// Pattern 2: Widespread negative feedback (>50% negative, ≥5 total entries).
	if totalEntries >= 5 && float64(totalNegative)/float64(totalEntries) > 0.5 {
		patterns = append(patterns, FeedbackPattern{
			Pattern:     "widespread_negative",
			Frequency:   totalNegative,
			RequestHash: "",
			Suggestion: fmt.Sprintf(
				"Overall negative rate is %.0f%% (%d/%d entries) — consider revising prompts",
				float64(totalNegative)/float64(totalEntries)*100, totalNegative, totalEntries,
			),
			Severity: "high",
		})
	}

	// Pattern 3: Frequent column corrections (≥3 entries include column_fixes JSON).
	const correctionThreshold = 3
	if totalWithFixes >= correctionThreshold {
		patterns = append(patterns, FeedbackPattern{
			Pattern:     "column_correction_needed",
			Frequency:   totalWithFixes,
			RequestHash: "",
			Suggestion: fmt.Sprintf(
				"%d feedback entries include column corrections — run LearnFromFeedback to improve column mapping examples",
				totalWithFixes,
			),
			Severity: "medium",
		})
	}

	return patterns, nil
}

// GetTopCorrections returns the [limit] most frequently seen column mapping corrections
// across all feedback entries that contain column_fixes JSON. Corrections with the
// same (source_header, wrong_mapping, correct_mapping) tuple are aggregated.
// A non-positive limit defaults to 10.
func (a *Analyzer) GetTopCorrections(limit int) ([]ColumnCorrection, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := a.store.db.Query(
		`SELECT column_fixes FROM feedback WHERE column_fixes != ''`,
	)
	if err != nil {
		return nil, fmt.Errorf("feedback: get top corrections: %w", err)
	}
	defer rows.Close()

	// Aggregate by (source_header, wrong_mapping, correct_mapping).
	type corrKey struct {
		sourceHeader   string
		wrongMapping   string
		correctMapping string
	}
	counts := make(map[corrKey]int)

	for rows.Next() {
		var fixesJSON string
		if err := rows.Scan(&fixesJSON); err != nil {
			return nil, fmt.Errorf("feedback: scan column_fixes: %w", err)
		}

		var fixes []ColumnCorrection
		if err := json.Unmarshal([]byte(fixesJSON), &fixes); err != nil {
			// Skip malformed JSON — don't let one bad row abort the whole query.
			continue
		}

		for _, fix := range fixes {
			key := corrKey{
				sourceHeader:   fix.SourceHeader,
				wrongMapping:   fix.WrongMapping,
				correctMapping: fix.CorrectMapping,
			}
			counts[key]++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("feedback: correction rows: %w", err)
	}

	// Convert to slice and sort descending by frequency.
	result := make([]ColumnCorrection, 0, len(counts))
	for key, freq := range counts {
		result = append(result, ColumnCorrection{
			SourceHeader:   key.sourceHeader,
			WrongMapping:   key.wrongMapping,
			CorrectMapping: key.correctMapping,
			Frequency:      freq,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Frequency != result[j].Frequency {
			return result[i].Frequency > result[j].Frequency
		}
		// Stable secondary sort to make tests deterministic.
		return result[i].SourceHeader < result[j].SourceHeader
	})

	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

// GenerateExampleFromCorrections converts a list of ColumnCorrections into
// ExampleSuggestions that can be registered in an ai.ExampleStore to improve
// future column-mapping prompts.
//
// Each correction produces one ExampleSuggestion with:
//   - Operation  = "column_mapping"
//   - SchemaType = "user_correction"
//   - Headers    = [correction.SourceHeader]
//   - Source     = "user_feedback"
func (a *Analyzer) GenerateExampleFromCorrections(corrections []ColumnCorrection) []ExampleSuggestion {
	if len(corrections) == 0 {
		return nil
	}

	suggestions := make([]ExampleSuggestion, 0, len(corrections))
	for _, c := range corrections {
		c := c // capture loop variable
		suggestions = append(suggestions, ExampleSuggestion{
			Operation:   "column_mapping",
			SchemaType:  "user_correction",
			Language:    "",
			Headers:     []string{c.SourceHeader},
			Corrections: []ColumnCorrection{c},
			Source:      "user_feedback",
		})
	}
	return suggestions
}
