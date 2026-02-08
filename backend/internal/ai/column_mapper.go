package ai

import (
	"context"
)

// ColumnMapperService maps spreadsheet headers to canonical fields using LLM
type ColumnMapperService struct {
	aiService Service
}

// NewColumnMapperService creates a new column mapper service
func NewColumnMapperService(aiService Service) *ColumnMapperService {
	return &ColumnMapperService{
		aiService: aiService,
	}
}

// MapColumnsRequest represents the input for column mapping
type MapColumnsRequest struct {
	Headers    []string   `json:"headers"`      // Column headers from spreadsheet
	SampleRows [][]string `json:"sample_rows"` // 3-5 rows for context
	Format     string     `json:"format"`      // "spec" | "table"
	FileType   string     `json:"file_type"`   // "csv", "xlsx", etc.
	SourceLang string     `json:"source_lang"` // Language code: "en", "ja", "vi", etc.
}

// MapColumns maps source headers to canonical fields using LLM
func (s *ColumnMapperService) MapColumns(ctx context.Context, req MapColumnsRequest) (*ColumnMappingResult, error) {
	if s.aiService == nil {
		return nil, ErrAIUnavailable
	}

	// For "table" format, return all columns as-is (no mapping)
	if req.Format == "table" {
		return s.tableFormatMapping(req)
	}

	// For "spec" format, use LLM to map to canonical fields
	return s.aiService.MapColumns(ctx, MapColumnsRequest{
		Headers:    req.Headers,
		SampleRows: req.SampleRows,
		Format:     req.Format,
		FileType:   req.FileType,
		SourceLang: req.SourceLang,
	})
}

// tableFormatMapping returns unmapped columns (table format preserves all columns)
func (s *ColumnMapperService) tableFormatMapping(req MapColumnsRequest) (*ColumnMappingResult, error) {
	result := &ColumnMappingResult{
		SchemaVersion:   SchemaVersionColumnMapping,
		CanonicalFields: make([]CanonicalFieldMapping, 0),
		ExtraColumns:    make([]ExtraColumnMapping, 0),
		Meta: MappingMeta{
			DetectedType:    "table",
			SourceLanguage:  "unknown",
			TotalColumns:    len(req.Headers),
			MappedColumns:   0,
			UnmappedColumns: len(req.Headers),
			AvgConfidence:   0,
		},
	}

	// All columns go to extra_columns for table format
	for i, header := range req.Headers {
		result.ExtraColumns = append(result.ExtraColumns, ExtraColumnMapping{
			Name:         header,
			SemanticRole: "table_column",
			ColumnIndex:  i,
			Confidence:   1.0,
		})
	}

	return result, nil
}

// GetCanonicalName returns the canonical field name from a mapping result
func GetCanonicalName(result *ColumnMappingResult, index int) string {
	for _, mapping := range result.CanonicalFields {
		if mapping.ColumnIndex == index {
			return mapping.CanonicalName
		}
	}
	return ""
}

// GetExtraColumnName returns the extra column name from a mapping result
func GetExtraColumnName(result *ColumnMappingResult, index int) string {
	for _, extra := range result.ExtraColumns {
		if extra.ColumnIndex == index {
			return extra.Name
		}
	}
	return ""
}
