package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
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

// Test helpers: prefer testing through public handler APIs instead of exposing internals.
// The helpers below are kept only for backward compatibility; avoid adding new ones.
// TODO: Refactor tests to use public handler methods instead of internal helpers.

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

// ParseGoogleSheetURL exposes parseGoogleSheetURL for external tests.
func ParseGoogleSheetURL(urlStr string) (sheetID string, gid string, ok bool) {
	return parseGoogleSheetURL(urlStr)
}

// SelectGID exposes selectGID for external tests.
func SelectGID(requestGID string, urlGID string) string {
	return selectGID(requestGID, urlGID)
}

// ValidateGID exposes validateGID for external tests.
func ValidateGID(gid string) error {
	return validateGID(gid)
}

// GetBearerToken exposes getBearerToken for external tests.
func GetBearerToken(c *gin.Context) string {
	return getBearerToken(c)
}

// IsAuthError exposes isAuthError for external tests.
func IsAuthError(err error) bool {
	return isAuthError(err)
}

// AnalyzeSelectedMatrix exposes analyzeSelectedMatrix for external tests.
func AnalyzeSelectedMatrix(matrix converter.CellMatrix) ConvertValidationStats {
	return analyzeSelectedMatrix(matrix)
}

// BuildCoreFieldCoverage exposes buildCoreFieldCoverage for external tests.
func BuildCoreFieldCoverage(colMap converter.ColumnMap) map[string]bool {
	return buildCoreFieldCoverage(colMap)
}

// GSheetValuesResult exposes gsheetValuesResult for external tests.
type GSheetValuesResult = gsheetValuesResult

// SetGSheetHTTPClientForTest injects HTTP client for testing (e.g. mock transport).
func SetGSheetHTTPClientForTest(h *GSheetHandler, c *http.Client) {
	h.gsheetHTTPClient = c
}

// GSheetHandlerConfigForTest returns config for assertions.
func GSheetHandlerConfigForTest(h *GSheetHandler) *config.Config {
	return h.cfg
}

// GSheetHandlerHasGetAIServiceForTest returns true if getAIService is set.
func GSheetHandlerHasGetAIServiceForTest(h *GSheetHandler) bool {
	return h.getAIService != nil
}

// AIServiceProviderHasBYOKCacheForTest returns true if byokCache is set.
func AIServiceProviderHasBYOKCacheForTest(p *AIServiceProvider) bool {
	return p.byokCache != nil
}

// AIServiceProviderHasDefaultAIForTest returns true if defaultAI is set.
func AIServiceProviderHasDefaultAIForTest(p *AIServiceProvider) bool {
	return p.defaultAI != nil
}

// ConvertHandlerConverterForTest returns the internal converter for tests.
func ConvertHandlerConverterForTest(h *ConvertHandler) *converter.Converter {
	return h.converter
}

// EmptyTablePreviewForTest exposes emptyTablePreview for external tests.
func EmptyTablePreviewForTest(h *PreviewHandler, c *gin.Context, confidence int, inputType string) PreviewResponse {
	return h.emptyTablePreview(c, confidence, inputType)
}
