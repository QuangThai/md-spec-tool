package converter

import (
	"fmt"
	"strings"
)

// Renderer is the interface for rendering Table to markdown
type Renderer interface {
	Render(table *Table) (markdown string, warnings []string, err error)
}

// GenericTableRenderer renders any Table to a standard Markdown table format
// It supports any number of columns and preserves all data
type GenericTableRenderer struct {
	escapeSpecialChars bool
	maxCellWidth       int // 0 = unlimited
}

// NewGenericTableRenderer creates a new GenericTableRenderer
func NewGenericTableRenderer() *GenericTableRenderer {
	return &GenericTableRenderer{
		escapeSpecialChars: true,
		maxCellWidth:       0, // No width limit by default
	}
}

// WithMaxCellWidth sets the maximum width for cells (for line wrapping)
// 0 means unlimited
func (r *GenericTableRenderer) WithMaxCellWidth(width int) *GenericTableRenderer {
	r.maxCellWidth = width
	return r
}

// Render converts a Table to Markdown table format
func (r *GenericTableRenderer) Render(table *Table) (string, []string, error) {
	if table == nil {
		return "", nil, fmt.Errorf("table is nil")
	}

	warnings := []string{}

	// Validate table
	if len(table.Headers) == 0 {
		return "", warnings, fmt.Errorf("table has no headers")
	}

	if len(table.Rows) == 0 {
		// Empty table - return header only
		return r.renderHeaderOnly(table.Headers), warnings, nil
	}

	// Calculate column widths
	colWidths := r.calculateColumnWidths(table)

	// Build markdown
	var buf strings.Builder

	// Write header
	r.writeHeaderRow(&buf, table.Headers, colWidths)
	buf.WriteString("\n")

	// Write separator
	r.writeSeparatorRow(&buf, table.Headers, colWidths)
	buf.WriteString("\n")

	// Write data rows
	for i, row := range table.Rows {
		if len(row.Cells) != len(table.Headers) {
			warnings = append(warnings, fmt.Sprintf("Row %d has %d cells, expected %d", i, len(row.Cells), len(table.Headers)))
		}

		r.writeDataRow(&buf, row.Cells, table.Headers, colWidths)
		buf.WriteString("\n")
	}

	// Collect metadata warnings
	if table.Meta.BlankHeaderCount > 0 {
		warnings = append(warnings, fmt.Sprintf("Found %d blank headers, auto-renamed", table.Meta.BlankHeaderCount))
	}

	if table.Meta.DuplicateHeaders > 0 {
		warnings = append(warnings, fmt.Sprintf("Found %d duplicate headers, renamed with (N) suffix", table.Meta.DuplicateHeaders))
	}

	if len(table.Meta.Warnings) > 0 {
		warnings = append(warnings, table.Meta.Warnings...)
	}

	return buf.String(), warnings, nil
}

// calculateColumnWidths determines the width of each column
// Width = max(header length, max cell length in column)
func (r *GenericTableRenderer) calculateColumnWidths(table *Table) []int {
	widths := make([]int, len(table.Headers))

	// Initialize with header widths
	for i, header := range table.Headers {
		widths[i] = len(header)
	}

	// Check cell widths
	for _, row := range table.Rows {
		for i, cell := range row.Cells {
			if i < len(widths) {
				cellLen := len(cell)
				if cellLen > widths[i] {
					widths[i] = cellLen
				}
			}
		}
	}

	// Apply max width limit if set
	if r.maxCellWidth > 0 {
		for i := range widths {
			if widths[i] > r.maxCellWidth {
				widths[i] = r.maxCellWidth
			}
		}
	}

	return widths
}

// writeHeaderRow writes the header row with proper spacing
func (r *GenericTableRenderer) writeHeaderRow(buf *strings.Builder, headers []string, widths []int) {
	buf.WriteString("| ")
	for i, header := range headers {
		padded := r.padCell(header, widths[i])
		buf.WriteString(padded)
		buf.WriteString(" | ")
	}
}

// writeSeparatorRow writes the separator row with dashes
func (r *GenericTableRenderer) writeSeparatorRow(buf *strings.Builder, headers []string, widths []int) {
	buf.WriteString("|")
	for i := range headers {
		buf.WriteString(" ")
		buf.WriteString(strings.Repeat("-", widths[i]))
		buf.WriteString(" |")
	}
}

// writeDataRow writes a data row with proper escaping and spacing
func (r *GenericTableRenderer) writeDataRow(buf *strings.Builder, cells []string, headers []string, widths []int) {
	buf.WriteString("| ")
	for i := 0; i < len(headers); i++ {
		var cell string
		if i < len(cells) {
			cell = cells[i]
		}

		// Escape and normalize
		cell = r.escapeCell(cell)

		padded := r.padCell(cell, widths[i])
		buf.WriteString(padded)
		buf.WriteString(" | ")
	}
}

// escapeCell escapes special Markdown characters in a cell
func (r *GenericTableRenderer) escapeCell(cell string) string {
	if !r.escapeSpecialChars {
		return cell
	}

	// Escape pipes - they break Markdown table syntax
	cell = strings.ReplaceAll(cell, "|", "\\|")

	// Replace newlines with <br> tags for multi-line content
	// Must process CRLF first, then remaining CR/LF to handle Windows files correctly
	cell = strings.ReplaceAll(cell, "\r\n", "<br>")
	cell = strings.ReplaceAll(cell, "\r", "<br>")
	cell = strings.ReplaceAll(cell, "\n", "<br>")

	return cell
}

// padCell pads a cell value to match column width
func (r *GenericTableRenderer) padCell(cell string, width int) string {
	if len(cell) >= width {
		return cell[:width]
	}
	return cell + strings.Repeat(" ", width-len(cell))
}

// renderHeaderOnly returns a table with just headers and no data rows
func (r *GenericTableRenderer) renderHeaderOnly(headers []string) string {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}

	var buf strings.Builder
	r.writeHeaderRow(&buf, headers, widths)
	buf.WriteString("\n")
	r.writeSeparatorRow(&buf, headers, widths)
	buf.WriteString("\n")
	return buf.String()
}
