package converter

// BlockSelectionCandidate holds lightweight scoring signals for block ranking.
type BlockSelectionCandidate struct {
	EnglishScore float64
	QualityScore float64
	RowCount     int
	ColumnCount  int
}

// SelectPreferredBlock picks the default block for preview.
// It prefers structurally rich blocks (enough rows/columns) and mapping quality,
// while staying language-neutral for multilingual sheets.
func SelectPreferredBlock(candidates []BlockSelectionCandidate) int {
	if len(candidates) == 0 {
		return 0
	}

	hasRichRows := false
	for _, c := range candidates {
		if c.RowCount >= 2 {
			hasRichRows = true
			break
		}
	}

	hasWideStructuredBlock := false
	for _, c := range candidates {
		if c.RowCount >= 2 && c.ColumnCount >= 4 {
			hasWideStructuredBlock = true
			break
		}
	}

	selectedIdx := 0
	bestComposite := -1.0

	for i, candidate := range candidates {
		if hasRichRows && candidate.RowCount < 2 {
			continue
		}
		if hasWideStructuredBlock && candidate.ColumnCount < 4 {
			continue
		}

		rowNorm := normalizeCount(candidate.RowCount, 8)
		colNorm := normalizeCount(candidate.ColumnCount, 6)
		richness := (rowNorm * 0.6) + (colNorm * 0.4)
		composite := (candidate.QualityScore * 0.7) + (richness * 0.3)

		if composite > bestComposite {
			bestComposite = composite
			selectedIdx = i
			continue
		}

		if composite == bestComposite {
			if candidate.QualityScore > candidates[selectedIdx].QualityScore {
				selectedIdx = i
				continue
			}
			if candidate.RowCount > candidates[selectedIdx].RowCount {
				selectedIdx = i
				continue
			}
			if candidate.ColumnCount > candidates[selectedIdx].ColumnCount {
				selectedIdx = i
				continue
			}
			if candidate.EnglishScore > candidates[selectedIdx].EnglishScore {
				selectedIdx = i
			}
		}
	}

	if bestComposite < 0 {
		return 0
	}

	return selectedIdx
}

func normalizeCount(value int, cap int) float64 {
	if value <= 0 {
		return 0
	}
	if value >= cap {
		return 1
	}
	return float64(value) / float64(cap)
}
