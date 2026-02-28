package ai

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
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
	GetModel() string
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
	Type       string `json:"type"`     // "ambiguous", "incomplete", "inconsistent", "missing_context"
	Severity   string `json:"severity"` // "info", "warn", "error"
	Message    string `json:"message"`
	RowRef     int    `json:"row_ref,omitempty"`
	Field      string `json:"field,omitempty"`
	Suggestion string `json:"suggestion"`
}

// Config holds service configuration
type Config struct {
	Model               string        // OpenAI model name (e.g., "gpt-4o-mini")
	PromptProfile       string        // Prompt profile: static_v3 (default) or legacy_v2
	CacheTTL            time.Duration // Cache time-to-live
	MaxCacheSize        int           // Maximum cache entries
	RequestTimeout      time.Duration // Timeout for individual requests
	MaxRetries          int           // Number of retry attempts
	APIKey              string        // OpenAI API key (required)
	RetryBaseDelay      time.Duration // Base delay between retries
	DisableCache        bool          // When true (BYOK), skip cache to avoid cross-user pollution
	MaxCompletionTokens int           // Guardrail: maximum completion tokens per request
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Model:               "gpt-4o-mini",
		PromptProfile:       PromptProfileStaticV3,
		CacheTTL:            1 * time.Hour,
		MaxCacheSize:        1000,
		RequestTimeout:      120 * time.Second,
		MaxRetries:          3,
		RetryBaseDelay:      1 * time.Second,
		MaxCompletionTokens: 1200,
	}
}

// AnalyzePasteRequest is the input for paste analysis
type AnalyzePasteRequest struct {
	Content string `json:"content"`  // Pasted content to analyze
	MaxSize int    `json:"max_size"` // Max content size to process (bytes)
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	client         *Client
	cache          CacheLayer
	validator      *Validator
	model          string
	promptProfile  string
	disableCache   bool // BYOK: skip cache to isolate per-user results
	promptRegistry *PromptRegistry
	cacheMetrics   *CacheMetrics
	cacheCleanup   func() // called on shutdown to close persistent cache

	// Phase 4: Observability
	tracer        *AITracer
	costTracker   *CostTracker
	budgetManager *BudgetManager
	aiMetrics     *AIMetrics
}

// NewService creates a new AI service instance
func NewService(config Config) (*ServiceImpl, error) {
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}

	// Build cache stack from service config
	cacheCfg := CacheConfigFromServiceConfig(config)
	cacheStack, cleanup, cacheErr := BuildCacheStack(cacheCfg)
	if cacheErr != nil {
		return nil, fmt.Errorf("cache setup failed: %w", cacheErr)
	}

	// Attach cache metrics if the stack is a MultiLevelCache
	var cacheMetrics *CacheMetrics
	if multi, ok := cacheStack.(*MultiLevelCache); ok {
		cacheMetrics = AttachMetrics(multi)
	}

	// Phase 4: Initialize observability stack
	aiMetrics := NewAIMetrics()
	costCalc := NewCostCalculator()
	costTracker := NewCostTracker()
	tracer := NewAITracer(aiMetrics, costCalc, costTracker)
	budgetMgr := NewBudgetManager(DefaultBudgetConfig())

	return &ServiceImpl{
		client:         client,
		cache:          cacheStack,
		validator:      NewValidator(),
		model:          config.Model,
		promptProfile:  NormalizePromptProfile(config.PromptProfile),
		disableCache:   config.DisableCache,
		promptRegistry: DefaultPromptRegistry(),
		cacheMetrics:   cacheMetrics,
		cacheCleanup:   cleanup,
		tracer:         tracer,
		costTracker:    costTracker,
		budgetManager:  budgetMgr,
		aiMetrics:      aiMetrics,
	}, nil
}

// GetMode returns "on" when service is active (used for metadata)
func (s *ServiceImpl) GetMode() string {
	return "on"
}

func (s *ServiceImpl) GetModel() string {
	return s.model
}

// MapColumns maps source headers to canonical fields
func (s *ServiceImpl) MapColumns(ctx context.Context, req MapColumnsRequest) (*ColumnMappingResult, error) {
	// Check cache first
	var cacheKey string
	if !s.disableCache {
		var err error
		cacheKey, err = MakeCacheKey(CacheKeyScopeMapColumns, s.model, s.promptCacheVersion(PromptIDColumnMapping, ColumnMappingPromptVersion(s.promptProfile)), SchemaVersionColumnMapping, req)
		if err == nil {
			if cached, ok := s.cache.Get(cacheKey); ok {
				s.recordCacheHit(CacheKeyScopeMapColumns)
				return cached.(*ColumnMappingResult), nil
			}
		}
	}

	// Check budget
	if err := s.checkBudget(CacheKeyScopeMapColumns); err != nil {
		return nil, err
	}

	// Trace the AI call
	var result *ColumnMappingResult
	trace, err := s.tracer.TraceCall(ctx, TraceInput{
		Operation: CacheKeyScopeMapColumns,
		Model:     s.model,
	}, func(ctx context.Context) (*TraceOutput, error) {
		r, usage, callErr := s.client.MapColumns(ctx, req)
		if callErr != nil {
			return nil, callErr
		}
		result = r
		out := &TraceOutput{Confidence: r.Meta.AvgConfidence}
		if usage != nil {
			out.InputTokens = usage.InputTokens
			out.OutputTokens = usage.OutputTokens
		}
		return out, nil
	})

	// Log the call
	s.logAICall(trace, err)

	if err != nil {
		return nil, err
	}

	// Record cost in budget
	s.recordSpend(trace.Cost.TotalCost)

	// Validate result with header count for column_index range check
	if err := s.validator.ValidateColumnMappingWithHeaders(result, len(req.Headers)); err != nil {
		slog.Warn("ai.MapColumns validation failed", "error", err, "headers_count", len(req.Headers), "mapped_count", len(result.CanonicalFields))
		return nil, err
	}

	if !s.disableCache && cacheKey != "" {
		s.cache.Set(cacheKey, result)
	}

	return result, nil
}

// AnalyzePaste analyzes pasted content
func (s *ServiceImpl) AnalyzePaste(ctx context.Context, req AnalyzePasteRequest) (*PasteAnalysis, error) {
	var cacheKey string
	if !s.disableCache {
		var err error
		cacheKey, err = MakeCacheKey(CacheKeyScopeAnalyzePaste, s.model, s.promptCacheVersion(PromptIDPasteAnalysis, PromptVersionPasteAnalysis), SchemaVersionPasteAnalysis, req)
		if err == nil {
			if cached, ok := s.cache.Get(cacheKey); ok {
				s.recordCacheHit(CacheKeyScopeAnalyzePaste)
				return cached.(*PasteAnalysis), nil
			}
		}
	}

	if err := s.checkBudget(CacheKeyScopeAnalyzePaste); err != nil {
		return nil, err
	}

	var result *PasteAnalysis
	trace, err := s.tracer.TraceCall(ctx, TraceInput{
		Operation: CacheKeyScopeAnalyzePaste,
		Model:     s.model,
	}, func(ctx context.Context) (*TraceOutput, error) {
		r, usage, callErr := s.client.AnalyzePaste(ctx, req)
		if callErr != nil {
			return nil, callErr
		}
		result = r
		out := &TraceOutput{Confidence: r.Confidence}
		if usage != nil {
			out.InputTokens = usage.InputTokens
			out.OutputTokens = usage.OutputTokens
		}
		return out, nil
	})
	s.logAICall(trace, err)
	if err != nil {
		return nil, err
	}
	s.recordSpend(trace.Cost.TotalCost)

	if err := s.validator.ValidatePasteAnalysis(result); err != nil {
		slog.Warn("ai.AnalyzePaste validation failed", "error", err)
		return nil, err
	}

	if !s.disableCache && cacheKey != "" {
		s.cache.Set(cacheKey, result)
	}

	return result, nil
}

// GetSuggestions analyzes spec content and returns improvement suggestions
func (s *ServiceImpl) GetSuggestions(ctx context.Context, req SuggestionsRequest) (*SuggestionsResult, error) {
	var cacheKey string
	if !s.disableCache {
		var err error
		cacheKey, err = MakeCacheKey(CacheKeyScopeSuggestions, s.model, s.promptCacheVersion(PromptIDSuggestions, SuggestionsPromptVersion(s.promptProfile)), SchemaVersionSuggestions, req)
		if err == nil {
			if cached, ok := s.cache.Get(cacheKey); ok {
				s.recordCacheHit(CacheKeyScopeSuggestions)
				return cached.(*SuggestionsResult), nil
			}
		}
	}

	if err := s.checkBudget(CacheKeyScopeSuggestions); err != nil {
		return nil, err
	}

	var result *SuggestionsResult
	trace, err := s.tracer.TraceCall(ctx, TraceInput{
		Operation: CacheKeyScopeSuggestions,
		Model:     s.model,
	}, func(ctx context.Context) (*TraceOutput, error) {
		r, usage, callErr := s.client.GetSuggestions(ctx, req)
		if callErr != nil {
			return nil, callErr
		}
		result = r
		out := &TraceOutput{}
		if usage != nil {
			out.InputTokens = usage.InputTokens
			out.OutputTokens = usage.OutputTokens
		}
		return out, nil
	})
	s.logAICall(trace, err)
	if err != nil {
		return nil, err
	}
	s.recordSpend(trace.Cost.TotalCost)

	if !s.disableCache && cacheKey != "" {
		s.cache.Set(cacheKey, result)
	}

	return result, nil
}

// SummarizeDiff generates AI-powered summary of changes between two documents
func (s *ServiceImpl) SummarizeDiff(ctx context.Context, req SummarizeDiffRequest) (*DiffSummary, error) {
	var cacheKey string
	if !s.disableCache {
		var err error
		cacheKey, err = MakeCacheKey(CacheKeyScopeSummarizeDiff, s.model, s.promptCacheVersion(PromptIDDiffSummary, PromptVersionDiffSummary), SchemaVersionDiffSummary, req)
		if err == nil {
			if cached, ok := s.cache.Get(cacheKey); ok {
				s.recordCacheHit(CacheKeyScopeSummarizeDiff)
				return cached.(*DiffSummary), nil
			}
		}
	}

	if err := s.checkBudget(CacheKeyScopeSummarizeDiff); err != nil {
		return nil, err
	}

	var result *DiffSummary
	trace, err := s.tracer.TraceCall(ctx, TraceInput{
		Operation: CacheKeyScopeSummarizeDiff,
		Model:     s.model,
	}, func(ctx context.Context) (*TraceOutput, error) {
		r, usage, callErr := s.client.SummarizeDiff(ctx, req)
		if callErr != nil {
			return nil, callErr
		}
		result = r
		out := &TraceOutput{Confidence: r.Confidence}
		if usage != nil {
			out.InputTokens = usage.InputTokens
			out.OutputTokens = usage.OutputTokens
		}
		return out, nil
	})
	s.logAICall(trace, err)
	if err != nil {
		return nil, err
	}
	s.recordSpend(trace.Cost.TotalCost)

	if !s.disableCache && cacheKey != "" {
		s.cache.Set(cacheKey, result)
	}

	return result, nil
}

// ValidateSemantic performs AI-powered semantic validation of spec content
func (s *ServiceImpl) ValidateSemantic(ctx context.Context, req SemanticValidationRequest) (*SemanticValidationResult, error) {
	var cacheKey string
	if !s.disableCache {
		var err error
		cacheKey, err = MakeCacheKey(CacheKeyScopeValidateSemantic, s.model, s.promptCacheVersion(PromptIDSemanticValidation, PromptVersionSemanticValidation), SchemaVersionSemanticValidation, req)
		if err == nil {
			if cached, ok := s.cache.Get(cacheKey); ok {
				s.recordCacheHit(CacheKeyScopeValidateSemantic)
				return cached.(*SemanticValidationResult), nil
			}
		}
	}

	if err := s.checkBudget(CacheKeyScopeValidateSemantic); err != nil {
		return nil, err
	}

	var result *SemanticValidationResult
	trace, err := s.tracer.TraceCall(ctx, TraceInput{
		Operation: CacheKeyScopeValidateSemantic,
		Model:     s.model,
	}, func(ctx context.Context) (*TraceOutput, error) {
		r, usage, callErr := s.client.ValidateSemantic(ctx, req)
		if callErr != nil {
			return nil, callErr
		}
		result = r
		out := &TraceOutput{Confidence: r.Confidence}
		if usage != nil {
			out.InputTokens = usage.InputTokens
			out.OutputTokens = usage.OutputTokens
		}
		return out, nil
	})
	s.logAICall(trace, err)
	if err != nil {
		return nil, err
	}
	s.recordSpend(trace.Cost.TotalCost)

	if !s.disableCache && cacheKey != "" {
		s.cache.Set(cacheKey, result)
	}

	return result, nil
}

// GetMappingWithFallback returns the column mapping result with confidence-based fallback orchestration.
// If average confidence is below 0.6, it attempts to refine the mapping.
// Confidence < 0.4 mappings are moved to extra_columns (never lost).
// Returns the best result available (original, refined, or fallback with extra_columns).
func (s *ServiceImpl) GetMappingWithFallback(ctx context.Context, req MapColumnsRequest) (*ColumnMappingResult, error) {
	// Get initial mapping
	result, err := s.MapColumns(ctx, req)
	if err != nil {
		return nil, err
	}

	// Check if refinement is needed
	if result.Meta.AvgConfidence < 0.6 {
		slog.Info("low confidence mapping, attempting refinement", "avg_confidence", result.Meta.AvgConfidence)

		// Try to refine with additional context
		refined, refineErr := s.RefineMapping(ctx, result, req)
		if refineErr == nil && refined != nil {
			// Use refined result if successful
			return refined, nil
		}
		slog.Warn("refinement failed, using fallback with extra_columns", "error", refineErr)
	}

	// Apply conservative fallback: move low-confidence mappings to extra_columns
	return s.applyConfidenceFallback(result), nil
}

// RefineMapping attempts to improve low-confidence mappings through prompt chaining.
// It analyzes the original mapping, identifies ambiguous fields, and requests refinement.
func (s *ServiceImpl) RefineMapping(ctx context.Context, original *ColumnMappingResult, originalReq MapColumnsRequest) (*ColumnMappingResult, error) {
	// Build refinement request with context about ambiguous fields
	ambiguousFields := []string{}
	for _, m := range original.CanonicalFields {
		if m.Confidence < 0.7 {
			ambiguousFields = append(ambiguousFields, m.SourceHeader)
		}
	}

	if len(ambiguousFields) == 0 {
		// Nothing to refine
		return original, nil
	}

	// Create refinement prompt with original context and identified ambiguous fields
	sourceLang := originalReq.SourceLang
	if sourceLang == "" {
		sourceLang = originalReq.Language
	}
	refinementReq := MapColumnsRequest{
		Headers:    originalReq.Headers,
		SampleRows: originalReq.SampleRows,
		SchemaHint: originalReq.SchemaHint,
		SourceLang: sourceLang,
		Language:   originalReq.Language,
		RefinementContext: fmt.Sprintf("Previous attempt mapped these headers with low confidence <%s>. Please reconsider these mappings with higher scrutiny, using sample data patterns and semantic analysis. If truly ambiguous, move to extra_columns rather than force an incorrect mapping.",
			strings.Join(ambiguousFields, ", ")),
	}

	// Call client refinement method
	refined, _, err := s.client.RefineMapping(ctx, refinementReq)
	if err != nil {
		return nil, err
	}

	// Validate refined result
	if err := s.validator.ValidateColumnMappingWithHeaders(refined, len(originalReq.Headers)); err != nil {
		slog.Warn("refined mapping validation failed", "error", err)
		return nil, err
	}

	return refined, nil
}

// applyConfidenceFallback moves mappings with confidence < 0.4 to extra_columns (conservative fallback)
// This ensures no data is lost while separating uncertain mappings.
func (s *ServiceImpl) applyConfidenceFallback(result *ColumnMappingResult) *ColumnMappingResult {
	var validMappings []CanonicalFieldMapping
	var extraFallbacks []ExtraColumnMapping

	for _, m := range result.CanonicalFields {
		if m.Confidence < 0.4 {
			// Move to extra_columns with semantic_role hint
			extraFallbacks = append(extraFallbacks, ExtraColumnMapping{
				Name:         m.SourceHeader,
				SemanticRole: fmt.Sprintf("possible_%s (confidence: %.1f%%)", m.CanonicalName, m.Confidence*100),
				ColumnIndex:  m.ColumnIndex,
				Confidence:   m.Confidence,
			})
		} else {
			validMappings = append(validMappings, m)
		}
	}

	// Add existing extra_columns
	result.ExtraColumns = append(extraFallbacks, result.ExtraColumns...)
	result.CanonicalFields = validMappings

	// Recalculate metadata
	if len(validMappings) > 0 {
		sum := 0.0
		for _, m := range validMappings {
			sum += m.Confidence
		}
		result.Meta.AvgConfidence = sum / float64(len(validMappings))
	} else {
		result.Meta.AvgConfidence = 0
	}
	result.Meta.MappedColumns = len(validMappings)
	result.Meta.UnmappedColumns = len(result.ExtraColumns)

	return result
}

// promptCacheVersion returns the hash-based cache version for an operation.
// Falls back to legacy version string if registry lookup fails.
func (s *ServiceImpl) promptCacheVersion(promptID string, fallbackVersion string) string {
	if s.promptRegistry != nil {
		if entry, ok := s.promptRegistry.Get(promptID); ok {
			return entry.CacheVersion()
		}
	}
	return fallbackVersion
}

// --- Phase 4: Observability helpers ---

// checkBudget verifies the AI budget has not been exceeded.
// Returns ErrAIUnavailable-wrapped error if over budget.
func (s *ServiceImpl) checkBudget(operation string) error {
	if s.budgetManager == nil {
		return nil
	}
	ok, remaining := s.budgetManager.CheckBudget()
	if !ok {
		slog.Warn("ai_budget_exceeded", "operation", operation, "remaining", remaining)
		return fmt.Errorf("%w: daily AI budget exceeded", ErrAIUnavailable)
	}
	return nil
}

// recordSpend records cost in the budget manager
func (s *ServiceImpl) recordSpend(cost float64) {
	if s.budgetManager != nil && cost > 0 {
		s.budgetManager.RecordSpend(cost)
	}
}

// recordCacheHit records a cache hit in AI metrics
func (s *ServiceImpl) recordCacheHit(operation string) {
	if s.tracer != nil {
		s.tracer.TraceCall(context.Background(), TraceInput{
			Operation: operation,
			Model:     s.model,
			CacheHit:  true,
		}, func(ctx context.Context) (*TraceOutput, error) {
			return &TraceOutput{}, nil
		})
	}
}

// logAICall logs a structured slog entry for an AI call trace
func (s *ServiceImpl) logAICall(trace AICallTrace, err error) {
	attrs := []any{
		"operation", trace.Operation,
		"model", trace.Model,
		"latency_ms", trace.Latency.Milliseconds(),
		"input_tokens", trace.InputTokens,
		"output_tokens", trace.OutputTokens,
		"cost_usd", fmt.Sprintf("%.6f", trace.Cost.TotalCost),
		"confidence", fmt.Sprintf("%.2f", trace.Confidence),
		"cache_hit", trace.CacheHit,
	}

	if err != nil {
		attrs = append(attrs, "error", err.Error())
		slog.Error("ai_call_failed", attrs...)
	} else {
		slog.Info("ai_call_completed", attrs...)
	}
}

// GetAIMetrics returns the AI pipeline metrics snapshot
func (s *ServiceImpl) GetAIMetrics() AIMetricsSnapshot {
	if s.aiMetrics == nil {
		return AIMetricsSnapshot{}
	}
	return s.aiMetrics.GetSnapshot()
}

// GetCostSummary returns the cost tracking summary
func (s *ServiceImpl) GetCostSummary() CostSummary {
	if s.costTracker == nil {
		return CostSummary{}
	}
	return s.costTracker.GetSummary()
}

// GetBudgetStatus returns the current budget status
func (s *ServiceImpl) GetBudgetStatus() BudgetStatus {
	if s.budgetManager == nil {
		return BudgetStatus{}
	}
	return s.budgetManager.GetStatus()
}

// GetAIMetricsPrometheus returns metrics in Prometheus text format
func (s *ServiceImpl) GetAIMetricsPrometheus() string {
	if s.aiMetrics == nil {
		return ""
	}
	return s.aiMetrics.PrometheusFormat()
}

// PromptInfo contains metadata about a registered prompt (for diagnostics)
type PromptInfo struct {
	ID           string `json:"id"`
	Version      string `json:"version"`
	Hash         string `json:"hash"`
	CacheVersion string `json:"cache_version"`
}

// GetPromptInfo returns metadata about all registered prompts
func (s *ServiceImpl) GetPromptInfo() []PromptInfo {
	if s.promptRegistry == nil {
		return nil
	}
	entries := s.promptRegistry.List()
	info := make([]PromptInfo, len(entries))
	for i, e := range entries {
		info[i] = PromptInfo{
			ID:           e.ID,
			Version:      e.Version,
			Hash:         e.Hash,
			CacheVersion: e.CacheVersion(),
		}
	}
	return info
}

// GetCacheMetrics returns a point-in-time snapshot of cache performance metrics
func (s *ServiceImpl) GetCacheMetrics() CacheMetricsSnapshot {
	if s.cacheMetrics == nil {
		return CacheMetricsSnapshot{}
	}
	return s.cacheMetrics.GetStats()
}

// Close releases resources held by the service (e.g., persistent cache DB).
// Should be called on server shutdown.
func (s *ServiceImpl) Close() error {
	if s.cacheCleanup != nil {
		s.cacheCleanup()
	}
	return nil
}
