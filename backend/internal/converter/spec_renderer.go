package converter

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// SpecRenderer renders data to AGENTS.md compatible spec format
type SpecRenderer struct{}

// NewSpecRenderer creates a new SpecRenderer
func NewSpecRenderer() *SpecRenderer {
	return &SpecRenderer{}
}

// SpecRenderInput holds data for spec rendering
type SpecRenderInput struct {
	Title           string
	Rows            []SpecRow
	Headers         []string
	ColumnMappings  map[string]interface{} // AI mapping metadata
	AIMode          string                 // "on", "shadow", "off"
	Degraded        bool                   // True if fallback was used
	SchemaVersion   string
	AvgConfidence   float64
	MappedColumns   int
	UnmappedColumns int
}

// Render implements Renderer interface for Table format
// Converts Table to AGENTS.md spec format
func (r *SpecRenderer) Render(table *Table) (string, []string, error) {
	if table == nil {
		return "", nil, fmt.Errorf("table is nil")
	}

	// Convert Table to SpecRows
	rows := r.tableToSpecRows(table)

	title := table.SheetName
	if title == "" {
		title = "Converted Specification"
	}

	aiMode := table.Meta.AIMode
	if aiMode == "" {
		aiMode = "off"
	}
	mappedColumns := table.Meta.AIMappedColumns
	unmappedColumns := table.Meta.AIUnmappedColumns
	if mappedColumns == 0 && unmappedColumns == 0 {
		mappedColumns = len(table.Headers)
	}

	input := SpecRenderInput{
		Title:           title,
		Rows:            rows,
		Headers:         table.Headers,
		SchemaVersion:   "v1",
		AIMode:          aiMode,
		Degraded:        table.Meta.AIDegraded,
		AvgConfidence:   table.Meta.AIAvgConfidence,
		MappedColumns:   mappedColumns,
		UnmappedColumns: unmappedColumns,
	}

	output := r.renderSpec(input)
	return output, []string{}, nil
}

// renderSpec is the internal render implementation
func (r *SpecRenderer) renderSpec(input SpecRenderInput) string {
	var buf bytes.Buffer

	// YAML front matter (AGENTS.md compatible)
	buf.WriteString(r.renderFrontMatter(input))

	// Title & metadata
	buf.WriteString(fmt.Sprintf("# %s\n\n", input.Title))

	// Summary section
	buf.WriteString(r.renderSummary(input))

	// Column mappings section (if available)
	if len(input.ColumnMappings) > 0 {
		buf.WriteString(r.renderMappingsSummary(input))
	}

	// Specifications section (grouped by feature/category)
	buf.WriteString(r.renderSpecifications(input))

	return buf.String()
}

// tableToSpecRows converts a Table to SpecRow array
// Maps table columns to SpecRow fields using HeaderSynonyms
func (r *SpecRenderer) tableToSpecRows(table *Table) []SpecRow {
	var rows []SpecRow
	useColumnMap := len(table.Meta.ColumnMap) > 0
	var mappedIndices map[int]bool
	if useColumnMap {
		mappedIndices = make(map[int]bool, len(table.Meta.ColumnMap))
		for _, idx := range table.Meta.ColumnMap {
			mappedIndices[idx] = true
		}
	}

	for _, tableRow := range table.Rows {
		row := SpecRow{
			Metadata: make(map[string]string),
		}

		if useColumnMap {
			for field, idx := range table.Meta.ColumnMap {
				if idx < 0 || idx >= len(tableRow.Cells) {
					continue
				}
				cell := tableRow.Cells[idx]
				switch field {
				case FieldID:
					row.ID = cell
				case FieldFeature:
					row.Feature = cell
				case FieldScenario:
					row.Scenario = cell
				case FieldInstructions:
					row.Instructions = cell
				case FieldInputs:
					row.Inputs = cell
				case FieldExpected:
					row.Expected = cell
				case FieldPrecondition:
					row.Precondition = cell
				case FieldPriority:
					row.Priority = cell
				case FieldType:
					row.Type = cell
				case FieldStatus:
					row.Status = cell
				case FieldEndpoint:
					row.Endpoint = cell
				case FieldNotes:
					row.Notes = cell
				case FieldNo:
					row.No = cell
				case FieldItemName:
					row.ItemName = cell
				case FieldItemType:
					row.ItemType = cell
				case FieldRequiredOptional:
					row.RequiredOptional = cell
				case FieldInputRestrictions:
					row.InputRestrictions = cell
				case FieldDisplayConditions:
					row.DisplayConditions = cell
				case FieldAction:
					row.Action = cell
				case FieldNavigationDest:
					row.NavigationDest = cell
				}
			}

			for i, header := range table.Headers {
				if i >= len(tableRow.Cells) {
					break
				}
				if mappedIndices[i] {
					continue
				}
				if tableRow.Cells[i] == "" {
					continue
				}
				row.Metadata[header] = tableRow.Cells[i]
			}
		} else {
			for i, cell := range tableRow.Cells {
				if i >= len(table.Headers) {
					break
				}

				header := table.Headers[i]
				normalized := normalizeHeader(header)

				// Try to map to canonical field using HeaderSynonyms
				if field, ok := HeaderSynonyms[normalized]; ok {
					switch field {
					case FieldID:
						row.ID = cell
					case FieldFeature:
						row.Feature = cell
					case FieldScenario:
						row.Scenario = cell
					case FieldInstructions:
						row.Instructions = cell
					case FieldInputs:
						row.Inputs = cell
					case FieldExpected:
						row.Expected = cell
					case FieldPrecondition:
						row.Precondition = cell
					case FieldPriority:
						row.Priority = cell
					case FieldType:
						row.Type = cell
					case FieldStatus:
						row.Status = cell
					case FieldEndpoint:
						row.Endpoint = cell
					case FieldNotes:
						row.Notes = cell
					case FieldNo:
						row.No = cell
					case FieldItemName:
						row.ItemName = cell
					case FieldItemType:
						row.ItemType = cell
					case FieldRequiredOptional:
						row.RequiredOptional = cell
					case FieldInputRestrictions:
						row.InputRestrictions = cell
					case FieldDisplayConditions:
						row.DisplayConditions = cell
					case FieldAction:
						row.Action = cell
					case FieldNavigationDest:
						row.NavigationDest = cell
					}
				} else {
					// Unmapped columns go to metadata
					row.Metadata[header] = cell
				}
			}
		}

		rows = append(rows, row)
	}

	return rows
}

// renderFrontMatter outputs YAML front matter
func (r *SpecRenderer) renderFrontMatter(input SpecRenderInput) string {
	var buf bytes.Buffer

	buf.WriteString("---\n")
	buf.WriteString("name: \"Specification\"\n")
	buf.WriteString("version: \"1.0\"\n")
	buf.WriteString(fmt.Sprintf("generated_at: \"%s\"\n", time.Now().Format("2006-01-02")))
	buf.WriteString("type: \"specification\"\n")

	buf.WriteString("---\n\n")

	return buf.String()
}

// renderSummary outputs metadata summary table
func (r *SpecRenderer) renderSummary(input SpecRenderInput) string {
	var buf bytes.Buffer

	buf.WriteString("## Summary\n\n")
	buf.WriteString("| Metric | Value |\n")
	buf.WriteString("|--------|-------|\n")
	buf.WriteString(fmt.Sprintf("| Total Items | %d |\n", len(input.Rows)))
	buf.WriteString(fmt.Sprintf("| Mapped Columns | %d |\n", input.MappedColumns))
	buf.WriteString(fmt.Sprintf("| Extra Columns | %d |\n", input.UnmappedColumns))

	confidencePercent := int(input.AvgConfidence * 100.0)
	buf.WriteString(fmt.Sprintf("| Avg Confidence | %d%% |\n", confidencePercent))

	if input.Degraded {
		buf.WriteString("| Status | ⚠️ Degraded (fallback mapper used) |\n")
	}

	buf.WriteString("\n")

	return buf.String()
}

// renderMappingsSummary outputs column mapping details
func (r *SpecRenderer) renderMappingsSummary(input SpecRenderInput) string {
	var buf bytes.Buffer

	buf.WriteString("## Column Mappings\n\n")
	buf.WriteString("| Original Header | Mapped To | Confidence |\n")
	buf.WriteString("|-----------------|-----------|------------|\n")

	// For now, just show headers if available
	// In full implementation, would iterate through canonicalFields + extraColumns from input.ColumnMappings
	if len(input.Headers) > 0 {
		for i, h := range input.Headers {
			if i < len(input.Rows) && input.Rows[i].Feature != "" {
				// Simple heuristic: map first header to feature
				confidence := input.AvgConfidence
				buf.WriteString(fmt.Sprintf("| %s | feature | %.0f%% |\n",
					escapeTableCell(h), confidence*100))
			}
		}
	}

	buf.WriteString("\n")

	return buf.String()
}

// renderSpecifications renders all specifications grouped by feature/category
func (r *SpecRenderer) renderSpecifications(input SpecRenderInput) string {
	var buf bytes.Buffer

	buf.WriteString("## Specifications\n\n")

	// Group by feature
	featureGroups := groupRowsByFeature(input.Rows)

	for _, feature := range getSortedFeatures(featureGroups) {
		rows := featureGroups[feature]

		buf.WriteString(fmt.Sprintf("### %s\n\n", feature))

		for _, row := range rows {
			buf.WriteString(r.renderSpecItem(row))
			buf.WriteString("\n")
		}
	}

	return buf.String()
}

// renderSpecItem renders a single specification item in AGENTS.md style
func (r *SpecRenderer) renderSpecItem(row SpecRow) string {
	var buf bytes.Buffer

	// Item header - try multiple sources
	title := row.Scenario
	if title == "" {
		title = row.Feature
	}
	if title == "" {
		// Fallback for spec-table style rows (No, ItemName, ItemType)
		title = row.ItemName
	}
	if title == "" && row.No != "" {
		// Last resort: use No if nothing else available
		title = row.No
	}

	if row.ID != "" {
		title = fmt.Sprintf("%s: %s", row.ID, title)
	}

	buf.WriteString(fmt.Sprintf("#### %s\n\n", title))

	// Metadata table
	var metaRows [][]string
	if row.ID != "" {
		metaRows = append(metaRows, []string{"id", row.ID})
	}
	if row.Type != "" {
		metaRows = append(metaRows, []string{"type", row.Type})
	}
	if row.Priority != "" {
		metaRows = append(metaRows, []string{"priority", row.Priority})
	}
	if row.Status != "" {
		metaRows = append(metaRows, []string{"status", row.Status})
	}

	if len(metaRows) > 0 {
		buf.WriteString("| Field | Value |\n")
		buf.WriteString("|-------|-------|\n")
		for _, m := range metaRows {
			buf.WriteString(fmt.Sprintf("| %s | %s |\n", m[0], escapeTableCell(m[1])))
		}
		buf.WriteString("\n")
	}

	// Description
	if row.Scenario != "" && row.Scenario != row.Feature {
		buf.WriteString("**Description:**\n")
		buf.WriteString(fmt.Sprintf("%s\n\n", row.Scenario))
	}

	// Precondition
	if row.Precondition != "" {
		buf.WriteString("**Precondition:**\n")
		buf.WriteString(fmt.Sprintf("%s\n\n", formatAsMultiLine(row.Precondition)))
	}

	// Steps/Instructions
	if row.Instructions != "" {
		buf.WriteString("**Steps:**\n")
		buf.WriteString(formatAsNumberedList(row.Instructions))
		buf.WriteString("\n")
	}

	// Expected Result
	if row.Expected != "" {
		buf.WriteString("**Expected Result:**\n")
		buf.WriteString(fmt.Sprintf("%s\n\n", formatAsMultiLine(row.Expected)))
	}

	// Inputs/Test Data
	if row.Inputs != "" {
		buf.WriteString("**Test Data:**\n")
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", row.Inputs))
	}

	// Endpoint
	if row.Endpoint != "" {
		buf.WriteString("**API/Endpoint:**\n")
		buf.WriteString(fmt.Sprintf("`%s`\n\n", row.Endpoint))
	}

	// Extra spec table fields
	extraFields := r.renderExtraFields(row)
	if extraFields != "" {
		buf.WriteString(extraFields)
	}

	// Notes
	if row.Notes != "" {
		buf.WriteString("**Notes:**\n")
		buf.WriteString(fmt.Sprintf("%s\n\n", row.Notes))
	}

	// Metadata
	if len(row.Metadata) > 0 {
		buf.WriteString("**Additional Fields:**\n")
		for k, v := range row.Metadata {
			buf.WriteString(fmt.Sprintf("- **%s**: %s\n", k, escapeMarkdown(v)))
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

// renderExtraFields renders Phase 3 spec table fields if present
func (r *SpecRenderer) renderExtraFields(row SpecRow) string {
	// Only output if at least one field is set
	if row.No == "" && row.ItemName == "" && row.ItemType == "" &&
		row.RequiredOptional == "" && row.InputRestrictions == "" &&
		row.DisplayConditions == "" && row.Action == "" &&
		row.NavigationDest == "" {
		return ""
	}

	var buf bytes.Buffer
	buf.WriteString("**Field Specification:**\n")

	if row.No != "" {
		buf.WriteString(fmt.Sprintf("- **No**: %s\n", row.No))
	}
	if row.ItemName != "" {
		buf.WriteString(fmt.Sprintf("- **Item Name**: %s\n", escapeMarkdown(row.ItemName)))
	}
	if row.ItemType != "" {
		buf.WriteString(fmt.Sprintf("- **Item Type**: %s\n", row.ItemType))
	}
	if row.RequiredOptional != "" {
		buf.WriteString(fmt.Sprintf("- **Required/Optional**: %s\n", row.RequiredOptional))
	}
	if row.InputRestrictions != "" {
		buf.WriteString(fmt.Sprintf("- **Input Restrictions**: %s\n", escapeMarkdown(row.InputRestrictions)))
	}
	if row.DisplayConditions != "" {
		buf.WriteString(fmt.Sprintf("- **Display Conditions**: %s\n", escapeMarkdown(row.DisplayConditions)))
	}
	if row.Action != "" {
		buf.WriteString(fmt.Sprintf("- **Action**: %s\n", escapeMarkdown(row.Action)))
	}
	if row.NavigationDest != "" {
		buf.WriteString(fmt.Sprintf("- **Navigation Destination**: %s\n", row.NavigationDest))
	}

	buf.WriteString("\n")

	return buf.String()
}

// Helper functions

func groupRowsByFeature(rows []SpecRow) map[string][]SpecRow {
	groups := make(map[string][]SpecRow)
	for _, row := range rows {
		feature := row.Feature
		if feature == "" {
			feature = "Uncategorized"
		}
		groups[feature] = append(groups[feature], row)
	}
	return groups
}

func getSortedFeatures(groups map[string][]SpecRow) []string {
	var features []string
	for f := range groups {
		features = append(features, f)
	}
	// Simple sort (stable for consistent output)
	for i := 0; i < len(features)-1; i++ {
		for j := i + 1; j < len(features); j++ {
			if features[j] < features[i] {
				features[i], features[j] = features[j], features[i]
			}
		}
	}
	return features
}

func formatAsMultiLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && trimmed != "-" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return ""
	}
	return strings.Join(result, "\n")
}

func formatAsNumberedList(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	var result []string
	num := 1
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "-" {
			continue
		}
		// Remove existing numbering if present
		trimmed = strings.TrimPrefix(trimmed, "- ")
		for trimmed != "" && trimmed[0] >= '0' && trimmed[0] <= '9' {
			// Skip leading digits and punctuation
			if len(trimmed) > 1 && (trimmed[1] == '.' || trimmed[1] == ')') {
				trimmed = strings.TrimSpace(trimmed[2:])
				break
			}
			trimmed = trimmed[1:]
		}
		if trimmed != "" {
			result = append(result, fmt.Sprintf("%d. %s", num, trimmed))
			num++
		}
	}
	if len(result) == 0 {
		return ""
	}
	return strings.Join(result, "\n") + "\n"
}

func escapeYAMLValue(s string) string {
	if strings.Contains(s, ":") || strings.Contains(s, "#") || strings.Contains(s, "\n") {
		// Use quoted string for safety
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		return fmt.Sprintf(`"%s"`, s)
	}
	return s
}

func escapeTableCell(s string) string {
	// Escape pipe and newline characters
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}

func escapeMarkdown(s string) string {
	// Escape markdown special characters
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `*`, `\*`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	s = strings.ReplaceAll(s, `[`, `\[`)
	s = strings.ReplaceAll(s, `]`, `\]`)
	return s
}
