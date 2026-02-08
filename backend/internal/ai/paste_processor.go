package ai

import (
	"context"
	"strings"
)

// PasteProcessorService handles intelligent analysis and processing of pasted content
type PasteProcessorService struct {
	aiService Service
	cache     *Cache
	validator *Validator
	config    Config
}

// NewPasteProcessorService creates a new paste processor service
func NewPasteProcessorService(aiService Service, cache *Cache, validator *Validator, config Config) *PasteProcessorService {
	return &PasteProcessorService{
		aiService: aiService,
		cache:     cache,
		validator: validator,
		config:    config,
	}
}

// AnalyzePaste processes pasted content with intelligent format detection
func (s *PasteProcessorService) AnalyzePaste(ctx context.Context, req AnalyzePasteRequest) (*PasteAnalysis, error) {
	// Heuristic pre-check: if clearly tabular, skip AI
	if result := s.quickTableDetect(req.Content); result != nil {
		return result, nil
	}

	// Truncate for large pastes
	content := req.Content
	maxLen := req.MaxSize
	if maxLen == 0 {
		maxLen = 10000 // Default 10KB
	}
	if len(content) > maxLen {
		content = content[:maxLen] + "\n...[truncated]"
	}

	// Check cache
	cacheKey, err := MakeCacheKey("analyze_paste", s.config.Model, PromptVersionPasteAnalysis, SchemaVersionPasteAnalysis, req)
	if err == nil {
		if cached, ok := s.cache.Get(cacheKey); ok {
			return cached.(*PasteAnalysis), nil
		}
	}

	// Call AI service
	result, err := s.aiService.AnalyzePaste(ctx, AnalyzePasteRequest{
		Content: content,
		MaxSize: maxLen,
	})
	if err != nil {
		return nil, err
	}

	// Validate
	if err := s.validator.ValidatePasteAnalysis(result); err != nil {
		return nil, err
	}

	// Ensure schema version is set
	if result.SchemaVersion == "" {
		result.SchemaVersion = SchemaVersionPasteAnalysis
	}

	// Cache successful result
	if cacheKey != "" {
		s.cache.Set(cacheKey, result)
	}

	return result, nil
}

// quickTableDetect attempts fast heuristic detection for common table formats
func (s *PasteProcessorService) quickTableDetect(content string) *PasteAnalysis {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return nil
	}

	// Check for tab-separated (TSV)
	firstLine := lines[0]
	if strings.Count(firstLine, "\t") >= 2 {
		headers := strings.Split(firstLine, "\t")
		rows := [][]string{headers}
		for _, line := range lines[1:] {
			if strings.TrimSpace(line) == "" {
				continue
			}
			rows = append(rows, strings.Split(line, "\t"))
		}
		return &PasteAnalysis{
			SchemaVersion:   SchemaVersionPasteAnalysis,
			InputType:       "table",
			DetectedFormat:  "tsv",
			NormalizedTable: rows,
			DetectedColumns: headers,
			SuggestedOutput: "spec",
			Confidence:      0.95,
		}
	}

	// Check for markdown table
	if len(lines) > 1 && strings.Contains(firstLine, "|") && strings.Contains(lines[1], "-") {
		result := s.parseMarkdownTable(lines)
		if result != nil {
			return result
		}
	}

	// Check for CSV (comma-separated, not too many lines)
	if strings.Count(firstLine, ",") >= 2 && len(lines) < 100 {
		// Be conservative with CSV detection - could be prose with commas
		if detectCSVReliability(lines) > 0.8 {
			headers := strings.Split(firstLine, ",")
			rows := [][]string{headers}
			for _, line := range lines[1:] {
				if strings.TrimSpace(line) == "" {
					continue
				}
				rows = append(rows, strings.Split(line, ","))
			}
			return &PasteAnalysis{
				SchemaVersion:   SchemaVersionPasteAnalysis,
				InputType:       "table",
				DetectedFormat:  "csv",
				NormalizedTable: rows,
				DetectedColumns: headers,
				SuggestedOutput: "spec",
				Confidence:      0.85,
			}
		}
	}

	return nil // Let AI handle ambiguous cases
}

// parseMarkdownTable extracts data from markdown table format
func (s *PasteProcessorService) parseMarkdownTable(lines []string) *PasteAnalysis {
	if len(lines) < 2 {
		return nil
	}

	// Parse header row
	headerLine := lines[0]
	if !strings.Contains(headerLine, "|") {
		return nil
	}

	headers := parseMarkdownRow(headerLine)
	if len(headers) == 0 {
		return nil
	}

	// Skip separator line (lines[1])
	rows := [][]string{headers}

	// Parse data rows
	for i := 2; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		row := parseMarkdownRow(line)
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	return &PasteAnalysis{
		SchemaVersion:   SchemaVersionPasteAnalysis,
		InputType:       "table",
		DetectedFormat:  "markdown_table",
		NormalizedTable: rows,
		DetectedColumns: headers,
		SuggestedOutput: "spec",
		Confidence:      0.95,
	}
}

// parseMarkdownRow extracts cells from a markdown table row
func parseMarkdownRow(line string) []string {
	// Remove leading/trailing pipes
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")

	// Split by pipe and trim whitespace
	parts := strings.Split(line, "|")
	var cells []string
	for _, part := range parts {
		cells = append(cells, strings.TrimSpace(part))
	}
	return cells
}

// detectCSVReliability checks if content is likely CSV format
func detectCSVReliability(lines []string) float64 {
	if len(lines) < 2 {
		return 0
	}

	firstLineCommas := strings.Count(lines[0], ",")
	if firstLineCommas < 2 {
		return 0
	}

	// Check if other lines have similar comma counts (exactly matching)
	matchingLines := 0
	nonEmptyLines := 0
	for i := 1; i < len(lines) && i < 10; i++ { // Check first 10 lines
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		nonEmptyLines++
		commas := strings.Count(line, ",")
		// Strict: must match exactly
		if commas == firstLineCommas {
			matchingLines++
		}
	}

	if nonEmptyLines == 0 {
		return 0
	}

	// If 80% of sampled non-empty lines match exactly, it's likely CSV
	return float64(matchingLines) / float64(nonEmptyLines)
}

// GetNormalizedTable returns the normalized table data from analysis
func (s *PasteProcessorService) GetNormalizedTable(analysis *PasteAnalysis) [][]string {
	if analysis == nil || len(analysis.NormalizedTable) == 0 {
		return nil
	}
	return analysis.NormalizedTable
}

// GetHeaders returns detected headers from analysis
func (s *PasteProcessorService) GetHeaders(analysis *PasteAnalysis) []string {
	if analysis == nil {
		return nil
	}
	if len(analysis.NormalizedTable) > 0 {
		return analysis.NormalizedTable[0]
	}
	return analysis.DetectedColumns
}

// GetDataRows returns data rows (excluding header) from analysis
func (s *PasteProcessorService) GetDataRows(analysis *PasteAnalysis) [][]string {
	if analysis == nil || len(analysis.NormalizedTable) < 2 {
		return nil
	}
	return analysis.NormalizedTable[1:]
}
