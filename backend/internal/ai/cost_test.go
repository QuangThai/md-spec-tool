package ai

import (
	"testing"
)

func TestCalculateCost_GPT4oMini(t *testing.T) {
	calc := NewCostCalculator()
	result := calc.CalculateCost("gpt-4o-mini", 1000, 500)

	// Input: 1000 tokens * $0.15/1M = $0.00015
	// Output: 500 tokens * $0.60/1M = $0.0003
	// Total: $0.00045
	expectedTotal := 0.00045
	tolerance := 0.000001

	if diff := result.TotalCost - expectedTotal; diff > tolerance || diff < -tolerance {
		t.Errorf("expected total cost %.6f, got %.6f", expectedTotal, result.TotalCost)
	}
	if result.InputCost == 0 {
		t.Error("expected non-zero input cost")
	}
	if result.OutputCost == 0 {
		t.Error("expected non-zero output cost")
	}
	if result.Model != "gpt-4o-mini" {
		t.Errorf("expected model gpt-4o-mini, got %s", result.Model)
	}
}

func TestCalculateCost_GPT4o(t *testing.T) {
	calc := NewCostCalculator()
	result := calc.CalculateCost("gpt-4o", 1000, 500)

	// Input: 1000 * $2.50/1M = $0.0025
	// Output: 500 * $10.00/1M = $0.005
	// Total: $0.0075
	expectedTotal := 0.0075
	tolerance := 0.000001

	if diff := result.TotalCost - expectedTotal; diff > tolerance || diff < -tolerance {
		t.Errorf("expected total cost %.6f, got %.6f", expectedTotal, result.TotalCost)
	}
}

func TestCalculateCost_UnknownModel(t *testing.T) {
	calc := NewCostCalculator()
	result := calc.CalculateCost("unknown-model", 1000, 500)

	// Should fall back to gpt-4o-mini pricing
	if result.TotalCost == 0 {
		t.Error("expected non-zero cost even for unknown model (fallback pricing)")
	}
	if result.Model != "unknown-model" {
		t.Errorf("expected model unknown-model, got %s", result.Model)
	}
}

func TestCalculateCost_ZeroTokens(t *testing.T) {
	calc := NewCostCalculator()
	result := calc.CalculateCost("gpt-4o-mini", 0, 0)

	if result.TotalCost != 0 {
		t.Errorf("expected zero cost for zero tokens, got %.6f", result.TotalCost)
	}
}

func TestCostTracker_TrackByOperation(t *testing.T) {
	tracker := NewCostTracker()
	tracker.Record("map_columns", CostResult{
		Model:        "gpt-4o-mini",
		InputCost:    0.00015,
		OutputCost:   0.0003,
		TotalCost:    0.00045,
		InputTokens:  1000,
		OutputTokens: 500,
	})
	tracker.Record("map_columns", CostResult{
		TotalCost:    0.00045,
		InputTokens:  1000,
		OutputTokens: 500,
	})
	tracker.Record("suggestions", CostResult{
		TotalCost:    0.0001,
		InputTokens:  500,
		OutputTokens: 200,
	})

	summary := tracker.GetSummary()
	if summary.TotalCost == 0 {
		t.Error("expected non-zero total cost")
	}
	if summary.TotalRequests != 3 {
		t.Errorf("expected 3 total requests, got %d", summary.TotalRequests)
	}
	if len(summary.ByOperation) != 2 {
		t.Errorf("expected 2 operations, got %d", len(summary.ByOperation))
	}
	if summary.ByOperation["map_columns"].RequestCount != 2 {
		t.Errorf("expected 2 map_columns requests, got %d", summary.ByOperation["map_columns"].RequestCount)
	}
}

func TestCostTracker_Reset(t *testing.T) {
	tracker := NewCostTracker()
	tracker.Record("test", CostResult{TotalCost: 0.001})
	tracker.Reset()

	summary := tracker.GetSummary()
	if summary.TotalCost != 0 {
		t.Errorf("expected zero cost after reset, got %.6f", summary.TotalCost)
	}
}

func TestCostTracker_ConcurrentSafety(t *testing.T) {
	tracker := NewCostTracker()
	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			tracker.Record("test", CostResult{TotalCost: 0.001})
			tracker.GetSummary()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
	summary := tracker.GetSummary()
	if summary.TotalRequests != 50 {
		t.Errorf("expected 50 requests, got %d", summary.TotalRequests)
	}
}

func TestModelPricing_AllModels(t *testing.T) {
	models := []string{"gpt-4o-mini", "gpt-4o", "gpt-4o-2024-11-20"}
	calc := NewCostCalculator()
	for _, model := range models {
		result := calc.CalculateCost(model, 1000, 500)
		if result.TotalCost == 0 {
			t.Errorf("expected non-zero cost for model %s", model)
		}
	}
}
