package converter

import (
	"regexp"
	"strings"
)

// InputType represents the detected content type
type InputType string

const (
	InputTypeTable    InputType = "table"
	InputTypeMarkdown InputType = "markdown"
	InputTypeUnknown  InputType = "unknown"
)

// InputAnalysis contains the result of content type detection
type InputAnalysis struct {
	Type       InputType
	Confidence int    // 0-100
	Reason     string // Debug info
}

// DetectInputType analyzes text and determines if it's table data or markdown content
func DetectInputType(text string) InputAnalysis {
	if strings.TrimSpace(text) == "" {
		return InputAnalysis{
			Type:       InputTypeUnknown,
			Confidence: 0,
			Reason:     "Empty input",
		}
	}

	mdScore, mdReasons := calculateMarkdownScore(text)
	tableScore, tableReasons := calculateTableScore(text)

	// Decision logic per plan
	if mdScore > tableScore && mdScore >= 30 {
		return InputAnalysis{
			Type:       InputTypeMarkdown,
			Confidence: min(mdScore, 100),
			Reason:     "Markdown signals: " + strings.Join(mdReasons, ", "),
		}
	}
	if tableScore > mdScore && tableScore >= 40 {
		return InputAnalysis{
			Type:       InputTypeTable,
			Confidence: min(tableScore, 100),
			Reason:     "Table signals: " + strings.Join(tableReasons, ", "),
		}
	}
	if tableScore == mdScore && tableScore >= 40 {
		return InputAnalysis{
			Type:       InputTypeTable,
			Confidence: min(tableScore, 100),
			Reason:     "Table signals: " + strings.Join(tableReasons, ", "),
		}
	}

	// Default to markdown when ambiguous (safer - won't produce garbage output)
	return InputAnalysis{
		Type:       InputTypeMarkdown,
		Confidence: 50,
		Reason:     "Ambiguous input, defaulting to markdown",
	}
}

// calculateMarkdownScore returns a score and reasons for markdown signals
func calculateMarkdownScore(text string) (int, []string) {
	score := 0
	var reasons []string
	lines := strings.Split(text, "\n")

	// Regex patterns
	headingRegex := regexp.MustCompile(`^>?\s*#{1,6}\s+`)
	blockquoteRegex := regexp.MustCompile(`^>\s*`)
	codeFenceRegex := regexp.MustCompile("^```")
	bulletListRegex := regexp.MustCompile(`^>?\s*[-*]\s+`)
	numberedListRegex := regexp.MustCompile(`^>?\s*\d+\.\s+`)

	headingCount := 0
	blockquoteCount := 0
	codeFenceCount := 0
	bulletCount := 0
	numberedCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check for headings (# ## ### etc.)
		if headingRegex.MatchString(line) {
			headingCount++
		}

		// Check for blockquotes (>)
		if blockquoteRegex.MatchString(line) {
			blockquoteCount++
		}

		// Check for code fences (```)
		if codeFenceRegex.MatchString(trimmed) {
			codeFenceCount++
		}

		// Check for bullet lists (- or *)
		if bulletListRegex.MatchString(line) {
			bulletCount++
		}

		// Check for numbered lists (1. 2. etc.)
		if numberedListRegex.MatchString(line) {
			numberedCount++
		}
	}

	// Apply scoring weights per plan
	if headingCount > 0 {
		score += 30
		reasons = append(reasons, "headings found")
	}
	if blockquoteCount >= 2 {
		score += 25
		reasons = append(reasons, "blockquotes found")
	}
	if codeFenceCount >= 2 { // Need at least open and close
		score += 40
		reasons = append(reasons, "code fences found")
	}
	if bulletCount > 0 {
		score += 15
		reasons = append(reasons, "bullet lists found")
	}
	if numberedCount > 0 {
		score += 10
		reasons = append(reasons, "numbered lists found")
	}

	return score, reasons
}

// calculateTableScore returns a score and reasons for table signals
func calculateTableScore(text string) (int, []string) {
	score := 0
	var reasons []string
	lines := strings.Split(text, "\n")

	// Count lines with tabs and/or commas (CSV from Google Sheets etc.)
	tabLineCount := 0
	commaLineCount := 0
	columnCounts := make(map[int]int) // column count -> occurrence

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check for tabs (TSV)
		if strings.Contains(line, "\t") {
			tabLineCount++
			cols := len(strings.Split(line, "\t"))
			columnCounts[cols]++
		}
		// Check for commas (CSV) - e.g. Google Sheet export
		if strings.Contains(line, ",") {
			commaLineCount++
			cols := len(strings.Split(line, ","))
			if cols > 1 {
				columnCounts[cols]++
			}
		}
	}

	if tabLineCount > 0 {
		score += 20
		reasons = append(reasons, "tabs found")
	}
	if commaLineCount > 0 {
		score += 20
		reasons = append(reasons, "commas found (CSV)")
	}

	// Check for 3+ consistent columns
	maxColCount := 0
	maxColOccurrence := 0
	for cols, count := range columnCounts {
		if cols >= 3 && count > maxColOccurrence {
			maxColCount = cols
			maxColOccurrence = count
		}
	}

	if maxColCount >= 3 {
		score += 40
		reasons = append(reasons, "3+ consistent columns")
	}

	if maxColOccurrence >= 2 {
		score += 30
		reasons = append(reasons, "2+ rows with same column count")
	}

	return score, reasons
}
