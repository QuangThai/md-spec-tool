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
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

func resolveConvertOptions(includeMetadata *bool, numberRows *bool) converter.ConvertOptions {
	options := converter.DefaultConvertOptions()
	if includeMetadata != nil {
		options.IncludeMetadata = *includeMetadata
	}
	if numberRows != nil {
		options.NumberRows = *numberRows
	}
	return options
}

func parseOptionalFormBool(c *gin.Context, field string) (*bool, error) {
	raw := strings.TrimSpace(c.PostForm(field))
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid %s value", field)
	}
	return &v, nil
}

// ConvertHandler handles conversion endpoints (Paste, TSV, XLSX)
type ConvertHandler struct {
	converter *converter.Converter
	cfg       *config.Config
	byokCache *AIServiceProvider
}

// NewConvertHandler creates a new ConvertHandler
func NewConvertHandler(conv *converter.Converter, cfg *config.Config, byokCache *AIServiceProvider) *ConvertHandler {
	if conv == nil {
		conv = converter.NewConverter()
	}
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	if byokCache == nil {
		byokCache = NewAIServiceProvider(cfg)
	}
	return &ConvertHandler{
		converter: conv,
		cfg:       cfg,
		byokCache: byokCache,
	}
}

// ConvertPaste handles POST /api/mdflow/paste
// If detect_only=true query param, returns input type analysis
// Otherwise converts pasted TSV/CSV text to MDFlow format
func (h *ConvertHandler) ConvertPaste(c *gin.Context) {
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
	ctx, cancel := context.WithTimeout(c.Request.Context(), 150*time.Second)
	defer cancel()

	conv := h.byokCache.GetConverterForRequest(c, h.converter)
	columnOverrides := req.ColumnOverrides
	if len(columnOverrides) == 0 {
		columnOverrides = nil
	}
	options := resolveConvertOptions(req.IncludeMetadata, req.NumberRows)
	result, err := conv.ConvertPasteWithOverridesAndOptions(ctx, req.PasteText, req.Template, req.Format, columnOverrides, options)
	if err != nil {
		slog.Error("mdflow.ConvertPaste failed", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to convert input"})
		return
	}

	warnings := result.Warnings
	if req.ValidationRules != nil && hasValidationRules(req.ValidationRules) {
		specDoc, buildErr := converter.BuildSpecDocFromPaste(req.PasteText)
		if buildErr == nil {
			valResult := converter.Validate(specDoc, req.ValidationRules)
			if len(valResult.Warnings) > 0 {
				warnings = append(warnings, valResult.Warnings...)
			}
		}
	}

	slog.Info("mdflow.ConvertPaste ai", "ai_mode", result.Meta.AIMode, "ai_used", result.Meta.AIUsed, "ai_confidence", result.Meta.AIAvgConfidence)

	c.JSON(http.StatusOK, MDFlowConvertResponse{
		MDFlow:   result.MDFlow,
		Warnings: warnings,
		Meta:     result.Meta,
		Format:   req.Format,
		Template: req.Template,
	})
}

// ConvertXLSX handles POST /api/mdflow/xlsx
// Converts uploaded XLSX file to MDFlow format
func (h *ConvertHandler) ConvertXLSX(c *gin.Context) {
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
	includeMetadata, err := parseOptionalFormBool(c, "include_metadata")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	numberRows, err := parseOptionalFormBool(c, "number_rows")
	if err != nil {
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

	conv := h.byokCache.GetConverterForRequest(c, h.converter)
	matrix, err := conv.ParseXLSX(tempName, sheetName)
	if err != nil {
		slog.Error("mdflow.ConvertXLSX parse error", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to parse file"})
		return
	}

	// Create context with timeout to prevent hanging on slow AI calls
	ctx, cancel := context.WithTimeout(c.Request.Context(), 130*time.Second)
	defer cancel()

	options := resolveConvertOptions(includeMetadata, numberRows)
	result, err := conv.ConvertMatrixWithOverridesAndOptions(ctx, matrix, sheetName, template, format, columnOverrides, options)
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
func (h *ConvertHandler) ConvertTSV(c *gin.Context) {
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
	includeMetadata, err := parseOptionalFormBool(c, "include_metadata")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	numberRows, err := parseOptionalFormBool(c, "number_rows")
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

	conv := h.byokCache.GetConverterForRequest(c, h.converter)

	// Create context with timeout to prevent hanging on slow AI calls
	ctx, cancel := context.WithTimeout(c.Request.Context(), 130*time.Second)
	defer cancel()

	options := resolveConvertOptions(includeMetadata, numberRows)
	result, err := conv.ConvertPasteWithOverridesAndOptions(ctx, string(content), template, format, columnOverrides, options)
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
func (h *ConvertHandler) GetXLSXSheets(c *gin.Context) {
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

func hasValidationRules(rules *converter.ValidationRules) bool {
	if rules == nil {
		return false
	}
	if len(rules.RequiredFields) > 0 {
		return true
	}
	if rules.FormatRules != nil {
		if rules.FormatRules.IDPattern != "" {
			return true
		}
		if len(rules.FormatRules.EmailFields) > 0 || len(rules.FormatRules.URLFields) > 0 {
			return true
		}
	}
	if len(rules.CrossField) > 0 {
		return true
	}
	return false
}
