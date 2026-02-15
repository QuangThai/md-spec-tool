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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/config"
	"github.com/yourorg/md-spec-tool/internal/converter"
)

const maxPreviewRows = 20

// PreviewHandler handles all preview endpoints for various input types
type PreviewHandler struct {
	converter  *converter.Converter
	cfg        *config.Config
	byokCache  *AIServiceProvider
}

// NewPreviewHandler creates a new PreviewHandler
func NewPreviewHandler(conv *converter.Converter, cfg *config.Config, byokCache *AIServiceProvider) *PreviewHandler {
	if conv == nil {
		conv = converter.NewConverter()
	}
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	if byokCache == nil {
		byokCache = NewAIServiceProvider(cfg)
	}
	return &PreviewHandler{
		converter:  conv,
		cfg:        cfg,
		byokCache:  byokCache,
	}
}

// buildPreviewFromMatrix builds a PreviewResponse from a parsed CellMatrix.
// Shared logic for PreviewPaste, PreviewTSV, PreviewXLSX to avoid duplication.
func (h *PreviewHandler) buildPreviewFromMatrix(c *gin.Context, matrix converter.CellMatrix, templateName string) PreviewResponse {
	headerDetector := converter.NewHeaderDetector()
	headerRow, confidence := headerDetector.DetectHeaderRow(matrix)
	headers := matrix.GetRow(headerRow)

	skipAI := c.Query("skip_ai") != "false"

	var columnMapping map[string]string
	var unmapped []string
	if skipAI {
		columnMapping, unmapped = h.converter.GetPreviewColumnMappingRuleBased(headers, templateName)
	} else {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
		defer cancel()
		conv := h.byokCache.GetConverterForRequest(c, h.converter)
		dataRows := matrix.SliceRows(headerRow+1, matrix.RowCount())
		columnMapping, unmapped = conv.GetPreviewColumnMappingWithContext(ctx, headers, dataRows, templateName, "")
	}

	dataRows := matrix.SliceRows(headerRow+1, matrix.RowCount())
	quality := converter.BuildPreviewMappingQuality(confidence, headers, dataRows, columnMapping, unmapped)

	totalDataRows := matrix.RowCount() - headerRow - 1
	previewCount := totalDataRows
	if previewCount > maxPreviewRows {
		previewCount = maxPreviewRows
	}

	rows := make([][]string, 0, previewCount)
	for i := headerRow + 1; i < headerRow+1+previewCount && i < matrix.RowCount(); i++ {
		rows = append(rows, matrix.GetRow(i))
	}

	return PreviewResponse{
		Headers:        headers,
		Rows:           rows,
		TotalRows:      totalDataRows,
		PreviewRows:    previewCount,
		HeaderRow:      headerRow,
		Confidence:     confidence,
		ColumnMapping:  columnMapping,
		UnmappedCols:   unmapped,
		MappingQuality: &quality,
		InputType:      "table",
		AIAvailable:    h.byokCache.HasAIForRequest(c),
	}
}

// emptyTablePreview returns an empty PreviewResponse for table/markdown edge cases.
func (h *PreviewHandler) emptyTablePreview(c *gin.Context, confidence int, inputType string) PreviewResponse {
	resp := PreviewResponse{
		Headers:       []string{},
		Rows:          [][]string{},
		TotalRows:     0,
		PreviewRows:   0,
		HeaderRow:     -1,
		Confidence:    confidence,
		ColumnMapping: map[string]string{},
		UnmappedCols:  []string{},
		InputType:     inputType,
		AIAvailable:   h.byokCache.HasAIForRequest(c),
	}
	return resp
}

// PreviewPaste handles POST /api/mdflow/preview
// Returns a preview of the parsed table data before conversion
func (h *PreviewHandler) PreviewPaste(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.cfg.MaxPasteBytes+4<<10)

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

	if int64(len(req.PasteText)) > h.cfg.MaxPasteBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: fmt.Sprintf("paste_text exceeds %s limit", humanSize(h.cfg.MaxPasteBytes))})
		return
	}

	// Try parsing as table first (CSV/TSV). If we get a valid multi-column table, use it.
	// This fixes Google Sheet CSV being misclassified as markdown by DetectInputType.
	parser := converter.NewPasteParser()
	matrix, parseErr := parser.Parse(req.PasteText)
	hasTable := parseErr == nil && matrix.RowCount() >= 1 && matrix.ColCount() >= 2

	if !hasTable {
		// Empty, parse failed, or single-column: use DetectInputType (e.g. real markdown)
		analysis := converter.DetectInputType(req.PasteText)
		if analysis.Type == converter.InputTypeMarkdown {
			c.JSON(http.StatusOK, h.emptyTablePreview(c, analysis.Confidence, "markdown"))
			return
		}
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to parse input"})
			return
		}
		if matrix.RowCount() == 0 {
			c.JSON(http.StatusOK, h.emptyTablePreview(c, 0, "table"))
			return
		}
	}

	templateName := strings.TrimSpace(req.Template)
	c.JSON(http.StatusOK, h.buildPreviewFromMatrix(c, matrix, templateName))
}

// PreviewTSV handles POST /api/mdflow/tsv/preview
// Returns a preview of the uploaded TSV file before conversion
func (h *PreviewHandler) PreviewTSV(c *gin.Context) {
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
		c.JSON(http.StatusOK, h.emptyTablePreview(c, 0, "table"))
		return
	}

	templateName := strings.TrimSpace(c.PostForm("template"))
	c.JSON(http.StatusOK, h.buildPreviewFromMatrix(c, matrix, templateName))
}

// PreviewXLSX handles POST /api/mdflow/xlsx/preview
// Returns a preview of the uploaded XLSX file before conversion
func (h *PreviewHandler) PreviewXLSX(c *gin.Context) {
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

	// Parse XLSX
	matrix, err := h.converter.ParseXLSX(tempName, sheetName)
	if err != nil {
		slog.Error("mdflow.PreviewXLSX parse error", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to parse file"})
		return
	}

	if len(matrix) == 0 {
		c.JSON(http.StatusOK, h.emptyTablePreview(c, 0, "table"))
		return
	}

	templateName := strings.TrimSpace(c.PostForm("template"))
	c.JSON(http.StatusOK, h.buildPreviewFromMatrix(c, matrix, templateName))
}
