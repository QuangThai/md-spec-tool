package ai

import (
	"testing"
	"time"
)

func TestServiceImpl_HasObservabilityStack(t *testing.T) {
	cache := NewMultiLevelCache(NewMemoryCache(100, time.Hour))
	svc := &ServiceImpl{
		model:          "gpt-4o-mini",
		promptProfile:  PromptProfileStaticV3,
		cache:          cache,
		promptRegistry: DefaultPromptRegistry(),
		tracer:         NewAITracer(NewAIMetrics(), NewCostCalculator(), NewCostTracker()),
		costTracker:    NewCostTracker(),
		budgetManager:  NewBudgetManager(DefaultBudgetConfig()),
		aiMetrics:      NewAIMetrics(),
	}

	if svc.tracer == nil {
		t.Error("expected non-nil tracer")
	}
	if svc.costTracker == nil {
		t.Error("expected non-nil cost tracker")
	}
	if svc.budgetManager == nil {
		t.Error("expected non-nil budget manager")
	}
	if svc.aiMetrics == nil {
		t.Error("expected non-nil AI metrics")
	}
}

func TestServiceImpl_GetAIMetrics(t *testing.T) {
	metrics := NewAIMetrics()
	metrics.RecordCall(AICallMetric{
		Operation:    "map_columns",
		Model:        "gpt-4o-mini",
		Latency:      200 * time.Millisecond,
		InputTokens:  500,
		OutputTokens: 300,
		Cost:         0.00035,
	})

	svc := &ServiceImpl{
		model:          "gpt-4o-mini",
		promptProfile:  PromptProfileStaticV3,
		cache:          NewMultiLevelCache(NewMemoryCache(100, time.Hour)),
		promptRegistry: DefaultPromptRegistry(),
		aiMetrics:      metrics,
	}

	snap := svc.GetAIMetrics()
	if snap.TotalCalls != 1 {
		t.Errorf("expected 1 call, got %d", snap.TotalCalls)
	}
	if snap.TotalInputTokens != 500 {
		t.Errorf("expected 500 input tokens, got %d", snap.TotalInputTokens)
	}
}

func TestServiceImpl_GetCostSummary(t *testing.T) {
	tracker := NewCostTracker()
	tracker.Record("map_columns", CostResult{
		TotalCost:    0.001,
		InputTokens:  1000,
		OutputTokens: 500,
	})

	svc := &ServiceImpl{
		model:          "gpt-4o-mini",
		promptProfile:  PromptProfileStaticV3,
		cache:          NewMultiLevelCache(NewMemoryCache(100, time.Hour)),
		promptRegistry: DefaultPromptRegistry(),
		costTracker:    tracker,
	}

	summary := svc.GetCostSummary()
	if summary.TotalCost == 0 {
		t.Error("expected non-zero cost")
	}
	if summary.TotalRequests != 1 {
		t.Errorf("expected 1 request, got %d", summary.TotalRequests)
	}
}

func TestServiceImpl_GetBudgetStatus(t *testing.T) {
	bm := NewBudgetManager(BudgetConfig{
		DailyBudget:       10.00,
		WarningThreshold:  0.80,
		HardStopThreshold: 1.00,
	})
	bm.RecordSpend(3.00)

	svc := &ServiceImpl{
		model:          "gpt-4o-mini",
		promptProfile:  PromptProfileStaticV3,
		cache:          NewMultiLevelCache(NewMemoryCache(100, time.Hour)),
		promptRegistry: DefaultPromptRegistry(),
		budgetManager:  bm,
	}

	status := svc.GetBudgetStatus()
	if status.Spent != 3.00 {
		t.Errorf("expected $3.00 spent, got $%.2f", status.Spent)
	}
	if status.Remaining != 7.00 {
		t.Errorf("expected $7.00 remaining, got $%.2f", status.Remaining)
	}
}

func TestServiceImpl_GetAIMetricsPrometheus(t *testing.T) {
	metrics := NewAIMetrics()
	metrics.RecordCall(AICallMetric{
		Operation:    "map_columns",
		Model:        "gpt-4o-mini",
		InputTokens:  500,
		OutputTokens: 300,
		Cost:         0.00035,
	})

	svc := &ServiceImpl{
		model:          "gpt-4o-mini",
		promptProfile:  PromptProfileStaticV3,
		cache:          NewMultiLevelCache(NewMemoryCache(100, time.Hour)),
		promptRegistry: DefaultPromptRegistry(),
		aiMetrics:      metrics,
	}

	output := svc.GetAIMetricsPrometheus()
	if output == "" {
		t.Error("expected non-empty Prometheus output")
	}
}

func TestServiceImpl_CheckBudget_RejectsOverBudget(t *testing.T) {
	bm := NewBudgetManager(BudgetConfig{
		DailyBudget:       0.01, // very small budget
		WarningThreshold:  0.80,
		HardStopThreshold: 1.00,
	})
	bm.RecordSpend(0.02) // over budget

	svc := &ServiceImpl{
		model:          "gpt-4o-mini",
		promptProfile:  PromptProfileStaticV3,
		cache:          NewMultiLevelCache(NewMemoryCache(100, time.Hour)),
		promptRegistry: DefaultPromptRegistry(),
		budgetManager:  bm,
	}

	err := svc.checkBudget("map_columns")
	if err == nil {
		t.Error("expected error when over budget")
	}
}

func TestServiceImpl_CheckBudget_AllowsUnderBudget(t *testing.T) {
	bm := NewBudgetManager(BudgetConfig{
		DailyBudget:       10.00,
		WarningThreshold:  0.80,
		HardStopThreshold: 1.00,
	})

	svc := &ServiceImpl{
		model:          "gpt-4o-mini",
		promptProfile:  PromptProfileStaticV3,
		cache:          NewMultiLevelCache(NewMemoryCache(100, time.Hour)),
		promptRegistry: DefaultPromptRegistry(),
		budgetManager:  bm,
	}

	err := svc.checkBudget("map_columns")
	if err != nil {
		t.Errorf("expected no error when under budget, got: %v", err)
	}
}

func TestServiceImpl_NilObservability(t *testing.T) {
	// All observability fields nil — should not panic
	svc := &ServiceImpl{
		model:          "gpt-4o-mini",
		promptProfile:  PromptProfileStaticV3,
		cache:          NewMultiLevelCache(NewMemoryCache(100, time.Hour)),
		promptRegistry: DefaultPromptRegistry(),
	}

	snap := svc.GetAIMetrics()
	if snap.TotalCalls != 0 {
		t.Error("expected 0 calls for nil metrics")
	}

	summary := svc.GetCostSummary()
	if summary.TotalCost != 0 {
		t.Error("expected 0 cost for nil tracker")
	}

	status := svc.GetBudgetStatus()
	if status.DailyBudget != 0 {
		t.Error("expected 0 budget for nil manager")
	}

	output := svc.GetAIMetricsPrometheus()
	if output != "" {
		t.Error("expected empty prometheus output for nil metrics")
	}

	// checkBudget should pass with nil manager
	err := svc.checkBudget("test")
	if err != nil {
		t.Errorf("expected nil error for nil budget manager, got: %v", err)
	}
}
