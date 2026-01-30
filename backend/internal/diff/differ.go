package diff

import (
	"fmt"
	"strings"
)

type DiffLine struct {
	Type    string `json:"type"`     // "add", "remove", "context"
	LineNum int    `json:"line_num"`
	Content string `json:"content"`
}

type DiffHunk struct {
	OldStart int        `json:"old_start"`
	OldCount int        `json:"old_count"`
	NewStart int        `json:"new_start"`
	NewCount int        `json:"new_count"`
	Lines    []DiffLine `json:"lines"`
}

type UnifiedDiff struct {
	Hunks   []DiffHunk `json:"hunks"`
	Added   int        `json:"added_lines"`
	Removed int        `json:"removed_lines"`
}

// Diff computes unified diff between two texts
func Diff(oldText, newText string) *UnifiedDiff {
	oldLines := strings.Split(oldText, "\n")
	newLines := strings.Split(newText, "\n")

	hunks := computeHunks(oldLines, newLines)

	added := 0
	removed := 0
	for _, hunk := range hunks {
		for _, line := range hunk.Lines {
			if line.Type == "add" {
				added++
			} else if line.Type == "remove" {
				removed++
			}
		}
	}

	return &UnifiedDiff{
		Hunks:   hunks,
		Added:   added,
		Removed: removed,
	}
}

// computeHunks creates diff hunks using simple line matching
func computeHunks(oldLines, newLines []string) []DiffHunk {
	hunk := DiffHunk{
		OldStart: 1,
		OldCount: len(oldLines),
		NewStart: 1,
		NewCount: len(newLines),
	}

	// Create maps for quick lookup
	oldMap := make(map[string]int)
	for i, line := range oldLines {
		if line != "" {
			oldMap[line] = i
		}
	}

	newMap := make(map[string]int)
	for i, line := range newLines {
		if line != "" {
			newMap[line] = i
		}
	}

	// Mark removed lines
	for i, line := range oldLines {
		if _, exists := newMap[line]; !exists {
			hunk.Lines = append(hunk.Lines, DiffLine{
				Type:    "remove",
				LineNum: i + 1,
				Content: line,
			})
		}
	}

	// Mark added lines
	for i, line := range newLines {
		if _, exists := oldMap[line]; !exists {
			hunk.Lines = append(hunk.Lines, DiffLine{
				Type:    "add",
				LineNum: i + 1,
				Content: line,
			})
		}
	}

	return []DiffHunk{hunk}
}

// FormatUnified returns unified diff format (text)
func FormatUnified(d *UnifiedDiff) string {
	var buf strings.Builder

	buf.WriteString("--- original\n")
	buf.WriteString("+++ modified\n")

	for _, hunk := range d.Hunks {
		buf.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
			hunk.OldStart, hunk.OldCount, hunk.NewStart, hunk.NewCount))

		for _, line := range hunk.Lines {
			switch line.Type {
			case "remove":
				buf.WriteString(fmt.Sprintf("-%s\n", line.Content))
			case "add":
				buf.WriteString(fmt.Sprintf("+%s\n", line.Content))
			case "context":
				buf.WriteString(fmt.Sprintf(" %s\n", line.Content))
			}
		}
	}

	return buf.String()
}
