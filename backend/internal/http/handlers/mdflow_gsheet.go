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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/converter"
	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// GoogleSheetRequest represents the request for Google Sheet import
type GoogleSheetRequest struct {
	URL      string `json:"url" binding:"required"`
	Template string `json:"template"`
	Format   string `json:"format"` // "spec" | "table"
	GID      string `json:"gid,omitempty"`
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

// parseGoogleSheetURL extracts the sheet ID from various Google Sheets URL formats
// Returns sheetID, gid, and a boolean indicating success
func parseGoogleSheetURL(urlStr string) (sheetID string, gid string, ok bool) {
	// Format 1: https://docs.google.com/spreadsheets/d/SHEET_ID/edit#gid=GID
	// Format 2: https://docs.google.com/spreadsheets/d/SHEET_ID/edit
	// Format 3: https://docs.google.com/spreadsheets/d/SHEET_ID

	// Parse URL properly
	u, err := url.Parse(urlStr)
	if err != nil {
		slog.Warn("Invalid URL format", "url", urlStr, "error", err)
		return "", "", false
	}

	// Verify host is docs.google.com
	if u.Host != "docs.google.com" {
		slog.Warn("Not a Google Docs URL", "host", u.Host)
		return "", "", false
	}

	// Path should contain /spreadsheets/d/{id}
	// Use regex to extract SHEET_ID pattern (alphanumeric, hyphens, underscores)
	sheetIDPattern := regexp.MustCompile(`/spreadsheets/d/([a-zA-Z0-9\-_]+)`)
	matches := sheetIDPattern.FindStringSubmatch(u.Path)

	if len(matches) < 2 {
		slog.Warn("Sheet ID not found in URL path", "path", u.Path)
		return "", "", false
	}

	sheetID = matches[1]

	// Validate sheet ID length (Google Sheet IDs are typically 40+ chars but vary)
	if len(sheetID) == 0 {
		return "", "", false
	}

	// Extract gid from fragment or query parameter
	// Fragment takes precedence (e.g., #gid=123)
	if u.Fragment != "" {
		gidPattern := regexp.MustCompile(`gid=(\d+)`)
		gidMatches := gidPattern.FindStringSubmatch(u.Fragment)
		if len(gidMatches) >= 2 {
			gid = gidMatches[1]
		}
	}

	// If not found in fragment, check query parameters
	if gid == "" {
		gid = u.Query().Get("gid")
	}

	return sheetID, gid, true
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

func selectGID(requestGID string, urlGID string) string {
	requestGID = strings.TrimSpace(requestGID)
	if requestGID != "" {
		return requestGID
	}
	return urlGID
}

func validateGID(gid string) error {
	if gid == "" {
		return nil
	}
	if _, err := strconv.ParseInt(gid, 10, 64); err != nil {
		return fmt.Errorf("gid must be numeric")
	}
	return nil
}

func getGoogleSheetTabsWithService(service *sheets.Service, spreadsheetID string) ([]GoogleSheetTab, error) {
	resp, err := service.Spreadsheets.Get(spreadsheetID).
		Fields("sheets.properties.sheetId,sheets.properties.title").
		Do()
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
func (h *MDFlowHandler) fetchGoogleSheetWithService(service *sheets.Service, sheetID string, gid string) ([]byte, error) {
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
	rangeStr := sheetRange(sheetTitle)
	resp, err := service.Spreadsheets.Values.Get(sheetID, rangeStr).Do()
	if err != nil {
		return nil, err
	}
	rows := convertValuesToRows(resp.Values)
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	for _, row := range rows {
		_ = w.Write(row)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (h *MDFlowHandler) fetchGoogleSheetCSV(sheetID string, gid string) ([]byte, int, error) {
	exportURL := fmt.Sprintf(googleSheetsExportURLFmt, sheetID)
	if gid != "" {
		exportURL += "&gid=" + url.QueryEscape(gid)
	}

	resp, err := h.httpClient.Get(exportURL)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("google sheets export returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, h.cfg.MaxUploadBytes))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return body, resp.StatusCode, nil
}

func (h *MDFlowHandler) convertGoogleSheetWithService(ctx context.Context, conv *converter.Converter, service *sheets.Service, sheetID string, gid string, templateName string, outputFormat string) (*converter.ConvertResponse, error) {
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
	rangeStr := sheetRange(sheetTitle)
	resp, err := service.Spreadsheets.Values.Get(sheetID, rangeStr).Do()
	if err != nil {
		return nil, err
	}
	rows := convertValuesToRows(resp.Values)
	matrix := converter.NewCellMatrix(rows).Normalize()
	return conv.ConvertMatrixWithFormatContext(ctx, matrix, sheetTitle, templateName, outputFormat)
}

func (h *MDFlowHandler) convertGoogleSheetWithAPI(ctx context.Context, conv *converter.Converter, sheetID string, gid string, templateName string, outputFormat string) (*converter.ConvertResponse, error) {
	service, err := h.getSheetsService()
	if err != nil {
		return nil, err
	}

	return h.convertGoogleSheetWithService(ctx, conv, service, sheetID, gid, templateName, outputFormat)
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
			body, err := h.fetchGoogleSheetWithService(service, sheetID, gid)
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
	body, statusCode, err := h.fetchGoogleSheetCSV(sheetID, gid)
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

	var body []byte

	// Prefer OAuth when client sends Bearer token
	if accessToken := getBearerToken(c); accessToken != "" {
		service, err := h.getSheetsServiceWithToken(accessToken)
		if err == nil {
			if fetchedBody, err := h.fetchGoogleSheetWithService(service, sheetID, gid); err == nil {
				body = fetchedBody
			} else if isAuthError(err) {
				slog.Warn("mdflow.PreviewGoogleSheet auth error", "error", err)
			}
		}
	}

	// Fallback: public export URL
	if body == nil {
		fetchedBody, statusCode, err := h.fetchGoogleSheetCSV(sheetID, gid)
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
		body = fetchedBody
	}

	// Parse as table
	parser := converter.NewPasteParser()
	matrix, err := parser.Parse(string(body))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to parse sheet data"})
		return
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

	// Detect header row
	headerDetector := converter.NewHeaderDetector()
	headerRow, confidence := headerDetector.DetectHeaderRow(matrix)
	headers := matrix.GetRow(headerRow)

	// Map columns via template-driven HeaderResolver with AI (Phase 3)
	templateName := req.Template
	dataRows := matrix.SliceRows(headerRow+1, matrix.RowCount())

	// AI-enabled preview: use shorter timeout (15s) to maintain responsiveness
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	conv := h.getConverterForRequest(c)
	columnMapping, unmapped := conv.GetPreviewColumnMappingWithContext(ctx, headers, dataRows, templateName, "")

	// Get preview rows (after header)
	totalDataRows := matrix.RowCount() - headerRow - 1
	previewCount := totalDataRows
	if previewCount > maxPreviewRows {
		previewCount = maxPreviewRows
	}

	rows := make([][]string, 0, previewCount)
	for i := headerRow + 1; i < headerRow+1+previewCount && i < matrix.RowCount(); i++ {
		rows = append(rows, matrix.GetRow(i))
	}

	c.JSON(http.StatusOK, PreviewResponse{
		Headers:       headers,
		Rows:          rows,
		TotalRows:     totalDataRows,
		PreviewRows:   previewCount,
		HeaderRow:     headerRow,
		Confidence:    confidence,
		ColumnMapping: columnMapping,
		UnmappedCols:  unmapped,
		InputType:     "table",
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
			if result, err := h.convertGoogleSheetWithService(c.Request.Context(), conv, service, sheetID, gid, req.Template, req.Format); err == nil {
				result.Meta.SourceURL = req.URL
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

	result, err := h.convertGoogleSheetWithAPI(c.Request.Context(), conv, sheetID, gid, req.Template, req.Format)
	if err == nil {
		result.Meta.SourceURL = req.URL
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
	body, statusCode, fetchErr := h.fetchGoogleSheetCSV(sheetID, gid)
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

	result, err = conv.ConvertPasteWithFormatContext(ctx, string(body), req.Template, req.Format)
	if err != nil {
		slog.Error("mdflow.ConvertGoogleSheet conversion error", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to convert data"})
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
