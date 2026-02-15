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
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/ai"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/converter"
	"github.com/yourorg/md-spec-tool/internal/gsheetutils"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// GSheetHandler handles Google Sheets specific operations
// Extracted from MDFlowHandler for better separation of concerns
type GSheetHandler struct {
	converter        *converter.Converter
	renderer         *converter.MDFlowRenderer
	httpClient       *http.Client
	gsheetHTTPClient *http.Client
	gsheetClientOnce sync.Once
	cfg              *config.Config
	sheetsService    *sheets.Service
	sheetsInitOnce   sync.Once
	sheetsInitErr    error
	byokCache        *BYOKServiceCacheInterface    // Placeholder for dependency injection
	getAIService     func(string) (Service, error) // Injected AI service factory
}

// BYOKServiceCacheInterface defines the contract for BYOK caching
type BYOKServiceCacheInterface interface {
	Get(ctx context.Context, apiKey string) (Service, error)
	Close()
}

// Service defines the contract for AI services
type Service = ai.Service

// NewGSheetHandler creates a new GSheetHandler with injected dependencies
func NewGSheetHandler(
	conv *converter.Converter,
	rend *converter.MDFlowRenderer,
	httpClient *http.Client,
	cfg *config.Config,
	getAIService func(string) (Service, error),
) *GSheetHandler {
	if conv == nil {
		conv = converter.NewConverter()
	}
	if rend == nil {
		rend = converter.NewMDFlowRenderer()
	}
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	if cfg == nil {
		cfg = config.LoadConfig()
	}

	return &GSheetHandler{
		converter:     conv,
		renderer:      rend,
		httpClient:    httpClient,
		cfg:           cfg,
		getAIService:  getAIService,
		sheetsInitErr: nil,
		sheetsService: nil,
	}
}

// FetchGoogleSheet handles GET /api/v1/gsheet/fetch
// Fetches a Google Sheet and returns metadata
func (h *GSheetHandler) FetchGoogleSheet(c *gin.Context) {
	var req GoogleSheetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "url is required"})
		return
	}

	sheetID, gid, ok := gsheetutils.ParseGoogleSheetURL(req.URL)
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid Google Sheets URL"})
		return
	}

	gid = gsheetutils.SelectGID(req.GID, gid)
	if err := gsheetutils.ValidateGID(gid); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	slog.Info("gsheet.Fetch", "sheetID", sheetID, "gid", gid)

	// If access token provided, use authenticated request
	if accessToken := getBearerToken(c); accessToken != "" {
		service, err := h.getSheetsServiceWithToken(accessToken)
		if err == nil {
			// Attempt authenticated fetch
			if body, err := h.fetchGoogleSheetWithService(c.Request.Context(), service, sheetID, gid, req.Range); err == nil {
				c.JSON(http.StatusOK, GoogleSheetResponse{
					SheetID:   sheetID,
					SheetName: gid,
					Data:      string(body),
				})
				return
			}
			slog.Warn("gsheet.Fetch auth error", "error", err)
		}
	}

	// Fallback: public export URL
	body, statusCode, err := h.fetchGoogleSheetCSV(sheetID, gid, req.Range)
	if err != nil && statusCode == 0 {
		slog.Error("gsheet.Fetch fetch error", "error", err)
		c.JSON(http.StatusBadGateway, ErrorResponse{Error: "failed to fetch Google Sheet"})
		return
	}
	if statusCode != http.StatusOK {
		slog.Warn("gsheet.Fetch non-200", "status", statusCode, "error", err)
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

// GetGoogleSheetSheets handles GET /api/v1/gsheet/sheets
// Lists all sheets in a Google Spreadsheet
func (h *GSheetHandler) GetGoogleSheetSheets(c *gin.Context) {
	var req GoogleSheetSheetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "url is required"})
		return
	}

	sheetID, _, ok := gsheetutils.ParseGoogleSheetURL(req.URL)
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid Google Sheets URL"})
		return
	}

	slog.Info("gsheet.GetSheets", "sheetID", sheetID)

	// If access token provided, use authenticated request
	if accessToken := getBearerToken(c); accessToken != "" {
		service, err := h.getSheetsServiceWithToken(accessToken)
		if err == nil {
			if tabs, activeGID, err := h.getGoogleSheetTabsWithToken(c.Request.Context(), service, sheetID); err == nil {
				c.JSON(http.StatusOK, GoogleSheetSheetsResponse{
					Sheets:    tabs,
					ActiveGID: activeGID,
				})
				return
			}
			slog.Warn("gsheet.GetSheets auth error", "error", err)
		}
	}

	// Fallback: try public fetch
	service, err := h.getSheetsService()
	if err != nil {
		slog.Error("gsheet.GetSheets service error", "error", err)
		c.JSON(http.StatusBadGateway, ErrorResponse{Error: "Google Sheets not configured"})
		return
	}

	tabs, activeGID, err := h.getGoogleSheetTabsWithToken(c.Request.Context(), service, sheetID)
	if err != nil {
		slog.Error("gsheet.GetSheets error", "error", err)
		if isAuthError(err) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Google Sheet is private; connect Google in the app to access it"})
		} else {
			c.JSON(http.StatusBadGateway, ErrorResponse{Error: "failed to fetch sheet list"})
		}
		return
	}

	c.JSON(http.StatusOK, GoogleSheetSheetsResponse{
		Sheets:    tabs,
		ActiveGID: activeGID,
	})
}

// PreviewGoogleSheet handles POST /api/v1/gsheet/preview
// Generates a preview of a Google Sheet with block detection and column mapping
func (h *GSheetHandler) PreviewGoogleSheet(c *gin.Context) {
	var req GoogleSheetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "url is required"})
		return
	}

	normalizedTemplate, normalizedFormat, err := normalizeTemplateAndFormat(req.Template, req.Format)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	req.Template = normalizedTemplate
	req.Format = normalizedFormat

	sheetID, gid, ok := gsheetutils.ParseGoogleSheetURL(req.URL)
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid Google Sheets URL"})
		return
	}

	gid = gsheetutils.SelectGID(req.GID, gid)
	if err := gsheetutils.ValidateGID(gid); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	slog.Info("gsheet.Preview", "sheetID", sheetID, "gid", gid, "template", req.Template, "format", req.Format)
	conv := h.getConverterForRequest(c)

	// If access token provided, use authenticated request
	if accessToken := getBearerToken(c); accessToken != "" {
		service, err := h.getSheetsServiceWithToken(accessToken)
		if err == nil {
			if result, err := h.previewGoogleSheetWithService(c.Request.Context(), conv, service, sheetID, gid, req.Template, req.Format, req.Range, req.SelectedBlockID); err == nil {
				c.JSON(http.StatusOK, result)
				return
			}
			slog.Warn("gsheet.Preview auth error", "error", err)
		}
	}

	// Fallback: try backend API service when configured
	if service, err := h.getSheetsService(); err == nil {
		if result, err := h.previewGoogleSheetWithService(c.Request.Context(), conv, service, sheetID, gid, req.Template, req.Format, req.Range, req.SelectedBlockID); err == nil {
			c.JSON(http.StatusOK, result)
			return
		} else if isAuthError(err) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Google Sheet is private; connect Google in the app to access it"})
			return
		} else {
			slog.Warn("gsheet.Preview service fallback failed", "error", err)
		}
	} else if !errors.Is(err, errSheetsNotConfigured) {
		slog.Warn("gsheet.Preview service unavailable", "error", err)
	}

	// Last fallback: public export URL
	fetchedBody, statusCode, err := h.fetchGoogleSheetCSV(sheetID, gid, req.Range)
	if err != nil && statusCode == 0 {
		slog.Error("gsheet.Preview fetch error", "error", err)
		c.JSON(http.StatusBadGateway, ErrorResponse{Error: "failed to fetch Google Sheet"})
		return
	}
	if statusCode != http.StatusOK {
		slog.Warn("gsheet.Preview non-200", "status", statusCode, "error", err)
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
	matrix, parseErr := parser.Parse(string(fetchedBody))
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to parse sheet data"})
		return
	}

	preview := h.buildPreviewResponse(c.Request.Context(), conv, matrix, req.Template, req.SelectedBlockID, "", 0, 0)
	c.JSON(http.StatusOK, preview)
}

// ConvertGoogleSheet handles POST /api/v1/gsheet/convert
// Fetches and converts a Google Sheet to MDFlow
func (h *GSheetHandler) ConvertGoogleSheet(c *gin.Context) {
	var req GoogleSheetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "url is required"})
		return
	}
	columnOverrides := req.ColumnOverrides
	convertOptions := resolveConvertOptions(req.IncludeMetadata, req.NumberRows)

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

	slog.Info("gsheet.Convert", "sheetID", sheetID, "gid", gid, "template", req.Template, "format", req.Format)
	conv := h.getConverterForRequest(c)

	// If access token provided, use authenticated request
	if accessToken := getBearerToken(c); accessToken != "" {
		service, err := h.getSheetsServiceWithToken(accessToken)
		if err == nil {
			if result, stats, err := h.convertGoogleSheetWithService(c.Request.Context(), conv, service, sheetID, gid, req.Template, req.Format, req.Range, req.SelectedBlockID, columnOverrides, convertOptions); err == nil {
				result.Meta.QualityReport = h.buildQualityReport(stats, result)
				if validationErr := h.buildConvertValidationError(req.Format, stats, result); validationErr != nil {
					validationErr.Details["quality_report"] = result.Meta.QualityReport
					c.JSON(http.StatusUnprocessableEntity, validationErr)
					return
				}
				result.Meta.SourceURL = req.URL
				c.JSON(http.StatusOK, MDFlowConvertResponse{
					MDFlow:   result.MDFlow,
					Warnings: result.Warnings,
					Meta:     result.Meta,
					Format:   req.Format,
					Template: req.Template,
				})
				return
			}
			slog.Warn("gsheet.Convert auth error", "error", err)
		}
	}

	// Fallback: try public fetch
	result, stats, err := h.convertGoogleSheetWithFallback(c.Request.Context(), conv, sheetID, gid, req.URL, req.Template, req.Format, req.Range, req.SelectedBlockID, columnOverrides, convertOptions)
	if err != nil {
		slog.Error("gsheet.Convert error", "error", err)
		c.JSON(http.StatusBadGateway, ErrorResponse{Error: err.Error()})
		return
	}

	result.Meta.QualityReport = h.buildQualityReport(stats, result)
	if validationErr := h.buildConvertValidationError(req.Format, stats, result); validationErr != nil {
		validationErr.Details["quality_report"] = result.Meta.QualityReport
		c.JSON(http.StatusUnprocessableEntity, validationErr)
		return
	}
	result.Meta.SourceURL = req.URL

	c.JSON(http.StatusOK, MDFlowConvertResponse{
		MDFlow:   result.MDFlow,
		Warnings: result.Warnings,
		Meta:     result.Meta,
		Format:   req.Format,
		Template: req.Template,
	})
}

// getSheetsService returns a shared Google Sheets service (lazy initialized)
func (h *GSheetHandler) getSheetsService() (*sheets.Service, error) {
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

// getSheetsServiceWithToken returns a service with a user's access token
func (h *GSheetHandler) getSheetsServiceWithToken(accessToken string) (*sheets.Service, error) {
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

// fetchGoogleSheetWithService fetches sheet data and returns CSV bytes
func (h *GSheetHandler) fetchGoogleSheetWithService(ctx context.Context, service *sheets.Service, sheetID, gid, rangeOverride string) ([]byte, error) {
	valuesResult, err := h.fetchGoogleSheetValuesWithService(ctx, service, sheetID, gid, rangeOverride)
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

// getGoogleSheetTabsWithToken retrieves sheet tabs and active GID
func (h *GSheetHandler) getGoogleSheetTabsWithToken(ctx context.Context, service *sheets.Service, sheetID string) ([]GoogleSheetTab, string, error) {
	tabs, err := getGoogleSheetTabsWithService(service, sheetID)
	if err != nil {
		return nil, "", err
	}

	// Default to first tab as active
	activeGID := ""
	if len(tabs) > 0 {
		activeGID = tabs[0].GID
	}

	return tabs, activeGID, nil
}

// previewGoogleSheetWithService generates a preview with block detection
func (h *GSheetHandler) previewGoogleSheetWithService(ctx context.Context, conv *converter.Converter, service *sheets.Service, sheetID, gid, template, format, rangeStr, selectedBlockID string) (interface{}, error) {
	valuesResult, err := h.fetchGoogleSheetValuesWithService(ctx, service, sheetID, gid, rangeStr)
	if err != nil {
		return nil, err
	}

	matrix := converter.NewCellMatrix(valuesResult.Rows).Normalize()
	return h.buildPreviewResponse(ctx, conv, matrix, template, selectedBlockID, valuesResult.SheetName, valuesResult.StartCol, valuesResult.StartRow), nil
}

func (h *GSheetHandler) buildPreviewResponse(ctx context.Context, conv *converter.Converter, matrix converter.CellMatrix, template, selectedBlockID, rangeSheetName string, rangeOffsetCol, rangeOffsetRow int) PreviewResponse {
	if matrix.RowCount() == 0 {
		return PreviewResponse{
			Headers:       []string{},
			Rows:          [][]string{},
			TotalRows:     0,
			PreviewRows:   0,
			HeaderRow:     -1,
			Confidence:    0,
			ColumnMapping: map[string]string{},
			UnmappedCols:  []string{},
			InputType:     "table",
		}
	}

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
		columnMapping, unmapped := conv.GetPreviewColumnMappingWithContext(ctx, headers, dataRows, template, "")
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
		return PreviewResponse{
			Headers:       []string{},
			Rows:          [][]string{},
			TotalRows:     0,
			PreviewRows:   0,
			HeaderRow:     -1,
			Confidence:    0,
			ColumnMapping: map[string]string{},
			UnmappedCols:  []string{},
			InputType:     "table",
		}
	}

	selectedIdx := -1
	if strings.TrimSpace(selectedBlockID) != "" {
		for idx, candidate := range previews {
			if candidate.block.ID == selectedBlockID {
				selectedIdx = idx
				break
			}
		}
	}

	if selectedIdx == -1 {
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
		selectedIdx = eligible[0]
		if selectedLocalIdx >= 0 && selectedLocalIdx < len(eligible) {
			selectedIdx = eligible[selectedLocalIdx]
		}
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
	return PreviewResponse{
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
	}
}

// convertGoogleSheetWithService converts a sheet using authenticated service
func (h *GSheetHandler) convertGoogleSheetWithService(ctx context.Context, conv *converter.Converter, service *sheets.Service, sheetID, gid, template, format, rangeStr, selectedBlockID string, columnOverrides map[string]string, options converter.ConvertOptions) (*converter.ConvertResponse, convertValidationStats, error) {
	valuesResult, err := h.fetchGoogleSheetValuesWithService(ctx, service, sheetID, gid, rangeStr)
	if err != nil {
		return nil, convertValidationStats{}, err
	}

	matrix := converter.NewCellMatrix(valuesResult.Rows).Normalize()
	selected := selectMatrixForConvert(ctx, conv, matrix, template, selectedBlockID, rangeStr)
	stats := analyzeSelectedMatrix(selected)

	result, err := conv.ConvertMatrixWithOverridesAndOptions(ctx, selected, valuesResult.SheetName, template, format, columnOverrides, options)
	if err != nil {
		return nil, stats, err
	}

	return result, stats, nil
}

// convertGoogleSheetWithFallback converts using public CSV export fallback
func (h *GSheetHandler) convertGoogleSheetWithFallback(ctx context.Context, conv *converter.Converter, sheetID, gid, urlStr, template, format, rangeStr, selectedBlockID string, columnOverrides map[string]string, options converter.ConvertOptions) (*converter.ConvertResponse, convertValidationStats, error) {
	body, statusCode, fetchErr := h.fetchGoogleSheetCSV(sheetID, gid, rangeStr)
	if fetchErr != nil && statusCode == 0 {
		slog.Error("gsheet.ConvertFallback fetch error", "error", fetchErr)
		return nil, convertValidationStats{}, fmt.Errorf("failed to fetch Google Sheet")
	}
	if statusCode != http.StatusOK {
		slog.Warn("gsheet.ConvertFallback non-200", "status", statusCode, "error", fetchErr)
		if statusCode == http.StatusNotFound {
			return nil, convertValidationStats{}, fmt.Errorf("Google Sheet not found or not public")
		}
		if statusCode == http.StatusUnauthorized {
			return nil, convertValidationStats{}, fmt.Errorf("Google Sheet is private; connect Google in the app to access it")
		}
		return nil, convertValidationStats{}, fmt.Errorf("Google Sheets returned status %d", statusCode)
	}

	parser := converter.NewPasteParser()
	matrix, parseErr := parser.Parse(string(body))
	if parseErr != nil {
		slog.Error("gsheet.ConvertFallback parse error", "error", parseErr)
		return nil, convertValidationStats{}, fmt.Errorf("failed to parse sheet data")
	}

	selected := selectMatrixForConvert(ctx, conv, matrix, template, selectedBlockID, rangeStr)
	stats := analyzeSelectedMatrix(selected)
	result, err := conv.ConvertMatrixWithOverridesAndOptions(ctx, selected, gid, template, format, columnOverrides, options)
	if err != nil {
		slog.Error("gsheet.ConvertFallback conversion error", "error", err)
		return nil, stats, fmt.Errorf("failed to convert data")
	}

	return result, stats, nil
}

func (h *GSheetHandler) getConverterForRequest(c *gin.Context) *converter.Converter {
	userKey := getUserAPIKey(c)
	if userKey == "" {
		return h.converter
	}
	if h.getAIService == nil {
		return converter.NewConverter()
	}

	aiService, err := h.getAIService(userKey)
	if err != nil || aiService == nil {
		slog.Warn("BYOK: failed to create AI service for gsheet", "error", err)
		return converter.NewConverter()
	}

	conv := converter.NewConverter()
	conv.WithAIService(aiService)
	return conv
}

func (h *GSheetHandler) buildQualityReport(stats convertValidationStats, result *converter.ConvertResponse) *converter.QualityReport {
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

func (h *GSheetHandler) buildConvertValidationError(format string, stats convertValidationStats, result *converter.ConvertResponse) *ErrorResponse {
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

// fetchGoogleSheetValuesWithService fetches raw cell values from a sheet
func (h *GSheetHandler) fetchGoogleSheetValuesWithService(ctx context.Context, service *sheets.Service, sheetID, gid, rangeOverride string) (*gsheetValuesResult, error) {
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

// fetchGoogleSheetCSV fetches a sheet as CSV via the public export URL
func (h *GSheetHandler) fetchGoogleSheetCSV(sheetID string, gid string, rangeOverride string) ([]byte, int, error) {
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

		lastStatus = resp.StatusCode
		lastErr = fmt.Errorf("status %d", resp.StatusCode)
		resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusUnauthorized {
			return nil, resp.StatusCode, lastErr
		}
	}

	if lastStatus == 0 {
		return nil, 0, lastErr
	}
	return nil, lastStatus, lastErr
}

// getGSheetHTTPClient returns a client with appropriate timeout for sheet exports
func (h *GSheetHandler) getGSheetHTTPClient() *http.Client {
	if h.gsheetHTTPClient != nil {
		return h.gsheetHTTPClient
	}

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
	return h.gsheetHTTPClient
}
