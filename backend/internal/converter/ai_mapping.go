package converter

import (
	"context"
	"strings"

	"github.com/yourorg/md-spec-tool/internal/ai"
)

const (
	aiMinAvgConfidence = 0.75
	aiMinMappedRatio   = 0.60
	aiSampleRows       = 5
)

// AIMappingMeta captures AI mapping summary for metadata
type AIMappingMeta struct {
	Mode            string
	Used            bool
	Degraded        bool
	AvgConfidence   float64
	MappedColumns   int
	UnmappedColumns int
}

func (c *Converter) resolveColumnMapping(ctx context.Context, headers []string, dataRows [][]string, format string) (ColumnMap, []string, []Warning, *AIMappingMeta) {
	return c.resolveColumnMappingWithFallback(ctx, headers, dataRows, format, false, func(h []string) (ColumnMap, []string) {
		return c.columnMapper.MapColumns(h)
	})
}

// resolveColumnMappingRuleBasedOnly resolves column mapping using only rule-based fallback, never AI.
// Used by preview endpoints to guarantee fast response times.
func (c *Converter) resolveColumnMappingRuleBasedOnly(ctx context.Context, headers []string, dataRows [][]string, format string, fallback func([]string) (ColumnMap, []string)) (ColumnMap, []string, []Warning, *AIMappingMeta) {
	return c.resolveColumnMappingWithFallback(ctx, headers, dataRows, format, true, fallback)
}

func (c *Converter) resolveColumnMappingWithFallback(ctx context.Context, headers []string, dataRows [][]string, format string, skipAI bool, fallback func([]string) (ColumnMap, []string)) (ColumnMap, []string, []Warning, *AIMappingMeta) {
	meta := &AIMappingMeta{Mode: "off"}

	// For table format, always use fallback (no AI needed)
	if format == "table" {
		colMap, unmapped := fallback(headers)
		return colMap, unmapped, nil, meta
	}

	// Skip AI when explicitly requested (e.g. preview endpoints for fast response)
	if skipAI {
		colMap, unmapped := fallback(headers)
		meta.Mode = "skipped"
		return colMap, unmapped, nil, meta
	}

	// If AI service not available, use fallback
	if c.aiService == nil || c.aiMapper == nil {
		colMap, unmapped := fallback(headers)
		return colMap, unmapped, nil, meta
	}

	meta.Mode = "on"

	cleanHeaders := normalizeHeaders(headers)
	sampleRows := buildSampleRows(dataRows, aiSampleRows)
	sourceLang := DetectLanguageHint(EstimateEnglishScore(headers, dataRows), headers, dataRows)
	schemaHint := inferSchemaHint(headers, dataRows)

	result, err := c.aiMapper.MapColumns(ctx, ai.MapColumnsRequest{
		Headers:    cleanHeaders,
		SampleRows: sampleRows,
		Format:     "spec",
		FileType:   "table",
		SourceLang: sourceLang,
		SchemaHint: schemaHint,
	})
	if err != nil {
		meta.Degraded = true
		colMap, unmapped := fallback(headers)
		return colMap, unmapped, []Warning{newWarning(
			"MAPPING_AI_FAILED",
			SeverityWarn,
			CatMapping,
			"AI mapping failed; using fallback mapping.",
			"Check your AI configuration or retry conversion.",
			map[string]any{"error": err.Error()},
		)}, meta
	}

	meta.Used = true
	meta.AvgConfidence = result.Meta.AvgConfidence
	meta.MappedColumns = result.Meta.MappedColumns
	meta.UnmappedColumns = result.Meta.UnmappedColumns

	colMap, unmapped, mappingWarnings := aiMappingToColumnMap(headers, result)

	// Recalculate mapped count after filtering invalid/unknown fields
	meta.MappedColumns = len(colMap)

	// Fall back if confidence is too low
	if !aiMappingMeetsThreshold(meta, len(headers)) || len(colMap) == 0 {
		meta.Degraded = true
		fallbackMap, fallbackUnmapped := fallback(headers)
		mappingWarnings = append(mappingWarnings, newWarning(
			"MAPPING_AI_LOW_CONFIDENCE",
			SeverityWarn,
			CatMapping,
			"AI mapping confidence is low; using fallback mapping.",
			"Verify header names or provide clearer column labels.",
			map[string]any{"avg_confidence": meta.AvgConfidence, "mapped_columns": meta.MappedColumns},
		))
		return fallbackMap, fallbackUnmapped, mappingWarnings, meta
	}

	return colMap, unmapped, mappingWarnings, meta
}

// inferSchemaHint detects likely schema from headers and sample data.
// Returns "test_case", "ui_spec", "product_backlog", "issue_tracker", "api_spec", or "auto".
func inferSchemaHint(headers []string, dataRows [][]string) string {
	lower := make([]string, len(headers))
	for i, h := range headers {
		lower[i] = strings.ToLower(strings.TrimSpace(h))
	}
	joined := " " + strings.Join(lower, " ") + " "

	// API spec: endpoint, method, parameters, response, status
	if strings.Contains(joined, " endpoint") || strings.Contains(joined, " method") ||
		strings.Contains(joined, " parameters") || strings.Contains(joined, " response") ||
		strings.Contains(joined, " status_code") {
		return "api_spec"
	}
	// UI spec: item name, item type, display conditions, action
	if strings.Contains(joined, " item_name") || strings.Contains(joined, " item type") ||
		strings.Contains(joined, " display_conditions") || strings.Contains(joined, " action") {
		return "ui_spec"
	}
	// Product backlog: story points, sprint, acceptance criteria
	if strings.Contains(joined, " story") || strings.Contains(joined, " sprint") ||
		strings.Contains(joined, " acceptance") || strings.Contains(joined, " backlog") {
		return "product_backlog"
	}
	// Issue tracker: assignee, component, priority, status
	if strings.Contains(joined, " assignee") || strings.Contains(joined, " component") ||
		strings.Contains(joined, " issue") || strings.Contains(joined, " ticket") {
		return "issue_tracker"
	}
	// Test case: feature, scenario, steps, expected
	if strings.Contains(joined, " feature") || strings.Contains(joined, " scenario") ||
		strings.Contains(joined, " steps") || strings.Contains(joined, " expected") ||
		strings.Contains(joined, " test case") {
		return "test_case"
	}
	return "auto"
}

func aiMappingMeetsThreshold(meta *AIMappingMeta, totalColumns int) bool {
	if meta == nil {
		return false
	}
	if meta.AvgConfidence < aiMinAvgConfidence {
		return false
	}
	if totalColumns == 0 {
		return false
	}
	mappedRatio := float64(meta.MappedColumns) / float64(totalColumns)
	return mappedRatio >= aiMinMappedRatio
}

func buildSampleRows(rows [][]string, limit int) [][]string {
	if limit <= 0 {
		return nil
	}
	result := make([][]string, 0, limit)
	for _, row := range rows {
		if len(result) >= limit {
			break
		}
		if isRowEmpty(row) {
			continue
		}
		result = append(result, trimCells(row))
	}
	return result
}

func isRowEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func trimCells(row []string) []string {
	trimmed := make([]string, len(row))
	for i, cell := range row {
		trimmed[i] = strings.TrimSpace(cell)
	}
	return trimmed
}

func normalizeHeaders(headers []string) []string {
	clean := make([]string, len(headers))
	for i, header := range headers {
		clean[i] = strings.TrimSpace(header)
	}
	return clean
}

func aiMappingToColumnMap(headers []string, result *ai.ColumnMappingResult) (ColumnMap, []string, []Warning) {
	colMap := make(ColumnMap)
	var warnings []Warning
	usedIndices := make(map[int]bool)

	for _, mapping := range result.CanonicalFields {
		field, ok := mapAICanonicalField(mapping.CanonicalName)
		if !ok {
			if !aiExtraFieldNames[mapping.CanonicalName] {
				warnings = append(warnings, newWarning(
					"MAPPING_AI_UNKNOWN_FIELD",
					SeverityInfo,
					CatMapping,
					"AI returned an unknown canonical field; ignoring mapping.",
					"Update AI schema if this field should be supported.",
					map[string]any{"canonical_name": mapping.CanonicalName},
				))
			}
			continue
		}
		if mapping.ColumnIndex < 0 || mapping.ColumnIndex >= len(headers) {
			warnings = append(warnings, newWarning(
				"MAPPING_AI_INVALID_COLUMN",
				SeverityWarn,
				CatMapping,
				"AI returned an invalid column index; ignoring mapping.",
				"Ensure headers align with the provided table data.",
				map[string]any{"canonical_name": mapping.CanonicalName, "column_index": mapping.ColumnIndex},
			))
			continue
		}
		if _, exists := colMap[field]; exists {
			warnings = append(warnings, newWarning(
				"MAPPING_AI_DUPLICATE_FIELD",
				SeverityInfo,
				CatMapping,
				"AI returned duplicate mappings for the same field; keeping first.",
				"Review headers for duplicate meanings.",
				map[string]any{"canonical_name": mapping.CanonicalName},
			))
			continue
		}
		colMap[field] = mapping.ColumnIndex
		usedIndices[mapping.ColumnIndex] = true
	}

	var unmapped []string
	if len(result.ExtraColumns) > 0 {
		for _, extra := range result.ExtraColumns {
			if extra.ColumnIndex >= 0 && extra.ColumnIndex < len(headers) {
				unmapped = append(unmapped, headers[extra.ColumnIndex])
			}
		}
	} else {
		for i, header := range headers {
			if !usedIndices[i] {
				unmapped = append(unmapped, header)
			}
		}
	}

	return colMap, unmapped, warnings
}

// aiCanonicalAliases maps AI-returned canonical names (including common
// variations the model may produce) to the internal CanonicalField values.
var aiCanonicalAliases = map[string]CanonicalField{
	// exact canonical names
	"id":                     FieldID,
	"title":                  FieldTitle,
	"description":            FieldDescription,
	"acceptance_criteria":    FieldAcceptance,
	"feature":                FieldFeature,
	"scenario":               FieldScenario,
	"instructions":           FieldInstructions,
	"inputs":                 FieldInputs,
	"expected":               FieldExpected,
	"precondition":           FieldPrecondition,
	"priority":               FieldPriority,
	"type":                   FieldType,
	"status":                 FieldStatus,
	"endpoint":               FieldEndpoint,
	"method":                 FieldMethod,
	"parameters":             FieldParameters,
	"response":               FieldResponse,
	"status_code":            FieldStatusCode,
	"notes":                  FieldNotes,
	"component":              FieldComponent,
	"assignee":               FieldAssignee,
	"category":               FieldCategory,
	"no":                     FieldNo,
	"item_name":              FieldItemName,
	"item_type":              FieldItemType,
	"required_optional":      FieldRequiredOptional,
	"input_restrictions":     FieldInputRestrictions,
	"display_conditions":     FieldDisplayConditions,
	"action":                 FieldAction,
	"navigation_destination": FieldNavigationDest,

	// --- aliases the AI model may return ---
	// title / feature aliases
	"name":           FieldTitle,
	"summary":        FieldTitle,
	"requirement":    FieldFeature,
	"test_case":      FieldScenario,
	"test_case_name": FieldScenario,

	// instructions aliases
	"steps":      FieldInstructions,
	"test_steps": FieldInstructions,
	"procedure":  FieldInstructions,

	// expected aliases
	"expected_result": FieldExpected,
	"result":          FieldExpected,
	"outcome":         FieldExpected,
	"acceptance":      FieldAcceptance,
	"criteria":        FieldAcceptance,

	// precondition aliases
	"pre_condition": FieldPrecondition,
	"prerequisites": FieldPrecondition,

	// method aliases
	"http_method": FieldMethod,
	"verb":        FieldMethod,

	// response aliases
	"response_body":   FieldResponse,
	"response_json":   FieldResponse,
	"response_schema": FieldResponse,

	// parameters aliases
	"params":       FieldParameters,
	"request":      FieldParameters,
	"request_body": FieldParameters,

	// status code aliases
	"status code": FieldStatusCode,
	"http_status": FieldStatusCode,

	// no aliases
	"number": FieldNo,
	"seq":    FieldNo,
	"index":  FieldNo,

	// item_name / item_type aliases
	"field_name":   FieldItemName,
	"label":        FieldItemName,
	"control_type": FieldItemType,
	"widget":       FieldItemType,
	"field_type":   FieldItemType,

	// notes aliases
	"remarks": FieldNotes,
	"comment": FieldNotes,
	"memo":    FieldNotes,

	// display_conditions aliases
	"visibility":     FieldDisplayConditions,
	"show_condition": FieldDisplayConditions,

	// action aliases
	"trigger":  FieldAction,
	"event":    FieldAction,
	"on_click": FieldAction,

	// navigation_destination aliases
	"target":      FieldNavigationDest,
	"redirect":    FieldNavigationDest,
	"next_screen": FieldNavigationDest,
	"link":        FieldNavigationDest,
	"destination": FieldNavigationDest,

	// required_optional aliases
	"mandatory": FieldRequiredOptional,
	"required":  FieldRequiredOptional,
	"optional":  FieldRequiredOptional,

	// input_restrictions aliases
	"constraint": FieldInputRestrictions,
	"validation": FieldInputRestrictions,
	"rule":       FieldInputRestrictions,
}

// aiExtraFieldNames are canonical names the AI may return that are valid
// metadata but don't map to any CanonicalField. They should not produce
// warnings.
var aiExtraFieldNames = map[string]bool{
	"assignee":  true,
	"component": true,
	"category":  true,
}

func mapAICanonicalField(name string) (CanonicalField, bool) {
	if field, ok := aiCanonicalAliases[name]; ok {
		return field, true
	}
	if aiExtraFieldNames[name] {
		return "", false
	}
	return "", false
}

func applyAIMeta(meta *SpecDocMeta, aiMeta *AIMappingMeta) {
	if meta == nil || aiMeta == nil {
		return
	}
	meta.AIMode = aiMeta.Mode
	meta.AIUsed = aiMeta.Used
	meta.AIDegraded = aiMeta.Degraded
	meta.AIAvgConfidence = aiMeta.AvgConfidence
	meta.AIMappedColumns = aiMeta.MappedColumns
	meta.AIUnmappedColumns = aiMeta.UnmappedColumns
}

func applyAIMetaToTableMeta(meta *TableMeta, aiMeta *AIMappingMeta) {
	if meta == nil || aiMeta == nil {
		return
	}
	meta.AIMode = aiMeta.Mode
	meta.AIUsed = aiMeta.Used
	meta.AIDegraded = aiMeta.Degraded
	meta.AIAvgConfidence = aiMeta.AvgConfidence
	meta.AIMappedColumns = aiMeta.MappedColumns
	meta.AIUnmappedColumns = aiMeta.UnmappedColumns
}

func preferAIMapping(candidate *AIMappingMeta, current *AIMappingMeta) bool {
	if candidate == nil {
		return false
	}
	if current == nil {
		return true
	}
	if candidate.Degraded && !current.Degraded {
		return false
	}
	if candidate.MappedColumns > current.MappedColumns {
		return true
	}
	if candidate.MappedColumns == current.MappedColumns {
		return candidate.AvgConfidence > current.AvgConfidence
	}
	return false
}
