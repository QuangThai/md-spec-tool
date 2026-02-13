package converter

import "testing"

func TestDetectTableBlocks_SplitsParallelTables(t *testing.T) {
	matrix := NewCellMatrix([][]string{
		{"No", "Item Name", "", "No", "項目名"},
		{"1", "Subscribed Channels", "", "1", "登録チャンネル"},
		{"2", "Side Menu", "", "2", "サイドメニュー"},
	})

	blocks := DetectTableBlocks(matrix.Normalize())
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}

	if blocks[0].StartCol != 0 || blocks[0].EndCol != 1 {
		t.Fatalf("unexpected first block range: %d-%d", blocks[0].StartCol, blocks[0].EndCol)
	}
	if blocks[1].StartCol != 3 || blocks[1].EndCol != 4 {
		t.Fatalf("unexpected second block range: %d-%d", blocks[1].StartCol, blocks[1].EndCol)
	}
}

func TestEstimateEnglishScore_EnglishHigherThanJapanese(t *testing.T) {
	headers := []string{"No", "Item Name", "Item Type", "Display"}
	rows := [][]string{{"1", "Subscribed Channels", "text", "Display tags"}}

	score := EstimateEnglishScore(headers, rows)
	if score < 0.7 {
		t.Fatalf("expected high english score, got %.2f", score)
	}
}
