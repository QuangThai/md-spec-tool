package converter

import (
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

// XLSXParser parses Excel files
type XLSXParser struct{}

// NewXLSXParser creates a new XLSXParser
func NewXLSXParser() *XLSXParser {
	return &XLSXParser{}
}

// ParseFile parses an Excel file from path
func (p *XLSXParser) ParseFile(filePath string) (*XLSXResult, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	return p.parseExcelFile(f)
}

// ParseReader parses an Excel file from io.Reader
func (p *XLSXParser) ParseReader(reader io.Reader) (*XLSXResult, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read excel data: %w", err)
	}
	defer f.Close()

	return p.parseExcelFile(f)
}

// XLSXResult contains parsed XLSX data
type XLSXResult struct {
	Sheets     []string              `json:"sheets"`
	SheetData  map[string]CellMatrix `json:"-"`
	ActiveSheet string               `json:"active_sheet"`
}

// GetMatrix returns the CellMatrix for a specific sheet
func (r *XLSXResult) GetMatrix(sheetName string) CellMatrix {
	if r.SheetData == nil {
		return nil
	}
	return r.SheetData[sheetName]
}

func (p *XLSXParser) parseExcelFile(f *excelize.File) (*XLSXResult, error) {
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in excel file")
	}

	result := &XLSXResult{
		Sheets:      sheets,
		SheetData:   make(map[string]CellMatrix),
		ActiveSheet: sheets[0],
	}

	for _, sheetName := range sheets {
		rows, err := f.GetRows(sheetName)
		if err != nil {
			continue // Skip sheets that can't be read
		}

		matrix := NewCellMatrix(rows).Normalize()
		result.SheetData[sheetName] = matrix
	}

	return result, nil
}

// ParseSheet parses a specific sheet from a file
func (p *XLSXParser) ParseSheet(filePath string, sheetName string) (CellMatrix, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows from sheet %s: %w", sheetName, err)
	}

	return NewCellMatrix(rows).Normalize(), nil
}

// ParseSheetFromReader parses a specific sheet from reader
func (p *XLSXParser) ParseSheetFromReader(reader io.Reader, sheetName string) (CellMatrix, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read excel data: %w", err)
	}
	defer f.Close()

	// If no sheet name provided, use first sheet
	if sheetName == "" {
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, fmt.Errorf("no sheets found in excel file")
		}
		sheetName = sheets[0]
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows from sheet %s: %w", sheetName, err)
	}

	return NewCellMatrix(rows).Normalize(), nil
}
