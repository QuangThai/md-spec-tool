package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

type MDFlowHandler struct {
	converter *converter.Converter
	renderer  *converter.MDFlowRenderer
}

func NewMDFlowHandler() *MDFlowHandler {
	return &MDFlowHandler{
		converter: converter.NewConverter(),
		renderer:  converter.NewMDFlowRenderer(),
	}
}

const (
	maxPasteBytes     = 1 << 20
	maxPasteBodyBytes = maxPasteBytes + (4 << 10)
	maxUploadBytes    = 10 << 20
	maxUploadBody     = maxUploadBytes + (1 << 20)
	maxTemplateLen    = 64
	maxSheetNameLen   = 128
)

var xlsxMagic = []byte{0x50, 0x4B, 0x03, 0x04}

// PasteConvertRequest represents the request for paste conversion
type PasteConvertRequest struct {
	PasteText string `json:"paste_text" binding:"required"`
	Template  string `json:"template"`
}

// XLSXConvertRequest represents the request for XLSX sheet conversion
type XLSXConvertRequest struct {
	SheetName string `json:"sheet_name"`
	Template  string `json:"template"`
}

// MDFlowConvertResponse represents the conversion response
type MDFlowConvertResponse struct {
	MDFlow   string                `json:"mdflow"`
	Warnings []converter.Warning   `json:"warnings"`
	Meta     converter.SpecDocMeta `json:"meta"`
}

// InputAnalysisResponse represents the input type detection response
type InputAnalysisResponse struct {
	Type       string  `json:"type"` // 'markdown' | 'table' | 'unknown'
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason,omitempty"`
}

// SheetsResponse represents the list of sheets
type SheetsResponse struct {
	Sheets      []string `json:"sheets"`
	ActiveSheet string   `json:"active_sheet"`
}

// PreviewResponse represents the table preview before conversion
type PreviewResponse struct {
	Headers       []string            `json:"headers"`
	Rows          [][]string          `json:"rows"`
	TotalRows     int                 `json:"total_rows"`
	PreviewRows   int                 `json:"preview_rows"`
	HeaderRow     int                 `json:"header_row"`
	Confidence    int                 `json:"confidence"`
	ColumnMapping map[string]string   `json:"column_mapping"`
	UnmappedCols  []string            `json:"unmapped_columns"`
	InputType     string              `json:"input_type"`
}

// ConvertPaste handles POST /api/mdflow/paste
// If detect_only=true query param, returns input type analysis
// Otherwise converts pasted TSV/CSV text to MDFlow format
func (h *MDFlowHandler) ConvertPaste(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxPasteBodyBytes)

	var req PasteConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "request body exceeds limit"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "paste_text is required"})
		return
	}

	// Check size limit (1MB)
	if len(req.PasteText) > maxPasteBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "paste_text exceeds 1MB limit"})
		return
	}

	req.Template = strings.TrimSpace(req.Template)
	if err := h.validateTemplate(req.Template); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Detect input type FIRST regardless of request
	analysis := converter.DetectInputType(req.PasteText)

	// Check if this is detection-only request
	detectOnly, err := strconv.ParseBool(c.DefaultQuery("detect_only", "false"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid detect_only value"})
		return
	}

	log.Printf(
		"mdflow.ConvertPaste detectOnly=%t template=%q pasteBytes=%d type=%s confidence=%d",
		detectOnly,
		req.Template,
		len(req.PasteText),
		string(analysis.Type),
		analysis.Confidence,
	)
	if detectOnly {
		typeStr := "unknown"
		switch analysis.Type {
		case converter.InputTypeMarkdown:
			typeStr = "markdown"
		case converter.InputTypeTable:
			typeStr = "table"
		}

		c.JSON(http.StatusOK, InputAnalysisResponse{
			Type:       typeStr,
			Confidence: float64(analysis.Confidence),
			Reason:     analysis.Reason,
		})
		return
	}

	// Full conversion
	result, err := h.converter.ConvertPaste(req.PasteText, req.Template)
	if err != nil {
		log.Printf("mdflow.ConvertPaste failed: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to convert input"})
		return
	}

	c.JSON(http.StatusOK, MDFlowConvertResponse{
		MDFlow:   result.MDFlow,
		Warnings: result.Warnings,
		Meta:     result.Meta,
	})
}

// ConvertXLSX handles POST /api/mdflow/xlsx
// Converts uploaded XLSX file to MDFlow format
func (h *MDFlowHandler) ConvertXLSX(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBody)

	// Get file from multipart form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "file exceeds 10MB limit"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	// Check file size limit (10MB)
	if header.Size > maxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "file exceeds 10MB limit"})
		return
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".xlsx" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "only .xlsx files are supported"})
		return
	}

	// Get optional parameters
	sheetName := strings.TrimSpace(c.PostForm("sheet_name"))
	template := strings.TrimSpace(c.PostForm("template"))

	if err := h.validateTemplate(template); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := validateSheetName(sheetName); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	buf := make([]byte, 4)
	n, err := io.ReadFull(file, buf)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is empty"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to read file"})
		return
	}
	if !bytes.HasPrefix(buf[:n], xlsxMagic) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid xlsx file"})
		return
	}

	reader := io.MultiReader(bytes.NewReader(buf[:n]), file)

	// Create temp file
	tempFile, err := os.CreateTemp("", "upload-*.xlsx")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to process file"})
		return
	}
	tempName := tempFile.Name()
	defer os.Remove(tempName)

	// Copy uploaded file to temp
	bytesCopied, err := io.Copy(tempFile, io.LimitReader(reader, maxUploadBytes+1))
	if err != nil {
		_ = tempFile.Close()
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save file"})
		return
	}
	if bytesCopied > maxUploadBytes {
		_ = tempFile.Close()
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "file exceeds 10MB limit"})
		return
	}
	if err := tempFile.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to process file"})
		return
	}

	result, err := h.converter.ConvertXLSX(tempName, sheetName, template)
	if err != nil {
		log.Printf("mdflow.ConvertXLSX failed: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to convert file"})
		return
	}

	c.JSON(http.StatusOK, MDFlowConvertResponse{
		MDFlow:   result.MDFlow,
		Warnings: result.Warnings,
		Meta:     result.Meta,
	})
}

// ConvertTSV handles POST /api/mdflow/tsv
// Converts uploaded TSV file to MDFlow format
func (h *MDFlowHandler) ConvertTSV(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBody)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "file exceeds 10MB limit"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	if header.Size > maxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "file exceeds 10MB limit"})
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".tsv" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "only .tsv files are supported"})
		return
	}

	template := strings.TrimSpace(c.PostForm("template"))
	if err := h.validateTemplate(template); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	content, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to read file"})
		return
	}
	if len(content) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is empty"})
		return
	}
	if len(content) > maxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "file exceeds 10MB limit"})
		return
	}

	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		content = content[3:]
	}

	result, err := h.converter.ConvertPaste(string(content), template)
	if err != nil {
		log.Printf("mdflow.ConvertTSV failed: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to convert file"})
		return
	}

	c.JSON(http.StatusOK, MDFlowConvertResponse{
		MDFlow:   result.MDFlow,
		Warnings: result.Warnings,
		Meta:     result.Meta,
	})
}

// GetXLSXSheets handles POST /api/mdflow/xlsx/sheets
// Returns list of sheets in uploaded XLSX file
func (h *MDFlowHandler) GetXLSXSheets(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBody)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "file exceeds 10MB limit"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	// Check file size limit
	if header.Size > maxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "file exceeds 10MB limit"})
		return
	}

	buf := make([]byte, 4)
	n, err := io.ReadFull(file, buf)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is empty"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to read file"})
		return
	}
	if !bytes.HasPrefix(buf[:n], xlsxMagic) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid xlsx file"})
		return
	}

	reader := io.MultiReader(bytes.NewReader(buf[:n]), file)

	// Create temp file
	tempFile, err := os.CreateTemp("", "upload-*.xlsx")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to process file"})
		return
	}
	tempName := tempFile.Name()
	defer os.Remove(tempName)

	bytesCopied, err := io.Copy(tempFile, io.LimitReader(reader, maxUploadBytes+1))
	if err != nil {
		_ = tempFile.Close()
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save file"})
		return
	}
	if bytesCopied > maxUploadBytes {
		_ = tempFile.Close()
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "file exceeds 10MB limit"})
		return
	}
	if err := tempFile.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to process file"})
		return
	}

	sheets, err := h.converter.GetXLSXSheets(tempName)
	if err != nil {
		log.Printf("mdflow.GetXLSXSheets failed: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read file"})
		return
	}

	activeSheet := ""
	if len(sheets) > 0 {
		activeSheet = sheets[0]
	}

	c.JSON(http.StatusOK, SheetsResponse{
		Sheets:      sheets,
		ActiveSheet: activeSheet,
	})
}

// ValidateRequest represents the request for validation with custom rules
type ValidateRequest struct {
	PasteText       string                  `json:"paste_text" binding:"required"`
	ValidationRules *converter.ValidationRules `json:"validation_rules"`
}

// Validate handles POST /api/mdflow/validate
// Builds SpecDoc from paste_text and runs custom validation rules
func (h *MDFlowHandler) Validate(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxPasteBodyBytes)

	var req ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "paste_text is required"})
		return
	}

	if len(req.PasteText) > maxPasteBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "paste_text exceeds 1MB limit"})
		return
	}

	specDoc, err := converter.BuildSpecDocFromPaste(req.PasteText)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to parse input: " + err.Error()})
		return
	}

	rules := req.ValidationRules
	if rules == nil {
		rules = &converter.ValidationRules{}
	}

	result := converter.Validate(specDoc, rules)
	c.JSON(http.StatusOK, result)
}

// GetTemplates handles GET /api/mdflow/templates
// Returns available MDFlow templates
func (h *MDFlowHandler) GetTemplates(c *gin.Context) {
	templates := h.renderer.GetTemplateNames()

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
	})
}

// TemplatePreviewRequest represents the request for custom template preview
type TemplatePreviewRequest struct {
	TemplateContent string `json:"template_content" binding:"required"`
	SampleData      string `json:"sample_data"`
}

// TemplatePreviewResponse represents the response for custom template preview
type TemplatePreviewResponse struct {
	Output   string            `json:"output"`
	Error    string            `json:"error,omitempty"`
	Warnings []converter.Warning `json:"warnings"`
}

// PreviewTemplate handles POST /api/mdflow/templates/preview
// Renders sample data using a custom template
func (h *MDFlowHandler) PreviewTemplate(c *gin.Context) {
	var req TemplatePreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "template_content is required"})
		return
	}

	// Limit template size
	if len(req.TemplateContent) > 50000 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "template_content exceeds 50KB limit"})
		return
	}

	// Use sample data or default sample
	sampleData := req.SampleData
	if sampleData == "" {
		sampleData = defaultSampleData
	}

	// Parse sample data to SpecDoc
	specDoc, err := converter.BuildSpecDocFromPaste(sampleData)
	if err != nil {
		c.JSON(http.StatusOK, TemplatePreviewResponse{
			Output: "",
			Error:  "Failed to parse sample data: " + err.Error(),
		})
		return
	}

	// Render with custom template
	output, err := h.renderer.RenderCustom(specDoc, req.TemplateContent)
	if err != nil {
		c.JSON(http.StatusOK, TemplatePreviewResponse{
			Output: "",
			Error:  "Template error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, TemplatePreviewResponse{
		Output:   output,
		Warnings: specDoc.Warnings,
	})
}

// GetTemplateInfo handles GET /api/mdflow/templates/info
// Returns available template variables and functions
func (h *MDFlowHandler) GetTemplateInfo(c *gin.Context) {
	info := h.renderer.GetTemplateInfo()
	c.JSON(http.StatusOK, info)
}

// GetTemplateContent handles GET /api/mdflow/templates/:name
// Returns the content of a built-in template
func (h *MDFlowHandler) GetTemplateContent(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "template name is required"})
		return
	}

	content := h.renderer.GetTemplateContent(name)
	if content == "" {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "template not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":    name,
		"content": content,
	})
}

// Default sample data for template preview
const defaultSampleData = `Feature	Scenario	Instructions	Expected	Priority	Type	Notes
User Authentication	Valid Login	1. Enter username
2. Enter password
3. Click login button	Dashboard should display with user name	High	Positive	Core feature
User Authentication	Invalid Password	1. Enter valid username
2. Enter wrong password
3. Click login button	Error message: "Invalid credentials"	High	Negative	Security test
Profile Management	Update Profile	1. Go to settings
2. Change display name
3. Click save	Profile updated successfully message shown	Medium	Positive	
Profile Management	Upload Avatar	1. Click avatar
2. Select image file
3. Confirm upload	New avatar displayed	Low	Positive	Max 5MB`

func (h *MDFlowHandler) validateTemplate(template string) error {
	if template == "" {
		return nil
	}
	if len(template) > maxTemplateLen {
		return fmt.Errorf("template exceeds %d characters", maxTemplateLen)
	}
	if !h.renderer.HasTemplate(template) {
		return fmt.Errorf("unknown template")
	}
	return nil
}

func validateSheetName(sheetName string) error {
	if sheetName == "" {
		return nil
	}
	if len(sheetName) > maxSheetNameLen {
		return fmt.Errorf("sheet_name exceeds %d characters", maxSheetNameLen)
	}
	if strings.IndexFunc(sheetName, unicode.IsControl) != -1 {
		return fmt.Errorf("sheet_name contains invalid characters")
	}
	return nil
}

const maxPreviewRows = 20

// GoogleSheetRequest represents the request for Google Sheet import
type GoogleSheetRequest struct {
	URL      string `json:"url" binding:"required"`
	Template string `json:"template"`
}

// GoogleSheetResponse represents the response for Google Sheet parsing
type GoogleSheetResponse struct {
	SheetID   string `json:"sheet_id"`
	SheetName string `json:"sheet_name,omitempty"`
	Data      string `json:"data"`
}

// parseGoogleSheetURL extracts the sheet ID from various Google Sheets URL formats
func parseGoogleSheetURL(url string) (sheetID string, gid string, ok bool) {
	// Format 1: https://docs.google.com/spreadsheets/d/SHEET_ID/edit#gid=GID
	// Format 2: https://docs.google.com/spreadsheets/d/SHEET_ID/edit
	// Format 3: https://docs.google.com/spreadsheets/d/SHEET_ID
	
	if !strings.Contains(url, "docs.google.com/spreadsheets") {
		return "", "", false
	}
	
	// Extract sheet ID
	parts := strings.Split(url, "/d/")
	if len(parts) < 2 {
		return "", "", false
	}
	
	idPart := parts[1]
	// Remove anything after the ID (like /edit, /export, etc.)
	if idx := strings.Index(idPart, "/"); idx != -1 {
		idPart = idPart[:idx]
	}
	
	sheetID = idPart
	
	// Extract gid if present
	if idx := strings.Index(url, "gid="); idx != -1 {
		gidPart := url[idx+4:]
		if endIdx := strings.IndexAny(gidPart, "&# "); endIdx != -1 {
			gid = gidPart[:endIdx]
		} else {
			gid = gidPart
		}
	}
	
	return sheetID, gid, true
}

// FetchGoogleSheet handles POST /api/mdflow/gsheet
// Fetches data from a public Google Sheet and returns it as CSV/TSV
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
	
	// Build export URL for CSV
	exportURL := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/export?format=csv", sheetID)
	if gid != "" {
		exportURL += "&gid=" + gid
	}
	
	log.Printf("mdflow.FetchGoogleSheet sheetID=%s gid=%s", sheetID, gid)
	
	// Fetch the sheet
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Get(exportURL)
	if err != nil {
		log.Printf("mdflow.FetchGoogleSheet fetch error: %v", err)
		c.JSON(http.StatusBadGateway, ErrorResponse{Error: "failed to fetch Google Sheet"})
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Google Sheet not found or not public"})
			return
		}
		c.JSON(http.StatusBadGateway, ErrorResponse{Error: fmt.Sprintf("Google Sheets returned status %d", resp.StatusCode)})
		return
	}
	
	// Read body with size limit
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxUploadBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read response"})
		return
	}
	
	// Return the CSV data
	c.JSON(http.StatusOK, GoogleSheetResponse{
		SheetID:   sheetID,
		SheetName: gid,
		Data:      string(body),
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
	
	if err := h.validateTemplate(req.Template); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	
	sheetID, gid, ok := parseGoogleSheetURL(req.URL)
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid Google Sheets URL"})
		return
	}
	
	// Build export URL for CSV
	exportURL := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/export?format=csv", sheetID)
	if gid != "" {
		exportURL += "&gid=" + gid
	}
	
	log.Printf("mdflow.ConvertGoogleSheet sheetID=%s gid=%s template=%s", sheetID, gid, req.Template)
	
	// Fetch the sheet
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Get(exportURL)
	if err != nil {
		log.Printf("mdflow.ConvertGoogleSheet fetch error: %v", err)
		c.JSON(http.StatusBadGateway, ErrorResponse{Error: "failed to fetch Google Sheet"})
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Google Sheet not found or not public"})
			return
		}
		c.JSON(http.StatusBadGateway, ErrorResponse{Error: fmt.Sprintf("Google Sheets returned status %d", resp.StatusCode)})
		return
	}
	
	// Read body with size limit
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxUploadBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to read response"})
		return
	}
	
	// Convert the CSV data
	result, err := h.converter.ConvertPaste(string(body), req.Template)
	if err != nil {
		log.Printf("mdflow.ConvertGoogleSheet conversion error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to convert data"})
		return
	}
	
	// Add source URL to meta
	result.Meta.SourceURL = req.URL
	
	c.JSON(http.StatusOK, MDFlowConvertResponse{
		MDFlow:   result.MDFlow,
		Warnings: result.Warnings,
		Meta:     result.Meta,
	})
}

// PreviewPaste handles POST /api/mdflow/preview
// Returns a preview of the parsed table data before conversion
func (h *MDFlowHandler) PreviewPaste(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxPasteBodyBytes)

	var req PasteConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "request body exceeds limit"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "paste_text is required"})
		return
	}

	if len(req.PasteText) > maxPasteBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "paste_text exceeds 1MB limit"})
		return
	}

	// Detect input type
	analysis := converter.DetectInputType(req.PasteText)
	inputType := "table"
	if analysis.Type == converter.InputTypeMarkdown {
		inputType = "markdown"
		// For markdown, return minimal preview
		c.JSON(http.StatusOK, PreviewResponse{
			Headers:       []string{},
			Rows:          [][]string{},
			TotalRows:     0,
			PreviewRows:   0,
			HeaderRow:     -1,
			Confidence:    analysis.Confidence,
			ColumnMapping: map[string]string{},
			UnmappedCols:  []string{},
			InputType:     inputType,
		})
		return
	}

	// Parse as table
	parser := converter.NewPasteParser()
	matrix, err := parser.Parse(req.PasteText)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to parse input"})
		return
	}

	if len(matrix) == 0 {
		c.JSON(http.StatusOK, PreviewResponse{
			Headers:       []string{},
			Rows:          [][]string{},
			TotalRows:     0,
			PreviewRows:   0,
			HeaderRow:     -1,
			Confidence:    0,
			ColumnMapping: map[string]string{},
			UnmappedCols:  []string{},
			InputType:     inputType,
		})
		return
	}

	// Detect header row
	headerDetector := converter.NewHeaderDetector()
	headerRow, confidence := headerDetector.DetectHeaderRow(matrix)
	headers := matrix.GetRow(headerRow)

	// Map columns
	columnMapper := converter.NewColumnMapper()
	colMap, unmapped := columnMapper.MapColumns(headers)

	// Build column mapping as string map for JSON
	columnMapping := make(map[string]string)
	for field, idx := range colMap {
		if idx >= 0 && idx < len(headers) {
			columnMapping[headers[idx]] = string(field)
		}
	}

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
		InputType:     inputType,
	})
}

// PreviewTSV handles POST /api/mdflow/tsv/preview
// Returns a preview of the uploaded TSV file before conversion
func (h *MDFlowHandler) PreviewTSV(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBody)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "file exceeds 10MB limit"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	if header.Size > maxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "file exceeds 10MB limit"})
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".tsv" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "only .tsv files are supported"})
		return
	}

	content, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to read file"})
		return
	}
	if len(content) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is empty"})
		return
	}
	if len(content) > maxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "file exceeds 10MB limit"})
		return
	}

	// Remove BOM if present
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		content = content[3:]
	}

	// Parse as table
	parser := converter.NewPasteParser()
	matrix, err := parser.Parse(string(content))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to parse file"})
		return
	}

	if len(matrix) == 0 {
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

	// Map columns
	columnMapper := converter.NewColumnMapper()
	colMap, unmapped := columnMapper.MapColumns(headers)

	// Build column mapping as string map for JSON
	columnMapping := make(map[string]string)
	for field, idx := range colMap {
		if idx >= 0 && idx < len(headers) {
			columnMapping[headers[idx]] = string(field)
		}
	}

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

// PreviewXLSX handles POST /api/mdflow/xlsx/preview
// Returns a preview of the uploaded XLSX file before conversion
func (h *MDFlowHandler) PreviewXLSX(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBody)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "file exceeds 10MB limit"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	if header.Size > maxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "file exceeds 10MB limit"})
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".xlsx" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "only .xlsx files are supported"})
		return
	}

	sheetName := strings.TrimSpace(c.PostForm("sheet_name"))

	buf := make([]byte, 4)
	n, err := io.ReadFull(file, buf)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is empty"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to read file"})
		return
	}
	if !bytes.HasPrefix(buf[:n], xlsxMagic) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid xlsx file"})
		return
	}

	reader := io.MultiReader(bytes.NewReader(buf[:n]), file)

	// Create temp file
	tempFile, err := os.CreateTemp("", "preview-*.xlsx")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to process file"})
		return
	}
	tempName := tempFile.Name()
	defer os.Remove(tempName)

	bytesCopied, err := io.Copy(tempFile, io.LimitReader(reader, maxUploadBytes+1))
	if err != nil {
		_ = tempFile.Close()
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save file"})
		return
	}
	if bytesCopied > maxUploadBytes {
		_ = tempFile.Close()
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "file exceeds 10MB limit"})
		return
	}
	if err := tempFile.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to process file"})
		return
	}

	// Parse XLSX
	matrix, err := h.converter.ParseXLSX(tempName, sheetName)
	if err != nil {
		log.Printf("mdflow.PreviewXLSX parse error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to parse file"})
		return
	}

	if len(matrix) == 0 {
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

	// Map columns
	columnMapper := converter.NewColumnMapper()
	colMap, unmapped := columnMapper.MapColumns(headers)

	// Build column mapping as string map for JSON
	columnMapping := make(map[string]string)
	for field, idx := range colMap {
		if idx >= 0 && idx < len(headers) {
			columnMapping[headers[idx]] = string(field)
		}
	}

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
