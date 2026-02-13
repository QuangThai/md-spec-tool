package converter

import (
	"encoding/csv"
	"strings"
)

// PasteParser parses pasted text (TSV or CSV)
type PasteParser struct{}

// NewPasteParser creates a new PasteParser
func NewPasteParser() *PasteParser {
	return &PasteParser{}
}

// Parse converts pasted text to CellMatrix
func (p *PasteParser) Parse(text string) (CellMatrix, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}

	// Detect format: TSV vs CSV
	// If tabs are present, treat as TSV
	if strings.Contains(text, "\t") {
		return p.parseTSV(text)
	}

	// Otherwise try CSV
	return p.parseCSV(text)
}

// parseTSV parses tab-separated values
func (p *PasteParser) parseTSV(text string) (CellMatrix, error) {
	reader := csv.NewReader(strings.NewReader(text))
	reader.FieldsPerRecord = -1 // Allow variable field counts
	reader.LazyQuotes = true
	reader.Comma = '\t'

	records, err := reader.ReadAll()
	if err != nil {
		return p.parseSimple(text)
	}
	records = processTSVRecords(records)

	return NewCellMatrix(records).Normalize(), nil
}

// parseCSV parses comma-separated values
func (p *PasteParser) parseCSV(text string) (CellMatrix, error) {
	reader := csv.NewReader(strings.NewReader(text))
	reader.FieldsPerRecord = -1 // Allow variable field counts
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		// If CSV parsing fails, fall back to simple line splitting
		return p.parseSimple(text)
	}
	records = processTSVRecords(records)

	return NewCellMatrix(records).Normalize(), nil
}

// parseSimple is a fallback for malformed input
func (p *PasteParser) parseSimple(text string) (CellMatrix, error) {
	lines := strings.Split(text, "\n")
	var matrix CellMatrix
	delimiter := detectLikelyDelimiter(lines)

	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		cells := splitSimpleLine(line, delimiter)
		matrix = append(matrix, cells)
	}

	return matrix.Normalize(), nil
}

func detectLikelyDelimiter(lines []string) rune {
	candidates := []rune{'\t', ',', ';', '|'}
	type score struct {
		delimiter   rune
		multiCol    int
		consistency int
	}
	best := score{delimiter: ','}

	for _, delimiter := range candidates {
		counts := make(map[int]int)
		multiCol := 0
		for _, line := range lines {
			line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
			if line == "" {
				continue
			}
			cols := len(strings.Split(line, string(delimiter)))
			if cols > 1 {
				multiCol++
				counts[cols]++
			}
		}

		consistency := 0
		for _, count := range counts {
			if count > consistency {
				consistency = count
			}
		}

		if multiCol > best.multiCol || (multiCol == best.multiCol && consistency > best.consistency) {
			best = score{delimiter: delimiter, multiCol: multiCol, consistency: consistency}
		}
	}

	if best.multiCol == 0 {
		return ','
	}
	return best.delimiter
}

func splitSimpleLine(line string, delimiter rune) []string {
	if strings.ContainsRune(line, delimiter) {
		return strings.Split(line, string(delimiter))
	}
	if delimiter != ',' && strings.Contains(line, ",") {
		return strings.Split(line, ",")
	}
	if delimiter != ';' && strings.Contains(line, ";") {
		return strings.Split(line, ";")
	}
	return []string{line}
}

// processTSVRecords handles non-RFC compliant TSV files where multiline cell
// content appears as separate rows (instead of being quoted per RFC 4180).
//
// Algorithm:
// 1. Use first row (header) to determine expected column count
// 2. For each subsequent row:
//
//   - If row has only 0-1 non-empty cells → it's a continuation of previous row
//
//   - Otherwise → it's a new data row (normalize to expectedCols)
//
//     3. Continuation rows are merged into the previous row's "primary content column"
//     (the column with the longest text, which is most likely the one that was split)
func processTSVRecords(records [][]string) [][]string {
	if len(records) == 0 {
		return records
	}

	// Step 1: Determine expected column count from header
	expectedCols := len(records[0])
	if expectedCols <= 1 {
		return records
	}

	// Step 2: Process records - merge continuations, normalize columns
	result := make([][]string, 0, len(records))
	var current []string

	for _, record := range records {
		if len(record) == 0 {
			continue
		}

		// Check if this is a continuation row (≤1 non-empty cell)
		if current != nil && isContinuationRow(record) {
			// Merge into current row
			current = mergeContinuation(current, record)
			continue
		}

		// This is a new data row
		if current != nil {
			result = append(result, current)
		}

		// Normalize column count
		current = normalizeColumns(record, expectedCols)
	}

	if current != nil {
		result = append(result, current)
	}

	return result
}

// isContinuationRow returns true if the row has at most 1 non-empty cell.
// Such rows are typically continuation lines from multiline cell content
// that wasn't properly quoted in the source file.
func isContinuationRow(record []string) bool {
	nonEmpty := 0
	for _, cell := range record {
		if !isEmptyCell(strings.TrimSpace(cell)) {
			nonEmpty++
			if nonEmpty > 1 {
				return false
			}
		}
	}
	return true
}

// mergeContinuation appends continuation row content to the current row.
// Content is merged into the column with the longest existing text,
// which is the most likely candidate for a split multiline cell.
func mergeContinuation(current, continuation []string) []string {
	// Collect non-empty content from continuation
	var parts []string
	for _, cell := range continuation {
		cell = strings.TrimSpace(cell)
		if !isEmptyCell(cell) {
			parts = append(parts, cell)
		}
	}
	if len(parts) == 0 {
		return current
	}

	// Find the column with the longest content (most likely split cell)
	targetIdx := findLongestContentColumn(current)
	if targetIdx < 0 {
		targetIdx = len(current) - 1
	}

	// Merge content
	joined := strings.Join(parts, "\n")
	current[targetIdx] = strings.TrimRight(current[targetIdx], " \t") + "\n" + joined
	return current
}

// normalizeColumns adjusts record to have exactly expectedCols columns.
// - Short rows: pad with empty strings, then try to fix misalignment
// - Long rows: collapse interior empty cells first, then append overflow to last column
func normalizeColumns(record []string, expectedCols int) []string {
	originalLen := len(record)

	// Pad short rows
	for len(record) < expectedCols {
		record = append(record, "")
	}

	// Handle excess columns
	if len(record) > expectedCols {
		// Try to collapse interior empty cells (spurious tabs)
		record = collapseInteriorEmpties(record, len(record)-expectedCols)

		// If still too many, append overflow to last column
		if len(record) > expectedCols {
			var overflow []string
			for _, cell := range record[expectedCols:] {
				cell = strings.TrimSpace(cell)
				if cell != "" && cell != "-" && cell != "–" {
					overflow = append(overflow, cell)
				}
			}
			record = record[:expectedCols]
			if len(overflow) > 0 {
				last := strings.TrimSpace(record[expectedCols-1])
				record[expectedCols-1] = last + " " + strings.Join(overflow, " ")
			}
		}
	}

	// Fix misalignment in short rows: if original row had fewer columns than expected,
	// interior empty cells are likely spurious tabs. Shift data left to realign.
	if originalLen < expectedCols {
		record = shiftLeftForShortRow(record, expectedCols-originalLen)
	}

	return record
}

// shiftLeftForShortRow attempts to fix column misalignment in rows that were
// originally shorter than expected. Interior empty cells in such rows are likely
// spurious tabs that should be collapsed to shift data left.
//
// Example: Row with 8 fields, expectedCols=9, empty at index 2:
// Before pad: ["22", "Clear Filter", "", "button", "-", "-", "Always", "Init..."]
// After pad:  ["22", "Clear Filter", "", "button", "-", "-", "Always", "Init...", ""]
// After shift: ["22", "Clear Filter", "button", "-", "-", "Always", "Init...", "", ""]
func shiftLeftForShortRow(record []string, paddedCount int) []string {
	if len(record) < 3 || paddedCount <= 0 {
		return record
	}

	// Count trailing empty cells (these are from padding)
	trailingEmpties := 0
	for i := len(record) - 1; i >= 0; i-- {
		if strings.TrimSpace(record[i]) == "" {
			trailingEmpties++
		} else {
			break
		}
	}

	// Only shift if we have trailing empties to "absorb" the collapse
	if trailingEmpties == 0 {
		return record
	}

	// Limit shifts to min(paddedCount, trailingEmpties)
	maxShifts := paddedCount
	if trailingEmpties < maxShifts {
		maxShifts = trailingEmpties
	}

	// Find and collapse interior empties (not first col, not in trailing zone)
	dataEnd := len(record) - trailingEmpties
	for shifted := 0; shifted < maxShifts; {
		found := false
		for i := dataEnd - 1; i >= 1; i-- {
			if strings.TrimSpace(record[i]) == "" {
				// Remove this empty, append empty at end to maintain length
				record = append(record[:i], record[i+1:]...)
				record = append(record, "")
				shifted++
				dataEnd--
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	return record
}

// collapseInteriorEmpties removes up to maxCollapse empty cells from interior
// positions (not first or last). This handles spurious extra tabs in rows.
func collapseInteriorEmpties(record []string, maxCollapse int) []string {
	for collapsed := 0; collapsed < maxCollapse; {
		found := false
		// Search from right to left (prefer removing trailing interior empties)
		for i := len(record) - 2; i >= 1; i-- {
			if strings.TrimSpace(record[i]) == "" {
				record = append(record[:i], record[i+1:]...)
				collapsed++
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	return record
}

// findLongestContentColumn returns the index of the column with the longest
// non-empty content. This heuristic identifies the cell most likely to have
// been split across multiple lines during TSV export.
func findLongestContentColumn(row []string) int {
	bestIdx := -1
	bestLen := 0
	for i, cell := range row {
		cell = strings.TrimSpace(cell)
		if isEmptyCell(cell) {
			continue
		}
		if len(cell) > bestLen {
			bestLen = len(cell)
			bestIdx = i
		}
	}
	return bestIdx
}

// isEmptyCell returns true if the cell is empty or contains only placeholder text
func isEmptyCell(cell string) bool {
	switch strings.TrimSpace(cell) {
	case "", "-", "–":
		return true
	default:
		return false
	}
}
