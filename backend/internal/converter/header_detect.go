package converter

import (
	"strings"
)

// HeaderDetector detects the header row in a matrix
type HeaderDetector struct {
	mapper *ColumnMapper
}

// NewHeaderDetector creates a new HeaderDetector
func NewHeaderDetector() *HeaderDetector {
	return &HeaderDetector{
		mapper: NewColumnMapper(),
	}
}

// DetectHeaderRow finds the most likely header row
// Returns the row index (0-based) and a confidence score (0-100)
func (d *HeaderDetector) DetectHeaderRow(matrix CellMatrix) (int, int) {
	if len(matrix) == 0 {
		return 0, 0
	}

	bestRow := 0
	bestScore := 0

	// Check first few rows (headers are usually in first 5 rows)
	maxCheck := min(5, len(matrix))
	for i := 0; i < maxCheck; i++ {
		score := d.scoreRow(matrix[i])
		if score > bestScore {
			bestScore = score
			bestRow = i
		}
	}

	return bestRow, bestScore
}

// scoreRow calculates how likely a row is to be a header
func (d *HeaderDetector) scoreRow(row []string) int {
	if len(row) == 0 {
		return 0
	}

	// Penalize rows with markdown markers
	for _, cell := range row {
		if d.hasMarkdownMarkers(cell) {
			return 0 // Definitely not a header row
		}
	}

	// Penalize rows with too few non-empty cells
	nonEmpty := 0
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			nonEmpty++
		}
	}
	if nonEmpty < 2 {
		return 0
	}

	score := 0
	matchedFields := 0

	for _, cell := range row {
		normalized := strings.ToLower(strings.TrimSpace(cell))
		if _, ok := HeaderSynonyms[normalized]; ok {
			matchedFields++
			score += 25 // Each matched header synonym adds 25 points
		}

		// Bonus for typical header characteristics
		if d.looksLikeHeader(cell) {
			score += 5
		}
	}

	// Bonus for having multiple recognized headers
	if matchedFields >= 2 {
		score += 20
	}
	if matchedFields >= 3 {
		score += 30
	}

	return min(score, 100)
}

// hasMarkdownMarkers checks if a cell contains markdown indicators
func (d *HeaderDetector) hasMarkdownMarkers(cell string) bool {
	trimmed := strings.TrimSpace(cell)
	return strings.HasPrefix(trimmed, "#") ||
		strings.HasPrefix(trimmed, ">") ||
		strings.HasPrefix(trimmed, "```") ||
		strings.HasPrefix(trimmed, "- ") ||
		strings.HasPrefix(trimmed, "* ")
}

// looksLikeHeader checks if a cell value looks like a header
func (d *HeaderDetector) looksLikeHeader(cell string) bool {
	cell = strings.TrimSpace(cell)
	if cell == "" {
		return false
	}

	// Headers typically:
	// - Don't start with numbers
	// - Are relatively short
	// - Don't contain long sentences
	// - Are often capitalized

	if len(cell) > 50 {
		return false
	}

	// Check if it starts with a number (data rows often do)
	if len(cell) > 0 && cell[0] >= '0' && cell[0] <= '9' {
		return false
	}

	// Check for sentence-like structure (data often has this)
	if strings.Contains(cell, ". ") {
		return false
	}

	// Headers often have underscores or are single/few words
	wordCount := len(strings.Fields(cell))
	if wordCount <= 3 {
		return true
	}

	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
