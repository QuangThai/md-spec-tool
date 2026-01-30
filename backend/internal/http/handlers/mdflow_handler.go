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
	Warnings []string              `json:"warnings"`
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

// GetTemplates handles GET /api/mdflow/templates
// Returns available MDFlow templates
func (h *MDFlowHandler) GetTemplates(c *gin.Context) {
	templates := h.renderer.GetTemplateNames()

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
	})
}

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
