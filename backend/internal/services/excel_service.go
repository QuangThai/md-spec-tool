package services

import (
	"fmt"

	"github.com/xuri/excelize/v2"
	"github.com/yourorg/md-spec-tool/internal/models"
)

type ExcelService struct{}

func NewExcelService() *ExcelService {
	return &ExcelService{}
}

// ParseExcel reads an Excel file and returns table data
func (s *ExcelService) ParseExcel(filePath string) (*models.TableData, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	// Get first sheet name
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in excel file")
	}

	sheetName := sheets[0]

	// Get rows
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("excel file must have headers and at least one data row")
	}

	// First row is headers
	headers := rows[0]

	// Data rows
	var dataRows []map[string]interface{}
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		rowMap := make(map[string]interface{})

		for j, header := range headers {
			value := ""
			if j < len(row) {
				value = row[j]
			}
			rowMap[header] = value
		}

		dataRows = append(dataRows, rowMap)
	}

	return &models.TableData{
		Headers:   headers,
		Rows:      dataRows,
		SheetName: sheetName,
	}, nil
}
