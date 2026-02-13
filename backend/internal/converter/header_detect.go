package converter

import (
	"strings"
)

// Scoring constants for header detection heuristics
const (
	scoreSynonymMatch     = 25  // Points for each matched header synonym
	scoreHeaderFormat     = 5   // Points for header-like characteristics
	scoreMultiHeader2     = 20  // Bonus for having 2+ recognized headers
	scoreMultiHeader3     = 30  // Bonus for having 3+ recognized headers
	maxRowsToCheck        = 5   // Maximum rows to check for header detection
	minNonEmptyCells      = 2   // Minimum non-empty cells to consider as header
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

	// Check first few rows (headers are usually in first maxRowsToCheck rows)
	maxCheck := min(maxRowsToCheck, len(matrix))
	for i := 0; i < maxCheck; i++ {
		score := d.scoreRow(matrix[i])
		if i+1 < len(matrix) {
			score += d.scoreHeaderDataSeparation(matrix[i], matrix[i+1])
		}
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
	if nonEmpty < minNonEmptyCells {
		return 0
	}

	score := 0
	matchedFields := 0
	dataLikeCount := 0 // Count cells that look like data, not headers

	for _, cell := range row {
		trimmed := strings.TrimSpace(cell)
		
		// Normalize and check for known synonyms
		normalized := normalizeHeader(cell)
		if _, ok := HeaderSynonyms[normalized]; ok {
			matchedFields++
			score += scoreSynonymMatch // Each matched header synonym adds scoreSynonymMatch points
		}

		// Check for header-like characteristics
		if d.looksLikeHeader(cell) {
			score += scoreHeaderFormat
		} else if d.looksLikeData(trimmed) {
			dataLikeCount++
		}
	}

	// Bonus for having multiple recognized headers
	if matchedFields >= 2 {
		score += scoreMultiHeader2
	}
	if matchedFields >= 3 {
		score += scoreMultiHeader3
	}

	// Penalty if too many cells look like data rather than headers
	if dataLikeCount > nonEmpty/2 && matchedFields == 0 {
		return 0
	}

	return min(score, 100)
}

func (d *HeaderDetector) scoreHeaderDataSeparation(header []string, sample []string) int {
	if len(header) == 0 || len(sample) == 0 {
		return 0
	}

	limit := min(len(header), len(sample))
	if limit == 0 {
		return 0
	}

	bonus := 0
	for i := 0; i < limit; i++ {
		h := strings.TrimSpace(header[i])
		s := strings.TrimSpace(sample[i])
		if h == "" || s == "" {
			continue
		}

		if looksLikeNumber(s) && !looksLikeNumber(h) {
			bonus += 3
		}
		if len(s) > len(h)+8 {
			bonus += 2
		}
		if h == strings.ToUpper(h) && h != s {
			bonus += 1
		}
	}

	if bonus > 20 {
		bonus = 20
	}
	return bonus
}

func looksLikeNumber(value string) bool {
	if value == "" {
		return false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
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
	// - Contain descriptive/categorical terms

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

// looksLikeData checks if a cell value looks like actual data (not a header)
func (d *HeaderDetector) looksLikeData(cell string) bool {
	if cell == "" {
		return false
	}

	// Data characteristics:
	// - Starts with a number (ID, date, or numeric value)
	// - Contains newlines (multi-line content)
	// - Is very long (paragraph of text)
	// - Contains special characters that indicate structured data

	// Check if it's purely numeric (IDs, dates, counts)
	if d.isPureNumeric(cell) {
		return true
	}

	// Check for multiple lines
	if strings.Contains(cell, "\n") {
		return true
	}

	// Check for very long text (data, not header)
	if len(cell) > 100 {
		return true
	}

	// Check for typical data patterns like URLs, emails
	if strings.Contains(cell, "@") || strings.Contains(cell, "http") {
		return true
	}

	return false
}

// isPureNumeric checks if a cell contains only numeric characters
func (d *HeaderDetector) isPureNumeric(cell string) bool {
	cell = strings.TrimSpace(cell)
	if cell == "" {
		return false
	}
	for _, ch := range cell {
		if (ch < '0' || ch > '9') && ch != '.' && ch != '-' && ch != '+' {
			return false
		}
	}
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
