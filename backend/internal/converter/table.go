package converter

import "fmt"

// Table represents a normalized, schema-agnostic table structure
// This is the new core model that replaces direct SpecDoc building
type Table struct {
	SheetName string     // Optional sheet/source name
	Headers   []string   // Normalized column headers (guaranteed unique)
	Rows      []TableRow // Data rows (aligned to Headers length)
	Meta      TableMeta  // Parsing metadata and warnings
}

// TableRow represents a single row with cells aligned to headers
type TableRow struct {
	Cells []string // Values aligned to Table.Headers (same length)
}

// TableMeta contains metadata about the parsed table
type TableMeta struct {
	HeaderRowIndex    int       // 0-based index of header row in source
	SourceURL         string    // Optional source URL (for Google Sheets)
	Warnings          []string  // Parsing warnings (e.g., "Duplicate header 'Status' renamed to 'Status (2)'")
	TotalSourceRows   int       // Total rows in source (before filtering)
	BlankHeaderCount  int       // Count of blank headers encountered
	DuplicateHeaders  int       // Count of duplicate headers encountered
	ColumnMap         ColumnMap // Canonical field -> column index mapping
	AIMode            string    // "on", "shadow", "off"
	AIUsed            bool      // True if AI mapping was used
	AIDegraded        bool      // True if fallback was used
	AIAvgConfidence   float64   // Average AI confidence
	AIMappedColumns   int       // Count of mapped columns by AI
	AIUnmappedColumns int       // Count of unmapped columns by AI
}

// NewTable creates a new Table with basic validation
func NewTable(sheetName string, headers []string, rows []TableRow) *Table {
	return &Table{
		SheetName: sheetName,
		Headers:   headers,
		Rows:      rows,
		Meta: TableMeta{
			HeaderRowIndex:  0,
			TotalSourceRows: len(rows),
		},
	}
}

// RowCount returns the number of data rows
func (t *Table) RowCount() int {
	return len(t.Rows)
}

// ColumnCount returns the number of columns (headers)
func (t *Table) ColumnCount() int {
	return len(t.Headers)
}

// GetCell safely retrieves a cell value by row and column index
func (t *Table) GetCell(rowIdx, colIdx int) string {
	if rowIdx < 0 || rowIdx >= len(t.Rows) {
		return ""
	}
	if colIdx < 0 || colIdx >= len(t.Rows[rowIdx].Cells) {
		return ""
	}
	return t.Rows[rowIdx].Cells[colIdx]
}

// GetCellByHeader retrieves a cell value by row index and header name
func (t *Table) GetCellByHeader(rowIdx int, headerName string) string {
	colIdx := t.FindHeaderIndex(headerName)
	if colIdx == -1 {
		return ""
	}
	return t.GetCell(rowIdx, colIdx)
}

// FindHeaderIndex finds the column index for a given header name
// Returns -1 if not found
func (t *Table) FindHeaderIndex(headerName string) int {
	for i, h := range t.Headers {
		if h == headerName {
			return i
		}
	}
	return -1
}

// HasHeader checks if a header exists
func (t *Table) HasHeader(headerName string) bool {
	return t.FindHeaderIndex(headerName) != -1
}

// GetRow returns a complete row by index
func (t *Table) GetRow(rowIdx int) TableRow {
	if rowIdx < 0 || rowIdx >= len(t.Rows) {
		return TableRow{Cells: []string{}}
	}
	return t.Rows[rowIdx]
}

// AddWarning adds a parsing warning to metadata
func (t *Table) AddWarning(warning string) {
	t.Meta.Warnings = append(t.Meta.Warnings, warning)
}

// Validate checks table integrity
func (t *Table) Validate() []string {
	var errors []string

	// Check all rows have correct cell count
	expectedCols := len(t.Headers)
	for i, row := range t.Rows {
		if len(row.Cells) != expectedCols {
			errors = append(errors,
				fmt.Sprintf("Row %d has %d cells, expected %d",
					i, len(row.Cells), expectedCols))
		}
	}

	return errors
}

// ToMap converts a row to a map of header -> value
func (tr TableRow) ToMap(headers []string) map[string]string {
	m := make(map[string]string)
	for i, header := range headers {
		if i < len(tr.Cells) {
			m[header] = tr.Cells[i]
		} else {
			m[header] = ""
		}
	}
	return m
}

// NewTableRow creates a new TableRow with specified cells
func NewTableRow(cells []string) TableRow {
	return TableRow{Cells: cells}
}

// PadTo ensures the row has at least n cells (padding with empty strings)
func (tr *TableRow) PadTo(n int) {
	if len(tr.Cells) >= n {
		return
	}
	padding := make([]string, n-len(tr.Cells))
	tr.Cells = append(tr.Cells, padding...)
}

// TruncateTo truncates the row to exactly n cells
func (tr *TableRow) TruncateTo(n int) {
	if len(tr.Cells) <= n {
		return
	}
	tr.Cells = tr.Cells[:n]
}

// AlignTo ensures the row has exactly n cells (pad or truncate)
func (tr *TableRow) AlignTo(n int) {
	if len(tr.Cells) < n {
		tr.PadTo(n)
	} else if len(tr.Cells) > n {
		tr.TruncateTo(n)
	}
}
