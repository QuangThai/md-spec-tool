package converter

import (
	"strings"
)

// TableToSpecDocAdapter converts a Table to SpecDoc using the existing column mapping logic
// This maintains backward compatibility while using the new Table model internally
type TableToSpecDocAdapter struct {
	columnMapper *ColumnMapper
}

// NewTableToSpecDocAdapter creates a new adapter
func NewTableToSpecDocAdapter() *TableToSpecDocAdapter {
	return &TableToSpecDocAdapter{
		columnMapper: NewColumnMapper(),
	}
}

// Convert converts a Table to SpecDoc
func (a *TableToSpecDocAdapter) Convert(table *Table) *SpecDoc {
	if table == nil {
		return &SpecDoc{
			Title: "Converted Spec",
			Rows:  []SpecRow{},
			Meta:  SpecDocMeta{},
		}
	}
	if table.RowCount() == 0 {
		return &SpecDoc{
			Title: table.SheetName,
			Rows:  []SpecRow{},
			Meta:  SpecDocMeta{},
		}
	}

	// Map headers to canonical fields
	colMap, unmapped := a.columnMapper.MapColumns(table.Headers)

	// Build SpecRows
	var rows []SpecRow
	rowsByFeature := make(map[string]int)

	for rowIdx := 0; rowIdx < table.RowCount(); rowIdx++ {
		tableRow := table.GetRow(rowIdx)
		rowMap := tableRow.ToMap(table.Headers)

		specRow := a.buildSpecRow(rowMap, colMap, table.Headers)

		// Apply same continuation logic as original
		if shouldAppendContinuation(rows, specRow) {
			continue
		}

		// Skip completely empty rows (same logic as original)
		if !hasMeaningfulFields(specRow) {
			continue
		}

		// Apply spec-table field mapping (same as original)
		a.applySpecTableMapping(&specRow)

		rows = append(rows, specRow)

		// Track row count by feature
		if specRow.Feature != "" {
			rowsByFeature[specRow.Feature]++
		}
	}

	// Determine title
	title := table.SheetName
	if title == "" {
		title = "Converted Spec"
	}

	// Build warnings from table metadata and column mapping
	var warnings []Warning

	// Add table parsing warnings
	for _, w := range table.Meta.Warnings {
		warnings = append(warnings, newWarning(
			"PARSING_WARNING",
			SeverityInfo,
			CatHeader,
			w,
			"",
			nil,
		))
	}

	// Add unmapped column warning
	if len(unmapped) > 0 {
		warnings = append(warnings, newWarning(
			"MAPPING_UNMAPPED_COLUMNS",
			SeverityWarn,
			CatMapping,
			"Some columns could not be mapped to known fields.",
			"Rename columns to match expected headers or choose a different template.",
			map[string]any{"unmapped_columns": unmapped},
		))
	}

	return &SpecDoc{
		Title:    title,
		Rows:     rows,
		Headers:  table.Headers,
		Warnings: warnings,
		Meta: SpecDocMeta{
			SheetName:       table.SheetName,
			HeaderRow:       table.Meta.HeaderRowIndex,
			ColumnMap:       colMap,
			UnmappedColumns: unmapped,
			TotalRows:       len(rows),
			RowsByFeature:   rowsByFeature,
			SourceURL:       table.Meta.SourceURL,
		},
	}
}

// buildSpecRow creates a SpecRow from a row map and column mapping
func (a *TableToSpecDocAdapter) buildSpecRow(rowMap map[string]string, colMap ColumnMap, headers []string) SpecRow {
	// Helper to get field value by canonical field
	getField := func(field CanonicalField) string {
		if colIdx, ok := colMap[field]; ok && colIdx < len(headers) {
			return rowMap[headers[colIdx]]
		}
		return ""
	}

	specRow := SpecRow{
		ID:           normalizeCell(getField(FieldID)),
		Feature:      normalizeCell(getField(FieldFeature)),
		Scenario:     normalizeCell(getField(FieldScenario)),
		Instructions: normalizeCell(getField(FieldInstructions)),
		Inputs:       normalizeCell(getField(FieldInputs)),
		Expected:     normalizeCell(getField(FieldExpected)),
		Precondition: normalizeCell(getField(FieldPrecondition)),
		Priority:     normalizeCell(getField(FieldPriority)),
		Type:         normalizeCell(getField(FieldType)),
		Status:       normalizeCell(getField(FieldStatus)),
		Endpoint:     normalizeCell(getField(FieldEndpoint)),
		Notes:        normalizeCell(getField(FieldNotes)),

		// Phase 3 fields
		No:                normalizeCell(getField(FieldNo)),
		ItemName:          normalizeCell(getField(FieldItemName)),
		ItemType:          normalizeCell(getField(FieldItemType)),
		RequiredOptional:  normalizeCell(getField(FieldRequiredOptional)),
		InputRestrictions: normalizeCell(getField(FieldInputRestrictions)),
		DisplayConditions: normalizeCell(getField(FieldDisplayConditions)),
		Action:            normalizeCell(getField(FieldAction)),
		NavigationDest:    normalizeCell(getField(FieldNavigationDest)),

		Metadata: make(map[string]string),
	}

	// Store unmapped columns in metadata
	for _, header := range headers {
		value := rowMap[header]
		if value == "" {
			continue
		}

		// Check if this column is mapped
		isMapped := false
		for field, _ := range colMap {
			mappedHeader := headers[colMap[field]]
			if mappedHeader == header {
				isMapped = true
				break
			}
		}

		if !isMapped {
			specRow.Metadata[header] = value
		}
	}

	return specRow
}

// applySpecTableMapping applies spec-table field mapping logic
func (a *TableToSpecDocAdapter) applySpecTableMapping(specRow *SpecRow) {
	// If this is a spec-table style row, map ItemName into Feature/Scenario
	if specRow.Feature == "" && specRow.ItemName != "" {
		specRow.Feature = specRow.ItemName
		if specRow.Scenario == "" {
			specRow.Scenario = specRow.ItemName
		}
	}

	// Populate Instructions from spec-table fields when missing
	if specRow.Instructions == "" {
		var parts []string
		if specRow.DisplayConditions != "" {
			parts = append(parts, "Display Conditions: "+specRow.DisplayConditions)
		}
		if specRow.InputRestrictions != "" {
			parts = append(parts, "Input Restrictions: "+specRow.InputRestrictions)
		}
		if specRow.Action != "" {
			parts = append(parts, "Action: "+specRow.Action)
		}
		if len(parts) > 0 {
			specRow.Instructions = strings.Join(parts, "\n")
		}
	}

	// Append spec-table fields to Instructions if they're not already there
	if specRow.Instructions != "" {
		if specRow.DisplayConditions != "" && !strings.Contains(specRow.Instructions, "Display Conditions:") {
			specRow.Instructions += "\nDisplay Conditions: " + specRow.DisplayConditions
		}
		if specRow.InputRestrictions != "" && !strings.Contains(specRow.Instructions, "Input Restrictions:") {
			specRow.Instructions += "\nInput Restrictions: " + specRow.InputRestrictions
		}
		if specRow.Action != "" && !strings.Contains(specRow.Instructions, "Action:") {
			specRow.Instructions += "\nAction: " + specRow.Action
		}
	}

	// Populate Expected from NavigationDest
	if specRow.Expected == "" && specRow.NavigationDest != "" {
		specRow.Expected = "Navigation: " + specRow.NavigationDest
	}
}
