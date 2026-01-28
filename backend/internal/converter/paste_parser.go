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
	lines := strings.Split(text, "\n")
	var matrix CellMatrix

	for _, line := range lines {
		// Handle Windows line endings
		line = strings.TrimSuffix(line, "\r")
		
		// Split by tabs
		cells := strings.Split(line, "\t")
		matrix = append(matrix, cells)
	}

	return matrix.Normalize(), nil
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

	return NewCellMatrix(records).Normalize(), nil
}

// parseSimple is a fallback for malformed input
func (p *PasteParser) parseSimple(text string) (CellMatrix, error) {
	lines := strings.Split(text, "\n")
	var matrix CellMatrix

	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		// Try comma first, then semicolon
		var cells []string
		if strings.Contains(line, ",") {
			cells = strings.Split(line, ",")
		} else if strings.Contains(line, ";") {
			cells = strings.Split(line, ";")
		} else {
			cells = []string{line}
		}
		matrix = append(matrix, cells)
	}

	return matrix.Normalize(), nil
}
