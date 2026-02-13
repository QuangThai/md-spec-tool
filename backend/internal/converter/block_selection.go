package converter

// BlockSelectionCandidate holds lightweight scoring signals for block ranking.
type BlockSelectionCandidate struct {
	EnglishScore float64
	QualityScore float64
}

// SelectPreferredBlock picks the default block for preview.
// It prefers English-heavy blocks and falls back to best quality.
func SelectPreferredBlock(candidates []BlockSelectionCandidate) int {
	if len(candidates) == 0 {
		return 0
	}

	selectedIdx := 0
	bestEnglish := -1.0
	bestEnglishQuality := -1.0

	for i, candidate := range candidates {
		if candidate.EnglishScore > bestEnglish || (candidate.EnglishScore == bestEnglish && candidate.QualityScore > bestEnglishQuality) {
			bestEnglish = candidate.EnglishScore
			bestEnglishQuality = candidate.QualityScore
			selectedIdx = i
		}
	}

	if bestEnglish >= 0.2 {
		return selectedIdx
	}

	bestQuality := candidates[0].QualityScore
	selectedIdx = 0
	for i := 1; i < len(candidates); i++ {
		if candidates[i].QualityScore > bestQuality {
			bestQuality = candidates[i].QualityScore
			selectedIdx = i
		}
	}

	return selectedIdx
}
