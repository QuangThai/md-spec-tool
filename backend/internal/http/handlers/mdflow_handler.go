package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

type MDFlowHandler struct {
	converter *converter.Converter
}

func NewMDFlowHandler() *MDFlowHandler {
	return &MDFlowHandler{
		converter: converter.NewConverter(),
	}
}

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
	Type       string  `json:"type"`       // 'markdown' | 'table' | 'unknown'
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
	var req PasteConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "paste_text is required"})
		return
	}

	// Check size limit (1MB)
	if len(req.PasteText) > 1*1024*1024 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "paste_text exceeds 1MB limit"})
		return
	}

	// Log incoming request
	println("DEBUG: ConvertPaste called")
	println("DEBUG: template='" + req.Template + "'")
	println("DEBUG: paste_text length=" + fmt.Sprintf("%d", len(req.PasteText)))

	// Detect input type FIRST regardless of request
	analysis := converter.DetectInputType(req.PasteText)
	println("DEBUG: detected type=" + string(analysis.Type))
	println("DEBUG: detection confidence=" + fmt.Sprintf("%d", analysis.Confidence))
	println("DEBUG: detection reason=" + analysis.Reason)

	// Check if this is detection-only request
	detectOnly := c.Query("detect_only") == "true"
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
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
	// Get file from multipart form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	// Check file size limit (10MB)
	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file exceeds 10MB limit"})
		return
	}

	// Check file extension
	ext := filepath.Ext(header.Filename)
	if ext != ".xlsx" && ext != ".xls" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "only .xlsx files are supported"})
		return
	}

	// Get optional parameters
	sheetName := c.PostForm("sheet_name")
	template := c.PostForm("template")

	// Create temp file
	tempFile, err := os.CreateTemp("", "upload-*.xlsx")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to process file"})
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy uploaded file to temp
	if _, err := io.Copy(tempFile, file); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save file"})
		return
	}

	result, err := h.converter.ConvertXLSX(tempFile.Name(), sheetName, template)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
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
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	// Check file size limit
	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file exceeds 10MB limit"})
		return
	}

	// Create temp file
	tempFile, err := os.CreateTemp("", "upload-*.xlsx")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to process file"})
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, file); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save file"})
		return
	}

	sheets, err := h.converter.GetXLSXSheets(tempFile.Name())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
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
	renderer := converter.NewMDFlowRenderer()
	templates := renderer.GetTemplateNames()

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
	})
}
