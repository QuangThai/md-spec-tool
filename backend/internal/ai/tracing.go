package ai

import (
	"context"
	"errors"
	"time"
)

// TraceInput is the input to TraceCall
type TraceInput struct {
	Operation string // e.g., "map_columns"
	Model     string // e.g., "gpt-4o-mini"
	CacheHit  bool   // Was result served from cache?
}

// TraceOutput is returned by the traced function with token/confidence info
type TraceOutput struct {
	InputTokens  int64
	OutputTokens int64
	Confidence   float64
}

// AICallTrace contains the full trace of an AI call
type AICallTrace struct {
	Operation    string        `json:"operation"`
	Model        string        `json:"model"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Latency      time.Duration `json:"latency_ns"`
	InputTokens  int64         `json:"input_tokens"`
	OutputTokens int64         `json:"output_tokens"`
	Cost         CostResult    `json:"cost"`
	Confidence   float64       `json:"confidence"`
	CacheHit     bool          `json:"cache_hit"`
	Error        string        `json:"error,omitempty"`
}

// AITracer wraps AI operations with tracing, cost calculation, and metrics
type AITracer struct {
	metrics     *AIMetrics
	costCalc    *CostCalculator
	costTracker *CostTracker
}

// NewAITracer creates a new tracer. All parameters are optional (nil-safe).
func NewAITracer(metrics *AIMetrics, costCalc *CostCalculator, costTracker *CostTracker) *AITracer {
	return &AITracer{
		metrics:     metrics,
		costCalc:    costCalc,
		costTracker: costTracker,
	}
}

// TraceCall wraps an AI operation with tracing, timing, cost calculation, and metrics recording.
// The fn function should perform the actual AI call and return token/confidence info.
func (t *AITracer) TraceCall(ctx context.Context, input TraceInput, fn func(ctx context.Context) (*TraceOutput, error)) (AICallTrace, error) {
	trace := AICallTrace{
		Operation: input.Operation,
		Model:     input.Model,
		StartTime: time.Now(),
		CacheHit:  input.CacheHit,
	}

	// Execute the wrapped function
	output, err := fn(ctx)

	trace.EndTime = time.Now()
	trace.Latency = trace.EndTime.Sub(trace.StartTime)

	// Capture output info
	if output != nil {
		trace.InputTokens = output.InputTokens
		trace.OutputTokens = output.OutputTokens
		trace.Confidence = output.Confidence
	}

	// Determine error type for metrics
	var errType string
	if err != nil {
		trace.Error = err.Error()
		errType = classifyErrorType(err)
	}

	// Calculate cost
	if t.costCalc != nil && (trace.InputTokens > 0 || trace.OutputTokens > 0) {
		trace.Cost = t.costCalc.CalculateCost(trace.Model, trace.InputTokens, trace.OutputTokens)
	}

	// Record in metrics
	if t.metrics != nil {
		t.metrics.RecordCall(AICallMetric{
			Operation:    trace.Operation,
			Model:        trace.Model,
			Latency:      trace.Latency,
			InputTokens:  trace.InputTokens,
			OutputTokens: trace.OutputTokens,
			Cost:         trace.Cost.TotalCost,
			Confidence:   trace.Confidence,
			CacheHit:     trace.CacheHit,
			Error:        errType,
		})
	}

	// Record in cost tracker
	if t.costTracker != nil && trace.Cost.TotalCost > 0 {
		t.costTracker.Record(trace.Operation, trace.Cost)
	}

	return trace, err
}

// classifyErrorType extracts the error category for metrics recording
func classifyErrorType(err error) string {
	if err == nil {
		return ""
	}

	var classified *ClassifiedError
	if errors.As(err, &classified) {
		return string(classified.Category)
	}

	// Check for known sentinel errors
	switch {
	case errors.Is(err, ErrAIRefused):
		return string(ErrorCategoryContent)
	case errors.Is(err, ErrAITruncated):
		return string(ErrorCategoryContent)
	case errors.Is(err, ErrAIContentFiltered):
		return string(ErrorCategoryContent)
	case errors.Is(err, ErrAIUnavailable):
		return string(ErrorCategoryTransient)
	case errors.Is(err, ErrAIRateLimited):
		return string(ErrorCategoryTransient)
	default:
		return "unknown"
	}
}
