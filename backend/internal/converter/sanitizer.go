package converter

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const (
	MaxColumnCount = 50
	MaxSampleRows  = 5 // representative selection
	MaxCellLength  = 1000
	MaxInputRows   = 100 // for AI, not for full conversion
)

// SanitizeHeaders trims whitespace and limits column count
func SanitizeHeaders(headers []string) []string {
	if len(headers) == 0 {
		return headers
	}
	if len(headers) > MaxColumnCount {
		headers = headers[:MaxColumnCount]
	}
	result := make([]string, len(headers))
	for i, h := range headers {
		result[i] = strings.TrimSpace(h)
	}
	return result
}

// SanitizeSampleRows selects representative rows: first 2 + 2 middle + last 1
func SanitizeSampleRows(rows [][]string) [][]string {
	if len(rows) <= MaxSampleRows {
		return sanitizeRowContents(rows)
	}

	selected := make([][]string, 0, MaxSampleRows)
	n := len(rows)

	// First 2
	selected = append(selected, rows[0], rows[1])
	// 2 middle
	mid := n / 2
	selected = append(selected, rows[mid-1], rows[mid])
	// Last 1
	selected = append(selected, rows[n-1])

	return sanitizeRowContents(selected)
}

func sanitizeRowContents(rows [][]string) [][]string {
	result := make([][]string, len(rows))
	for i, row := range rows {
		sanitized := make([]string, len(row))
		for j, cell := range row {
			sanitized[j] = SanitizeCellContent(cell)
		}
		result[i] = sanitized
	}
	return result
}

// SanitizeCellContent truncates and normalizes a cell value
func SanitizeCellContent(s string) string {
	s = NormalizeUnicode(s)
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > MaxCellLength {
		runes := []rune(s)
		s = string(runes[:MaxCellLength]) + "..."
	}
	return s
}

// NormalizeUnicode applies NFKC normalization
func NormalizeUnicode(s string) string {
	return norm.NFKC.String(s)
}
