package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const (
	// Truncation limits for prompts
	MaxSuggestionsContentBytes    = 8000
	MaxSemanticValidationBytes    = 8000
	MaxDiffBeforeBytes            = 4000
	MaxDiffAfterBytes             = 4000
	MaxDiffTextBytes              = 2000

	// Default retry after for rate limiting
	DefaultRetryAfterSeconds = 60
)

// Client wraps the OpenAI API with structured output support
type Client struct {
	client     openai.Client
	model      string
	config     Config
	maxRetries int
	retryDelay time.Duration
}

// NewClient creates a new OpenAI client
func NewClient(config Config) (*Client, error) {
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	// Apply defaults for missing config values
	defaults := DefaultConfig()
	if config.Model == "" {
		config.Model = defaults.Model
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = defaults.MaxRetries
	}
	if config.RetryBaseDelay <= 0 {
		config.RetryBaseDelay = defaults.RetryBaseDelay
	}

	var clientOpts []option.RequestOption
	clientOpts = append(clientOpts, option.WithAPIKey(apiKey))

	client := openai.NewClient(clientOpts...)

	return &Client{
		client:     client,
		model:      config.Model,
		config:     config,
		maxRetries: config.MaxRetries,
		retryDelay: config.RetryBaseDelay,
	}, nil
}

// MapColumns performs column header mapping with structured output
func (c *Client) MapColumns(ctx context.Context, req MapColumnsRequest) (*ColumnMappingResult, error) {
	userContent := formatMapColumnsPrompt(req)
	result := &ColumnMappingResult{}

	// Build JSON schema for structured output
	schema := c.buildColumnMappingSchema()

	err := c.callStructured(ctx, SystemPromptColumnMapping, userContent, schema, result)
	if err != nil {
		return nil, err
	}

	// Ensure schema version is set
	if result.SchemaVersion == "" {
		result.SchemaVersion = SchemaVersionColumnMapping
	}

	return result, nil
}

// AnalyzePaste analyzes pasted content with structured output
func (c *Client) AnalyzePaste(ctx context.Context, req AnalyzePasteRequest) (*PasteAnalysis, error) {
	userContent := formatAnalyzePastePrompt(req)
	result := &PasteAnalysis{}

	// Build JSON schema for structured output
	schema := c.buildPasteAnalysisSchema()

	err := c.callStructured(ctx, SystemPromptPasteAnalysis, userContent, schema, result)
	if err != nil {
		return nil, err
	}

	// Ensure schema version is set
	if result.SchemaVersion == "" {
		result.SchemaVersion = SchemaVersionPasteAnalysis
	}

	return result, nil
}

// callStructured makes a structured output call with retry logic
func (c *Client) callStructured(ctx context.Context, systemPrompt, userContent string, schema interface{}, out interface{}) error {
	var lastErr error

	// Calculate max attempts: 1 initial + maxRetries attempts
	// If maxRetries is 0, we get 1 total attempt
	// If maxRetries is 3, we get 4 total attempts (1 initial + 3 retries)
	maxAttempts := 1 + c.maxRetries
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			delay := c.retryDelayFor(attempt, lastErr)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Create context with timeout if configured
		reqCtx := ctx
		var cancel context.CancelFunc
		if c.config.RequestTimeout > 0 {
			reqCtx, cancel = context.WithTimeout(ctx, c.config.RequestTimeout)
		}

		// Make the API call
		resp, err := c.client.Chat.Completions.New(reqCtx, openai.ChatCompletionNewParams{
			Model: openai.ChatModel(c.model),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(systemPrompt),
				openai.UserMessage(userContent),
			},
			ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
					JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
						Name:   "response",
						Schema: schema,
						Strict: openai.Bool(true),
					},
				},
			},
		})
		if cancel != nil {
			cancel()
		}

		if err != nil {
			lastErr = c.translateError(err)
			if !c.isRetryable(lastErr) {
				return lastErr
			}
			continue
		}

		// Parse response
		if len(resp.Choices) == 0 {
			lastErr = ErrAIInvalidOutput
			continue
		}

		content := resp.Choices[0].Message.Content
		if content == "" {
			lastErr = ErrAIInvalidOutput
			continue
		}

		// Unmarshal JSON into output
		if err := json.Unmarshal([]byte(content), out); err != nil {
			lastErr = fmt.Errorf("%w: %v", ErrAIInvalidOutput, err)
			if !c.isRetryable(lastErr) {
				return lastErr
			}
			continue
		}

		return nil
	}

	return lastErr
}

func (c *Client) retryDelayFor(attempt int, lastErr error) time.Duration {
	if attempt <= 0 {
		return 0
	}
	base := c.retryDelay * time.Duration(1<<uint(attempt-1))
	var aiErr *AIError
	if errors.As(lastErr, &aiErr) && aiErr.RetryAfter > 0 {
		base = time.Duration(aiErr.RetryAfter) * time.Second
	}
	return base + jitterDuration(base/4)
}

func jitterDuration(base time.Duration) time.Duration {
	if base <= 0 {
		return 0
	}
	// Note: As of Go 1.20, the global rand is automatically seeded
	maxJitter := int64(base)
	if maxJitter <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(maxJitter + 1))
}

// translateError converts OpenAI errors to domain errors
func (c *Client) translateError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Rate limit errors
	if apiErr, ok := err.(*openai.Error); ok {
		if apiErr.StatusCode == 429 {
			return &AIError{
				Err:        ErrAIRateLimited,
				Message:    "Rate limited by OpenAI",
				RetryAfter: DefaultRetryAfterSeconds,
			}
		}
		// Server errors
		if apiErr.StatusCode >= 500 {
			return &AIError{
				Err:     ErrAIUnavailable,
				Message: fmt.Sprintf("OpenAI server error: %d", apiErr.StatusCode),
			}
		}
	}

	// Network/timeout errors
	if isTimeoutError(err) {
		return &AIError{
			Err:     ErrAIUnavailable,
			Message: "Request timeout",
		}
	}

	// Default to unavailable
	return &AIError{
		Err:     ErrAIUnavailable,
		Message: errMsg,
	}
}

// isRetryable determines if an error should trigger a retry
func (c *Client) isRetryable(err error) bool {
	// Check if it's an AIError
	var aiErr *AIError
	if errors.As(err, &aiErr) {
		return aiErr.Err == ErrAIRateLimited || aiErr.Err == ErrAIUnavailable
	}
	// Check if it's one of our domain errors
	return errors.Is(err, ErrAIRateLimited) || errors.Is(err, ErrAIUnavailable)
}

// isTimeoutError checks if error is a timeout
func isTimeoutError(err error) bool {
	if err == context.DeadlineExceeded {
		return true
	}
	// Check for net.Error with Timeout() method
	type timeoutError interface {
		Timeout() bool
	}
	if te, ok := err.(timeoutError); ok {
		return te.Timeout()
	}
	return false
}

// buildColumnMappingSchema builds the JSON schema for column mapping results
func (c *Client) buildColumnMappingSchema() interface{} {
	canonicalEnum := make([]string, 0, len(CanonicalFields))
	for field := range CanonicalFields {
		canonicalEnum = append(canonicalEnum, field)
	}

	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"schema_version": map[string]interface{}{
				"type": "string",
				"enum": []string{SchemaVersionColumnMapping},
			},
			"canonical_fields": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"canonical_name": map[string]interface{}{"type": "string", "enum": canonicalEnum},
						"source_header":  map[string]interface{}{"type": "string"},
						"column_index":   map[string]interface{}{"type": "integer", "minimum": 0},
						"confidence":     map[string]interface{}{"type": "number", "minimum": 0, "maximum": 1},
						"reasoning":      map[string]interface{}{"type": "string", "maxLength": 256},
						"alternatives": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"source_header": map[string]interface{}{"type": "string"},
									"column_index":  map[string]interface{}{"type": "integer", "minimum": 0},
									"confidence":    map[string]interface{}{"type": "number", "minimum": 0, "maximum": 1},
								},
								"required":             []string{"source_header", "column_index", "confidence"},
								"additionalProperties": false,
							},
						},
					},
					"required":             []string{"canonical_name", "source_header", "column_index", "confidence"},
					"additionalProperties": false,
				},
			},
			"extra_columns": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":          map[string]interface{}{"type": "string"},
						"semantic_role": map[string]interface{}{"type": "string"},
						"column_index":  map[string]interface{}{"type": "integer", "minimum": 0},
						"confidence":    map[string]interface{}{"type": "number", "minimum": 0, "maximum": 1},
					},
					"required":             []string{"name", "semantic_role", "column_index", "confidence"},
					"additionalProperties": false,
				},
			},
			"meta": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"detected_type":    map[string]interface{}{"type": "string"},
					"source_language":  map[string]interface{}{"type": "string"},
					"total_columns":    map[string]interface{}{"type": "integer", "minimum": 0},
					"mapped_columns":   map[string]interface{}{"type": "integer", "minimum": 0},
					"unmapped_columns": map[string]interface{}{"type": "integer", "minimum": 0},
					"avg_confidence":   map[string]interface{}{"type": "number", "minimum": 0, "maximum": 1},
				},
				"required":             []string{"detected_type", "source_language", "total_columns", "mapped_columns", "unmapped_columns", "avg_confidence"},
				"additionalProperties": false,
			},
		},
		"required":             []string{"schema_version", "canonical_fields", "meta"},
		"additionalProperties": false,
	}
}

// buildPasteAnalysisSchema builds the JSON schema for paste analysis results
func (c *Client) buildPasteAnalysisSchema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"schema_version": map[string]interface{}{
				"type": "string",
				"enum": []string{SchemaVersionPasteAnalysis},
			},
			"input_type": map[string]interface{}{
				"type": "string",
				"enum": []string{"table", "backlog_list", "test_cases", "prose", "mixed", "unknown"},
			},
			"detected_format": map[string]interface{}{
				"type": "string",
				"enum": []string{"csv", "tsv", "markdown_table", "free_text", "mixed"},
			},
			"normalized_table": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "string"},
				},
			},
			"detected_columns": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string"},
			},
			"suggested_output": map[string]interface{}{"type": "string", "enum": []string{"spec", "table"}},
			"confidence":       map[string]interface{}{"type": "number", "minimum": 0, "maximum": 1},
			"notes":            map[string]interface{}{"type": "string"},
		},
		"required":             []string{"schema_version", "input_type", "detected_format", "suggested_output", "confidence"},
		"additionalProperties": false,
	}
}

// formatMapColumnsPrompt formats the user prompt for column mapping
func formatMapColumnsPrompt(req MapColumnsRequest) string {
	prompt := fmt.Sprintf(`Analyze the following spreadsheet headers and map them to canonical fields.

Headers: %v

`, req.Headers)

	if len(req.SampleRows) > 0 {
		prompt += "Sample data rows:\n"
		for i, row := range req.SampleRows {
			prompt += fmt.Sprintf("Row %d: %v\n", i+1, row)
		}
		prompt += "\n"
	}

	if req.FileType != "" {
		prompt += fmt.Sprintf("Source file type: %s\n", req.FileType)
	}

	if req.SourceLang != "" {
		prompt += fmt.Sprintf("Source language: %s\n", req.SourceLang)
	}

	prompt += "\nReturn the mapping as JSON following the ColumnMappingResult schema."

	return prompt
}

// formatAnalyzePastePrompt formats the user prompt for paste analysis
func formatAnalyzePastePrompt(req AnalyzePasteRequest) string {
	content := req.Content
	if len(content) > 2000 {
		content = content[:2000] + "\n... (truncated)"
	}

	return fmt.Sprintf(`Analyze the following pasted content and determine its structure:

Content:
%s

Return the analysis as JSON following the PasteAnalysis schema.`, content)
}

// GetSuggestions analyzes spec content and returns improvement suggestions
func (c *Client) GetSuggestions(ctx context.Context, req SuggestionsRequest) (*SuggestionsResult, error) {
	userContent := formatSuggestionsPrompt(req)
	result := &SuggestionsResult{}

	// Build JSON schema for structured output
	schema := c.buildSuggestionsSchema()

	err := c.callStructured(ctx, SystemPromptSuggestions, userContent, schema, result)
	if err != nil {
		return nil, err
	}

	// Ensure schema version is set
	if result.SchemaVersion == "" {
		result.SchemaVersion = SchemaVersionSuggestions
	}

	// Validate and normalize suggestions
	validSuggestions := make([]Suggestion, 0, len(result.Suggestions))
	for _, s := range result.Suggestions {
		// Validate type
		switch s.Type {
		case SuggestionMissingField, SuggestionVagueDescription, SuggestionIncompleteSteps, SuggestionFormatting, SuggestionCoverage:
			// Valid type
		default:
			s.Type = SuggestionVagueDescription // Default fallback
		}

		// Validate severity
		switch s.Severity {
		case "info", "warn", "error":
			// Valid severity
		default:
			s.Severity = "info" // Default fallback
		}

		// Skip empty suggestions
		if s.Message == "" && s.Suggestion == "" {
			continue
		}

		validSuggestions = append(validSuggestions, s)
	}
	result.Suggestions = validSuggestions

	return result, nil
}

// formatSuggestionsPrompt formats the user prompt for suggestions
func formatSuggestionsPrompt(req SuggestionsRequest) string {
	content := req.SpecContent
	if len(content) > MaxSuggestionsContentBytes {
		content = content[:MaxSuggestionsContentBytes] + "\n... (truncated)"
	}

	return fmt.Sprintf(`Analyze the following test specification document and identify quality issues.

Template: %s
Row Count: %d

Spec Content:
%s

Return the analysis as JSON with a "suggestions" array following the SuggestionsResult schema.`,
		req.Template, req.RowCount, content)
}

// buildSuggestionsSchema builds the JSON schema for suggestions results
func (c *Client) buildSuggestionsSchema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"schema_version": map[string]interface{}{
				"type": "string",
				"enum": []string{SchemaVersionSuggestions},
			},
			"suggestions": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"type": map[string]interface{}{
							"type": "string",
							"enum": ValidSuggestionTypes(),
						},
						"severity": map[string]interface{}{
							"type": "string",
							"enum": []string{"info", "warn", "error"},
						},
						"message":    map[string]interface{}{"type": "string"},
						"row_ref":    map[string]interface{}{"type": "integer", "minimum": 1},
						"field":      map[string]interface{}{"type": "string"},
						"suggestion": map[string]interface{}{"type": "string"},
					},
					"required":             []string{"type", "severity", "message", "suggestion"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"schema_version", "suggestions"},
		"additionalProperties": false,
	}
}

// SummarizeDiff analyzes diff between two documents and returns AI-generated summary
func (c *Client) SummarizeDiff(ctx context.Context, req SummarizeDiffRequest) (*DiffSummary, error) {
	userContent := formatSummarizeDiffPrompt(req)
	result := &DiffSummary{}

	// Build JSON schema for structured output
	schema := c.buildDiffSummarySchema()

	err := c.callStructured(ctx, SystemPromptDiffSummary, userContent, schema, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// formatSummarizeDiffPrompt formats the user prompt for diff summarization
func formatSummarizeDiffPrompt(req SummarizeDiffRequest) string {
	before := req.Before
	if len(before) > MaxDiffBeforeBytes {
		before = before[:MaxDiffBeforeBytes] + "\n... (truncated)"
	}

	after := req.After
	if len(after) > MaxDiffAfterBytes {
		after = after[:MaxDiffAfterBytes] + "\n... (truncated)"
	}

	diffText := req.DiffText
	if len(diffText) > MaxDiffTextBytes {
		diffText = diffText[:MaxDiffTextBytes] + "\n... (truncated)"
	}

	return fmt.Sprintf(`Analyze the changes between two versions of a test specification document.

BEFORE:
%s

AFTER:
%s

DIFF:
%s

Provide a concise summary of the changes, listing key changes and their potential impact.`, before, after, diffText)
}

// buildDiffSummarySchema builds the JSON schema for diff summary results
func (c *Client) buildDiffSummarySchema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"summary": map[string]interface{}{
				"type":        "string",
				"description": "Brief summary of what changed",
			},
			"key_changes": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "List of key changes made",
			},
			"impact_analysis": map[string]interface{}{
				"type":        "string",
				"description": "Analysis of potential impact of changes",
			},
			"confidence": map[string]interface{}{
				"type":        "number",
				"minimum":     0,
				"maximum":     1,
				"description": "Confidence score for the analysis",
			},
		},
		"required":             []string{"summary", "key_changes", "confidence"},
		"additionalProperties": false,
	}
}

// ValidateSemantic performs AI-powered semantic validation of spec content
func (c *Client) ValidateSemantic(ctx context.Context, req SemanticValidationRequest) (*SemanticValidationResult, error) {
	userContent := formatSemanticValidationPrompt(req)
	result := &SemanticValidationResult{}

	// Build JSON schema for structured output
	schema := c.buildSemanticValidationSchema()

	err := c.callStructured(ctx, SystemPromptSemanticValidation, userContent, schema, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// formatSemanticValidationPrompt formats the user prompt for semantic validation
func formatSemanticValidationPrompt(req SemanticValidationRequest) string {
	content := req.SpecContent
	if len(content) > MaxSemanticValidationBytes {
		content = content[:MaxSemanticValidationBytes] + "\n... (truncated)"
	}

	return fmt.Sprintf(`Analyze the following test specification document for semantic quality issues.

Template: %s

Spec Content:
%s

Identify ambiguous, incomplete, inconsistent, or missing context issues.`, req.Template, content)
}

// buildSemanticValidationSchema builds the JSON schema for semantic validation results
func (c *Client) buildSemanticValidationSchema() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"issues": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"type": map[string]interface{}{
							"type": "string",
							"enum": []string{"ambiguous", "incomplete", "inconsistent", "missing_context"},
						},
						"severity": map[string]interface{}{
							"type": "string",
							"enum": []string{"info", "warn", "error"},
						},
						"message":    map[string]interface{}{"type": "string"},
						"row_ref":    map[string]interface{}{"type": "integer", "minimum": 1},
						"field":      map[string]interface{}{"type": "string"},
						"suggestion": map[string]interface{}{"type": "string"},
					},
					"required":             []string{"type", "severity", "message", "suggestion"},
					"additionalProperties": false,
				},
			},
			"overall": map[string]interface{}{
				"type": "string",
				"enum": []string{"good", "needs_improvement", "poor"},
			},
			"score": map[string]interface{}{
				"type":    "number",
				"minimum": 0,
				"maximum": 1,
			},
			"confidence": map[string]interface{}{
				"type":    "number",
				"minimum": 0,
				"maximum": 1,
			},
		},
		"required":             []string{"issues", "overall", "score", "confidence"},
		"additionalProperties": false,
	}
}
