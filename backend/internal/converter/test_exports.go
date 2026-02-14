package converter

// MappingQuality exposes mappingQuality for external tests.
type MappingQuality struct {
	Score        float64
	HeaderScore  float64
	MappedRatio  float64
	CoreCoverage float64
	CoreMapped   int
}

func toExportedMappingQuality(q mappingQuality) MappingQuality {
	return MappingQuality{
		Score:        q.Score,
		HeaderScore:  q.HeaderScore,
		MappedRatio:  q.MappedRatio,
		CoreCoverage: q.CoreCoverage,
		CoreMapped:   q.CoreMapped,
	}
}

func toInternalMappingQuality(q MappingQuality) mappingQuality {
	return mappingQuality{
		Score:        q.Score,
		HeaderScore:  q.HeaderScore,
		MappedRatio:  q.MappedRatio,
		CoreCoverage: q.CoreCoverage,
		CoreMapped:   q.CoreMapped,
	}
}

// EvaluateMappingQuality exposes evaluateMappingQuality for external tests.
func EvaluateMappingQuality(headerConfidence int, headers []string, colMap ColumnMap) MappingQuality {
	return toExportedMappingQuality(evaluateMappingQuality(headerConfidence, headers, colMap))
}

// ShouldFallbackToTable exposes shouldFallbackToTable for external tests.
func ShouldFallbackToTable(format string, quality MappingQuality) bool {
	return shouldFallbackToTable(format, toInternalMappingQuality(quality))
}

// EnhanceColumnMapping exposes enhanceColumnMapping for external tests.
func EnhanceColumnMapping(headers []string, dataRows [][]string, current ColumnMap) (ColumnMap, []string, []Warning) {
	return enhanceColumnMapping(headers, dataRows, current)
}

// DetectLikelyDelimiter exposes detectLikelyDelimiter for external tests.
func DetectLikelyDelimiter(lines []string) rune {
	return detectLikelyDelimiter(lines)
}

// ParseSimple exposes parseSimple for external tests.
func (p *PasteParser) ParseSimple(text string) (CellMatrix, error) {
	return p.parseSimple(text)
}

// GroupRowsByFeature exposes groupRowsByFeature for external tests.
func GroupRowsByFeature(rows []SpecRow) map[string][]SpecRow {
	return groupRowsByFeature(rows)
}

// FormatAsNumberedList exposes formatAsNumberedList for external tests.
func FormatAsNumberedList(s string) string {
	return formatAsNumberedList(s)
}

// EscapeTableCell exposes escapeTableCell for external tests.
func EscapeTableCell(s string) string {
	return escapeTableCell(s)
}

// InferHeaders exposes inferHeaders for external tests.
func (r *TableRenderer) InferHeaders(rows []SpecRow) []string {
	return r.inferHeaders(rows)
}

// GetCellValue exposes getCellValue for external tests.
func (r *TableRenderer) GetCellValue(row SpecRow, header string) string {
	return r.getCellValue(row, header)
}
