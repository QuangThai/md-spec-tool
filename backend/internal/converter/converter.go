package converter

import "strings"

// Helper function to create structured warnings
func newWarning(code string, severity WarningSeverity, category WarningCategory, message string, hint string, details map[string]any) Warning {
	return Warning{
		Code:     code,
		Severity: severity,
		Category: category,
		Message:  message,
		Hint:     hint,
		Details:  details,
	}
}

// Converter orchestrates the conversion process
type Converter struct {
	pasteParser    *PasteParser
	xlsxParser     *XLSXParser
	headerDetector *HeaderDetector
	columnMapper   *ColumnMapper
	renderer       *MDFlowRenderer
}

// NewConverter creates a new Converter
func NewConverter() *Converter {
	return &Converter{
		pasteParser:    NewPasteParser(),
		xlsxParser:     NewXLSXParser(),
		headerDetector: NewHeaderDetector(),
		columnMapper:   NewColumnMapper(),
		renderer:       NewMDFlowRenderer(),
	}
}

// BuildSpecDocFromPaste parses pasted content into a SpecDoc
func BuildSpecDocFromPaste(text string) (*SpecDoc, error) {
	analysis := DetectInputType(text)
	if analysis.Type == InputTypeMarkdown {
		return BuildMarkdownSpecDoc(text, "Specification"), nil
	}

	parser := NewPasteParser()
	matrix, err := parser.Parse(text)
	if err != nil {
		return nil, err
	}
	if len(matrix) == 0 {
		return &SpecDoc{Title: "Converted Spec"}, nil
	}

	converter := NewConverter()
	headerRow, _ := converter.headerDetector.DetectHeaderRow(matrix)
	headers := matrix.GetRow(headerRow)
	colMap, _ := converter.columnMapper.MapColumns(headers)

	specDoc := converter.buildSpecDoc(matrix, headerRow, headers, colMap, "")
	return specDoc, nil
}

// ConvertPaste converts pasted text to MDFlow
func (c *Converter) ConvertPaste(text string, template string) (*ConvertResponse, error) {
	// Phase 1: Detect input type first
	analysis := DetectInputType(text)

	if analysis.Type == InputTypeMarkdown {
		return c.convertMarkdown(text, template)
	}

	// Table path (existing behavior)
	matrix, err := c.pasteParser.Parse(text)
	if err != nil {
		return nil, err
	}

	return c.convertMatrix(matrix, "", template)
}

// convertMarkdown handles markdown/prose input without table parsing
func (c *Converter) convertMarkdown(text string, template string) (*ConvertResponse, error) {
	specDoc := BuildMarkdownSpecDoc(text, "Specification")

	mdflow, err := c.renderer.RenderMarkdown(specDoc, template)
	if err != nil {
		return nil, err
	}

	return &ConvertResponse{
		MDFlow:   mdflow,
		Warnings: []Warning{}, // No warnings for markdown
		Meta:     specDoc.Meta,
	}, nil
}

// ConvertXLSX converts an XLSX file to MDFlow
func (c *Converter) ConvertXLSX(filePath string, sheetName string, template string) (*ConvertResponse, error) {
	var matrix CellMatrix
	var err error

	if sheetName == "" {
		result, err := c.xlsxParser.ParseFile(filePath)
		if err != nil {
			return nil, err
		}
		sheetName = result.ActiveSheet
		matrix = result.GetMatrix(sheetName)
	} else {
		matrix, err = c.xlsxParser.ParseSheet(filePath, sheetName)
		if err != nil {
			return nil, err
		}
	}

	return c.convertMatrix(matrix, sheetName, template)
}

// GetXLSXSheets returns list of sheets in an XLSX file
func (c *Converter) GetXLSXSheets(filePath string) ([]string, error) {
	result, err := c.xlsxParser.ParseFile(filePath)
	if err != nil {
		return nil, err
	}
	return result.Sheets, nil
}

// convertMatrix converts a CellMatrix to MDFlow
func (c *Converter) convertMatrix(matrix CellMatrix, sheetName string, template string) (*ConvertResponse, error) {
	if len(matrix) == 0 {
		return &ConvertResponse{
			MDFlow:   "",
			Warnings: []Warning{newWarning("INPUT_EMPTY", SeverityWarn, CatInput, "Empty input.", "Paste a table or upload a file to convert.", nil)},
			Meta:     SpecDocMeta{},
		}, nil
	}

	// Detect header row
	headerRow, confidence := c.headerDetector.DetectHeaderRow(matrix)

	var warnings []Warning
	if confidence < 50 {
		warnings = append(warnings, newWarning(
			"HEADER_LOW_CONFIDENCE",
			SeverityWarn,
			CatHeader,
			"Low confidence in header detection; results may be inaccurate.",
			"Verify the header row and ensure column names are present.",
			map[string]any{"confidence": confidence, "header_row": headerRow},
		))
	}

	// Get headers
	headers := matrix.GetRow(headerRow)

	// Map columns
	colMap, unmapped := c.columnMapper.MapColumns(headers)

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

	// Build SpecDoc
	specDoc := c.buildSpecDoc(matrix, headerRow, headers, colMap, sheetName)
	specDoc.Warnings = warnings

	// Render to MDFlow
	mdflow, err := c.renderer.Render(specDoc, template)
	if err != nil {
		return nil, err
	}

	return &ConvertResponse{
		MDFlow:   mdflow,
		Warnings: warnings,
		Meta:     specDoc.Meta,
	}, nil
}

// buildSpecDoc constructs a SpecDoc from parsed data
func (c *Converter) buildSpecDoc(matrix CellMatrix, headerRow int, headers []string, colMap ColumnMap, sheetName string) *SpecDoc {
	// Count rows by feature
	rowsByFeature := make(map[string]int)

	var rows []SpecRow
	dataRows := matrix.SliceRows(headerRow+1, matrix.RowCount())

	for _, row := range dataRows {
		specRow := SpecRow{
			ID:           normalizeCell(GetFieldValue(row, colMap, FieldID)),
			Feature:      normalizeCell(GetFieldValue(row, colMap, FieldFeature)),
			Scenario:     normalizeCell(GetFieldValue(row, colMap, FieldScenario)),
			Instructions: normalizeCell(GetFieldValue(row, colMap, FieldInstructions)),
			Inputs:       normalizeCell(GetFieldValue(row, colMap, FieldInputs)),
			Expected:     normalizeCell(GetFieldValue(row, colMap, FieldExpected)),
			Precondition: normalizeCell(GetFieldValue(row, colMap, FieldPrecondition)),
			Priority:     normalizeCell(GetFieldValue(row, colMap, FieldPriority)),
			Type:         normalizeCell(GetFieldValue(row, colMap, FieldType)),
			Status:       normalizeCell(GetFieldValue(row, colMap, FieldStatus)),
			Endpoint:     normalizeCell(GetFieldValue(row, colMap, FieldEndpoint)),
			Notes:        normalizeCell(GetFieldValue(row, colMap, FieldNotes)),

			// Phase 3 fields
			No:                normalizeCell(GetFieldValue(row, colMap, FieldNo)),
			ItemName:          normalizeCell(GetFieldValue(row, colMap, FieldItemName)),
			ItemType:          normalizeCell(GetFieldValue(row, colMap, FieldItemType)),
			RequiredOptional:  normalizeCell(GetFieldValue(row, colMap, FieldRequiredOptional)),
			InputRestrictions: normalizeCell(GetFieldValue(row, colMap, FieldInputRestrictions)),
			DisplayConditions: normalizeCell(GetFieldValue(row, colMap, FieldDisplayConditions)),
			Action:            normalizeCell(GetFieldValue(row, colMap, FieldAction)),
			NavigationDest:    normalizeCell(GetFieldValue(row, colMap, FieldNavigationDest)),

			Metadata: make(map[string]string),
		}

		// Store unmapped columns in metadata
		for i, header := range headers {
			if i < len(row) {
				// Check if this column is mapped
				isMapped := false
				for _, idx := range colMap {
					if idx == i {
						isMapped = true
						break
					}
				}
				if !isMapped && row[i] != "" {
					specRow.Metadata[header] = row[i]
				}
			}
		}

		if shouldAppendContinuation(rows, specRow) {
			continue
		}

		// Skip completely empty rows (check both test case and spec table fields)
		if specRow.Feature == "" && specRow.Scenario == "" && specRow.Instructions == "" &&
			specRow.ItemName == "" && specRow.No == "" && specRow.Notes == "" {
			continue
		}

		// If this is a spec-table style row, map ItemName into Feature/Scenario for other templates
		if specRow.Feature == "" && specRow.ItemName != "" {
			specRow.Feature = specRow.ItemName
			if specRow.Scenario == "" {
				specRow.Scenario = specRow.ItemName
			}
		}

		// Populate Instructions/Expected from spec-table fields when missing
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
		if specRow.Expected == "" && specRow.NavigationDest != "" {
			specRow.Expected = "Navigation: " + specRow.NavigationDest
		}

		rows = append(rows, specRow)

		// Track row count by feature
		if specRow.Feature != "" {
			rowsByFeature[specRow.Feature]++
		}
	}

	// Determine title
	title := sheetName
	if title == "" {
		title = "Converted Spec"
	}

	_, unmapped := c.columnMapper.MapColumns(headers)

	return &SpecDoc{
		Title:   title,
		Rows:    rows,
		Headers: headers,
		Meta: SpecDocMeta{
			SheetName:       sheetName,
			HeaderRow:       headerRow,
			ColumnMap:       colMap,
			UnmappedColumns: unmapped,
			TotalRows:       len(rows),
			RowsByFeature:   rowsByFeature,
		},
	}
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func normalizeCell(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "-" {
		return ""
	}
	return trimmed
}

func hasMeaningfulFields(row SpecRow) bool {
	return row.ID != "" || row.Feature != "" || row.Scenario != "" || row.Instructions != "" ||
		row.Inputs != "" || row.Expected != "" || row.Precondition != "" || row.Priority != "" ||
		row.Type != "" || row.Status != "" || row.Endpoint != "" || row.Notes != "" ||
		row.ItemName != "" || row.ItemType != "" || row.RequiredOptional != "" ||
		row.InputRestrictions != "" || row.DisplayConditions != "" || row.Action != "" ||
		row.NavigationDest != ""
}

func shouldAppendContinuation(rows []SpecRow, row SpecRow) bool {
	if row.No == "" {
		return false
	}
	if hasMeaningfulFields(row) {
		return false
	}
	if len(rows) == 0 {
		return false
	}
	appendContinuation(&rows[len(rows)-1], row.No)
	return true
}

func appendContinuation(target *SpecRow, text string) {
	text = normalizeCell(text)
	if text == "" {
		return
	}
	if target.Notes != "" {
		target.Notes += "\n" + text
		return
	}
	if target.Expected != "" {
		target.Expected += "\n" + text
		return
	}
	if target.Instructions != "" {
		target.Instructions += "\n" + text
		return
	}
	if target.DisplayConditions != "" {
		target.DisplayConditions += "\n" + text
		return
	}
	target.Notes = text
}
