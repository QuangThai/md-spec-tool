package ai

// NewPasteProcessorForTest exposes a quick-detect processor for external tests.
func NewPasteProcessorForTest(config Config) *PasteProcessorService {
	return &PasteProcessorService{config: config}
}

// NewServiceForFallbackTest exposes a minimal service for fallback tests.
func NewServiceForFallbackTest() *ServiceImpl {
	return &ServiceImpl{validator: NewValidator()}
}

// QuickTableDetect exposes quickTableDetect for external tests.
func (s *PasteProcessorService) QuickTableDetect(content string) *PasteAnalysis {
	return s.quickTableDetect(content)
}

// ParseMarkdownRow exposes parseMarkdownRow for external tests.
func ParseMarkdownRow(line string) []string {
	return parseMarkdownRow(line)
}

// DetectCSVReliability exposes detectCSVReliability for external tests.
func DetectCSVReliability(lines []string) float64 {
	return detectCSVReliability(lines)
}

// ApplyConfidenceFallback exposes applyConfidenceFallback for external tests.
func (s *ServiceImpl) ApplyConfidenceFallback(result *ColumnMappingResult) *ColumnMappingResult {
	return s.applyConfidenceFallback(result)
}

// GetRequiredFieldsBySchema exposes getRequiredFieldsBySchema for external tests.
func GetRequiredFieldsBySchema(schema string) []string {
	return getRequiredFieldsBySchema(schema)
}
