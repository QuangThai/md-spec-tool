package handlers

import (
	"context"

	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

// ConvertValidationStats exposes convertValidationStats for external tests.
type ConvertValidationStats = convertValidationStats

// NewValidationTestHandler exposes a configurable handler for validation tests.
func NewValidationTestHandler(strict bool, minHeaderConfidence int, maxRowLossRatio float64) *MDFlowHandler {
	return &MDFlowHandler{cfg: &config.Config{
		SpecStrictMode:          strict,
		SpecMinHeaderConfidence: minHeaderConfidence,
		SpecMaxRowLossRatio:     maxRowLossRatio,
	}}
}

// NormalizeSheetRange exposes normalizeSheetRange for external tests.
func NormalizeSheetRange(sheetTitle string, rangeOverride string) string {
	return normalizeSheetRange(sheetTitle, rangeOverride)
}

// ParseA1Start exposes parseA1Start for external tests.
func ParseA1Start(rangeStr string) (startCol int, startRow int) {
	return parseA1Start(rangeStr)
}

// SelectPreferredBlockMatrix exposes selectPreferredBlockMatrix for external tests.
func SelectPreferredBlockMatrix(ctx context.Context, conv *converter.Converter, matrix converter.CellMatrix, templateName string) converter.CellMatrix {
	return selectPreferredBlockMatrix(ctx, conv, matrix, templateName)
}

// SelectMatrixForConvert exposes selectMatrixForConvert for external tests.
func SelectMatrixForConvert(ctx context.Context, conv *converter.Converter, matrix converter.CellMatrix, templateName string, selectedBlockID string, rangeOverride string) converter.CellMatrix {
	return selectMatrixForConvert(ctx, conv, matrix, templateName, selectedBlockID, rangeOverride)
}

// BuildConvertValidationError exposes buildConvertValidationError for external tests.
func BuildConvertValidationError(h *MDFlowHandler, format string, stats ConvertValidationStats, result *converter.ConvertResponse) *ErrorResponse {
	return h.buildConvertValidationError(format, stats, result)
}

// ResolveTemplateContentName exposes resolveTemplateContentName for external tests.
func ResolveTemplateContentName(name string) (canonicalName string, ok bool) {
	return resolveTemplateContentName(name)
}
