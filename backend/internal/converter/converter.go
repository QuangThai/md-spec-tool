package converter

import (
	"context"
	"fmt"
	"strings"

	"github.com/yourorg/md-spec-tool/internal/ai"
)

// OutputFormat represents supported output formats
type OutputFormat string

const (
	OutputFormatSpec  OutputFormat = "spec"
	OutputFormatTable OutputFormat = "table"
)

// DefaultTemplateName is the default template used when none is specified
const DefaultTemplateName = "spec"

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
	aiService      ai.Service
	aiMapper       *ai.ColumnMapperService

	// Phase 1: New schema-agnostic components
	tableParser  *TableParser
	tableAdapter *TableToSpecDocAdapter

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
		tableParser:  NewTableParser(),
		tableAdapter: NewTableToSpecDocAdapter(),

		// Phase 2: Generic renderer
		genericRenderer: NewGenericTableRenderer(),

		// Phase 3: Initialize template registry
		templateRegistry: templateRegistry,

		// Phase 4: Initialize renderer factory
		rendererFactory: NewRendererFactory(templateRegistry),
	}
}

// WithAIService injects an AI service for column mapping
func (c *Converter) WithAIService(service ai.Service) *Converter {
	c.aiService = service
	if service != nil {
		c.aiMapper = ai.NewColumnMapperService(service)
	}
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
	colMap, unmapped := converter.columnMapper.MapColumns(headers)

	specDoc := converter.buildSpecDoc(matrix, headerRow, headers, colMap, unmapped, "")
	return specDoc, nil
}

// ConvertPaste converts pasted text to MDFlow
func (c *Converter) ConvertPaste(text string, template string) (*ConvertResponse, error) {
	return c.ConvertPasteWithFormatContext(context.Background(), text, template, "")
}

// ConvertPasteWithFormat converts pasted text with format option
func (c *Converter) ConvertPasteWithFormat(text string, templateName string, outputFormat string) (*ConvertResponse, error) {
	return c.ConvertPasteWithFormatContext(context.Background(), text, templateName, outputFormat)
}

// ConvertPasteWithFormatContext converts pasted text with format option and context
// templateName: template identifier for rendering
// outputFormat: "spec" | "table" (output rendering format)
func (c *Converter) ConvertPasteWithFormatContext(ctx context.Context, text string, templateName string, outputFormat string) (*ConvertResponse, error) {
	// Phase 1: Detect input type first
	analysis := DetectInputType(text)

	if analysis.Type == InputTypeMarkdown {
		return c.convertMarkdown(text, templateName)
	}

	// Table path (existing behavior)
	matrix, err := c.pasteParser.Parse(text)
	if err != nil {
		return nil, err
	}

	return c.convertMatrixWithFormat(ctx, matrix, "", templateName, outputFormat)
}

// ConvertPasteWithOverrides applies column overrides before conversion.
func (c *Converter) ConvertPasteWithOverrides(ctx context.Context, text string, templateName string, outputFormat string, overrides map[string]string) (*ConvertResponse, error) {
	analysis := DetectInputType(text)
	if analysis.Type == InputTypeMarkdown {
		return c.convertMarkdown(text, templateName)
	}

	matrix, err := c.pasteParser.Parse(text)
	if err != nil {
		return nil, err
	}

	if len(overrides) > 0 {
		matrix = applyColumnOverrides(matrix, overrides)
	}

	return c.convertMatrixWithFormat(ctx, matrix, "", templateName, outputFormat)
}

// ConvertMatrix converts a CellMatrix to MDFlow (uses template-driven pipeline)
func (c *Converter) ConvertMatrix(matrix CellMatrix, sheetName string, templateName string) (*ConvertResponse, error) {
	return c.convertMatrixWithFormat(context.Background(), matrix, sheetName, templateName, "")
}

// ConvertMatrixWithFormat converts a CellMatrix with output format option
// outputFormat: "spec" | "table" (output rendering format)
// templateName: template identifier for rendering
func (c *Converter) ConvertMatrixWithFormat(matrix CellMatrix, sheetName string, templateName string, outputFormat string) (*ConvertResponse, error) {
	return c.convertMatrixWithFormat(context.Background(), matrix, sheetName, templateName, outputFormat)
}

// ConvertMatrixWithFormatContext converts a CellMatrix with output format option and context
// outputFormat: "spec" | "table" (output rendering format)
// templateName: template identifier for rendering
func (c *Converter) ConvertMatrixWithFormatContext(ctx context.Context, matrix CellMatrix, sheetName string, templateName string, outputFormat string) (*ConvertResponse, error) {
	return c.convertMatrixWithFormat(ctx, matrix, sheetName, templateName, outputFormat)
}

// ConvertMatrixWithOverrides applies column overrides before conversion.
func (c *Converter) ConvertMatrixWithOverrides(ctx context.Context, matrix CellMatrix, sheetName string, templateName string, outputFormat string, overrides map[string]string) (*ConvertResponse, error) {
	if len(overrides) > 0 {
		matrix = applyColumnOverrides(matrix, overrides)
	}
	return c.convertMatrixWithFormat(ctx, matrix, sheetName, templateName, outputFormat)
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

	return c.convertMatrixWithFormat(context.Background(), matrix, sheetName, template, "")
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

// convertMatrixWithFormat converts a CellMatrix to markdown with output format option
// outputFormat: "spec" | "table" (output rendering format)
// templateName: template identifier for rendering
func (c *Converter) convertMatrixWithFormat(ctx context.Context, matrix CellMatrix, sheetName string, templateName string, outputFormat string) (*ConvertResponse, error) {
	// Validate output format
	outputFormat = strings.ToLower(strings.TrimSpace(outputFormat))
	if OutputFormat(outputFormat) != OutputFormatSpec && OutputFormat(outputFormat) != OutputFormatTable && outputFormat != "" {
		return nil, fmt.Errorf("invalid output format '%s': must be 'spec' or 'table'", outputFormat)
	}

	// Default to "spec" if not specified
	if outputFormat == "" {
		outputFormat = string(OutputFormatSpec)
	}

	return c.convertWithTemplate(ctx, matrix, sheetName, templateName, outputFormat)
}

// convertWithTemplate converts using template-driven rendering (Phase 3-4)
// Supports all output types via template.Output.Type
func (c *Converter) convertWithTemplate(ctx context.Context, matrix CellMatrix, sheetName string, templateName string, format string) (*ConvertResponse, error) {
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
	headers := matrix.GetRow(headerRow)
	dataRows := matrix.SliceRows(headerRow+1, matrix.RowCount())
	colMap, unmapped, mappingWarnings, aiMeta := c.resolveColumnMapping(ctx, headers, dataRows, format)
	warnings = append(warnings, mappingWarnings...)
	colMap, unmapped, inferredWarnings := enhanceColumnMapping(headers, dataRows, colMap)
	warnings = append(warnings, inferredWarnings...)

	quality := evaluateMappingQuality(confidence, headers, colMap)
	if shouldFallbackToTable(format, quality) {
		warnings = append(warnings, newWarning(
			"MAPPING_LOW_CONFIDENCE_TABLE_FALLBACK",
			SeverityWarn,
			CatMapping,
			"Mapping confidence is low for non-standard schema; switching output to table format.",
			"Use preview column overrides if you want spec format for this input.",
			map[string]any{
				"quality_score":    quality.Score,
				"mapped_ratio":     quality.MappedRatio,
				"core_field_count": quality.CoreMapped,
			},
		))
		format = string(OutputFormatTable)
	}

	table := c.tableParser.MatrixToTable(headers, dataRows, sheetName)
	table.Meta.HeaderRowIndex = headerRow
	table.Meta.ColumnMap = colMap
	applyAIMetaToTableMeta(&table.Meta, aiMeta)

	// Load template - use format name if templateName is empty
	templateToUse := templateName
	if templateToUse == "" {
		templateToUse = format // Try to use format as template name
	}
	template := c.templateRegistry.LoadTemplateOrDefault(templateToUse)

	// Create renderer using factory (Phase 4)
	var renderer Renderer
	var err error
	if format == "table" {
		renderer, err = NewRendererSimple("table")
	} else {
		renderer, err = c.rendererFactory.CreateRenderer(template)
	}
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
	meta := SpecDocMeta{
		SheetName:       sheetName,
		HeaderRow:       headerRow,
		ColumnMap:       colMap,
		UnmappedColumns: unmapped,
		TotalRows:       table.RowCount(),
		OutputFormat:    format,
	}
	applyAIMeta(&meta, aiMeta)

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
func (c *Converter) buildSpecDoc(matrix CellMatrix, headerRow int, headers []string, colMap ColumnMap, unmapped []string, sheetName string) *SpecDoc {
	// Count rows by feature
	rowsByFeature := make(map[string]int)

	var rows []SpecRow
	dataRows := matrix.SliceRows(headerRow+1, matrix.RowCount())

	for _, row := range dataRows {
		specRow := SpecRow{
			ID:           normalizeCell(GetFieldValue(row, colMap, FieldID)),
			Title:        normalizeCell(GetFieldValue(row, colMap, FieldTitle)),
			Description:  normalizeCell(GetFieldValue(row, colMap, FieldDescription)),
			Acceptance:   normalizeCell(GetFieldValue(row, colMap, FieldAcceptance)),
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
			Method:       normalizeCell(GetFieldValue(row, colMap, FieldMethod)),
			Parameters:   normalizeCell(GetFieldValue(row, colMap, FieldParameters)),
			Response:     normalizeCell(GetFieldValue(row, colMap, FieldResponse)),
			StatusCode:   normalizeCell(GetFieldValue(row, colMap, FieldStatusCode)),
			Notes:        normalizeCell(GetFieldValue(row, colMap, FieldNotes)),
			Component:    normalizeCell(GetFieldValue(row, colMap, FieldComponent)),
			Assignee:     normalizeCell(GetFieldValue(row, colMap, FieldAssignee)),
			Category:     normalizeCell(GetFieldValue(row, colMap, FieldCategory)),

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

		// Skip completely empty rows (check test case, backlog, API, and spec table fields)
		if specRow.Feature == "" && specRow.Title == "" && specRow.Description == "" && specRow.Acceptance == "" &&
			specRow.Scenario == "" && specRow.Instructions == "" &&
			specRow.Endpoint == "" && specRow.Method == "" && specRow.Parameters == "" && specRow.Response == "" && specRow.StatusCode == "" &&
			specRow.ItemName == "" && specRow.No == "" && specRow.Notes == "" && len(specRow.Metadata) == 0 {
			continue
		}

		// If this is a spec-table style row, map ItemName into Feature/Scenario for other templates
		if specRow.Feature == "" && specRow.ItemName != "" {
			specRow.Feature = specRow.ItemName
			if specRow.Scenario == "" {
				specRow.Scenario = specRow.ItemName
			}
		}

		// If this is a backlog row, map Title into Feature/Scenario when missing
		if specRow.Feature == "" && specRow.Title != "" {
			specRow.Feature = specRow.Title
			if specRow.Scenario == "" {
				specRow.Scenario = specRow.Title
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
	return strings.Join(strs, sep)
}

func filterWarningsByCategory(warnings []Warning, category WarningCategory) []Warning {
	if len(warnings) == 0 {
		return warnings
	}
	filtered := make([]Warning, 0, len(warnings))
	for _, w := range warnings {
		if w.Category == category {
			continue
		}
		filtered = append(filtered, w)
	}
	return filtered
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
	// Only treat as continuation if No looks like wrapped text (not a section header or row number)
	if looksLikeRowNumber(row.No) {
		return false
	}
	if looksLikeSectionHeader(row.No) {
		return false
	}
	appendContinuation(&rows[len(rows)-1], row.No)
	return true
}

func looksLikeRowNumber(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func looksLikeSectionHeader(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			return true
		}
		if ch >= 0x3000 && ch <= 0x9FFF {
			return true
		}
		if ch >= 0xF900 && ch <= 0xFAFF {
			return true
		}
	}
	return false
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

// GetPreviewColumnMapping returns column mapping and unmapped headers for preview.
// Uses template-driven HeaderResolver when template is loaded; otherwise falls back to ColumnMapper.
// templateName can be empty (uses DefaultTemplateName). Returns header -> canonical field name for JSON.
func (c *Converter) GetPreviewColumnMapping(headers []string, templateName string) (columnMapping map[string]string, unmapped []string) {
	if templateName == "" {
		templateName = DefaultTemplateName
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

// GetPreviewColumnMappingWithContext returns column mapping using AI when available.
// Falls back to template-driven resolver when AI is off/low confidence.
func (c *Converter) GetPreviewColumnMappingWithContext(ctx context.Context, headers []string, dataRows [][]string, templateName string, format string) (columnMapping map[string]string, unmapped []string) {
	if templateName == "" {
		templateName = DefaultTemplateName
	}
	template := c.templateRegistry.LoadTemplateOrDefault(templateName)
	resolver := NewHeaderResolver(template)
	colMap, unmapped, _, _ := c.resolveColumnMappingWithFallback(ctx, headers, dataRows, format, false, func(h []string) (ColumnMap, []string) {
		resolved, unresolved, _ := resolver.ResolveHeaders(h)
		return resolved, unresolved
	})
	colMap, unmapped, _ = enhanceColumnMapping(headers, dataRows, colMap)
	columnMapping = make(map[string]string)
	for field, idx := range colMap {
		if idx >= 0 && idx < len(headers) {
			columnMapping[headers[idx]] = string(field)
		}
	}
	return columnMapping, unmapped
}

// GetPreviewColumnMappingRuleBased returns column mapping using only rule-based resolution.
// Never calls AI service. Used by preview endpoints when skip_ai=true for guaranteed fast response.
func (c *Converter) GetPreviewColumnMappingRuleBased(headers []string, templateName string) (columnMapping map[string]string, unmapped []string) {
	if templateName == "" {
		templateName = DefaultTemplateName
	}
	template := c.templateRegistry.LoadTemplateOrDefault(templateName)
	resolver := NewHeaderResolver(template)
	colMap, unmapped, _ := resolver.ResolveHeaders(headers)
	colMap, unmapped, _ = enhanceColumnMapping(headers, nil, colMap)
	columnMapping = make(map[string]string)
	for field, idx := range colMap {
		if idx >= 0 && idx < len(headers) {
			columnMapping[headers[idx]] = string(field)
		}
	}
	return columnMapping, unmapped
}

// HasAIService returns true if AI service is configured and available.
// Used by handlers to signal to the frontend whether AI is available for convert operations.
func (c *Converter) HasAIService() bool {
	return c.aiService != nil && c.aiMapper != nil
}

