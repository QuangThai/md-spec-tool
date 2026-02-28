package ai

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// AICallMetric represents a single AI call measurement
type AICallMetric struct {
	Operation    string        // e.g., "map_columns"
	Model        string        // e.g., "gpt-4o-mini"
	Latency      time.Duration // How long the call took
	InputTokens  int64         // Prompt tokens
	OutputTokens int64         // Completion tokens
	Cost         float64       // USD cost
	Confidence   float64       // Result confidence (0-1)
	CacheHit     bool          // Was this served from cache?
	Error        string        // Error type if failed ("transient", "permanent", "content", "")
}

// OperationMetrics tracks metrics for a single operation
type OperationMetrics struct {
	CallCount         int64   `json:"call_count"`
	ErrorCount        int64   `json:"error_count"`
	TotalCost         float64 `json:"total_cost"`
	TotalInputTokens  int64   `json:"total_input_tokens"`
	TotalOutputTokens int64   `json:"total_output_tokens"`
	TotalLatencyMs    float64 `json:"total_latency_ms"`
	AvgLatencyMs      float64 `json:"avg_latency_ms"`
	TotalConfidence   float64 `json:"total_confidence"`
	ConfidenceCount   int64   `json:"confidence_count"`
}

// AIMetricsSnapshot is a point-in-time snapshot of all metrics
type AIMetricsSnapshot struct {
	TotalCalls        int64                        `json:"total_calls"`
	TotalErrors       int64                        `json:"total_errors"`
	TotalInputTokens  int64                        `json:"total_input_tokens"`
	TotalOutputTokens int64                        `json:"total_output_tokens"`
	TotalCost         float64                      `json:"total_cost"`
	AvgConfidence     float64                      `json:"avg_confidence"`
	CacheHits         int64                        `json:"cache_hits"`
	CacheHitRate      float64                      `json:"cache_hit_rate"`
	ErrorsByType      map[string]int64             `json:"errors_by_type"`
	ByOperation       map[string]*OperationMetrics `json:"by_operation"`
}

// AIMetrics tracks all AI pipeline metrics
type AIMetrics struct {
	mu sync.RWMutex

	totalCalls        int64
	totalErrors       int64
	totalInputTokens  int64
	totalOutputTokens int64
	totalCost         float64
	totalConfidence   float64
	confidenceCount   int64
	cacheHits         int64

	errorsByType map[string]int64
	byOperation  map[string]*OperationMetrics
}

// NewAIMetrics creates a new metrics tracker
func NewAIMetrics() *AIMetrics {
	return &AIMetrics{
		errorsByType: make(map[string]int64),
		byOperation:  make(map[string]*OperationMetrics),
	}
}

// RecordCall records metrics for a single AI call
func (m *AIMetrics) RecordCall(metric AICallMetric) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalCalls++
	m.totalInputTokens += metric.InputTokens
	m.totalOutputTokens += metric.OutputTokens
	m.totalCost += metric.Cost

	if metric.Confidence > 0 {
		m.totalConfidence += metric.Confidence
		m.confidenceCount++
	}

	if metric.CacheHit {
		m.cacheHits++
	}

	if metric.Error != "" {
		m.totalErrors++
		m.errorsByType[metric.Error]++
	}

	// Per-operation tracking
	op, ok := m.byOperation[metric.Operation]
	if !ok {
		op = &OperationMetrics{}
		m.byOperation[metric.Operation] = op
	}
	op.CallCount++
	op.TotalCost += metric.Cost
	op.TotalInputTokens += metric.InputTokens
	op.TotalOutputTokens += metric.OutputTokens
	op.TotalLatencyMs += float64(metric.Latency.Milliseconds())
	if op.CallCount > 0 {
		op.AvgLatencyMs = op.TotalLatencyMs / float64(op.CallCount)
	}
	if metric.Confidence > 0 {
		op.TotalConfidence += metric.Confidence
		op.ConfidenceCount++
	}
	if metric.Error != "" {
		op.ErrorCount++
	}
}

// GetSnapshot returns a point-in-time copy of all metrics
func (m *AIMetrics) GetSnapshot() AIMetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	errorsCopy := make(map[string]int64, len(m.errorsByType))
	for k, v := range m.errorsByType {
		errorsCopy[k] = v
	}

	opCopy := make(map[string]*OperationMetrics, len(m.byOperation))
	for k, v := range m.byOperation {
		vc := *v
		opCopy[k] = &vc
	}

	var avgConfidence float64
	if m.confidenceCount > 0 {
		avgConfidence = m.totalConfidence / float64(m.confidenceCount)
	}

	var cacheHitRate float64
	if m.totalCalls > 0 {
		cacheHitRate = float64(m.cacheHits) / float64(m.totalCalls) * 100
	}

	return AIMetricsSnapshot{
		TotalCalls:        m.totalCalls,
		TotalErrors:       m.totalErrors,
		TotalInputTokens:  m.totalInputTokens,
		TotalOutputTokens: m.totalOutputTokens,
		TotalCost:         m.totalCost,
		AvgConfidence:     avgConfidence,
		CacheHits:         m.cacheHits,
		CacheHitRate:      cacheHitRate,
		ErrorsByType:      errorsCopy,
		ByOperation:       opCopy,
	}
}

// PrometheusFormat outputs metrics in Prometheus text exposition format
func (m *AIMetrics) PrometheusFormat() string {
	snap := m.GetSnapshot()
	var b strings.Builder

	// Total calls
	b.WriteString("# HELP ai_calls_total Total number of AI API calls\n")
	b.WriteString("# TYPE ai_calls_total counter\n")
	b.WriteString(fmt.Sprintf("ai_calls_total %d\n\n", snap.TotalCalls))

	// Input tokens
	b.WriteString("# HELP ai_tokens_input_total Total input tokens consumed\n")
	b.WriteString("# TYPE ai_tokens_input_total counter\n")
	b.WriteString(fmt.Sprintf("ai_tokens_input_total %d\n\n", snap.TotalInputTokens))

	// Output tokens
	b.WriteString("# HELP ai_tokens_output_total Total output tokens generated\n")
	b.WriteString("# TYPE ai_tokens_output_total counter\n")
	b.WriteString(fmt.Sprintf("ai_tokens_output_total %d\n\n", snap.TotalOutputTokens))

	// Cost
	b.WriteString("# HELP ai_cost_usd_total Total cost in USD\n")
	b.WriteString("# TYPE ai_cost_usd_total counter\n")
	b.WriteString(fmt.Sprintf("ai_cost_usd_total %.6f\n\n", snap.TotalCost))

	// Errors
	b.WriteString("# HELP ai_errors_total Total number of AI errors\n")
	b.WriteString("# TYPE ai_errors_total counter\n")
	b.WriteString(fmt.Sprintf("ai_errors_total %d\n", snap.TotalErrors))
	for errType, count := range snap.ErrorsByType {
		b.WriteString(fmt.Sprintf("ai_errors_total{type=%q} %d\n", errType, count))
	}
	b.WriteString("\n")

	// Latency per operation
	b.WriteString("# HELP ai_latency_seconds Average AI call latency in seconds\n")
	b.WriteString("# TYPE ai_latency_seconds gauge\n")
	for op, metrics := range snap.ByOperation {
		b.WriteString(fmt.Sprintf("ai_latency_seconds{operation=%q} %.3f\n", op, metrics.AvgLatencyMs/1000))
	}
	b.WriteString("\n")

	// Cache hits
	b.WriteString("# HELP ai_cache_hits_total Total cache hits\n")
	b.WriteString("# TYPE ai_cache_hits_total counter\n")
	b.WriteString(fmt.Sprintf("ai_cache_hits_total %d\n\n", snap.CacheHits))

	// Average confidence
	b.WriteString("# HELP ai_confidence_avg Average result confidence\n")
	b.WriteString("# TYPE ai_confidence_avg gauge\n")
	b.WriteString(fmt.Sprintf("ai_confidence_avg %.4f\n\n", snap.AvgConfidence))

	// Per-operation call counts
	b.WriteString("# HELP ai_operation_calls_total Calls per operation\n")
	b.WriteString("# TYPE ai_operation_calls_total counter\n")
	for op, metrics := range snap.ByOperation {
		b.WriteString(fmt.Sprintf("ai_operation_calls_total{operation=%q} %d\n", op, metrics.CallCount))
	}

	return b.String()
}

// ServeMetrics writes Prometheus-format metrics to an http.ResponseWriter
func (m *AIMetrics) ServeMetrics(w http.ResponseWriter) {
	output := m.PrometheusFormat()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = w.Write([]byte(output))
}

// Reset clears all metrics
func (m *AIMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalCalls = 0
	m.totalErrors = 0
	m.totalInputTokens = 0
	m.totalOutputTokens = 0
	m.totalCost = 0
	m.totalConfidence = 0
	m.confidenceCount = 0
	m.cacheHits = 0
	m.errorsByType = make(map[string]int64)
	m.byOperation = make(map[string]*OperationMetrics)
}
