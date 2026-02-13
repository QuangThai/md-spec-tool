package converter

import "strings"

func applyColumnOverrides(matrix CellMatrix, overrides map[string]string) CellMatrix {
	if len(matrix) == 0 {
		return matrix
	}
	detector := NewHeaderDetector()
	headerRowIndex, _ := detector.DetectHeaderRow(matrix)
	if headerRowIndex < 0 || headerRowIndex >= len(matrix) {
		return matrix
	}
	headerRow := matrix[headerRowIndex]
	if len(headerRow) == 0 {
		return matrix
	}

	updated := make(CellMatrix, len(matrix))
	for rowIdx := range matrix {
		updated[rowIdx] = make([]string, len(matrix[rowIdx]))
		copy(updated[rowIdx], matrix[rowIdx])
	}

	for i, header := range headerRow {
		trimmed := strings.TrimSpace(header)
		if trimmed == "" {
			continue
		}
		if override, ok := overrides[trimmed]; ok {
			override = strings.TrimSpace(override)
			if override != "" {
				updated[headerRowIndex][i] = override
				continue
			}
		}
		if override, ok := overrides[header]; ok {
			override = strings.TrimSpace(override)
			if override != "" {
				updated[headerRowIndex][i] = override
			}
		}
	}

	return updated
}
