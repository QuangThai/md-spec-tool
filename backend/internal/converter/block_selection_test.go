package converter

import "testing"

func TestSelectPreferredBlock_PrefersRicherStructureWhenQualityClose(t *testing.T) {
	candidates := []BlockSelectionCandidate{
		{QualityScore: 0.42, RowCount: 1, ColumnCount: 2},
		{QualityScore: 0.40, RowCount: 5, ColumnCount: 8},
	}

	selected := SelectPreferredBlock(candidates)
	if selected != 1 {
		t.Fatalf("expected richer block index 1, got %d", selected)
	}
}

func TestSelectPreferredBlock_PrefersHigherQualityWhenStructureComparable(t *testing.T) {
	candidates := []BlockSelectionCandidate{
		{QualityScore: 0.72, RowCount: 5, ColumnCount: 7},
		{QualityScore: 0.61, RowCount: 5, ColumnCount: 7},
	}

	selected := SelectPreferredBlock(candidates)
	if selected != 0 {
		t.Fatalf("expected higher-quality block index 0, got %d", selected)
	}
}

func TestSelectPreferredBlock_FiltersNarrowBlockWhenWideExists(t *testing.T) {
	candidates := []BlockSelectionCandidate{
		{QualityScore: 0.90, RowCount: 3, ColumnCount: 2},
		{QualityScore: 0.55, RowCount: 3, ColumnCount: 8},
	}

	selected := SelectPreferredBlock(candidates)
	if selected != 1 {
		t.Fatalf("expected wide structured block index 1, got %d", selected)
	}
}
