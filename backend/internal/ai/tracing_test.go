package ai

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTraceAICall_Success(t *testing.T) {
	metrics := NewAIMetrics()
	costCalc := NewCostCalculator()
	costTracker := NewCostTracker()
	tracer := NewAITracer(metrics, costCalc, costTracker)

	trace, err := tracer.TraceCall(context.Background(), TraceInput{
		Operation: "map_columns",
		Model:     "gpt-4o-mini",
	}, func(ctx context.Context) (*TraceOutput, error) {
		return &TraceOutput{
			InputTokens:  500,
			OutputTokens: 300,
			Confidence:   0.85,
		}, nil
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if trace.Latency <= 0 {
		t.Error("expected positive latency")
	}
	if trace.Cost.TotalCost == 0 {
		t.Error("expected non-zero cost")
	}
	if trace.InputTokens != 500 {
		t.Errorf("expected 500 input tokens, got %d", trace.InputTokens)
	}
	if trace.OutputTokens != 300 {
		t.Errorf("expected 300 output tokens, got %d", trace.OutputTokens)
	}

	// Should be recorded in metrics
	snap := metrics.GetSnapshot()
	if snap.TotalCalls != 1 {
		t.Errorf("expected 1 call in metrics, got %d", snap.TotalCalls)
	}

	// Should be recorded in cost tracker
	summary := costTracker.GetSummary()
	if summary.TotalCost == 0 {
		t.Error("expected non-zero cost in tracker")
	}
}

func TestTraceAICall_Error(t *testing.T) {
	metrics := NewAIMetrics()
	costCalc := NewCostCalculator()
	costTracker := NewCostTracker()
	tracer := NewAITracer(metrics, costCalc, costTracker)

	expectedErr := errors.New("api error")
	trace, err := tracer.TraceCall(context.Background(), TraceInput{
		Operation: "map_columns",
		Model:     "gpt-4o-mini",
	}, func(ctx context.Context) (*TraceOutput, error) {
		return nil, expectedErr
	})

	if err == nil {
		t.Fatal("expected error")
	}
	if trace.Error == "" {
		t.Error("expected non-empty error in trace")
	}

	// Error should be recorded in metrics
	snap := metrics.GetSnapshot()
	if snap.TotalErrors != 1 {
		t.Errorf("expected 1 error in metrics, got %d", snap.TotalErrors)
	}
}

func TestTraceAICall_ClassifiedError(t *testing.T) {
	metrics := NewAIMetrics()
	costCalc := NewCostCalculator()
	costTracker := NewCostTracker()
	tracer := NewAITracer(metrics, costCalc, costTracker)

	_, _ = tracer.TraceCall(context.Background(), TraceInput{
		Operation: "map_columns",
		Model:     "gpt-4o-mini",
	}, func(ctx context.Context) (*TraceOutput, error) {
		return nil, &ClassifiedError{
			Category: ErrorCategoryTransient,
			Original: errors.New("timeout"),
		}
	})

	snap := metrics.GetSnapshot()
	if snap.ErrorsByType["transient"] != 1 {
		t.Error("expected transient error to be tracked by type")
	}
}

func TestTraceAICall_CacheHit(t *testing.T) {
	metrics := NewAIMetrics()
	costCalc := NewCostCalculator()
	costTracker := NewCostTracker()
	tracer := NewAITracer(metrics, costCalc, costTracker)

	trace, _ := tracer.TraceCall(context.Background(), TraceInput{
		Operation: "map_columns",
		Model:     "gpt-4o-mini",
		CacheHit:  true,
	}, func(ctx context.Context) (*TraceOutput, error) {
		return &TraceOutput{}, nil
	})

	if !trace.CacheHit {
		t.Error("expected cache hit to be recorded")
	}

	snap := metrics.GetSnapshot()
	if snap.CacheHits != 1 {
		t.Errorf("expected 1 cache hit, got %d", snap.CacheHits)
	}
}

func TestTraceAICall_LatencyRecorded(t *testing.T) {
	metrics := NewAIMetrics()
	costCalc := NewCostCalculator()
	tracer := NewAITracer(metrics, costCalc, nil)

	trace, _ := tracer.TraceCall(context.Background(), TraceInput{
		Operation: "map_columns",
		Model:     "gpt-4o-mini",
	}, func(ctx context.Context) (*TraceOutput, error) {
		time.Sleep(10 * time.Millisecond)
		return &TraceOutput{InputTokens: 100, OutputTokens: 50}, nil
	})

	if trace.Latency < 10*time.Millisecond {
		t.Errorf("expected latency >= 10ms, got %v", trace.Latency)
	}
}

func TestTraceAICall_NilMetrics(t *testing.T) {
	// Should work even without metrics/tracker
	tracer := NewAITracer(nil, nil, nil)

	trace, err := tracer.TraceCall(context.Background(), TraceInput{
		Operation: "test",
		Model:     "gpt-4o-mini",
	}, func(ctx context.Context) (*TraceOutput, error) {
		return &TraceOutput{InputTokens: 100}, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trace.InputTokens != 100 {
		t.Errorf("expected 100 input tokens, got %d", trace.InputTokens)
	}
}

func TestAICallTrace_StructFields(t *testing.T) {
	trace := AICallTrace{
		Operation:    "map_columns",
		Model:        "gpt-4o-mini",
		StartTime:    time.Now().Add(-100 * time.Millisecond),
		EndTime:      time.Now(),
		Latency:      100 * time.Millisecond,
		InputTokens:  500,
		OutputTokens: 300,
		Confidence:   0.85,
		CacheHit:     false,
		Error:        "",
	}

	if trace.Operation != "map_columns" {
		t.Error("operation field mismatch")
	}
	if trace.Latency != 100*time.Millisecond {
		t.Error("latency field mismatch")
	}
}
