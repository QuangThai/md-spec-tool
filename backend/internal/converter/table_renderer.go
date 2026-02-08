package converter

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// TableRenderer renders data to simple markdown table format
type TableRenderer struct{}

// NewTableRenderer creates a new TableRenderer
func NewTableRenderer() *TableRenderer {
	return &TableRenderer{}
}

// TableRenderInput holds data for table rendering
type TableRenderInput struct {
	Title   string
	Headers []string
	Rows    []SpecRow
	// If headers not provided, uses row field names
}

// Render implements Renderer interface
// Output is a simple markdown table (columns as-is, no mapping to canonical fields)
func (r *TableRenderer) Render(table *Table) (string, []string, error) {
	if table == nil {
		return "", nil, fmt.Errorf("table is nil")
	}

	title := table.SheetName
	if title == "" {
		title = "Data Table"
	}

	input := TableRenderInput{
		Title:   title,
		Headers: table.Headers,
		Rows:    r.tableToSpecRows(table),
	}

	output := r.renderTable(input)
	return output, []string{}, nil
}

// renderTable is the internal render implementation
func (r *TableRenderer) renderTable(input TableRenderInput) string {
	var buf bytes.Buffer

	buf.WriteString("---\n")
	buf.WriteString("name: \"Specification\"\n")
	buf.WriteString("version: \"1.0\"\n")
	buf.WriteString(fmt.Sprintf("generated_at: \"%s\"\n", time.Now().Format("2006-01-02")))
	buf.WriteString("type: \"specification\"\n")
	buf.WriteString("---\n\n")

	// Title
	if input.Title != "" {
		buf.WriteString(fmt.Sprintf("# %s\n\n", input.Title))
	}

	// Use provided headers or infer from row structure
	headers := input.Headers
	if len(headers) == 0 {
		headers = r.inferHeaders(input.Rows)
	}

	if len(headers) == 0 || len(input.Rows) == 0 {
		buf.WriteString("No data to render.\n")
		return buf.String()
	}

	// Header row
	buf.WriteString("|")
	for _, h := range headers {
		buf.WriteString(fmt.Sprintf(" %s |", escapeTableCell(h)))
	}
	buf.WriteString("\n")

	// Separator row
	buf.WriteString("|")
	for range headers {
		buf.WriteString(" --- |")
	}
	buf.WriteString("\n")

	// Data rows
	for _, row := range input.Rows {
		buf.WriteString("|")
		for _, h := range headers {
			val := r.getCellValue(row, h)
			buf.WriteString(fmt.Sprintf(" %s |", val))
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

// tableToSpecRows converts a Table to SpecRow array
func (r *TableRenderer) tableToSpecRows(table *Table) []SpecRow {
	var rows []SpecRow

	for _, tableRow := range table.Rows {
		row := SpecRow{
			Metadata: make(map[string]string),
		}

		for i, cell := range tableRow.Cells {
			if i >= len(table.Headers) {
				break
			}

			header := table.Headers[i]
			normalized := strings.ToLower(strings.TrimSpace(header))

			// Try to map using basic normalization
			switch normalized {
			case "id":
				row.ID = cell
			case "feature":
				row.Feature = cell
			case "scenario":
				row.Scenario = cell
			case "type":
				row.Type = cell
			case "priority":
				row.Priority = cell
			case "status":
				row.Status = cell
			case "instructions", "steps":
				row.Instructions = cell
			case "expected", "expected_result":
				row.Expected = cell
			case "precondition", "preconditions":
				row.Precondition = cell
			case "inputs", "test_data":
				row.Inputs = cell
			case "endpoint":
				row.Endpoint = cell
			case "notes":
				row.Notes = cell
			case "no":
				row.No = cell
			case "item_name", "name":
				row.ItemName = cell
			case "item_type", "item type":
				row.ItemType = cell
			case "required_optional", "required":
				row.RequiredOptional = cell
			case "input_restrictions", "restrictions":
				row.InputRestrictions = cell
			case "display_conditions", "display":
				row.DisplayConditions = cell
			case "action":
				row.Action = cell
			case "navigation_destination", "navigation":
				row.NavigationDest = cell
			default:
				// Unmapped columns go to metadata
				row.Metadata[header] = cell
			}
		}

		rows = append(rows, row)
	}

	return rows
}

// inferHeaders attempts to determine available columns from rows
// Returns headers in order of commonality
func (r *TableRenderer) inferHeaders(rows []SpecRow) []string {
	if len(rows) == 0 {
		return nil
	}

	// Count which fields are populated
	fieldCounts := make(map[string]int)
	for _, row := range rows {
		if row.ID != "" {
			fieldCounts["id"]++
		}
		if row.Feature != "" {
			fieldCounts["feature"]++
		}
		if row.Scenario != "" {
			fieldCounts["scenario"]++
		}
		if row.Type != "" {
			fieldCounts["type"]++
		}
		if row.Priority != "" {
			fieldCounts["priority"]++
		}
		if row.Status != "" {
			fieldCounts["status"]++
		}
		if row.Instructions != "" {
			fieldCounts["instructions"]++
		}
		if row.Expected != "" {
			fieldCounts["expected"]++
		}
		if row.Precondition != "" {
			fieldCounts["precondition"]++
		}
		if row.Notes != "" {
			fieldCounts["notes"]++
		}
		// Phase 3 fields
		if row.ItemName != "" {
			fieldCounts["item_name"]++
		}
		if row.ItemType != "" {
			fieldCounts["item_type"]++
		}
		if row.RequiredOptional != "" {
			fieldCounts["required_optional"]++
		}
		if row.Action != "" {
			fieldCounts["action"]++
		}
	}

	// Return fields that have at least one value, in preferred order
	// Note: removed duplicate "notes" from original list
	preferred := []string{"id", "feature", "scenario", "type", "priority", "status",
		"instructions", "expected", "precondition", "notes",
		"item_name", "item_type", "required_optional", "action"}

	var headers []string
	for _, field := range preferred {
		if fieldCounts[field] > 0 {
			headers = append(headers, field)
		}
	}

	return headers
}

// getCellValue returns the value for a row cell by header name
func (r *TableRenderer) getCellValue(row SpecRow, header string) string {
	normalized := strings.ToLower(strings.TrimSpace(header))

	val := ""
	switch normalized {
	case "id":
		val = row.ID
	case "feature":
		val = row.Feature
	case "scenario":
		val = row.Scenario
	case "type":
		val = row.Type
	case "priority":
		val = row.Priority
	case "status":
		val = row.Status
	case "instructions", "steps":
		val = row.Instructions
	case "expected", "expected_result":
		val = row.Expected
	case "precondition", "preconditions":
		val = row.Precondition
	case "inputs", "test_data":
		val = row.Inputs
	case "endpoint":
		val = row.Endpoint
	case "notes":
		val = row.Notes
	case "no":
		val = row.No
	case "item_name", "name":
		val = row.ItemName
	case "item_type", "item type":
		val = row.ItemType
	case "required_optional", "required":
		val = row.RequiredOptional
	case "input_restrictions", "restrictions":
		val = row.InputRestrictions
	case "display_conditions", "display":
		val = row.DisplayConditions
	case "action":
		val = row.Action
	case "navigation_destination", "navigation":
		val = row.NavigationDest
	default:
		// Try metadata
		if row.Metadata != nil {
			if v, ok := row.Metadata[header]; ok {
				val = v
			}
		}
	}

	// Normalize: trim, replace newlines with spaces
	val = strings.TrimSpace(val)
	val = strings.ReplaceAll(val, "\r\n", " ")
	val = strings.ReplaceAll(val, "\n", " ")
	val = strings.ReplaceAll(val, "\r", " ")

	// Skip empty and dash-only values
	if val == "" || val == "-" {
		return ""
	}

	return escapeTableCell(val)
}
