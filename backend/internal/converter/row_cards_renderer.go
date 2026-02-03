package converter

import (
	"fmt"
	"strings"
)

// RowCardsRenderer renders a Table as per-row card/section format
// Each row becomes a card with a title and configurable sections
// This is useful for test cases, UI specs, API endpoints, etc.
type RowCardsRenderer struct {
	template       *TemplateConfig
	headerResolver *HeaderResolver
	cardConfig     *RowCardsConfig
}

// NewRowCardsRenderer creates a new RowCardsRenderer with a template
func NewRowCardsRenderer(template *TemplateConfig) *RowCardsRenderer {
	return &RowCardsRenderer{
		template:       template,
		headerResolver: NewHeaderResolver(template),
		cardConfig:     template.Output.RowCards,
	}
}

// Render converts a Table to row cards markdown format
// Implements the Renderer interface
func (r *RowCardsRenderer) Render(table *Table) (markdown string, warnings []string, err error) {
	if table == nil {
		return "", nil, fmt.Errorf("table is nil")
	}

	if len(table.Headers) == 0 {
		return "", []string{}, fmt.Errorf("table has no headers")
	}

	if r.cardConfig == nil {
		return "", []string{}, fmt.Errorf("row cards config not set")
	}

	// Build header map for resolving header references
	headerMap := r.template.BuildHeaderMap()

	var buf strings.Builder

	// Write frontmatter
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("name: \"%s\"\n", table.SheetName))
	buf.WriteString("version: \"1.0\"\n")
	buf.WriteString("type: \"row_cards\"\n")
	buf.WriteString("---\n\n")

	// Title
	title := table.SheetName
	if title == "" {
		title = "Cards"
	}
	buf.WriteString(fmt.Sprintf("# %s\n\n", title))

	if len(table.Rows) == 0 {
		warnings = append(warnings, "Table has no data rows")
		return buf.String(), warnings, nil
	}

	// Resolve title field index
	titleIdx := r.cardConfig.TitleFrom.Resolve(table.Headers, headerMap)
	if titleIdx < 0 {
		return "", []string{fmt.Sprintf("Could not resolve title_from field: %v", r.cardConfig.TitleFrom)}, 
			fmt.Errorf("title field not found")
	}

	// Render each row as a card
	var rowWarnings []string
	for rowNum, row := range table.Rows {
		// Extract title
		title := ""
		if titleIdx >= 0 && titleIdx < len(row.Cells) {
			title = strings.TrimSpace(row.Cells[titleIdx])
		}
		if title == "" {
			title = fmt.Sprintf("Row %d", rowNum+1)
		}

		buf.WriteString(fmt.Sprintf("## %s\n\n", title))

		// Render sections
		hasContent := false
		for _, section := range r.cardConfig.Sections {
			idx := section.From.Resolve(table.Headers, headerMap)
			if idx >= 0 && idx < len(row.Cells) {
				content := strings.TrimSpace(row.Cells[idx])
				if content != "" && content != "-" {
					buf.WriteString(fmt.Sprintf("**%s:**\n%s\n\n", section.Label, content))
					hasContent = true
				}
			}
		}

		// Handle unmapped columns if configured
		if r.cardConfig.Extras.Mode == "append_section" {
			unmappedContent := r.collectUnmappedContent(row, table.Headers, headerMap)
			if unmappedContent != "" {
				label := r.cardConfig.Extras.Label
				if label == "" {
					label = "Additional Fields"
				}
				buf.WriteString(fmt.Sprintf("**%s:**\n%s\n\n", label, unmappedContent))
				hasContent = true
			}
		}

		if !hasContent {
			buf.WriteString("*(No content)*\n\n")
		}
	}

	return buf.String(), rowWarnings, nil
}

// collectUnmappedContent gathers unmapped column values for a row
func (r *RowCardsRenderer) collectUnmappedContent(row TableRow, headers []string, headerMap map[string]string) string {
	var usedIndices map[int]bool = make(map[int]bool)

	// Mark indices used in configured sections
	for _, section := range r.cardConfig.Sections {
		idx := section.From.Resolve(headers, headerMap)
		if idx >= 0 {
			usedIndices[idx] = true
		}
	}

	// Mark title field as used
	titleIdx := r.cardConfig.TitleFrom.Resolve(headers, headerMap)
	if titleIdx >= 0 {
		usedIndices[titleIdx] = true
	}

	// Collect unmapped columns
	var extras []string
	for i, header := range headers {
		if usedIndices[i] {
			continue // Skip already used columns
		}

		if i < len(row.Cells) {
			content := strings.TrimSpace(row.Cells[i])
			if content != "" && content != "-" {
				extras = append(extras, fmt.Sprintf("- %s: %s", header, content))
			}
		}
	}

	if len(extras) == 0 {
		return ""
	}

	return strings.Join(extras, "\n")
}

// GetTemplate returns the underlying template
func (r *RowCardsRenderer) GetTemplate() *TemplateConfig {
	return r.template
}

// GetHeaderResolver returns the underlying header resolver
func (r *RowCardsRenderer) GetHeaderResolver() *HeaderResolver {
	return r.headerResolver
}
