package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/services"
)

type ImportHandler struct {
	excelService *services.ExcelService
}

func NewImportHandler(excelService *services.ExcelService) *ImportHandler {
	return &ImportHandler{excelService: excelService}
}

func (h *ImportHandler) UploadExcel(c *gin.Context) {
	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "No file provided"})
		return
	}

	// Validate file extension
	ext := filepath.Ext(file.Filename)
	if ext != ".xlsx" && ext != ".xls" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Only .xlsx and .xls files are supported"})
		return
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "File size exceeds 10MB limit"})
		return
	}

	// Save temporary file
	tmpDir := "/tmp"
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("upload_%d_%s", file.Size, file.Filename))

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to open file"})
		return
	}
	defer src.Close()

	dst, err := os.Create(tmpFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to save file"})
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to copy file"})
		return
	}

	// Parse Excel
	tableData, err := h.excelService.ParseExcel(tmpFile)
	if err != nil {
		os.Remove(tmpFile)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Clean up temp file
	defer os.Remove(tmpFile)

	c.JSON(http.StatusOK, tableData)
}
