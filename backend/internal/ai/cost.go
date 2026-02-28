package ai

import (
	"sync"
	"time"
)

// ModelPricing holds per-million-token prices for a model
type ModelPricing struct {
	InputPerMillion  float64 // USD per million input tokens
	OutputPerMillion float64 // USD per million output tokens
}

// Known model pricing (updated 2024-12)
var modelPricing = map[string]ModelPricing{
	"gpt-4o-mini":            {InputPerMillion: 0.15, OutputPerMillion: 0.60},
	"gpt-4o-mini-2024-07-18": {InputPerMillion: 0.15, OutputPerMillion: 0.60},
	"gpt-4o":                 {InputPerMillion: 2.50, OutputPerMillion: 10.00},
	"gpt-4o-2024-11-20":      {InputPerMillion: 2.50, OutputPerMillion: 10.00},
	"gpt-4o-2024-08-06":      {InputPerMillion: 2.50, OutputPerMillion: 10.00},
}

// fallbackPricing is used when the model is not recognized
var fallbackPricing = ModelPricing{InputPerMillion: 0.15, OutputPerMillion: 0.60}

// CostResult is the result of a single cost calculation
type CostResult struct {
	Model        string  `json:"model"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	TotalTokens  int64   `json:"total_tokens"`
	InputCost    float64 `json:"input_cost"`
	OutputCost   float64 `json:"output_cost"`
	TotalCost    float64 `json:"total_cost"`
}

// CostCalculator computes cost from token usage
type CostCalculator struct{}

// NewCostCalculator creates a new CostCalculator
func NewCostCalculator() *CostCalculator {
	return &CostCalculator{}
}

// CalculateCost computes USD cost for a given model and token usage.
// Falls back to gpt-4o-mini pricing for unknown models.
func (c *CostCalculator) CalculateCost(model string, inputTokens, outputTokens int64) CostResult {
	pricing, ok := modelPricing[model]
	if !ok {
		pricing = fallbackPricing
	}

	inputCost := float64(inputTokens) * pricing.InputPerMillion / 1_000_000
	outputCost := float64(outputTokens) * pricing.OutputPerMillion / 1_000_000

	return CostResult{
		Model:        model,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  inputTokens + outputTokens,
		InputCost:    inputCost,
		OutputCost:   outputCost,
		TotalCost:    inputCost + outputCost,
	}
}

// GetPricing returns the pricing for a model, falling back to default pricing
// if the model is not recognized.
func (c *CostCalculator) GetPricing(model string) ModelPricing {
	if p, ok := modelPricing[model]; ok {
		return p
	}
	return fallbackPricing
}

// --- Cost Tracker: cumulative cost tracking ---

// OperationCostSummary tracks aggregated cost for a single operation type
type OperationCostSummary struct {
	RequestCount int     `json:"request_count"`
	TotalCost    float64 `json:"total_cost"`
	TotalInput   int64   `json:"total_input_tokens"`
	TotalOutput  int64   `json:"total_output_tokens"`
}

// CostSummary is a snapshot of all tracked costs
type CostSummary struct {
	TotalCost     float64                          `json:"total_cost"`
	TotalRequests int                              `json:"total_requests"`
	TotalInput    int64                            `json:"total_input_tokens"`
	TotalOutput   int64                            `json:"total_output_tokens"`
	ByOperation   map[string]*OperationCostSummary `json:"by_operation"`
	Since         time.Time                        `json:"since"`
}

// CostTracker accumulates cost across all AI operations. It is thread-safe.
type CostTracker struct {
	mu          sync.RWMutex
	byOperation map[string]*OperationCostSummary
	totalCost   float64
	totalReqs   int
	totalInput  int64
	totalOutput int64
	since       time.Time
}

// NewCostTracker creates a new, empty CostTracker
func NewCostTracker() *CostTracker {
	return &CostTracker{
		byOperation: make(map[string]*OperationCostSummary),
		since:       time.Now(),
	}
}

// Record adds a CostResult for the named operation to the running totals.
func (t *CostTracker) Record(operation string, result CostResult) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.totalCost += result.TotalCost
	t.totalReqs++
	t.totalInput += result.InputTokens
	t.totalOutput += result.OutputTokens

	op, ok := t.byOperation[operation]
	if !ok {
		op = &OperationCostSummary{}
		t.byOperation[operation] = op
	}
	op.RequestCount++
	op.TotalCost += result.TotalCost
	op.TotalInput += result.InputTokens
	op.TotalOutput += result.OutputTokens
}

// GetSummary returns a point-in-time snapshot of all tracked costs.
// The returned map values are copies — safe to read without holding the lock.
func (t *CostTracker) GetSummary() CostSummary {
	t.mu.RLock()
	defer t.mu.RUnlock()

	byOp := make(map[string]*OperationCostSummary, len(t.byOperation))
	for k, v := range t.byOperation {
		copied := *v
		byOp[k] = &copied
	}

	return CostSummary{
		TotalCost:     t.totalCost,
		TotalRequests: t.totalReqs,
		TotalInput:    t.totalInput,
		TotalOutput:   t.totalOutput,
		ByOperation:   byOp,
		Since:         t.since,
	}
}

// Reset clears all tracked costs and resets the start time.
func (t *CostTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.byOperation = make(map[string]*OperationCostSummary)
	t.totalCost = 0
	t.totalReqs = 0
	t.totalInput = 0
	t.totalOutput = 0
	t.since = time.Now()
}
