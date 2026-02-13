package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

// PasteConvertRequest represents the request for paste conversion
type PasteConvertRequest struct {
	PasteText       string            `json:"paste_text" binding:"required"`
	Template        string            `json:"template"`
	Format          string            `json:"format"`
	ColumnOverrides map[string]string `json:"column_overrides,omitempty"`
}

// XLSXConvertRequest represents the request for XLSX sheet conversion
type XLSXConvertRequest struct {
	SheetName string `json:"sheet_name"`
	Template  string `json:"template"`
	Format    string `json:"format"`
}

// MDFlowConvertResponse represents the conversion response
type MDFlowConvertResponse struct {
	MDFlow   string                `json:"mdflow"`
	Warnings []converter.Warning   `json:"warnings"`
	Meta     converter.SpecDocMeta `json:"meta"`
	Format   string                `json:"format"`
	Template string                `json:"template"`
}

// InputAnalysisResponse represents the input type detection response
type InputAnalysisResponse struct {
	Type       string  `json:"type"` // 'markdown' | 'table' | 'unknown'
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason,omitempty"`
}

// ConvertPaste handles POST /api/mdflow/paste
// If detect_only=true query param, returns input type analysis
// Otherwise converts pasted TSV/CSV text to MDFlow format
func (h *MDFlowHandler) ConvertPaste(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.cfg.MaxPasteBytes+4<<10)

	var req PasteConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "request body exceeds limit"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	// Check if paste_text is empty after binding succeeded
	if strings.TrimSpace(req.PasteText) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "paste_text is required"})
		return
	}

	// Check size limit
	if int64(len(req.PasteText)) > h.cfg.MaxPasteBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: fmt.Sprintf("paste_text exceeds %s limit", humanSize(h.cfg.MaxPasteBytes))})
		return
	}

	normalizedTemplate, normalizedFormat, err := normalizeTemplateAndFormat(req.Template, req.Format)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	req.Template = normalizedTemplate
	req.Format = normalizedFormat

	// Detect input type FIRST regardless of request
	analysis := converter.DetectInputType(req.PasteText)

	// Check if this is detection-only request
	detectOnly, err := strconv.ParseBool(c.DefaultQuery("detect_only", "false"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid detect_only value"})
		return
	}

	slog.Info("mdflow.ConvertPaste", "detectOnly", detectOnly, "template", req.Template, "format", req.Format, "pasteBytes", len(req.PasteText), "type", string(analysis.Type), "confidence", analysis.Confidence)
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

	// Full conversion with format support (BYOK-aware)
	// Create context with timeout to prevent hanging on slow AI calls
	ctx, cancel := context.WithTimeout(c.Request.Context(), 150*time.Second)
	defer cancel()

	slog.Info("mdflow.ConvertPaste starting", "has_ai", h.aiService != nil)

	conv := h.getConverterForRequest(c)
	columnOverrides := req.ColumnOverrides
	if len(columnOverrides) == 0 {
		columnOverrides = nil
	}
	result, err := conv.ConvertPasteWithOverrides(ctx, req.PasteText, req.Template, req.Format, columnOverrides)
	if err != nil {
		slog.Error("mdflow.ConvertPaste failed", "error", err, "ai_enabled", h.aiService != nil)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to convert input"})
		return
	}

	slog.Info("mdflow.ConvertPaste ai", "ai_mode", result.Meta.AIMode, "ai_used", result.Meta.AIUsed, "ai_confidence", result.Meta.AIAvgConfidence)

	c.JSON(http.StatusOK, MDFlowConvertResponse{
		MDFlow:   result.MDFlow,
		Warnings: result.Warnings,
		Meta:     result.Meta,
		Format:   req.Format,
		Template: req.Template,
	})
}

// ConvertXLSX handles POST /api/mdflow/xlsx
// Converts uploaded XLSX file to MDFlow format
func (h *MDFlowHandler) ConvertXLSX(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.cfg.MaxUploadBytes+1<<20)

	// Get file from multipart form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: fmt.Sprintf("file exceeds %s limit", humanSize(h.cfg.MaxUploadBytes))})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	// Check file size limit
	if header.Size > h.cfg.MaxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: fmt.Sprintf("file exceeds %s limit", humanSize(h.cfg.MaxUploadBytes))})
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
	template, format, err := normalizeTemplateAndFormat(c.PostForm("template"), c.PostForm("format"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	columnOverrides, err := parseColumnOverrides(c.PostForm("column_overrides"))
	if err != nil {
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
	bytesCopied, err := io.Copy(tempFile, io.LimitReader(reader, h.cfg.MaxUploadBytes+1))
	if err != nil {
		_ = tempFile.Close()
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save file"})
		return
	}
	if bytesCopied > h.cfg.MaxUploadBytes {
		_ = tempFile.Close()
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: fmt.Sprintf("file exceeds %s limit", humanSize(h.cfg.MaxUploadBytes))})
		return
	}
	if err := tempFile.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to process file"})
		return
	}

	conv := h.getConverterForRequest(c)
	matrix, err := conv.ParseXLSX(tempName, sheetName)
	if err != nil {
		slog.Error("mdflow.ConvertXLSX parse error", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to parse file"})
		return
	}

	// Create context with timeout to prevent hanging on slow AI calls
	ctx, cancel := context.WithTimeout(c.Request.Context(), 130*time.Second)
	defer cancel()

	result, err := conv.ConvertMatrixWithOverrides(ctx, matrix, sheetName, template, format, columnOverrides)
	if err != nil {
		slog.Error("mdflow.ConvertXLSX failed", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to convert file"})
		return
	}

	slog.Info("mdflow.ConvertXLSX ai", "ai_mode", result.Meta.AIMode, "ai_used", result.Meta.AIUsed, "ai_confidence", result.Meta.AIAvgConfidence)

	c.JSON(http.StatusOK, MDFlowConvertResponse{
		MDFlow:   result.MDFlow,
		Warnings: result.Warnings,
		Meta:     result.Meta,
		Format:   format,
		Template: template,
	})
}

// ConvertTSV handles POST /api/mdflow/tsv
// Converts uploaded TSV file to MDFlow format
func (h *MDFlowHandler) ConvertTSV(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.cfg.MaxUploadBytes+1<<20)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: fmt.Sprintf("file exceeds %s limit", humanSize(h.cfg.MaxUploadBytes))})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	if header.Size > h.cfg.MaxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: fmt.Sprintf("file exceeds %s limit", humanSize(h.cfg.MaxUploadBytes))})
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".tsv" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "only .tsv files are supported"})
		return
	}

	template, format, err := normalizeTemplateAndFormat(c.PostForm("template"), c.PostForm("format"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	columnOverrides, err := parseColumnOverrides(c.PostForm("column_overrides"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	content, err := io.ReadAll(io.LimitReader(file, h.cfg.MaxUploadBytes+1))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to read file"})
		return
	}
	if len(content) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is empty"})
		return
	}
	if int64(len(content)) > h.cfg.MaxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: fmt.Sprintf("file exceeds %s limit", humanSize(h.cfg.MaxUploadBytes))})
		return
	}

	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		content = content[3:]
	}

	conv := h.getConverterForRequest(c)

	// Create context with timeout to prevent hanging on slow AI calls
	ctx, cancel := context.WithTimeout(c.Request.Context(), 130*time.Second)
	defer cancel()

	result, err := conv.ConvertPasteWithOverrides(ctx, string(content), template, format, columnOverrides)
	if err != nil {
		slog.Error("mdflow.ConvertTSV failed", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to convert file"})
		return
	}

	c.JSON(http.StatusOK, MDFlowConvertResponse{
		MDFlow:   result.MDFlow,
		Warnings: result.Warnings,
		Meta:     result.Meta,
		Format:   format,
		Template: template,
	})
}

// GetXLSXSheets handles POST /api/mdflow/xlsx/sheets
// Returns list of sheets in uploaded XLSX file
func (h *MDFlowHandler) GetXLSXSheets(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.cfg.MaxUploadBytes+1<<20)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: fmt.Sprintf("file exceeds %s limit", humanSize(h.cfg.MaxUploadBytes))})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	// Check file size limit
	if header.Size > h.cfg.MaxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: fmt.Sprintf("file exceeds %s limit", humanSize(h.cfg.MaxUploadBytes))})
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

	bytesCopied, err := io.Copy(tempFile, io.LimitReader(reader, h.cfg.MaxUploadBytes+1))
	if err != nil {
		_ = tempFile.Close()
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save file"})
		return
	}
	if bytesCopied > h.cfg.MaxUploadBytes {
		_ = tempFile.Close()
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: fmt.Sprintf("file exceeds %s limit", humanSize(h.cfg.MaxUploadBytes))})
		return
	}
	if err := tempFile.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to process file"})
		return
	}

	sheets, err := h.converter.GetXLSXSheets(tempName)
	if err != nil {
		slog.Error("mdflow.GetXLSXSheets failed", "error", err)
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

// SheetsResponse represents the list of sheets
type SheetsResponse struct {
	Sheets      []string `json:"sheets"`
	ActiveSheet string   `json:"active_sheet"`
}
