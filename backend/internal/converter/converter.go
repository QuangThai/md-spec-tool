package converter

// Converter orchestrates the conversion process
type Converter struct {
	pasteParser    *PasteParser
	xlsxParser     *XLSXParser
	headerDetector *HeaderDetector
	columnMapper   *ColumnMapper
	renderer       *MDFlowRenderer
}

// NewConverter creates a new Converter
func NewConverter() *Converter {
	return &Converter{
		pasteParser:    NewPasteParser(),
		xlsxParser:     NewXLSXParser(),
		headerDetector: NewHeaderDetector(),
		columnMapper:   NewColumnMapper(),
		renderer:       NewMDFlowRenderer(),
	}
}

// ConvertPaste converts pasted text to MDFlow
func (c *Converter) ConvertPaste(text string, template string) (*ConvertResponse, error) {
	matrix, err := c.pasteParser.Parse(text)
	if err != nil {
		return nil, err
	}

	return c.convertMatrix(matrix, "", template)
}

// ConvertXLSX converts an XLSX file to MDFlow
func (c *Converter) ConvertXLSX(filePath string, sheetName string, template string) (*ConvertResponse, error) {
	var matrix CellMatrix
	var err error

	if sheetName == "" {
		result, err := c.xlsxParser.ParseFile(filePath)
		if err != nil {
			return nil, err
		}
		sheetName = result.ActiveSheet
		matrix = result.GetMatrix(sheetName)
	} else {
		matrix, err = c.xlsxParser.ParseSheet(filePath, sheetName)
		if err != nil {
			return nil, err
		}
	}

	return c.convertMatrix(matrix, sheetName, template)
}

// GetXLSXSheets returns list of sheets in an XLSX file
func (c *Converter) GetXLSXSheets(filePath string) ([]string, error) {
	result, err := c.xlsxParser.ParseFile(filePath)
	if err != nil {
		return nil, err
	}
	return result.Sheets, nil
}

// convertMatrix converts a CellMatrix to MDFlow
func (c *Converter) convertMatrix(matrix CellMatrix, sheetName string, template string) (*ConvertResponse, error) {
	if len(matrix) == 0 {
		return &ConvertResponse{
			MDFlow:   "",
			Warnings: []string{"Empty input"},
			Meta:     SpecDocMeta{},
		}, nil
	}

	// Detect header row
	headerRow, confidence := c.headerDetector.DetectHeaderRow(matrix)
	
	var warnings []string
	if confidence < 50 {
		warnings = append(warnings, "Low confidence in header detection, results may be inaccurate")
	}

	// Get headers
	headers := matrix.GetRow(headerRow)

	// Map columns
	colMap, unmapped := c.columnMapper.MapColumns(headers)
	
	if len(unmapped) > 0 {
		warnings = append(warnings, "Unmapped columns: "+joinStrings(unmapped, ", "))
	}

	// Build SpecDoc
	specDoc := c.buildSpecDoc(matrix, headerRow, headers, colMap, sheetName)
	specDoc.Warnings = warnings

	// Render to MDFlow
	mdflow, err := c.renderer.Render(specDoc, template)
	if err != nil {
		return nil, err
	}

	return &ConvertResponse{
		MDFlow:   mdflow,
		Warnings: warnings,
		Meta:     specDoc.Meta,
	}, nil
}

// buildSpecDoc constructs a SpecDoc from parsed data
func (c *Converter) buildSpecDoc(matrix CellMatrix, headerRow int, headers []string, colMap ColumnMap, sheetName string) *SpecDoc {
	// Count rows by feature
	rowsByFeature := make(map[string]int)
	
	var rows []SpecRow
	dataRows := matrix.SliceRows(headerRow+1, matrix.RowCount())

	for _, row := range dataRows {
		specRow := SpecRow{
			ID:           GetFieldValue(row, colMap, FieldID),
			Feature:      GetFieldValue(row, colMap, FieldFeature),
			Scenario:     GetFieldValue(row, colMap, FieldScenario),
			Instructions: GetFieldValue(row, colMap, FieldInstructions),
			Inputs:       GetFieldValue(row, colMap, FieldInputs),
			Expected:     GetFieldValue(row, colMap, FieldExpected),
			Precondition: GetFieldValue(row, colMap, FieldPrecondition),
			Priority:     GetFieldValue(row, colMap, FieldPriority),
			Type:         GetFieldValue(row, colMap, FieldType),
			Status:       GetFieldValue(row, colMap, FieldStatus),
			Endpoint:     GetFieldValue(row, colMap, FieldEndpoint),
			Notes:        GetFieldValue(row, colMap, FieldNotes),
			Metadata:     make(map[string]string),
		}

		// Store unmapped columns in metadata
		for i, header := range headers {
			if i < len(row) {
				// Check if this column is mapped
				isMapped := false
				for _, idx := range colMap {
					if idx == i {
						isMapped = true
						break
					}
				}
				if !isMapped && row[i] != "" {
					specRow.Metadata[header] = row[i]
				}
			}
		}

		// Skip completely empty rows
		if specRow.Feature == "" && specRow.Scenario == "" && specRow.Instructions == "" {
			continue
		}

		rows = append(rows, specRow)
		
		// Track row count by feature
		if specRow.Feature != "" {
			rowsByFeature[specRow.Feature]++
		}
	}

	// Determine title
	title := sheetName
	if title == "" {
		title = "Converted Spec"
	}

	_, unmapped := c.columnMapper.MapColumns(headers)

	return &SpecDoc{
		Title:    title,
		Rows:     rows,
		Headers:  headers,
		Meta: SpecDocMeta{
			SheetName:       sheetName,
			HeaderRow:       headerRow,
			ColumnMap:       colMap,
			UnmappedColumns: unmapped,
			TotalRows:       len(rows),
			RowsByFeature:   rowsByFeature,
		},
	}
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
