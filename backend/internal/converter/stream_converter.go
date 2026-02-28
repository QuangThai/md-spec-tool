package converter

import (
	"context"
	"fmt"
	"strings"
)

// StreamEvent represents a single Server-Sent Event emitted during conversion.
// Event values: "progress", "partial", "complete", "error"
type StreamEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// ProgressData is the payload for "progress" and "complete" events.
type ProgressData struct {
	Phase   string `json:"phase"`             // "parsing" | "mapping" | "rendering" | "complete"
	Percent int    `json:"percent"`           // 0–100
	Message string `json:"message,omitempty"` // optional human-readable hint
}

// PartialResult is the payload for "partial" events carrying incremental markdown.
type PartialResult struct {
	Section string `json:"section"` // which logical section is ready
	Content string `json:"content"` // rendered markdown content for that section
}

// StreamCallback is invoked synchronously at each pipeline milestone.
// Implementations must be non-blocking (write to a channel or buffer);
// any slow operation inside the callback will delay the pipeline.
type StreamCallback func(event StreamEvent)

// ConvertPasteStreaming runs the full conversion pipeline for pasted text and
// fires a StreamCallback at each phase milestone.  Phases in order:
//
//	parsing  (20%) → mapping (50%) → rendering (80%) → complete (100%)
//
// It does NOT modify any existing Converter methods.
// Returns the final ConvertResponse, or an error (including context.Canceled).
func (c *Converter) ConvertPasteStreaming(
	ctx context.Context,
	content, templateName, outputFormat string,
	callback StreamCallback,
) (*ConvertResponse, error) {
	// Validate format early (before any work) so callers always get a fast,
	// predictable error for unsupported formats.
	normalized := strings.ToLower(strings.TrimSpace(outputFormat))
	if normalized != string(OutputFormatSpec) &&
		normalized != string(OutputFormatTable) &&
		normalized != "" {
		return nil, fmt.Errorf("invalid output format %q: must be 'spec' or 'table'", outputFormat)
	}
	// Default empty → "spec"; carry the normalised value forward.
	if normalized == "" {
		normalized = string(OutputFormatSpec)
	}
	outputFormat = normalized

	// ─── Phase 1: Parsing ────────────────────────────────────────────────────
	callback(StreamEvent{
		Event: "progress",
		Data:  ProgressData{Phase: "parsing", Percent: 20, Message: "Parsing input..."},
	})

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	analysis := DetectInputType(content)

	// Markdown short-circuit: no column-mapping needed
	if analysis.Type == InputTypeMarkdown {
		callback(StreamEvent{
			Event: "progress",
			Data:  ProgressData{Phase: "rendering", Percent: 80, Message: "Rendering markdown..."},
		})
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result, err := c.convertMarkdown(content, templateName)
		if err != nil {
			return nil, err
		}
		callback(StreamEvent{
			Event: "complete",
			Data:  ProgressData{Phase: "complete", Percent: 100},
		})
		return result, nil
	}

	// Table path ──────────────────────────────────────────────────────────────
	matrix, err := c.pasteParser.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("streaming parse failed: %w", err)
	}

	if len(matrix) == 0 {
		callback(StreamEvent{
			Event: "complete",
			Data:  ProgressData{Phase: "complete", Percent: 100},
		})
		return &ConvertResponse{
			MDFlow: "",
			Warnings: []Warning{
				newWarning("INPUT_EMPTY", SeverityWarn, CatInput,
					"Empty input.", "Paste a table or upload a file to convert.", nil),
			},
			Meta: SpecDocMeta{},
		}, nil
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Detect header row
	headerRow, confidence := c.headerDetector.DetectHeaderRow(matrix)
	headers := matrix.GetRow(headerRow)
	dataRows := matrix.SliceRows(headerRow+1, matrix.RowCount())

	var warnings []Warning
	if confidence < 50 {
		warnings = append(warnings, newWarning(
			"HEADER_LOW_CONFIDENCE", SeverityWarn, CatHeader,
			"Low confidence in header detection; results may be inaccurate.",
			"Verify the header row and ensure column names are present.",
			map[string]any{"confidence": confidence, "header_row": headerRow},
		))
	}

	// ─── Phase 2: AI column mapping ──────────────────────────────────────────
	callback(StreamEvent{
		Event: "progress",
		Data:  ProgressData{Phase: "mapping", Percent: 50, Message: "Mapping columns..."},
	})

	colMap, unmapped, mappingWarnings, aiMeta := c.resolveColumnMapping(ctx, headers, dataRows, outputFormat)
	warnings = append(warnings, mappingWarnings...)
	colMap, unmapped, inferredWarnings := enhanceColumnMapping(headers, dataRows, colMap)
	warnings = append(warnings, inferredWarnings...)

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	quality := evaluateMappingQuality(confidence, headers, colMap)
	if shouldFallbackToTable(outputFormat, quality) {
		warnings = append(warnings, newWarning(
			"MAPPING_LOW_CONFIDENCE_TABLE_FALLBACK", SeverityWarn, CatMapping,
			"Mapping confidence is low for non-standard schema; switching output to table format.",
			"Use preview column overrides if you want spec format for this input.",
			map[string]any{
				"quality_score":    quality.Score,
				"mapped_ratio":     quality.MappedRatio,
				"core_field_count": quality.CoreMapped,
			},
		))
		outputFormat = string(OutputFormatTable)
	}

	table := c.tableParser.MatrixToTable(headers, dataRows, "")
	table.Meta.HeaderRowIndex = headerRow
	table.Meta.ColumnMap = colMap
	table.Meta.IncludeMetadata = true
	applyAIMetaToTableMeta(&table.Meta, aiMeta)

	// Resolve template
	templateToUse := templateName
	if templateToUse == "" {
		templateToUse = outputFormat
	}
	tmpl := c.templateRegistry.LoadTemplateOrDefault(templateToUse)

	// ─── Phase 3: Rendering ───────────────────────────────────────────────────
	callback(StreamEvent{
		Event: "progress",
		Data:  ProgressData{Phase: "rendering", Percent: 80, Message: "Rendering output..."},
	})

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var renderer Renderer
	if outputFormat == "table" {
		renderer, err = NewRendererSimple("table")
	} else {
		renderer, err = c.rendererFactory.CreateRenderer(tmpl)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create renderer for format %q: %w", outputFormat, err)
	}

	mdflow, renderWarnings, err := renderer.Render(table)
	if err != nil {
		return nil, err
	}
	for _, w := range renderWarnings {
		warnings = append(warnings, newWarning("RENDER_WARNING", SeverityWarn, CatRender, w, "", nil))
	}

	meta := SpecDocMeta{
		HeaderRow:       headerRow,
		ColumnMap:       colMap,
		UnmappedColumns: unmapped,
		TotalRows:       table.RowCount(),
		OutputFormat:    outputFormat,
	}
	applyAIMeta(&meta, aiMeta)

	// ─── Phase 4: Complete ────────────────────────────────────────────────────
	callback(StreamEvent{
		Event: "complete",
		Data:  ProgressData{Phase: "complete", Percent: 100},
	})

	return &ConvertResponse{
		MDFlow:   mdflow,
		Warnings: warnings,
		Meta:     meta,
	}, nil
}
