package converter

import "sort"

// PreviewMappingQuality provides quality signals for preview UI.
type PreviewMappingQuality struct {
	Score                float64             `json:"score"`
	HeaderScore          float64             `json:"header_score"`
	MappedRatio          float64             `json:"mapped_ratio"`
	CoreCoverage         float64             `json:"core_coverage"`
	CoreMapped           int                 `json:"core_mapped"`
	RecommendedFormat    string              `json:"recommended_format"`
	LowConfidenceColumns []string            `json:"low_confidence_columns"`
	ColumnConfidence     map[string]float64  `json:"column_confidence"`
	ColumnReasons        map[string][]string `json:"column_reasons,omitempty"`
}

// BuildPreviewMappingQuality computes mapping quality and per-column confidence
// so the preview UI can highlight risky mappings.
func BuildPreviewMappingQuality(headerConfidence int, headers []string, dataRows [][]string, columnMapping map[string]string, unmapped []string) PreviewMappingQuality {
	colMap := toColumnMap(headers, columnMapping)
	quality := evaluateMappingQuality(headerConfidence, headers, colMap)
	recommended := string(OutputFormatSpec)
	if shouldFallbackToTable(string(OutputFormatSpec), quality) {
		recommended = string(OutputFormatTable)
	}

	columnConfidence := make(map[string]float64)
	columnReasons := make(map[string][]string)
	lowSet := make(map[string]bool)
	for _, h := range unmapped {
		columnConfidence[h] = 0
		columnReasons[h] = []string{"Unmapped column", "No reliable canonical match"}
		lowSet[h] = true
	}

	for field, idx := range colMap {
		if idx < 0 || idx >= len(headers) {
			continue
		}
		header := headers[idx]
		stats := collectColumnStats(dataRows, idx)
		conf, reasons := estimateHeaderFieldConfidence(header, field, stats)
		columnConfidence[header] = conf
		if len(reasons) > 0 {
			columnReasons[header] = reasons
		}
		if conf < 0.65 {
			lowSet[header] = true
		}
	}

	low := make([]string, 0, len(lowSet))
	for header := range lowSet {
		low = append(low, header)
	}
	sort.Strings(low)

	return PreviewMappingQuality{
		Score:                quality.Score,
		HeaderScore:          quality.HeaderScore,
		MappedRatio:          quality.MappedRatio,
		CoreCoverage:         quality.CoreCoverage,
		CoreMapped:           quality.CoreMapped,
		RecommendedFormat:    recommended,
		LowConfidenceColumns: low,
		ColumnConfidence:     columnConfidence,
		ColumnReasons:        columnReasons,
	}
}

func toColumnMap(headers []string, columnMapping map[string]string) ColumnMap {
	colMap := make(ColumnMap)
	indexByHeader := make(map[string]int, len(headers))
	for idx, header := range headers {
		indexByHeader[header] = idx
	}

	for header, fieldName := range columnMapping {
		idx, ok := indexByHeader[header]
		if !ok {
			continue
		}
		field := CanonicalField(fieldName)
		if !isCanonicalField(field) {
			continue
		}
		if _, exists := colMap[field]; !exists {
			colMap[field] = idx
		}
	}

	return colMap
}

func isCanonicalField(field CanonicalField) bool {
	for _, candidate := range allCanonicalFields() {
		if candidate == field {
			return true
		}
	}
	return false
}

func estimateHeaderFieldConfidence(header string, field CanonicalField, stats columnStats) (float64, []string) {
	normalized := normalizeHeader(header)
	if mapped, ok := HeaderSynonyms[normalized]; ok && mapped == field {
		return 0.95, []string{"Exact synonym match"}
	}

	reasons := make([]string, 0, 3)
	if hasHeaderKeyword(normalized, fieldHeaderKeywords(field)) {
		reasons = append(reasons, "Header keyword match")
	} else {
		reasons = append(reasons, "Weak header keyword signal")
	}

	score := scoreCandidate(field, normalized, stats)
	if stats.samples < 2 {
		reasons = append(reasons, "Limited sample rows")
	}
	if stats.longTextRatio < 0.25 && (field == FieldInstructions || field == FieldExpected) {
		reasons = append(reasons, "Cell pattern weak for long-text field")
	}
	if stats.statusRatio < 0.25 && field == FieldStatus {
		reasons = append(reasons, "Few status-like values")
	}
	if stats.priorityRatio < 0.25 && field == FieldPriority {
		reasons = append(reasons, "Few priority-like values")
	}
	if stats.methodRatio < 0.25 && field == FieldMethod {
		reasons = append(reasons, "Few HTTP method-like values")
	}
	if stats.statusCodeRatio < 0.25 && field == FieldStatusCode {
		reasons = append(reasons, "Few HTTP status code-like values")
	}

	if score <= 0 {
		return 0.45, reasons
	}
	if score > 1 {
		return 1, reasons
	}
	return score, reasons
}
