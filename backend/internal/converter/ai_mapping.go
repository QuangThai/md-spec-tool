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

	result, err := c.aiMapper.MapColumns(ctx, ai.MapColumnsRequest{
		Headers:    cleanHeaders,
		SampleRows: sampleRows,
		Format:     "spec",
		FileType:   "table",
		SourceLang: "unknown",
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
			warnings = append(warnings, newWarning(
				"MAPPING_AI_UNKNOWN_FIELD",
				SeverityInfo,
				CatMapping,
				"AI returned an unknown canonical field; ignoring mapping.",
				"Update AI schema if this field should be supported.",
				map[string]any{"canonical_name": mapping.CanonicalName},
			))
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

func mapAICanonicalField(name string) (CanonicalField, bool) {
	switch name {
	case "id":
		return FieldID, true
	case "feature":
		return FieldFeature, true
	case "scenario":
		return FieldScenario, true
	case "instructions":
		return FieldInstructions, true
	case "inputs":
		return FieldInputs, true
	case "expected":
		return FieldExpected, true
	case "precondition":
		return FieldPrecondition, true
	case "priority":
		return FieldPriority, true
	case "type":
		return FieldType, true
	case "status":
		return FieldStatus, true
	case "endpoint":
		return FieldEndpoint, true
	case "notes":
		return FieldNotes, true
	case "no":
		return FieldNo, true
	case "item_name":
		return FieldItemName, true
	case "item_type":
		return FieldItemType, true
	case "required_optional":
		return FieldRequiredOptional, true
	case "input_restrictions":
		return FieldInputRestrictions, true
	case "display_conditions":
		return FieldDisplayConditions, true
	case "action":
		return FieldAction, true
	case "navigation_destination":
		return FieldNavigationDest, true
	default:
		return "", false
	}
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
