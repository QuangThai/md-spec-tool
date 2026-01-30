package converter

import (
	"regexp"
	"strconv"
	"strings"
)

// ProseContent represents markdown/prose content with extracted sections
type ProseContent struct {
	OriginalMarkdown string        `json:"original_markdown"`
	Sections         []ProseSection `json:"sections"`
	RawMessage       string        `json:"raw_message,omitempty"`
}

// ProseSection represents a section extracted from markdown
type ProseSection struct {
	Heading string `json:"heading"`
	Level   int    `json:"level"`   // 1-6 for h1-h6
	Content string `json:"content"`
}

// BuildMarkdownSpecDoc creates a SpecDoc from markdown/prose input
func BuildMarkdownSpecDoc(text string, title string) *SpecDoc {
	sections := extractSections(text)
	rawMessage := extractRawMessage(text)

	return &SpecDoc{
		Title: title,
		Rows: []SpecRow{}, // Empty for prose
		Warnings: []Warning{},
		Meta: SpecDocMeta{
			ColumnMap:       ColumnMap{},
			UnmappedColumns: []string{},
			TotalRows:       0,
		},
		Headers: []string{},
		Prose: &ProseContent{
			OriginalMarkdown: text,
			Sections:         sections,
			RawMessage:       rawMessage,
		},
	}
}

func buildRowsFromSections(sections []ProseSection) []SpecRow {
	rows := make([]SpecRow, 0, len(sections))
	for i, section := range sections {
		rows = append(rows, SpecRow{
			Feature:            section.Heading,
			Scenario:           section.Heading,
			Instructions:       section.Content,
			No:                 strconv.Itoa(i + 1),
			ItemName:           section.Heading,
			DisplayConditions:  section.Content,
		})
	}
	return rows
}

// extractSections extracts markdown sections from text
// Handles both `## Heading` and `> ## Heading` formats
func extractSections(text string) []ProseSection {
	var sections []ProseSection
	lines := strings.Split(text, "\n")
	boldHeadingRegex := regexp.MustCompile(`^\*\*([^*]+)\*\*\s*$`)

	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Remove blockquote prefix if present
		if strings.HasPrefix(trimmed, ">") {
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
		}

		// Check for heading
		if strings.HasPrefix(trimmed, "##") {
			// Extract heading level and text
			level := 0
			heading := trimmed
			for j := 0; j < len(trimmed); j++ {
				if trimmed[j] == '#' {
					level++
				} else {
					heading = strings.TrimSpace(trimmed[j:])
					break
				}
			}

			// Collect content lines until next heading or end
			contentLines := []string{}
			i++
			for i < len(lines) {
				nextLine := lines[i]
				nextTrimmed := strings.TrimSpace(nextLine)
				
				// Remove blockquote prefix if present and save clean line
				if strings.HasPrefix(nextTrimmed, ">") {
					nextTrimmed = strings.TrimSpace(strings.TrimPrefix(nextTrimmed, ">"))
				}

				// Check if we hit another heading
				if strings.HasPrefix(nextTrimmed, "##") {
					break
				}

				// Check if we hit raw message marker
				if strings.HasPrefix(nextTrimmed, "Raw message:") {
					break
				}

				// Only add non-empty lines
				if strings.TrimSpace(nextTrimmed) != "" {
					contentLines = append(contentLines, strings.TrimSpace(nextTrimmed))
				}
				i++
			}

			// Join content, trimming empty lines at start and end
			content := strings.TrimSpace(strings.Join(contentLines, "\n"))

			sections = append(sections, ProseSection{
				Heading: heading,
				Level:   level,
				Content: content,
			})
			continue
		}

		// Check for bold heading lines (e.g. **Summary**)
		if matches := boldHeadingRegex.FindStringSubmatch(trimmed); len(matches) == 2 {
			heading := strings.TrimSpace(matches[1])
			level := 2
			contentLines := []string{}
			i++
			for i < len(lines) {
				nextLine := lines[i]
				nextTrimmed := strings.TrimSpace(nextLine)
				if strings.HasPrefix(nextTrimmed, ">") {
					nextTrimmed = strings.TrimSpace(strings.TrimPrefix(nextTrimmed, ">"))
				}
				if strings.HasPrefix(nextTrimmed, "##") || boldHeadingRegex.MatchString(nextTrimmed) {
					break
				}
				if strings.HasPrefix(nextTrimmed, "Raw message:") {
					break
				}
				if strings.TrimSpace(nextTrimmed) != "" {
					contentLines = append(contentLines, strings.TrimSpace(nextTrimmed))
				}
				i++
			}
			content := strings.TrimSpace(strings.Join(contentLines, "\n"))
			sections = append(sections, ProseSection{
				Heading: heading,
				Level:   level,
				Content: content,
			})
			continue
		}

		i++
	}

	return sections
}

// extractRawMessage extracts content after "Raw message:" marker
func extractRawMessage(text string) string {
	// Find the Raw message: marker
	rawIdx := strings.Index(strings.ToLower(text), "raw message:")
	if rawIdx == -1 {
		return ""
	}

	// Get everything after the marker
	afterMarker := text[rawIdx+len("raw message:"):]
	content := strings.TrimSpace(afterMarker)

	// Clean up blockquote prefixes if present
	lines := strings.Split(content, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ">") {
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
		}
		if trimmed != "" {
			cleanedLines = append(cleanedLines, trimmed)
		}
	}

	return strings.Join(cleanedLines, "\n")
}

// containsMarkdownSignals checks if text contains markdown indicators
func containsMarkdownSignals(text string) bool {
	signals := []string{
		"^#",      // headings
		"^>",      // blockquotes
		"^```",    // code fences
		"^- ",     // bullet lists
		"^\\* ",   // bullet lists
		"^[0-9]+\\.", // numbered lists
	}

	for _, signal := range signals {
		re := regexp.MustCompile(signal + "(?m)")
		if re.MatchString(text) {
			return true
		}
	}
	return false
}
