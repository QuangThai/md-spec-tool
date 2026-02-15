package diff

import (
	"fmt"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

type DiffLine struct {
	Type    string `json:"type"` // "add", "remove", "context"
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

// Diff computes unified diff between two texts using Patience algorithm
// Provides better contextual diffs with proper hunk support
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

// computeHunks creates diff hunks using the Patience diff algorithm
// Provides better context lines and more accurate change detection
func computeHunks(oldLines, newLines []string) []DiffHunk {
	// Use difflib.SequenceMatcher for accurate diff computation
	matcher := difflib.NewMatcher(oldLines, newLines)
	opcodes := matcher.GetOpCodes()

	var hunks []DiffHunk
	contextLines := 3 // Number of context lines to include

	for _, opcode := range opcodes {
		tag := string(opcode.Tag)
		oldStart := opcode.I1
		oldEnd := opcode.I2
		newStart := opcode.J1
		newEnd := opcode.J2

		// Skip if no actual changes
		if tag == "e" { // 'e' for equal in difflib
			continue
		}

		// Create hunk with context
		hunkStart := oldStart
		if hunkStart > contextLines {
			hunkStart -= contextLines
		}

		hunkEnd := oldEnd
		if hunkEnd+contextLines < len(oldLines) {
			hunkEnd += contextLines
		} else {
			hunkEnd = len(oldLines)
		}

		newHunkStart := newStart
		if newHunkStart > contextLines {
			newHunkStart -= contextLines
		}

		newHunkEnd := newEnd
		if newHunkEnd+contextLines < len(newLines) {
			newHunkEnd += contextLines
		} else {
			newHunkEnd = len(newLines)
		}

		hunk := DiffHunk{
			OldStart: hunkStart + 1,
			OldCount: hunkEnd - hunkStart,
			NewStart: newHunkStart + 1,
			NewCount: newHunkEnd - newHunkStart,
		}

		// Add context lines before change
		for i := hunkStart; i < oldStart && i < len(oldLines); i++ {
			hunk.Lines = append(hunk.Lines, DiffLine{
				Type:    "context",
				LineNum: i + 1,
				Content: oldLines[i],
			})
		}

		// Process the actual change
		// difflib uses: 'r' for replace, 'd' for delete, 'i' for insert, 'e' for equal
		switch tag {
		case "r": // replace
			// Add removed lines
			for i := oldStart; i < oldEnd; i++ {
				hunk.Lines = append(hunk.Lines, DiffLine{
					Type:    "remove",
					LineNum: i + 1,
					Content: oldLines[i],
				})
			}
			// Add new lines
			for i := newStart; i < newEnd; i++ {
				hunk.Lines = append(hunk.Lines, DiffLine{
					Type:    "add",
					LineNum: i + 1,
					Content: newLines[i],
				})
			}

		case "d": // delete
			for i := oldStart; i < oldEnd; i++ {
				hunk.Lines = append(hunk.Lines, DiffLine{
					Type:    "remove",
					LineNum: i + 1,
					Content: oldLines[i],
				})
			}

		case "i": // insert
			for i := newStart; i < newEnd; i++ {
				hunk.Lines = append(hunk.Lines, DiffLine{
					Type:    "add",
					LineNum: i + 1,
					Content: newLines[i],
				})
			}
		}

		// Add context lines after change
		for i := oldEnd; i < hunkEnd && i < len(oldLines); i++ {
			hunk.Lines = append(hunk.Lines, DiffLine{
				Type:    "context",
				LineNum: i + 1,
				Content: oldLines[i],
			})
		}

		hunks = append(hunks, hunk)
	}

	// If no hunks generated, return empty slice
	if len(hunks) == 0 {
		return []DiffHunk{}
	}

	return hunks
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
