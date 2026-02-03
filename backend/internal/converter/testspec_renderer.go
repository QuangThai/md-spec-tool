package converter

import (
	"fmt"
	"strings"
)

// TestSpecRenderer renders a Table to legacy test specification Markdown format
// using configuration-driven header mappings instead of hardcoded synonyms.
// It replaces the hardcoded logic from column_map.go and integrates with the
// template-based approach (Phase 3).
type TestSpecRenderer struct {
	template       *TemplateConfig
	headerResolver *HeaderResolver
}

// NewTestSpecRenderer creates a new TestSpecRenderer with a template
func NewTestSpecRenderer(template *TemplateConfig) *TestSpecRenderer {
	return &TestSpecRenderer{
		template:       template,
		headerResolver: NewHeaderResolver(template),
	}
}

// Render converts a Table to legacy test spec Markdown format
// Implements the Renderer interface
func (r *TestSpecRenderer) Render(table *Table) (markdown string, warnings []string, err error) {
	if table == nil {
		return "", nil, fmt.Errorf("table is nil")
	}

	if len(table.Headers) == 0 {
		return "", []string{}, fmt.Errorf("table has no headers")
	}

	// Resolve headers to canonical fields using template
	colMap, unmapped, resolveWarnings := r.headerResolver.ResolveHeaders(table.Headers)
	warnings = append(warnings, resolveWarnings...)

	if len(unmapped) > 0 {
		unmappedStr := strings.Join(unmapped, ", ")
		warnings = append(warnings, fmt.Sprintf("Unmapped columns: %s", unmappedStr))
	}

	// Build the SpecDoc-like structure from the Table
	specRows := r.convertTableToSpecRows(table, colMap, unmapped, table.Headers)

	// Render to markdown using the default MDFlow template
	// For now, we'll generate a simple markdown output
	// In a full implementation, this would use the template registry's rendering logic
	return r.renderSpecRows(specRows, table.SheetName, table.Headers, colMap), warnings, nil
}

// convertTableToSpecRows converts Table rows to SpecRow format using column mapping
func (r *TestSpecRenderer) convertTableToSpecRows(table *Table, colMap ColumnMap, unmapped []string, headers []string) []SpecRow {
	var specRows []SpecRow

	for _, tableRow := range table.Rows {
		specRow := SpecRow{
			ID:                r.getFieldValue(tableRow.Cells, colMap, FieldID),
			Feature:           r.getFieldValue(tableRow.Cells, colMap, FieldFeature),
			Scenario:          r.getFieldValue(tableRow.Cells, colMap, FieldScenario),
			Instructions:      r.getFieldValue(tableRow.Cells, colMap, FieldInstructions),
			Inputs:            r.getFieldValue(tableRow.Cells, colMap, FieldInputs),
			Expected:          r.getFieldValue(tableRow.Cells, colMap, FieldExpected),
			Precondition:      r.getFieldValue(tableRow.Cells, colMap, FieldPrecondition),
			Priority:          r.getFieldValue(tableRow.Cells, colMap, FieldPriority),
			Type:              r.getFieldValue(tableRow.Cells, colMap, FieldType),
			Status:            r.getFieldValue(tableRow.Cells, colMap, FieldStatus),
			Endpoint:          r.getFieldValue(tableRow.Cells, colMap, FieldEndpoint),
			Notes:             r.getFieldValue(tableRow.Cells, colMap, FieldNotes),
			No:                r.getFieldValue(tableRow.Cells, colMap, FieldNo),
			ItemName:          r.getFieldValue(tableRow.Cells, colMap, FieldItemName),
			ItemType:          r.getFieldValue(tableRow.Cells, colMap, FieldItemType),
			RequiredOptional:  r.getFieldValue(tableRow.Cells, colMap, FieldRequiredOptional),
			InputRestrictions: r.getFieldValue(tableRow.Cells, colMap, FieldInputRestrictions),
			DisplayConditions: r.getFieldValue(tableRow.Cells, colMap, FieldDisplayConditions),
			Action:            r.getFieldValue(tableRow.Cells, colMap, FieldAction),
			NavigationDest:    r.getFieldValue(tableRow.Cells, colMap, FieldNavigationDest),
			Metadata:          make(map[string]string),
		}

		// Store unmapped columns in metadata
		for _, unmappedHeader := range unmapped {
			idx := r.findHeaderIndex(unmappedHeader, headers)
			if idx >= 0 && idx < len(tableRow.Cells) && tableRow.Cells[idx] != "" {
				specRow.Metadata[unmappedHeader] = tableRow.Cells[idx]
			}
		}

		specRows = append(specRows, specRow)
	}

	return specRows
}

// getFieldValue extracts a normalized field value
func (r *TestSpecRenderer) getFieldValue(cells []string, colMap ColumnMap, field CanonicalField) string {
	if idx, ok := colMap[field]; ok && idx < len(cells) {
		val := cells[idx]
		// Normalize (trim spaces, handle dashes as empty)
		val = strings.TrimSpace(val)
		if val == "-" {
			return ""
		}
		return val
	}
	return ""
}

// findHeaderIndex finds the index of a header by name
func (r *TestSpecRenderer) findHeaderIndex(headerName string, headers []string) int {
	for i, h := range headers {
		if h == headerName {
			return i
		}
	}
	return -1
}

// renderSpecRows generates markdown from SpecRow data
// This is a simplified version that matches the default MDFlow output
func (r *TestSpecRenderer) renderSpecRows(rows []SpecRow, sheetName string, headers []string, colMap ColumnMap) string {
	var buf strings.Builder

	// Title
	title := sheetName
	if title == "" {
		title = "Test Specification"
	}
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("name: \"%s\"\n", title))
	buf.WriteString("version: \"1.0\"\n")
	buf.WriteString("type: \"test_spec_markdown\"\n")
	buf.WriteString("---\n\n")
	buf.WriteString(fmt.Sprintf("# %s\n\n", title))

	// Summary table
	buf.WriteString(fmt.Sprintf("**Total Test Cases:** %d\n\n", len(rows)))

	// Group by feature if present
	if r.hasFeatures(rows) {
		r.renderGroupedByFeature(&buf, rows)
	} else if r.hasItemNames(rows) {
		r.renderSpecTable(&buf, rows, headers, colMap)
	} else {
		r.renderSimpleList(&buf, rows)
	}

	return buf.String()
}

// hasFeatures checks if any row has a Feature field
func (r *TestSpecRenderer) hasFeatures(rows []SpecRow) bool {
	for _, row := range rows {
		if row.Feature != "" && row.Feature != "Uncategorized" {
			return true
		}
	}
	return false
}

// hasItemNames checks if any row has ItemName (spec table style)
func (r *TestSpecRenderer) hasItemNames(rows []SpecRow) bool {
	for _, row := range rows {
		if row.ItemName != "" {
			return true
		}
	}
	return false
}

// renderGroupedByFeature renders rows grouped by Feature
func (r *TestSpecRenderer) renderGroupedByFeature(buf *strings.Builder, rows []SpecRow) {
	groupMap := make(map[string][]SpecRow)
	var order []string

	for _, row := range rows {
		feature := row.Feature
		if feature == "" {
			feature = "Uncategorized"
		}
		if _, exists := groupMap[feature]; !exists {
			order = append(order, feature)
		}
		groupMap[feature] = append(groupMap[feature], row)
	}

	for _, feature := range order {
		buf.WriteString(fmt.Sprintf("## %s\n\n", feature))
		for _, row := range groupMap[feature] {
			r.renderRow(buf, row)
		}
	}
}

// renderSpecTable renders as a spec table (for UI/UX specs with No, ItemName, etc.)
func (r *TestSpecRenderer) renderSpecTable(buf *strings.Builder, rows []SpecRow, _ []string, _ ColumnMap) {
	buf.WriteString("## Summary\n\n")
	buf.WriteString("| No | Item Name | Item Type | Required | Display Conditions | Input Restrictions | Action | Navigation |\n")
	buf.WriteString("|----|-----------|-----------|---------|--------------------|-------------------|--------|------------|\n")

	for _, row := range rows {
		buf.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s | %s |\n",
			row.No, row.ItemName, row.ItemType, row.RequiredOptional,
			row.DisplayConditions, row.InputRestrictions, row.Action, row.NavigationDest))
	}

	buf.WriteString("\n## Item Details\n\n")
	for _, row := range rows {
		if row.ItemName != "" {
			buf.WriteString(fmt.Sprintf("### %s. %s\n\n", row.No, row.ItemName))
			if row.ItemType != "" {
				buf.WriteString(fmt.Sprintf("**Type:** %s\n\n", row.ItemType))
			}
			if row.RequiredOptional != "" {
				buf.WriteString(fmt.Sprintf("**Required:** %s\n\n", row.RequiredOptional))
			}
			if row.DisplayConditions != "" {
				buf.WriteString(fmt.Sprintf("**Display Conditions:**\n%s\n\n", row.DisplayConditions))
			}
			if row.InputRestrictions != "" {
				buf.WriteString(fmt.Sprintf("**Input Restrictions:**\n%s\n\n", row.InputRestrictions))
			}
			if row.Action != "" {
				buf.WriteString(fmt.Sprintf("**Action:** %s\n\n", row.Action))
			}
			if row.NavigationDest != "" {
				buf.WriteString(fmt.Sprintf("**Navigation:** %s\n\n", row.NavigationDest))
			}
			if row.Notes != "" {
				buf.WriteString(fmt.Sprintf("**Notes:** %s\n\n", row.Notes))
			}
			// Render unmapped/custom columns
			if len(row.Metadata) > 0 {
				buf.WriteString("**Additional Fields:**\n")
				for key, val := range row.Metadata {
					buf.WriteString(fmt.Sprintf("- %s: %s\n", key, val))
				}
				buf.WriteString("\n")
			}
		}
	}
}

// renderSimpleList renders rows as a simple list
func (r *TestSpecRenderer) renderSimpleList(buf *strings.Builder, rows []SpecRow) {
	for i, row := range rows {
		buf.WriteString(fmt.Sprintf("## Row %d\n\n", i+1))
		r.renderRow(buf, row)
	}
}

// renderRow renders a single SpecRow
func (r *TestSpecRenderer) renderRow(buf *strings.Builder, row SpecRow) {
	if row.Scenario != "" {
		buf.WriteString(fmt.Sprintf("**Scenario:** %s\n\n", row.Scenario))
	}
	if row.Instructions != "" {
		buf.WriteString(fmt.Sprintf("**Instructions:**\n%s\n\n", row.Instructions))
	}
	if row.Expected != "" {
		buf.WriteString(fmt.Sprintf("**Expected:**\n%s\n\n", row.Expected))
	}
	if row.Inputs != "" {
		buf.WriteString(fmt.Sprintf("**Inputs:**\n```\n%s\n```\n\n", row.Inputs))
	}
	if row.Notes != "" {
		buf.WriteString(fmt.Sprintf("**Notes:** %s\n\n", row.Notes))
	}
}

// GetTemplate returns the underlying template
func (r *TestSpecRenderer) GetTemplate() *TemplateConfig {
	return r.template
}

// GetHeaderResolver returns the underlying header resolver
func (r *TestSpecRenderer) GetHeaderResolver() *HeaderResolver {
	return r.headerResolver
}
