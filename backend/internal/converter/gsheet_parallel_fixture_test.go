package converter

import (
	"os"
	"strings"
	"testing"
)

func TestGSheetParallelJapaneseEnglishFixture_SelectsEnglishBlock(t *testing.T) {
	content, err := os.ReadFile("../../../use-cases/gsheet_parallel_jp_en.csv")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	parser := NewPasteParser()
	matrix, err := parser.Parse(string(content))
	if err != nil {
		t.Fatalf("failed to parse fixture: %v", err)
	}

	blocks := DetectTableBlocks(matrix)
	if len(blocks) < 2 {
		t.Fatalf("expected at least 2 blocks, got %d", len(blocks))
	}

	conv := NewConverter()
	candidates := make([]BlockSelectionCandidate, 0, len(blocks))
	selectedHeaders := make([][]string, 0, len(blocks))

	for _, block := range blocks {
		detector := NewHeaderDetector()
		headerRow, confidence := detector.DetectHeaderRow(block.Matrix)
		headers := block.Matrix.GetRow(headerRow)
		dataRows := block.Matrix.SliceRows(headerRow+1, block.Matrix.RowCount())

		mapping, unmapped := conv.GetPreviewColumnMappingRuleBased(headers, DefaultTemplateName)
		quality := BuildPreviewMappingQuality(confidence, headers, dataRows, mapping, unmapped)
		englishScore := EstimateEnglishScore(headers, dataRows)

		candidates = append(candidates, BlockSelectionCandidate{
			EnglishScore: englishScore,
			QualityScore: quality.Score,
			RowCount:     len(dataRows),
			ColumnCount:  len(headers),
		})
		selectedHeaders = append(selectedHeaders, headers)
	}

	selectedIdx := SelectPreferredBlock(candidates)
	if selectedIdx < 0 || selectedIdx >= len(selectedHeaders) {
		t.Fatalf("invalid selected index %d", selectedIdx)
	}

	headerLine := strings.ToLower(strings.Join(selectedHeaders[selectedIdx], " | "))
	if !strings.Contains(headerLine, "item name") {
		t.Fatalf("expected english block to be selected, got headers: %s", headerLine)
	}
}
