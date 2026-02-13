package converter

import (
	"fmt"
	"strings"
	"unicode"
)

// MatrixBlock represents a dense table-like region inside a sheet matrix.
type MatrixBlock struct {
	ID       string
	StartCol int
	EndCol   int
	StartRow int
	EndRow   int
	Matrix   CellMatrix
}

// DetectTableBlocks finds independent table regions separated by blank columns.
func DetectTableBlocks(matrix CellMatrix) []MatrixBlock {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return nil
	}

	rows := len(matrix)
	cols := len(matrix[0])
	active := make([]bool, cols)

	inspectRows := min(rows, 10)
	for c := 0; c < cols; c++ {
		nonEmpty := 0
		topNonEmpty := 0
		for r := 0; r < rows; r++ {
			if strings.TrimSpace(matrix[r][c]) != "" {
				nonEmpty++
				if r < inspectRows {
					topNonEmpty++
				}
			}
		}

		ratio := float64(nonEmpty) / float64(rows)
		active[c] = nonEmpty >= 2 && (topNonEmpty > 0 || ratio >= 0.2)
	}

	segments := make([][2]int, 0)
	for c := 0; c < cols; {
		if !active[c] {
			c++
			continue
		}
		start := c
		for c < cols && active[c] {
			c++
		}
		end := c - 1
		if end-start+1 >= 2 {
			segments = append(segments, [2]int{start, end})
		}
	}

	if len(segments) == 0 {
		segments = append(segments, [2]int{0, cols - 1})
	}

	blocks := make([]MatrixBlock, 0, len(segments))
	for i, seg := range segments {
		startCol, endCol := seg[0], seg[1]
		startRow, endRow := detectRowBounds(matrix, startCol, endCol)
		if startRow < 0 || endRow < startRow {
			continue
		}

		sub := sliceMatrix(matrix, startRow, endRow, startCol, endCol)
		norm := NewCellMatrix(sub).Normalize()
		if len(norm) < 2 || len(norm[0]) < 2 {
			continue
		}

		blocks = append(blocks, MatrixBlock{
			ID:       fmt.Sprintf("block_%d", i+1),
			StartCol: startCol,
			EndCol:   endCol,
			StartRow: startRow,
			EndRow:   endRow,
			Matrix:   norm,
		})
	}

	if len(blocks) == 0 {
		return []MatrixBlock{{
			ID:       "block_1",
			StartCol: 0,
			EndCol:   cols - 1,
			StartRow: 0,
			EndRow:   rows - 1,
			Matrix:   NewCellMatrix(matrix).Normalize(),
		}}
	}

	return blocks
}

func detectRowBounds(matrix CellMatrix, startCol, endCol int) (int, int) {
	startRow := -1
	endRow := -1
	for r := 0; r < len(matrix); r++ {
		hasContent := false
		for c := startCol; c <= endCol && c < len(matrix[r]); c++ {
			if strings.TrimSpace(matrix[r][c]) != "" {
				hasContent = true
				break
			}
		}
		if hasContent {
			if startRow == -1 {
				startRow = r
			}
			endRow = r
		}
	}
	return startRow, endRow
}

func sliceMatrix(matrix CellMatrix, startRow, endRow, startCol, endCol int) CellMatrix {
	out := make(CellMatrix, 0, endRow-startRow+1)
	for r := startRow; r <= endRow && r < len(matrix); r++ {
		row := make([]string, 0, endCol-startCol+1)
		for c := startCol; c <= endCol && c < len(matrix[r]); c++ {
			row = append(row, matrix[r][c])
		}
		out = append(out, row)
	}
	return out
}

// EstimateEnglishScore returns ratio of Latin script letters in text.
func EstimateEnglishScore(headers []string, rows [][]string) float64 {
	latin := 0
	japanese := 0
	otherLetters := 0

	count := func(text string) {
		for _, r := range text {
			if !unicode.IsLetter(r) {
				continue
			}
			switch {
			case unicode.In(r, unicode.Latin):
				latin++
			case unicode.In(r, unicode.Hiragana, unicode.Katakana, unicode.Han):
				japanese++
			default:
				otherLetters++
			}
		}
	}

	for _, h := range headers {
		count(h)
	}

	maxRows := min(len(rows), 30)
	for i := 0; i < maxRows; i++ {
		for _, cell := range rows[i] {
			count(cell)
		}
	}

	total := latin + japanese + otherLetters
	if total == 0 || latin < 3 {
		return 0
	}
	return float64(latin) / float64(total)
}

func DetectLanguageHint(englishScore float64, headers []string, rows [][]string) string {
	japaneseScore := estimateJapaneseScore(headers, rows)
	if englishScore >= 0.55 && englishScore > japaneseScore {
		return "english"
	}
	if japaneseScore >= 0.55 && japaneseScore > englishScore {
		return "japanese"
	}
	if englishScore > 0 || japaneseScore > 0 {
		return "mixed"
	}
	return "unknown"
}

func estimateJapaneseScore(headers []string, rows [][]string) float64 {
	latin := 0
	japanese := 0
	count := func(text string) {
		for _, r := range text {
			if !unicode.IsLetter(r) {
				continue
			}
			if unicode.In(r, unicode.Hiragana, unicode.Katakana, unicode.Han) {
				japanese++
			} else if unicode.In(r, unicode.Latin) {
				latin++
			}
		}
	}
	for _, h := range headers {
		count(h)
	}
	maxRows := min(len(rows), 30)
	for i := 0; i < maxRows; i++ {
		for _, cell := range rows[i] {
			count(cell)
		}
	}
	if japanese+latin == 0 {
		return 0
	}
	return float64(japanese) / float64(japanese+latin)
}
