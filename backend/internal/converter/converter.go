package converter

import (
	"fmt"
	"strings"
)

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

	// Phase 1: New schema-agnostic components
	tableParser    *TableParser
	tableAdapter   *TableToSpecDocAdapter
	useNewPipeline bool // Feature flag for gradual migration

	// Phase 2: Generic renderer
	genericRenderer *GenericTableRenderer

	// Phase 3: Template registry and config-driven renderers
	templateRegistry *TemplateRegistry

	// Phase 4: Renderer factory for output format abstraction
	rendererFactory *RendererFactory
}

// NewConverter creates a new Converter
func NewConverter() *Converter {
	templateRegistry := NewTemplateRegistry()
	return &Converter{
		pasteParser:    NewPasteParser(),
		xlsxParser:     NewXLSXParser(),
		headerDetector: NewHeaderDetector(),
		columnMapper:   NewColumnMapper(),
		renderer:       NewMDFlowRenderer(),

		// Phase 1: Initialize new components
		tableParser:    NewTableParser(),
		tableAdapter:   NewTableToSpecDocAdapter(),
		useNewPipeline: false, // Default to old pipeline for safety

		// Phase 2: Generic renderer
		genericRenderer: NewGenericTableRenderer(),

		// Phase 3: Initialize template registry
		templateRegistry: templateRegistry,

		// Phase 4: Initialize renderer factory
		rendererFactory: NewRendererFactory(templateRegistry),
	}
}

// WithNewPipeline enables the new Table-based pipeline
func (c *Converter) WithNewPipeline() *Converter {
	c.useNewPipeline = true
	return c
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
	return c.ConvertPasteWithFormat(text, template, "")
}

// ConvertPasteWithFormat converts pasted text with format option
func (c *Converter) ConvertPasteWithFormat(text string, template string, format string) (*ConvertResponse, error) {
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

	return c.convertMatrixWithFormat(matrix, "", template, format)
}

// ConvertMatrix converts a CellMatrix to MDFlow
func (c *Converter) ConvertMatrix(matrix CellMatrix, sheetName string, template string) (*ConvertResponse, error) {
	return c.convertMatrix(matrix, sheetName, template)
}

// ConvertMatrixWithFormat converts a CellMatrix with format option
func (c *Converter) ConvertMatrixWithFormat(matrix CellMatrix, sheetName string, template string, format string) (*ConvertResponse, error) {
	return c.convertMatrixWithFormat(matrix, sheetName, template, format)
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

// ParseXLSX parses an XLSX file and returns the cell matrix
func (c *Converter) ParseXLSX(filePath string, sheetName string) (CellMatrix, error) {
	if sheetName == "" {
		result, err := c.xlsxParser.ParseFile(filePath)
		if err != nil {
			return nil, err
		}
		sheetName = result.ActiveSheet
		return result.GetMatrix(sheetName), nil
	}
	return c.xlsxParser.ParseSheet(filePath, sheetName)
}

// convertMatrix converts a CellMatrix to MDFlow
func (c *Converter) convertMatrix(matrix CellMatrix, sheetName string, template string) (*ConvertResponse, error) {
	// NEW: Check if we should use the new Table pipeline
	if c.useNewPipeline {
		return c.convertMatrixViaTable(matrix, sheetName, template)
	}

	// OLD: Keep existing logic for backward compatibility
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

// convertMatrixWithFormat converts a CellMatrix to markdown with format option (Phase 2-4)
// Format can be: test_spec (legacy), generic_table, test_spec_v1 (template-based), row_cards, etc.
func (c *Converter) convertMatrixWithFormat(matrix CellMatrix, sheetName string, template string, format string) (*ConvertResponse, error) {
	// If no format specified or "test_spec", use default behavior
	if format == "" || format == "test_spec" {
		return c.convertMatrix(matrix, sheetName, template)
	}

	// Phase 2: Handle generic_table format
	if format == "generic_table" {
		return c.convertToGenericTable(matrix, sheetName)
	}

	// Phase 3-4: Handle template-driven formats
	// template-based formats: test_spec_v1, row_cards, narrative, etc.
	return c.convertWithTemplate(matrix, sheetName, template, format)
}

// convertWithTemplate converts using template-driven rendering (Phase 3-4)
// Supports all output types via template.Output.Type
func (c *Converter) convertWithTemplate(matrix CellMatrix, sheetName string, templateName string, format string) (*ConvertResponse, error) {
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

	// Parse matrix to Table (schema-agnostic)
	table := c.tableParser.MatrixToTable(matrix.GetRow(headerRow), matrix.SliceRows(headerRow+1, matrix.RowCount()), sheetName)

	// Load template - use format name if templateName is empty
	templateToUse := templateName
	if templateToUse == "" {
		templateToUse = format // Try to use format as template name
	}
	template := c.templateRegistry.LoadTemplateOrDefault(templateToUse)

	// Create renderer using factory (Phase 4)
	renderer, err := c.rendererFactory.CreateRenderer(template)
	if err != nil {
		return nil, fmt.Errorf("failed to create renderer for format '%s': %w", format, err)
	}

	// Render using template
	mdflow, renderWarnings, err := renderer.Render(table)
	if err != nil {
		return nil, err
	}

	// Add render warnings
	for _, w := range renderWarnings {
		warnings = append(warnings, newWarning(
			"RENDER_WARNING",
			SeverityWarn,
			CatRender,
			w,
			"",
			nil,
		))
	}

	// Build metadata
	resolver := NewHeaderResolver(template)
	colMap, unmapped, _ := resolver.ResolveHeaders(table.Headers)

	meta := SpecDocMeta{
		SheetName:       sheetName,
		HeaderRow:       headerRow,
		ColumnMap:       colMap,
		UnmappedColumns: unmapped,
		TotalRows:       table.RowCount(),
	}

	return &ConvertResponse{
		MDFlow:   mdflow,
		Warnings: warnings,
		Meta:     meta,
	}, nil
}

// convertToGenericTable converts matrix to simple Markdown table format (Phase 2)
func (c *Converter) convertToGenericTable(matrix CellMatrix, sheetName string) (*ConvertResponse, error) {
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

	// Parse to Table using TableParser (Phase 1)
	headers := matrix.GetRow(headerRow)
	dataRows := matrix.SliceRows(headerRow+1, matrix.RowCount())

	// Convert to Table
	table := c.tableParser.MatrixToTable(headers, dataRows, sheetName)
	table.Meta.HeaderRowIndex = headerRow

	// Render using GenericTableRenderer
	markdown, renderWarnings, err := c.genericRenderer.Render(table)
	if err != nil {
		return nil, err
	}

	// Combine warnings
	warnings = append(warnings, convertStringWarningsToWarningObjects(renderWarnings)...)

	return &ConvertResponse{
		MDFlow:   markdown,
		Warnings: warnings,
		Meta: SpecDocMeta{
			SheetName:       sheetName,
			HeaderRow:       headerRow,
			ColumnMap:       ColumnMap{}, // Not applicable for generic tables
			TotalRows:       table.RowCount(),
			UnmappedColumns: []string{}, // All columns preserved
		},
	}, nil
}

// convertStringWarningsToWarningObjects converts string warnings to Warning objects
func convertStringWarningsToWarningObjects(strWarnings []string) []Warning {
	var warnings []Warning
	for _, w := range strWarnings {
		warnings = append(warnings, newWarning(
			"RENDER_WARNING",
			SeverityWarn,
			CatRender,
			w,
			"",
			nil,
		))
	}
	return warnings
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

// convertMatrixViaTable converts a CellMatrix to MDFlow using the new Table pipeline
// This is the Phase 1 implementation that maintains backward compatibility
func (c *Converter) convertMatrixViaTable(matrix CellMatrix, sheetName string, template string) (*ConvertResponse, error) {
	if len(matrix) == 0 {
		return &ConvertResponse{
			MDFlow:   "",
			Warnings: []Warning{newWarning("INPUT_EMPTY", SeverityWarn, CatInput, "Empty input.", "Paste a table or upload a file to convert.", nil)},
			Meta:     SpecDocMeta{},
		}, nil
	}

	// Parse matrix to Table (schema-agnostic)
	table, err := c.tableParser.ParseMatrix(matrix, sheetName)
	if err != nil {
		return nil, err
	}

	// Convert Table to SpecDoc (using existing column mapping)
	specDoc := c.tableAdapter.Convert(table)

	// Add header confidence warning if needed
	if table.Meta.HeaderRowIndex > 0 {
		// Header was not in first row - might indicate detection issue
		// But this is actually normal, so we'll skip this warning for now
	}

	// Render to MDFlow
	mdflow, err := c.renderer.Render(specDoc, template)
	if err != nil {
		return nil, err
	}

	return &ConvertResponse{
		MDFlow:   mdflow,
		Warnings: specDoc.Warnings,
		Meta:     specDoc.Meta,
	}, nil
}

// ListTemplates returns all available templates as TemplateConfig objects
func (c *Converter) ListTemplates() []*TemplateConfig {
	names := c.templateRegistry.ListTemplates()
	templates := make([]*TemplateConfig, 0, len(names))
	for _, name := range names {
		if cfg, err := c.templateRegistry.LoadTemplate(name); err == nil {
			templates = append(templates, cfg)
		}
	}
	return templates
}

// GetPreviewColumnMapping returns column mapping and unmapped headers for preview (Phase 3).
// Uses template-driven HeaderResolver when template is loaded; otherwise falls back to ColumnMapper.
// templateName can be empty (uses "default"). Returns header -> canonical field name for JSON.
func (c *Converter) GetPreviewColumnMapping(headers []string, templateName string) (columnMapping map[string]string, unmapped []string) {
	if templateName == "" {
		templateName = "default"
	}
	template := c.templateRegistry.LoadTemplateOrDefault(templateName)
	resolver := NewHeaderResolver(template)
	colMap, unmapped, _ := resolver.ResolveHeaders(headers)
	columnMapping = make(map[string]string)
	for field, idx := range colMap {
		if idx >= 0 && idx < len(headers) {
			columnMapping[headers[idx]] = string(field)
		}
	}
	return columnMapping, unmapped
}

// convertToTestSpecWithTemplate converts a CellMatrix using config-driven template (Phase 3)
func (c *Converter) convertToTestSpecWithTemplate(matrix CellMatrix, sheetName string, format string) (*ConvertResponse, error) {
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

	// Parse matrix to Table (schema-agnostic)
	table := c.tableParser.MatrixToTable(matrix.GetRow(headerRow), matrix.SliceRows(headerRow+1, matrix.RowCount()), sheetName)

	// Load template (Phase 3)
	template := c.templateRegistry.LoadTemplateOrDefault(format)

	// Create renderer with template
	renderer := NewTestSpecRenderer(template)

	// Render using template
	mdflow, renderWarnings, err := renderer.Render(table)
	if err != nil {
		return nil, err
	}

	// Add render warnings
	for _, w := range renderWarnings {
		warnings = append(warnings, newWarning(
			"RENDER_WARNING",
			SeverityWarn,
			CatRender,
			w,
			"",
			nil,
		))
	}

	// Build metadata
	resolver := NewHeaderResolver(template)
	colMap, unmapped, _ := resolver.ResolveHeaders(table.Headers)

	meta := SpecDocMeta{
		SheetName:       sheetName,
		HeaderRow:       headerRow,
		ColumnMap:       colMap,
		UnmappedColumns: unmapped,
		TotalRows:       table.RowCount(),
	}

	return &ConvertResponse{
		MDFlow:   mdflow,
		Warnings: warnings,
		Meta:     meta,
	}, nil
}
