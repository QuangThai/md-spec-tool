package ai

import (
	"context"
	"time"
)

const (
	// Cache key scopes
	CacheKeyScopeMapColumns       = "map_columns"
	CacheKeyScopeAnalyzePaste     = "analyze_paste"
	CacheKeyScopeSuggestions      = "suggestions"
	CacheKeyScopeSummarizeDiff    = "summarize_diff"
	CacheKeyScopeValidateSemantic = "validate_semantic"

	// Schema versions for diff and semantic
	SchemaVersionDiffSummary        = "v1"
	SchemaVersionSemanticValidation = "v1"
)

// Service defines high-level AI operations for the converter domain
type Service interface {
	MapColumns(ctx context.Context, req MapColumnsRequest) (*ColumnMappingResult, error)
	AnalyzePaste(ctx context.Context, req AnalyzePasteRequest) (*PasteAnalysis, error)
	GetSuggestions(ctx context.Context, req SuggestionsRequest) (*SuggestionsResult, error)
	SummarizeDiff(ctx context.Context, req SummarizeDiffRequest) (*DiffSummary, error)
	ValidateSemantic(ctx context.Context, req SemanticValidationRequest) (*SemanticValidationResult, error)
	GetMode() string // Returns "on" when service is active
}

// SummarizeDiffRequest is the input for diff summarization
type SummarizeDiffRequest struct {
	Before   string `json:"before"`
	After    string `json:"after"`
	DiffText string `json:"diff_text"`
}

// DiffSummary is the AI-generated summary of changes
type DiffSummary struct {
	Summary        string   `json:"summary"`
	KeyChanges     []string `json:"key_changes"`
	ImpactAnalysis string   `json:"impact_analysis,omitempty"`
	Confidence     float64  `json:"confidence"`
}

// SemanticValidationRequest is the input for AI semantic validation
type SemanticValidationRequest struct {
	SpecContent string `json:"spec_content"`
	Template    string `json:"template"`
}

// SemanticValidationResult contains AI-detected semantic issues
type SemanticValidationResult struct {
	Issues     []SemanticIssue `json:"issues"`
	Overall    string          `json:"overall"` // "good", "needs_improvement", "poor"
	Score      float64         `json:"score"`   // 0-1 quality score
	Confidence float64         `json:"confidence"`
}

// SemanticIssue represents a semantic validation issue found by AI
type SemanticIssue struct {
	Type       string `json:"type"`        // "ambiguous", "incomplete", "inconsistent", "missing_context"
	Severity   string `json:"severity"`    // "info", "warn", "error"
	Message    string `json:"message"`
	RowRef     int    `json:"row_ref,omitempty"`
	Field      string `json:"field,omitempty"`
	Suggestion string `json:"suggestion"`
}

// Config holds service configuration
type Config struct {
	Model          string        // OpenAI model name (e.g., "gpt-4o-mini")
	CacheTTL       time.Duration // Cache time-to-live
	MaxCacheSize   int           // Maximum cache entries
	RequestTimeout time.Duration // Timeout for individual requests
	MaxRetries     int           // Number of retry attempts
	APIKey         string        // OpenAI API key (required)
	RetryBaseDelay time.Duration // Base delay between retries
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Model:          "gpt-4o-mini",
		CacheTTL:       1 * time.Hour,
		MaxCacheSize:   1000,
		RequestTimeout: 120 * time.Second,
		MaxRetries:     3,
		RetryBaseDelay: 1 * time.Second,
	}
}

// AnalyzePasteRequest is the input for paste analysis
type AnalyzePasteRequest struct {
	Content string `json:"content"`  // Pasted content to analyze
	MaxSize int    `json:"max_size"` // Max content size to process (bytes)
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	client    *Client
	cache     *Cache
	validator *Validator
	model     string
}

// NewService creates a new AI service instance
func NewService(config Config) (*ServiceImpl, error) {
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}

	return &ServiceImpl{
		client:    client,
		cache:     NewCache(config.MaxCacheSize, config.CacheTTL),
		validator: NewValidator(),
		model:     config.Model,
	}, nil
}

// GetMode returns "on" when service is active (used for metadata)
func (s *ServiceImpl) GetMode() string {
	return "on"
}

// MapColumns maps source headers to canonical fields
func (s *ServiceImpl) MapColumns(ctx context.Context, req MapColumnsRequest) (*ColumnMappingResult, error) {
	// Try cache first
	cacheKey, err := MakeCacheKey(CacheKeyScopeMapColumns, s.model, PromptVersionColumnMapping, SchemaVersionColumnMapping, req)
	if err == nil {
		if cached, ok := s.cache.Get(cacheKey); ok {
			return cached.(*ColumnMappingResult), nil
		}
	}

	result, err := s.client.MapColumns(ctx, req)
	if err != nil {
		return nil, err
	}

	// Validate result
	if err := s.validator.ValidateColumnMapping(result); err != nil {
		return nil, err
	}

	// Cache the result
	if cacheKey != "" {
		s.cache.Set(cacheKey, result)
	}

	return result, nil
}

// AnalyzePaste analyzes pasted content
func (s *ServiceImpl) AnalyzePaste(ctx context.Context, req AnalyzePasteRequest) (*PasteAnalysis, error) {
	// Try cache first
	cacheKey, err := MakeCacheKey(CacheKeyScopeAnalyzePaste, s.model, PromptVersionPasteAnalysis, SchemaVersionPasteAnalysis, req)
	if err == nil {
		if cached, ok := s.cache.Get(cacheKey); ok {
			return cached.(*PasteAnalysis), nil
		}
	}

	result, err := s.client.AnalyzePaste(ctx, req)
	if err != nil {
		return nil, err
	}

	// Validate result
	if err := s.validator.ValidatePasteAnalysis(result); err != nil {
		return nil, err
	}

	// Cache the result
	if cacheKey != "" {
		s.cache.Set(cacheKey, result)
	}

	return result, nil
}

// GetSuggestions analyzes spec content and returns improvement suggestions
func (s *ServiceImpl) GetSuggestions(ctx context.Context, req SuggestionsRequest) (*SuggestionsResult, error) {
	// Try cache first
	cacheKey, err := MakeCacheKey(CacheKeyScopeSuggestions, s.model, PromptVersionSuggestions, SchemaVersionSuggestions, req)
	if err == nil {
		if cached, ok := s.cache.Get(cacheKey); ok {
			return cached.(*SuggestionsResult), nil
		}
	}

	result, err := s.client.GetSuggestions(ctx, req)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if cacheKey != "" {
		s.cache.Set(cacheKey, result)
	}

	return result, nil
}

// SummarizeDiff generates AI-powered summary of changes between two documents
func (s *ServiceImpl) SummarizeDiff(ctx context.Context, req SummarizeDiffRequest) (*DiffSummary, error) {
	// Try cache first
	cacheKey, err := MakeCacheKey(CacheKeyScopeSummarizeDiff, s.model, PromptVersionDiffSummary, SchemaVersionDiffSummary, req)
	if err == nil {
		if cached, ok := s.cache.Get(cacheKey); ok {
			return cached.(*DiffSummary), nil
		}
	}

	result, err := s.client.SummarizeDiff(ctx, req)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if cacheKey != "" {
		s.cache.Set(cacheKey, result)
	}

	return result, nil
}

// ValidateSemantic performs AI-powered semantic validation of spec content
func (s *ServiceImpl) ValidateSemantic(ctx context.Context, req SemanticValidationRequest) (*SemanticValidationResult, error) {
	// Try cache first
	cacheKey, err := MakeCacheKey(CacheKeyScopeValidateSemantic, s.model, PromptVersionSemanticValidation, SchemaVersionSemanticValidation, req)
	if err == nil {
		if cached, ok := s.cache.Get(cacheKey); ok {
			return cached.(*SemanticValidationResult), nil
		}
	}

	result, err := s.client.ValidateSemantic(ctx, req)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if cacheKey != "" {
		s.cache.Set(cacheKey, result)
	}

	return result, nil
}
