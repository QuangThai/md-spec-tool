package handlers

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/converter"
	"github.com/yourorg/md-spec-tool/internal/gsheetutils"
	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// GoogleSheetRequest represents the request for Google Sheet import
type GoogleSheetRequest struct {
	URL             string            `json:"url" binding:"required"`
	Template        string            `json:"template"`
	Format          string            `json:"format"` // "spec" | "table"
	GID             string            `json:"gid,omitempty"`
	Range           string            `json:"range,omitempty"`
	SelectedBlockID string            `json:"selected_block_id,omitempty"`
	ColumnOverrides map[string]string `json:"column_overrides,omitempty"`
	IncludeMetadata *bool             `json:"include_metadata,omitempty"`
	NumberRows      *bool             `json:"number_rows,omitempty"`
}

// GoogleSheetSheetsRequest represents the request for sheet list
type GoogleSheetSheetsRequest struct {
	URL string `json:"url" binding:"required"`
}

// GoogleSheetTab represents a Google Sheet tab
type GoogleSheetTab struct {
	Title string `json:"title"`
	GID   string `json:"gid"`
}

// GoogleSheetSheetsResponse represents a sheet list response
type GoogleSheetSheetsResponse struct {
	Sheets    []GoogleSheetTab `json:"sheets"`
	ActiveGID string           `json:"active_gid"`
}

// GoogleSheetResponse represents the response for Google Sheet parsing
type GoogleSheetResponse struct {
	SheetID   string `json:"sheet_id"`
	SheetName string `json:"sheet_name,omitempty"`
	Data      string `json:"data"`
}

type gsheetValuesResult struct {
	Rows      [][]string
	SheetName string
	Range     string
	StartCol  int
	StartRow  int
}

type convertValidationStats struct {
	SourceRows       int
	HeaderRow        int
	HeaderConfidence int
	HeaderCount      int
}

// parseGoogleSheetURL is deprecated. Use gsheetutils.ParseGoogleSheetURL instead.
// This is kept for backward compatibility during migration.
func parseGoogleSheetURL(urlStr string) (sheetID string, gid string, ok bool) {
	return gsheetutils.ParseGoogleSheetURL(urlStr)
}

func (h *MDFlowHandler) getSheetsService() (*sheets.Service, error) {
	h.sheetsInitOnce.Do(func() {
		credsPath := strings.TrimSpace(os.Getenv(googleSheetsCredsEnv))
		if credsPath == "" {
			h.sheetsInitErr = errSheetsNotConfigured
			return
		}

		ctx := context.Background()
		service, err := sheets.NewService(
			ctx,
			option.WithCredentialsFile(credsPath),
			option.WithScopes(sheets.SpreadsheetsReadonlyScope),
		)
		if err != nil {
			h.sheetsInitErr = err
			return
		}
		h.sheetsService = service
	})

	if h.sheetsService == nil {
		if h.sheetsInitErr == nil {
			return nil, errSheetsNotConfigured
		}
		return nil, h.sheetsInitErr
	}

	return h.sheetsService, nil
}

func (h *MDFlowHandler) getSheetsServiceWithToken(accessToken string) (*sheets.Service, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, fmt.Errorf("missing access token")
	}

	ctx := context.Background()
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
	}))

	service, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return service, nil
}

// selectGID is deprecated. Use gsheetutils.SelectGID instead.
func selectGID(requestGID string, urlGID string) string {
	return gsheetutils.SelectGID(requestGID, urlGID)
}

// validateGID is deprecated. Use gsheetutils.ValidateGID instead.
func validateGID(gid string) error {
	return gsheetutils.ValidateGID(gid)
}

func getGoogleSheetTabsWithService(service *sheets.Service, spreadsheetID string) ([]GoogleSheetTab, error) {
	var resp *sheets.Spreadsheet
	err := retryGSheetAPI(func() error {
		var e error
		resp, e = service.Spreadsheets.Get(spreadsheetID).
			Fields("sheets.properties.sheetId,sheets.properties.title").
			Do()
		return e
	}, 2)
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.Sheets) == 0 {
		return nil, fmt.Errorf("no sheets found")
	}

	result := make([]GoogleSheetTab, 0, len(resp.Sheets))
	for _, sheet := range resp.Sheets {
		if sheet.Properties == nil {
			continue
		}
		gid := strconv.FormatInt(sheet.Properties.SheetId, 10)
		result = append(result, GoogleSheetTab{
			Title: sheet.Properties.Title,
			GID:   gid,
		})
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no sheets found")
	}

	return result, nil
}

func (h *MDFlowHandler) getGoogleSheetTabs(spreadsheetID string) ([]GoogleSheetTab, error) {
	service, err := h.getSheetsService()
	if err != nil {
		return nil, err
	}

	return getGoogleSheetTabsWithService(service, spreadsheetID)
}

// retryGSheetAPI retries fn on 429/503 with exponential backoff
func retryGSheetAPI(fn func() error, maxRetries int) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if backoff > 8*time.Second {
				backoff = 8 * time.Second
			}
			time.Sleep(backoff)
		}
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if gerr, ok := lastErr.(*googleapi.Error); ok && (gerr.Code == 429 || gerr.Code == 503) {
			slog.Warn("GSheet API retry", "code", gerr.Code, "attempt", attempt+1)
			continue
		}
		return lastErr
	}
	return lastErr
}

func findActiveGID(tabs []GoogleSheetTab, requestedGID string) string {
	requestedGID = strings.TrimSpace(requestedGID)
	if requestedGID != "" {
		for _, tab := range tabs {
			if tab.GID == requestedGID {
				return requestedGID
			}
		}
	}

	if len(tabs) > 0 {
		return tabs[0].GID
	}

	return ""
}

func findSheetTitleByGID(tabs []GoogleSheetTab, gid string) (string, bool) {
	for _, tab := range tabs {
		if tab.GID == gid {
			return tab.Title, true
		}
	}
	return "", false
}

// sheetRange returns the range string for Values.Get (whole sheet). Sheet names with spaces or special chars must be single-quoted.
func sheetRange(sheetTitle string) string {
	if sheetTitle == "" {
		return "A1"
	}
	if strings.ContainsAny(sheetTitle, " \t-'!,&") {
		escaped := strings.ReplaceAll(sheetTitle, "'", "''")
		return "'" + escaped + "'"
	}
	return sheetTitle
}

// normalizeSheetRange resolves user range input to a sheet-scoped A1 range.
// If rangeOverride is empty, it returns the whole selected sheet.
// If rangeOverride has no "!", it is treated as a relative A1 range for the selected sheet.
func normalizeSheetRange(sheetTitle string, rangeOverride string) string {
	base := sheetRange(sheetTitle)
	trimmed := strings.TrimSpace(rangeOverride)
	if trimmed == "" {
		return base
	}
	if strings.HasPrefix(trimmed, "!") {
		return base + trimmed
	}
	if strings.Contains(trimmed, "!") {
		return trimmed
	}
	return base + "!" + trimmed
}

func matrixBlockA1(startCol, endCol, startRow, endRow int) string {
	start := fmt.Sprintf("%s%d", columnToLetters(startCol), startRow+1)
	end := fmt.Sprintf("%s%d", columnToLetters(endCol), endRow+1)
	return start + ":" + end
}

func matrixBlockA1WithOffset(startCol, endCol, startRow, endRow, offsetCol, offsetRow int) string {
	start := fmt.Sprintf("%s%d", columnToLetters(startCol+offsetCol), startRow+offsetRow+1)
	end := fmt.Sprintf("%s%d", columnToLetters(endCol+offsetCol), endRow+offsetRow+1)
	return start + ":" + end
}

func parseA1Start(rangeStr string) (startCol int, startRow int) {
	trimmed := strings.TrimSpace(rangeStr)
	if trimmed == "" {
		return 0, 0
	}

	if bang := strings.LastIndex(trimmed, "!"); bang >= 0 {
		trimmed = trimmed[bang+1:]
	}
	if colon := strings.Index(trimmed, ":"); colon >= 0 {
		trimmed = trimmed[:colon]
	}
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return 0, 0
	}

	i := 0
	for i < len(trimmed) {
		ch := trimmed[i]
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
			i++
			continue
		}
		break
	}
	if i == 0 || i >= len(trimmed) {
		return 0, 0
	}
	if i > 3 {
		return 0, 0
	}

	colLetters := strings.ToUpper(trimmed[:i])
	rowText := trimmed[i:]
	rowNum, err := strconv.Atoi(rowText)
	if err != nil || rowNum < 1 {
		return 0, 0
	}

	colNum := 0
	for _, ch := range colLetters {
		colNum = colNum*26 + int(ch-'A'+1)
	}
	if colNum < 1 {
		return 0, 0
	}

	return colNum - 1, rowNum - 1
}

func qualifyRangeWithSheet(sheetName string, rangeA1 string) string {
	rangeA1 = strings.TrimSpace(rangeA1)
	if rangeA1 == "" || strings.Contains(rangeA1, "!") {
		return rangeA1
	}
	if strings.TrimSpace(sheetName) == "" {
		return rangeA1
	}
	return sheetRange(sheetName) + "!" + rangeA1
}

func columnToLetters(index int) string {
	if index < 0 {
		return "A"
	}
	n := index + 1
	letters := ""
	for n > 0 {
		rem := (n - 1) % 26
		letters = string(rune('A'+rem)) + letters
		n = (n - 1) / 26
	}
	return letters
}

func convertValuesToRows(values [][]interface{}) [][]string {
	if len(values) == 0 {
		return nil
	}

	rows := make([][]string, 0, len(values))
	for _, row := range values {
		converted := make([]string, len(row))
		for i, cell := range row {
			if cell == nil {
				continue
			}
			switch v := cell.(type) {
			case string:
				converted[i] = v
			default:
				converted[i] = fmt.Sprint(v)
			}
		}
		rows = append(rows, converted)
	}

	return rows
}

// fetchGoogleSheetWithService fetches sheet data via Sheets API (uses user OAuth token, works for private sheets).
func (h *MDFlowHandler) fetchGoogleSheetValuesWithService(service *sheets.Service, sheetID string, gid string, rangeOverride string) (*gsheetValuesResult, error) {
	tabs, err := getGoogleSheetTabsWithService(service, sheetID)
	if err != nil {
		return nil, err
	}
	activeGID := findActiveGID(tabs, gid)
	if activeGID == "" {
		return nil, fmt.Errorf("no sheets available")
	}
	sheetTitle, ok := findSheetTitleByGID(tabs, activeGID)
	if !ok {
		return nil, fmt.Errorf("sheet gid not found")
	}
	rangeStr := normalizeSheetRange(sheetTitle, rangeOverride)
	var resp *sheets.ValueRange
	err = retryGSheetAPI(func() error {
		var e error
		resp, e = service.Spreadsheets.Values.Get(sheetID, rangeStr).Do()
		return e
	}, 2)
	if err != nil {
		return nil, err
	}
	rows := convertValuesToRows(resp.Values)
	startCol, startRow := parseA1Start(resp.Range)

	return &gsheetValuesResult{
		Rows:      rows,
		SheetName: sheetTitle,
		Range:     resp.Range,
		StartCol:  startCol,
		StartRow:  startRow,
	}, nil
}

func (h *MDFlowHandler) fetchGoogleSheetWithService(service *sheets.Service, sheetID string, gid string, rangeOverride string) ([]byte, error) {
	valuesResult, err := h.fetchGoogleSheetValuesWithService(service, sheetID, gid, rangeOverride)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	for _, row := range valuesResult.Rows {
		_ = w.Write(row)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// getGSheetHTTPClient returns a client with GSheetHTTPTimeout for Sheets export fetches
func (h *MDFlowHandler) getGSheetHTTPClient() *http.Client {
	h.gsheetClientOnce.Do(func() {
		timeout := 45 * time.Second
		if h.cfg != nil {
			timeout = h.cfg.HTTPClientTimeout + 15*time.Second
			if h.cfg.GSheetHTTPTimeout > 0 {
				timeout = h.cfg.GSheetHTTPTimeout
			}
		}
		h.gsheetHTTPClient = &http.Client{Timeout: timeout}
	})
	if h.gsheetHTTPClient != nil {
		return h.gsheetHTTPClient
	}
	return h.httpClient
}

func (h *MDFlowHandler) fetchGoogleSheetCSV(sheetID string, gid string, rangeOverride string) ([]byte, int, error) {
	exportURL := fmt.Sprintf(googleSheetsExportURLFmt, sheetID)
	if gid != "" {
		exportURL += "&gid=" + url.QueryEscape(gid)
	}
	if strings.TrimSpace(rangeOverride) != "" {
		exportURL += "&range=" + url.QueryEscape(rangeOverride)
	}

	client := h.getGSheetHTTPClient()
	maxRetries := h.cfg.GSheetMaxRetries
	if maxRetries <= 0 {
		maxRetries = 2
	}

	var lastErr error
	var lastStatus int
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if backoff > 8*time.Second {
				backoff = 8 * time.Second
			}
			time.Sleep(backoff)
		}

		resp, err := client.Get(exportURL)
		if err != nil {
			lastErr = err
			lastStatus = 0
			continue
		}

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(io.LimitReader(resp.Body, h.cfg.MaxUploadBytes))
			resp.Body.Close()
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			return body, resp.StatusCode, nil
		}

		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		lastStatus = resp.StatusCode
		lastErr = fmt.Errorf("google sheets export returned status %d", resp.StatusCode)

		if resp.StatusCode != 429 && resp.StatusCode != 503 {
			return nil, resp.StatusCode, lastErr
		}
		slog.Warn("GSheet CSV fetch retry", "status", resp.StatusCode, "attempt", attempt+1)
	}

	if lastStatus != 0 {
		return nil, lastStatus, lastErr
	}
	return nil, 0, lastErr
}

func (h *MDFlowHandler) convertGoogleSheetWithService(ctx context.Context, conv *converter.Converter, service *sheets.Service, sheetID string, gid string, templateName string, outputFormat string, rangeOverride string, selectedBlockID string, columnOverrides map[string]string) (*converter.ConvertResponse, convertValidationStats, error) {
	valuesResult, err := h.fetchGoogleSheetValuesWithService(service, sheetID, gid, rangeOverride)
	if err != nil {
		return nil, convertValidationStats{}, err
	}
	rows := valuesResult.Rows
	matrix := converter.NewCellMatrix(rows).Normalize()
	selected := selectMatrixForConvert(ctx, conv, matrix, templateName, selectedBlockID, rangeOverride)
	stats := analyzeSelectedMatrix(selected)
	result, err := conv.ConvertMatrixWithOverrides(ctx, selected, valuesResult.SheetName, templateName, outputFormat, columnOverrides)
	if err != nil {
		return nil, stats, err
	}
	return result, stats, nil
}

func (h *MDFlowHandler) convertGoogleSheetWithAPI(ctx context.Context, conv *converter.Converter, sheetID string, gid string, templateName string, outputFormat string, rangeOverride string, selectedBlockID string, columnOverrides map[string]string) (*converter.ConvertResponse, convertValidationStats, error) {
	service, err := h.getSheetsService()
	if err != nil {
		return nil, convertValidationStats{}, err
	}

	return h.convertGoogleSheetWithService(ctx, conv, service, sheetID, gid, templateName, outputFormat, rangeOverride, selectedBlockID, columnOverrides)
}

func (h *MDFlowHandler) fetchGoogleSheetValuesWithAPI(sheetID string, gid string, rangeOverride string) (*gsheetValuesResult, error) {
	service, err := h.getSheetsService()
	if err != nil {
		return nil, err
	}

	return h.fetchGoogleSheetValuesWithService(service, sheetID, gid, rangeOverride)
}

func getBearerToken(c *gin.Context) string {
	value := strings.TrimSpace(c.GetHeader("Authorization"))
	if value == "" {
		return ""
	}
	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func isAuthError(err error) bool {
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		return apiErr.Code == http.StatusUnauthorized || apiErr.Code == http.StatusForbidden
	}
	return false
}

func selectPreferredBlockMatrix(ctx context.Context, conv *converter.Converter, matrix converter.CellMatrix, templateName string) converter.CellMatrix {
	blocks := converter.DetectTableBlocks(matrix)
	if len(blocks) == 0 {
		return matrix
	}
	if len(blocks) == 1 {
		return blocks[0].Matrix
	}

	hasMultiRowBlock := false
	for _, block := range blocks {
		headerDetector := converter.NewHeaderDetector()
		headerRow, _ := headerDetector.DetectHeaderRow(block.Matrix)
		dataRows := block.Matrix.SliceRows(headerRow+1, block.Matrix.RowCount())
		if len(dataRows) >= 2 {
			hasMultiRowBlock = true
			break
		}
	}

	indices := make([]int, 0, len(blocks))
	for idx, block := range blocks {
		headerDetector := converter.NewHeaderDetector()
		headerRow, _ := headerDetector.DetectHeaderRow(block.Matrix)
		dataRows := block.Matrix.SliceRows(headerRow+1, block.Matrix.RowCount())
		if !hasMultiRowBlock || len(dataRows) >= 2 {
			indices = append(indices, idx)
		}
	}

	if len(indices) == 0 {
		return blocks[0].Matrix
	}

	candidates := make([]converter.BlockSelectionCandidate, 0, len(indices))
	for _, idx := range indices {
		block := blocks[idx]
		headerDetector := converter.NewHeaderDetector()
		headerRow, confidence := headerDetector.DetectHeaderRow(block.Matrix)
		headers := block.Matrix.GetRow(headerRow)
		dataRows := block.Matrix.SliceRows(headerRow+1, block.Matrix.RowCount())
		columnMapping, unmapped := conv.GetPreviewColumnMappingWithContext(ctx, headers, dataRows, templateName, "")
		quality := converter.BuildPreviewMappingQuality(confidence, headers, dataRows, columnMapping, unmapped)
		englishScore := converter.EstimateEnglishScore(headers, dataRows)

		candidates = append(candidates, converter.BlockSelectionCandidate{
			EnglishScore: englishScore,
			QualityScore: quality.Score,
			RowCount:     len(dataRows),
			ColumnCount:  len(headers),
		})
	}

	selectedLocalIdx := converter.SelectPreferredBlock(candidates)
	if selectedLocalIdx < 0 || selectedLocalIdx >= len(indices) {
		return blocks[indices[0]].Matrix
	}

	return blocks[indices[selectedLocalIdx]].Matrix
}

func selectBlockMatrixByID(matrix converter.CellMatrix, blockID string) (converter.CellMatrix, bool) {
	if strings.TrimSpace(blockID) == "" {
		return nil, false
	}
	blocks := converter.DetectTableBlocks(matrix)
	for _, block := range blocks {
		if block.ID == blockID {
			return block.Matrix, true
		}
	}
	return nil, false
}

func selectMatrixForConvert(ctx context.Context, conv *converter.Converter, matrix converter.CellMatrix, templateName string, selectedBlockID string, rangeOverride string) converter.CellMatrix {
	if selected, ok := selectBlockMatrixByID(matrix, selectedBlockID); ok {
		return selected
	}
	if strings.TrimSpace(selectedBlockID) != "" {
		return matrix
	}
	if strings.TrimSpace(rangeOverride) != "" {
		return matrix
	}
	return selectPreferredBlockMatrix(ctx, conv, matrix, templateName)
}

func analyzeSelectedMatrix(matrix converter.CellMatrix) convertValidationStats {
	headerDetector := converter.NewHeaderDetector()
	headerRow, confidence := headerDetector.DetectHeaderRow(matrix)
	sourceRows := matrix.RowCount() - headerRow - 1
	if sourceRows < 0 {
		sourceRows = 0
	}
	return convertValidationStats{
		SourceRows:       sourceRows,
		HeaderRow:        headerRow,
		HeaderConfidence: confidence,
		HeaderCount:      len(matrix.GetRow(headerRow)),
	}
}

func buildCoreFieldCoverage(colMap converter.ColumnMap) map[string]bool {
	coverage := map[string]bool{
		string(converter.FieldFeature):           false,
		string(converter.FieldScenario):          false,
		string(converter.FieldDescription):       false,
		string(converter.FieldInstructions):      false,
		string(converter.FieldExpected):          false,
		string(converter.FieldItemName):          false,
		string(converter.FieldDisplayConditions): false,
		string(converter.FieldAction):            false,
		string(converter.FieldNavigationDest):    false,
	}
	for field := range colMap {
		key := string(field)
		if _, ok := coverage[key]; ok {
			coverage[key] = true
		}
	}
	return coverage
}

func (h *MDFlowHandler) buildQualityReport(stats convertValidationStats, result *converter.ConvertResponse) *converter.QualityReport {
	convertedRows := result.Meta.TotalRows
	mappedColumns := len(result.Meta.ColumnMap)
	mappedRatio := 0.0
	if stats.HeaderCount > 0 {
		mappedRatio = float64(mappedColumns) / float64(stats.HeaderCount)
	}
	rowLossRatio := 0.0
	if stats.SourceRows > 0 {
		rowLossRatio = 1 - (float64(convertedRows) / float64(stats.SourceRows))
		if rowLossRatio < 0 {
			rowLossRatio = 0
		}
	}

	validationPassed := true
	validationReason := ""
	if stats.HeaderConfidence < h.cfg.SpecMinHeaderConfidence {
		validationPassed = false
		validationReason = "low_header_confidence"
	} else if stats.SourceRows >= 2 && rowLossRatio > h.cfg.SpecMaxRowLossRatio {
		validationPassed = false
		validationReason = "row_loss"
	}

	return &converter.QualityReport{
		StrictMode:          h.cfg.SpecStrictMode,
		ValidationPassed:    validationPassed,
		ValidationReason:    validationReason,
		HeaderConfidence:    stats.HeaderConfidence,
		MinHeaderConfidence: h.cfg.SpecMinHeaderConfidence,
		SourceRows:          stats.SourceRows,
		ConvertedRows:       convertedRows,
		RowLossRatio:        rowLossRatio,
		MaxRowLossRatio:     h.cfg.SpecMaxRowLossRatio,
		HeaderCount:         stats.HeaderCount,
		MappedColumns:       mappedColumns,
		MappedRatio:         mappedRatio,
		CoreFieldCoverage:   buildCoreFieldCoverage(result.Meta.ColumnMap),
	}
}

func qualityReportLogArgs(report *converter.QualityReport) []any {
	if report == nil {
		return nil
	}
	return []any{
		"strict_mode", report.StrictMode,
		"validation_passed", report.ValidationPassed,
		"validation_reason", report.ValidationReason,
		"header_confidence", report.HeaderConfidence,
		"min_header_confidence", report.MinHeaderConfidence,
		"source_rows", report.SourceRows,
		"converted_rows", report.ConvertedRows,
		"row_loss_ratio", report.RowLossRatio,
		"max_row_loss_ratio", report.MaxRowLossRatio,
		"header_count", report.HeaderCount,
		"mapped_columns", report.MappedColumns,
		"mapped_ratio", report.MappedRatio,
	}
}

func (h *MDFlowHandler) buildConvertValidationError(format string, stats convertValidationStats, result *converter.ConvertResponse) *ErrorResponse {
	if format != string(converter.OutputFormatSpec) {
		return nil
	}
	if !h.cfg.SpecStrictMode {
		return nil
	}
	if stats.HeaderConfidence < h.cfg.SpecMinHeaderConfidence {
		return &ErrorResponse{
			Error:            "Header confidence is too low for reliable spec conversion",
			Code:             "CONVERT_VALIDATION_FAILED",
			ValidationReason: "low_header_confidence",
			Details: map[string]any{
				"validation_reason":  "low_header_confidence",
				"confidence":         stats.HeaderConfidence,
				"header_row":         stats.HeaderRow,
				"min_confidence":     h.cfg.SpecMinHeaderConfidence,
				"recommended_action": "Refine range/header or use Simple Table output",
			},
		}
	}
	convertedRows := result.Meta.TotalRows
	rowLossRatio := 0.0
	if stats.SourceRows > 0 {
		rowLossRatio = 1 - (float64(convertedRows) / float64(stats.SourceRows))
	}
	if stats.SourceRows >= 2 && rowLossRatio > h.cfg.SpecMaxRowLossRatio {
		return &ErrorResponse{
			Error:            "Conversion dropped too many rows to safely generate spec output",
			Code:             "CONVERT_VALIDATION_FAILED",
			ValidationReason: "row_loss",
			Details: map[string]any{
				"validation_reason":  "row_loss",
				"source_rows":        stats.SourceRows,
				"converted_rows":     convertedRows,
				"loss_ratio":         rowLossRatio,
				"max_allowed_loss":   h.cfg.SpecMaxRowLossRatio,
				"recommended_action": "Expand range or review column mapping before converting to spec",
			},
		}
	}
	return nil
}

// GetGoogleSheetSheets handles POST /api/mdflow/gsheet/sheets
// Returns available tabs for a Google Sheet
func (h *MDFlowHandler) GetGoogleSheetSheets(c *gin.Context) {
	var req GoogleSheetSheetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "url is required"})
		return
	}

	sheetID, gid, ok := parseGoogleSheetURL(req.URL)
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid Google Sheets URL"})
		return
	}

	if accessToken := getBearerToken(c); accessToken != "" {
		service, err := h.getSheetsServiceWithToken(accessToken)
		if err == nil {
			if tabs, err := getGoogleSheetTabsWithService(service, sheetID); err == nil {
				activeGID := findActiveGID(tabs, gid)
				c.JSON(http.StatusOK, GoogleSheetSheetsResponse{
					Sheets:    tabs,
					ActiveGID: activeGID,
				})
				return
			} else if isAuthError(err) {
				c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Google authorization expired"})
				return
			}
		}
	}

	tabs, err := h.getGoogleSheetTabs(sheetID)
	if err != nil {
		if errors.Is(err, errSheetsNotConfigured) {
			c.JSON(http.StatusNotImplemented, ErrorResponse{Error: "Google Sheets API not configured"})
			return
		}
		slog.Error("mdflow.GetGoogleSheetSheets error", "error", err)
		c.JSON(http.StatusBadGateway, ErrorResponse{Error: "failed to fetch Google Sheet tabs"})
		return
	}

	activeGID := findActiveGID(tabs, gid)
	c.JSON(http.StatusOK, GoogleSheetSheetsResponse{
		Sheets:    tabs,
		ActiveGID: activeGID,
	})
}

// FetchGoogleSheet handles POST /api/mdflow/gsheet
// Fetches sheet data as CSV. Uses OAuth when Authorization Bearer is present (private sheets); otherwise public export URL.
func (h *MDFlowHandler) FetchGoogleSheet(c *gin.Context) {
	var req GoogleSheetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "url is required"})
		return
	}

	sheetID, gid, ok := parseGoogleSheetURL(req.URL)
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid Google Sheets URL"})
		return
	}

	gid = selectGID(req.GID, gid)
	if err := validateGID(gid); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	slog.Info("mdflow.FetchGoogleSheet", "sheetID", sheetID, "gid", gid)

	// Prefer OAuth when client sends Bearer token (e.g. frontend with "Connect Google") so private sheets work
	if accessToken := getBearerToken(c); accessToken != "" {
		service, err := h.getSheetsServiceWithToken(accessToken)
		if err == nil {
			body, err := h.fetchGoogleSheetWithService(service, sheetID, gid, req.Range)
			if err == nil {
				c.JSON(http.StatusOK, GoogleSheetResponse{
					SheetID:   sheetID,
					SheetName: gid,
					Data:      string(body),
				})
				return
			}
			if isAuthError(err) {
				slog.Warn("mdflow.FetchGoogleSheet auth error", "error", err)
			}
		}
	}

	// Fallback: public export URL (fails with 401 for private sheets)
	body, statusCode, err := h.fetchGoogleSheetCSV(sheetID, gid, req.Range)
	if err != nil && statusCode == 0 {
		slog.Error("mdflow.FetchGoogleSheet fetch error", "error", err)
		c.JSON(http.StatusBadGateway, ErrorResponse{Error: "failed to fetch Google Sheet"})
		return
	}
	if statusCode != http.StatusOK {
		slog.Warn("mdflow.FetchGoogleSheet non-200", "status", statusCode, "error", err)
		if statusCode == http.StatusNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Google Sheet not found or not public"})
			return
		}
		if statusCode == http.StatusUnauthorized {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Google Sheet is private; connect Google in the app to access it"})
			return
		}
		c.JSON(http.StatusBadGateway, ErrorResponse{Error: fmt.Sprintf("Google Sheets returned status %d", statusCode)})
		return
	}

	c.JSON(http.StatusOK, GoogleSheetResponse{
		SheetID:   sheetID,
		SheetName: gid,
		Data:      string(body),
	})
}

// PreviewGoogleSheet handles POST /api/mdflow/gsheet/preview
// Returns a preview of the Google Sheet data before conversion (with AI column mapping)
func (h *MDFlowHandler) PreviewGoogleSheet(c *gin.Context) {
	var req GoogleSheetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "url is required"})
		return
	}

	normalizedTemplate, err := normalizeTemplate(req.Template)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	req.Template = normalizedTemplate

	sheetID, gid, ok := parseGoogleSheetURL(req.URL)
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid Google Sheets URL"})
		return
	}

	gid = selectGID(req.GID, gid)
	if err := validateGID(gid); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	slog.Info("mdflow.PreviewGoogleSheet", "sheetID", sheetID, "gid", gid, "template", req.Template)

	var matrix converter.CellMatrix
	rangeOffsetRow := 0
	rangeOffsetCol := 0
	rangeSheetName := ""

	// Prefer OAuth when client sends Bearer token
	if accessToken := getBearerToken(c); accessToken != "" {
		service, err := h.getSheetsServiceWithToken(accessToken)
		if err == nil {
			if valuesResult, err := h.fetchGoogleSheetValuesWithService(service, sheetID, gid, req.Range); err == nil {
				matrix = converter.NewCellMatrix(valuesResult.Rows).Normalize()
				rangeOffsetCol = valuesResult.StartCol
				rangeOffsetRow = valuesResult.StartRow
				rangeSheetName = valuesResult.SheetName
			} else if isAuthError(err) {
				slog.Warn("mdflow.PreviewGoogleSheet auth error", "error", err)
			}
		}
	}

	// Fallback to backend API service (configured credentials)
	if matrix == nil {
		if valuesResult, err := h.fetchGoogleSheetValuesWithAPI(sheetID, gid, req.Range); err == nil {
			matrix = converter.NewCellMatrix(valuesResult.Rows).Normalize()
			rangeOffsetCol = valuesResult.StartCol
			rangeOffsetRow = valuesResult.StartRow
			rangeSheetName = valuesResult.SheetName
		}
	}

	// Last fallback: public export URL
	if matrix == nil {
		fetchedBody, statusCode, err := h.fetchGoogleSheetCSV(sheetID, gid, req.Range)
		if err != nil && statusCode == 0 {
			slog.Error("mdflow.PreviewGoogleSheet fetch error", "error", err)
			c.JSON(http.StatusBadGateway, ErrorResponse{Error: "failed to fetch Google Sheet"})
			return
		}
		if statusCode != http.StatusOK {
			slog.Warn("mdflow.PreviewGoogleSheet non-200", "status", statusCode, "error", err)
			if statusCode == http.StatusNotFound {
				c.JSON(http.StatusNotFound, ErrorResponse{Error: "Google Sheet not found or not public"})
				return
			}
			if statusCode == http.StatusUnauthorized {
				c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Google Sheet is private; connect Google in the app to access it"})
				return
			}
			c.JSON(http.StatusBadGateway, ErrorResponse{Error: fmt.Sprintf("Google Sheets returned status %d", statusCode)})
			return
		}

		parser := converter.NewPasteParser()
		parsedMatrix, parseErr := parser.Parse(string(fetchedBody))
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to parse sheet data"})
			return
		}
		matrix = parsedMatrix
	}

	if matrix.RowCount() == 0 {
		c.JSON(http.StatusOK, PreviewResponse{
			Headers:       []string{},
			Rows:          [][]string{},
			TotalRows:     0,
			PreviewRows:   0,
			HeaderRow:     -1,
			Confidence:    0,
			ColumnMapping: map[string]string{},
			UnmappedCols:  []string{},
			InputType:     "table",
		})
		return
	}

	templateName := req.Template
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	conv := h.getConverterForRequest(c)
	blocks := converter.DetectTableBlocks(matrix)
	type blockPreview struct {
		block         converter.MatrixBlock
		headers       []string
		rows          [][]string
		totalRows     int
		headerRow     int
		confidence    int
		columnMapping map[string]string
		unmapped      []string
		quality       converter.PreviewMappingQuality
		englishScore  float64
		languageHint  string
		rangeA1       string
	}

	previews := make([]blockPreview, 0, len(blocks))
	for _, block := range blocks {
		headerDetector := converter.NewHeaderDetector()
		headerRow, confidence := headerDetector.DetectHeaderRow(block.Matrix)
		headers := block.Matrix.GetRow(headerRow)
		dataRows := block.Matrix.SliceRows(headerRow+1, block.Matrix.RowCount())
		columnMapping, unmapped := conv.GetPreviewColumnMappingWithContext(ctx, headers, dataRows, templateName, "")
		quality := converter.BuildPreviewMappingQuality(confidence, headers, dataRows, columnMapping, unmapped)
		englishScore := converter.EstimateEnglishScore(headers, dataRows)
		languageHint := converter.DetectLanguageHint(englishScore, headers, dataRows)

		previewRows := dataRows
		if len(previewRows) > maxPreviewRows {
			previewRows = previewRows[:maxPreviewRows]
		}

		previews = append(previews, blockPreview{
			block:         block,
			headers:       headers,
			rows:          previewRows,
			totalRows:     len(dataRows),
			headerRow:     headerRow,
			confidence:    confidence,
			columnMapping: columnMapping,
			unmapped:      unmapped,
			quality:       quality,
			englishScore:  englishScore,
			languageHint:  languageHint,
			rangeA1:       qualifyRangeWithSheet(rangeSheetName, matrixBlockA1WithOffset(block.StartCol, block.EndCol, block.StartRow, block.EndRow, rangeOffsetCol, rangeOffsetRow)),
		})
	}

	if len(previews) == 0 {
		c.JSON(http.StatusOK, PreviewResponse{
			Headers:       []string{},
			Rows:          [][]string{},
			TotalRows:     0,
			PreviewRows:   0,
			HeaderRow:     -1,
			Confidence:    0,
			ColumnMapping: map[string]string{},
			UnmappedCols:  []string{},
			InputType:     "table",
		})
		return
	}

	hasMultiRowBlock := false
	for _, candidate := range previews {
		if candidate.totalRows >= 2 {
			hasMultiRowBlock = true
			break
		}
	}

	eligible := make([]int, 0, len(previews))
	for idx, candidate := range previews {
		if !hasMultiRowBlock || candidate.totalRows >= 2 {
			eligible = append(eligible, idx)
		}
	}
	if len(eligible) == 0 {
		eligible = append(eligible, 0)
	}

	candidates := make([]converter.BlockSelectionCandidate, 0, len(eligible))
	for _, idx := range eligible {
		candidate := previews[idx]
		candidates = append(candidates, converter.BlockSelectionCandidate{
			EnglishScore: candidate.englishScore,
			QualityScore: candidate.quality.Score,
			RowCount:     candidate.totalRows,
			ColumnCount:  len(candidate.headers),
		})
	}

	selectedLocalIdx := converter.SelectPreferredBlock(candidates)
	selectedIdx := eligible[0]
	if selectedLocalIdx >= 0 && selectedLocalIdx < len(eligible) {
		selectedIdx = eligible[selectedLocalIdx]
	}
	selected := previews[selectedIdx]

	responseBlocks := make([]PreviewBlock, 0, len(previews))
	for _, candidate := range previews {
		quality := candidate.quality
		responseBlocks = append(responseBlocks, PreviewBlock{
			ID:             candidate.block.ID,
			Range:          candidate.rangeA1,
			TotalRows:      candidate.totalRows,
			TotalColumns:   len(candidate.headers),
			LanguageHint:   candidate.languageHint,
			EnglishScore:   candidate.englishScore,
			HeaderRow:      candidate.headerRow,
			Confidence:     candidate.confidence,
			MappingQuality: &quality,
		})
	}

	quality := selected.quality
	c.JSON(http.StatusOK, PreviewResponse{
		Headers:            selected.headers,
		Rows:               selected.rows,
		TotalRows:          selected.totalRows,
		PreviewRows:        len(selected.rows),
		HeaderRow:          selected.headerRow,
		Confidence:         selected.confidence,
		ColumnMapping:      selected.columnMapping,
		UnmappedCols:       selected.unmapped,
		MappingQuality:     &quality,
		Blocks:             responseBlocks,
		SelectedBlockID:    selected.block.ID,
		SelectedBlockRange: selected.rangeA1,
		InputType:          "table",
	})
}

// ConvertGoogleSheet handles POST /api/mdflow/gsheet/convert
// Fetches and converts a public Google Sheet to MDFlow
func (h *MDFlowHandler) ConvertGoogleSheet(c *gin.Context) {
	var req GoogleSheetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "url is required"})
		return
	}
	columnOverrides := req.ColumnOverrides

	normalizedTemplate, normalizedFormat, err := normalizeTemplateAndFormat(req.Template, req.Format)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	req.Template = normalizedTemplate
	req.Format = normalizedFormat

	sheetID, gid, ok := parseGoogleSheetURL(req.URL)
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid Google Sheets URL"})
		return
	}

	gid = selectGID(req.GID, gid)
	if err := validateGID(gid); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	slog.Info("mdflow.ConvertGoogleSheet", "sheetID", sheetID, "gid", gid, "template", req.Template, "format", req.Format)

	conv := h.getConverterForRequest(c)

	if accessToken := getBearerToken(c); accessToken != "" {
		service, err := h.getSheetsServiceWithToken(accessToken)
		if err == nil {
			if result, stats, err := h.convertGoogleSheetWithService(c.Request.Context(), conv, service, sheetID, gid, req.Template, req.Format, req.Range, req.SelectedBlockID, columnOverrides); err == nil {
				result.Meta.QualityReport = h.buildQualityReport(stats, result)
				if validationErr := h.buildConvertValidationError(req.Format, stats, result); validationErr != nil {
					validationErr.Details["quality_report"] = result.Meta.QualityReport
					slog.Warn("mdflow.ConvertGoogleSheet validation failed", qualityReportLogArgs(result.Meta.QualityReport)...)
					c.JSON(http.StatusUnprocessableEntity, validationErr)
					return
				}
				result.Meta.SourceURL = req.URL
				slog.Info("mdflow.ConvertGoogleSheet quality", qualityReportLogArgs(result.Meta.QualityReport)...)
				slog.Info("mdflow.ConvertGoogleSheet ai", "ai_mode", result.Meta.AIMode, "ai_used", result.Meta.AIUsed, "ai_confidence", result.Meta.AIAvgConfidence)
				c.JSON(http.StatusOK, MDFlowConvertResponse{
					MDFlow:   result.MDFlow,
					Warnings: result.Warnings,
					Meta:     result.Meta,
					Format:   req.Format,
					Template: req.Template,
				})
				return
			} else if isAuthError(err) {
				slog.Warn("mdflow.ConvertGoogleSheet auth error", "error", err)
			}
		}
	}

	result, stats, err := h.convertGoogleSheetWithAPI(c.Request.Context(), conv, sheetID, gid, req.Template, req.Format, req.Range, req.SelectedBlockID, columnOverrides)
	if err == nil {
		result.Meta.QualityReport = h.buildQualityReport(stats, result)
		if validationErr := h.buildConvertValidationError(req.Format, stats, result); validationErr != nil {
			validationErr.Details["quality_report"] = result.Meta.QualityReport
			slog.Warn("mdflow.ConvertGoogleSheet validation failed", qualityReportLogArgs(result.Meta.QualityReport)...)
			c.JSON(http.StatusUnprocessableEntity, validationErr)
			return
		}
		result.Meta.SourceURL = req.URL
		slog.Info("mdflow.ConvertGoogleSheet quality", qualityReportLogArgs(result.Meta.QualityReport)...)
		slog.Info("mdflow.ConvertGoogleSheet ai", "ai_mode", result.Meta.AIMode, "ai_used", result.Meta.AIUsed, "ai_confidence", result.Meta.AIAvgConfidence)
		c.JSON(http.StatusOK, MDFlowConvertResponse{
			MDFlow:   result.MDFlow,
			Warnings: result.Warnings,
			Meta:     result.Meta,
			Format:   req.Format,
			Template: req.Template,
		})
		return
	}

	slog.Warn("mdflow.ConvertGoogleSheet API fallback", "error", err)
	body, statusCode, fetchErr := h.fetchGoogleSheetCSV(sheetID, gid, req.Range)
	if fetchErr != nil && statusCode == 0 {
		slog.Error("mdflow.ConvertGoogleSheet fetch error", "error", fetchErr)
		c.JSON(http.StatusBadGateway, ErrorResponse{Error: "failed to fetch Google Sheet"})
		return
	}
	if statusCode != http.StatusOK {
		slog.Warn("mdflow.ConvertGoogleSheet non-200", "status", statusCode, "error", fetchErr)
		if statusCode == http.StatusNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Google Sheet not found or not public"})
			return
		}
		if statusCode == http.StatusUnauthorized {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Google Sheet is private; connect Google in the app to access it"})
			return
		}
		c.JSON(http.StatusBadGateway, ErrorResponse{Error: fmt.Sprintf("Google Sheets returned status %d", statusCode)})
		return
	}

	// Create context with timeout to prevent hanging on slow AI calls
	ctx, cancel := context.WithTimeout(c.Request.Context(), 130*time.Second)
	defer cancel()

	parser := converter.NewPasteParser()
	matrix, parseErr := parser.Parse(string(body))
	if parseErr != nil {
		slog.Error("mdflow.ConvertGoogleSheet parse error", "error", parseErr)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to parse sheet data"})
		return
	}

	selectedMatrix := selectMatrixForConvert(ctx, conv, matrix, req.Template, req.SelectedBlockID, req.Range)
	stats = analyzeSelectedMatrix(selectedMatrix)
	result, err = conv.ConvertMatrixWithOverrides(ctx, selectedMatrix, gid, req.Template, req.Format, columnOverrides)
	if err != nil {
		slog.Error("mdflow.ConvertGoogleSheet conversion error", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to convert data"})
		return
	}
	result.Meta.QualityReport = h.buildQualityReport(stats, result)
	if validationErr := h.buildConvertValidationError(req.Format, stats, result); validationErr != nil {
		validationErr.Details["quality_report"] = result.Meta.QualityReport
		slog.Warn("mdflow.ConvertGoogleSheet validation failed", qualityReportLogArgs(result.Meta.QualityReport)...)
		c.JSON(http.StatusUnprocessableEntity, validationErr)
		return
	}

	result.Meta.SourceURL = req.URL
	slog.Info("mdflow.ConvertGoogleSheet quality", qualityReportLogArgs(result.Meta.QualityReport)...)
	c.JSON(http.StatusOK, MDFlowConvertResponse{
		MDFlow:   result.MDFlow,
		Warnings: result.Warnings,
		Meta:     result.Meta,
		Format:   req.Format,
		Template: req.Template,
	})
}
