package converter

import (
	"strings"
)

// NewCellMatrix creates a new CellMatrix from raw rows
func NewCellMatrix(rows [][]string) CellMatrix {
	return CellMatrix(rows)
}

// Normalize cleans up the matrix by:
// - Trimming whitespace
// - Removing completely empty rows
// - Padding rows to have consistent column count
func (m CellMatrix) Normalize() CellMatrix {
	if len(m) == 0 {
		return m
	}

	// Find max column count
	maxCols := 0
	for _, row := range m {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	var result CellMatrix
	for _, row := range m {
		// Trim whitespace from all cells
		trimmedRow := make([]string, maxCols)
		for i := 0; i < maxCols; i++ {
			if i < len(row) {
				trimmedRow[i] = strings.TrimSpace(row[i])
			}
		}

		// Skip completely empty rows
		isEmpty := true
		for _, cell := range trimmedRow {
			if cell != "" {
				isEmpty = false
				break
			}
		}
		if !isEmpty {
			result = append(result, trimmedRow)
		}
	}

	return result
}

// GetColumn returns all values in a specific column
func (m CellMatrix) GetColumn(index int) []string {
	var result []string
	for _, row := range m {
		if index < len(row) {
			result = append(result, row[index])
		} else {
			result = append(result, "")
		}
	}
	return result
}

// GetRow returns a specific row
func (m CellMatrix) GetRow(index int) []string {
	if index < 0 || index >= len(m) {
		return nil
	}
	return m[index]
}

// RowCount returns the number of rows
func (m CellMatrix) RowCount() int {
	return len(m)
}

// ColCount returns the number of columns (based on first row)
func (m CellMatrix) ColCount() int {
	if len(m) == 0 {
		return 0
	}
	return len(m[0])
}

// SliceRows returns a subset of rows
func (m CellMatrix) SliceRows(start, end int) CellMatrix {
	if start < 0 {
		start = 0
	}
	if end > len(m) {
		end = len(m)
	}
	if start >= end {
		return nil
	}
	return CellMatrix(m[start:end])
}
