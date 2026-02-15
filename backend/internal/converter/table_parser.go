package converter

import (
	"fmt"
	"strings"
)

// TableParser converts CellMatrix ([][]string) to normalized Table
type TableParser struct {
	headerDetector *HeaderDetector
}

// NewTableParser creates a new TableParser
func NewTableParser() *TableParser {
	return &TableParser{
		headerDetector: NewHeaderDetector(),
	}
}

// MatrixToTable directly converts headers and data rows to Table (Phase 2)
// Used when headers are already detected
func (p *TableParser) MatrixToTable(headers []string, dataRows [][]string, sheetName string) *Table {
	if len(headers) == 0 {
		return &Table{
			SheetName: sheetName,
			Headers:   []string{},
			Rows:      []TableRow{},
			Meta: TableMeta{
				HeaderRowIndex:  0,
				TotalSourceRows: len(dataRows),
				IncludeMetadata: true,
			},
		}
	}

	// Normalize headers
	normalizedHeaders, headerMeta := p.normalizeHeaders(headers)

	// Parse rows with normalized header count
	tableRows := p.parseRowsFromSlice(dataRows, len(normalizedHeaders))

	return &Table{
		SheetName: sheetName,
		Headers:   normalizedHeaders,
		Rows:      tableRows,
		Meta: TableMeta{
			HeaderRowIndex:   0,
			TotalSourceRows:  len(dataRows),
			BlankHeaderCount: headerMeta.blankCount,
			DuplicateHeaders: headerMeta.duplicateCount,
			Warnings:         headerMeta.warnings,
			IncludeMetadata:  true,
		},
	}
}

// ParseMatrix converts a CellMatrix to a Table
// This is the new schema-agnostic parsing path
func (p *TableParser) ParseMatrix(matrix CellMatrix, sheetName string) (*Table, error) {
	if len(matrix) == 0 {
		return &Table{
			SheetName: sheetName,
			Headers:   []string{},
			Rows:      []TableRow{},
			Meta: TableMeta{
				HeaderRowIndex:  0,
				TotalSourceRows: 0,
				IncludeMetadata: true,
			},
		}, nil
	}

	// Detect header row
	headerRowIdx, confidence := p.headerDetector.DetectHeaderRow(matrix)

	// Extract and normalize headers
	rawHeaders := matrix.GetRow(headerRowIdx)
	headers, headerMeta := p.normalizeHeaders(rawHeaders)

	// Parse data rows
	dataRows := p.parseRows(matrix, headerRowIdx, len(headers))

	// Build table
	table := &Table{
		SheetName: sheetName,
		Headers:   headers,
		Rows:      dataRows,
		Meta: TableMeta{
			HeaderRowIndex:   headerRowIdx,
			TotalSourceRows:  len(matrix),
			BlankHeaderCount: headerMeta.blankCount,
			DuplicateHeaders: headerMeta.duplicateCount,
			Warnings:         headerMeta.warnings,
			IncludeMetadata:  true,
		},
	}

	// Add confidence warning if needed
	if confidence < 50 {
		table.AddWarning(fmt.Sprintf(
			"Low header detection confidence (%d%%). Header row might be incorrect (detected row %d)",
			confidence, headerRowIdx))
	}

	return table, nil
}

// headerMetadata tracks header normalization results
type headerMetadata struct {
	blankCount     int
	duplicateCount int
	warnings       []string
}

// normalizeHeaders processes raw headers and handles edge cases
func (p *TableParser) normalizeHeaders(rawHeaders []string) ([]string, headerMetadata) {
	meta := headerMetadata{
		warnings: []string{},
	}

	normalized := make([]string, len(rawHeaders))
	seen := make(map[string]int) // Track occurrences for duplicate handling

	for i, raw := range rawHeaders {
		// Normalize: trim whitespace, collapse multiple spaces
		header := strings.TrimSpace(raw)
		header = strings.Join(strings.Fields(header), " ")

		// Handle blank headers
		if header == "" {
			meta.blankCount++
			header = fmt.Sprintf("Column %d", i+1)
			meta.warnings = append(meta.warnings,
				fmt.Sprintf("Blank header at position %d renamed to '%s'", i, header))
		}

		// Handle duplicate headers
		if count, exists := seen[header]; exists {
			meta.duplicateCount++
			newName := fmt.Sprintf("%s (%d)", header, count+1)
			meta.warnings = append(meta.warnings,
				fmt.Sprintf("Duplicate header '%s' at position %d renamed to '%s'",
					header, i, newName))
			seen[header] = count + 1
			header = newName
		} else {
			seen[header] = 1
		}

		normalized[i] = header
	}

	return normalized, meta
}

// parseRows extracts data rows and aligns them to header count
func (p *TableParser) parseRows(matrix CellMatrix, headerRowIdx int, headerCount int) []TableRow {
	var rows []TableRow

	// Skip header row and process remaining rows
	for i := headerRowIdx + 1; i < len(matrix); i++ {
		rawRow := matrix[i]

		// Check if row is completely empty
		if p.isEmptyRow(rawRow) {
			continue
		}

		// Align row to header count
		cells := p.alignRow(rawRow, headerCount)

		rows = append(rows, TableRow{Cells: cells})
	}

	return rows
}

// parseRowsFromSlice extracts data rows from a slice (Phase 2)
func (p *TableParser) parseRowsFromSlice(dataRows [][]string, headerCount int) []TableRow {
	var rows []TableRow

	for _, rawRow := range dataRows {
		// Check if row is completely empty
		if p.isEmptyRow(rawRow) {
			continue
		}

		// Align row to header count
		cells := p.alignRow(rawRow, headerCount)

		rows = append(rows, TableRow{Cells: cells})
	}

	return rows
}

// isEmptyRow checks if a row has no meaningful content
func (p *TableParser) isEmptyRow(row []string) bool {
	for _, cell := range row {
		trimmed := strings.TrimSpace(cell)
		if trimmed != "" && trimmed != "-" {
			return false
		}
	}
	return true
}

// alignRow ensures row has exactly headerCount cells
// Pads with empty strings if too short, truncates if too long
func (p *TableParser) alignRow(row []string, headerCount int) []string {
	aligned := make([]string, headerCount)

	for i := 0; i < headerCount; i++ {
		if i < len(row) {
			aligned[i] = row[i]
		} else {
			aligned[i] = ""
		}
	}

	// Note: If row is longer than headers, extra cells are dropped
	// This is intentional to maintain alignment

	return aligned
}

// ParseMatrixWithHeaderRow parses a matrix with an explicitly specified header row
func (p *TableParser) ParseMatrixWithHeaderRow(matrix CellMatrix, sheetName string, headerRowIdx int) (*Table, error) {
	if len(matrix) == 0 {
		return &Table{
			SheetName: sheetName,
			Headers:   []string{},
			Rows:      []TableRow{},
			Meta: TableMeta{
				HeaderRowIndex:  0,
				TotalSourceRows: 0,
				IncludeMetadata: true,
			},
		}, nil
	}

	if headerRowIdx < 0 || headerRowIdx >= len(matrix) {
		return nil, fmt.Errorf("invalid header row index %d (matrix has %d rows)",
			headerRowIdx, len(matrix))
	}

	// Extract and normalize headers
	rawHeaders := matrix.GetRow(headerRowIdx)
	headers, headerMeta := p.normalizeHeaders(rawHeaders)

	// Parse data rows
	dataRows := p.parseRows(matrix, headerRowIdx, len(headers))

	// Build table
	table := &Table{
		SheetName: sheetName,
		Headers:   headers,
		Rows:      dataRows,
		Meta: TableMeta{
			HeaderRowIndex:   headerRowIdx,
			TotalSourceRows:  len(matrix),
			BlankHeaderCount: headerMeta.blankCount,
			DuplicateHeaders: headerMeta.duplicateCount,
			Warnings:         headerMeta.warnings,
			IncludeMetadata:  true,
		},
	}

	return table, nil
}
