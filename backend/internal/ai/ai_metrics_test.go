package ai

import (
	"strings"
	"testing"
	"time"
)

func TestAIMetrics_RecordCall(t *testing.T) {
	m := NewAIMetrics()
	m.RecordCall(AICallMetric{
		Operation:    "map_columns",
		Model:        "gpt-4o-mini",
		Latency:      200 * time.Millisecond,
		InputTokens:  500,
		OutputTokens: 300,
		Cost:         0.00035,
		Confidence:   0.85,
		CacheHit:     false,
		Error:        "",
	})

	snapshot := m.GetSnapshot()
	if snapshot.TotalCalls != 1 {
		t.Errorf("expected 1 call, got %d", snapshot.TotalCalls)
	}
	if snapshot.TotalInputTokens != 500 {
		t.Errorf("expected 500 input tokens, got %d", snapshot.TotalInputTokens)
	}
	if snapshot.TotalOutputTokens != 300 {
		t.Errorf("expected 300 output tokens, got %d", snapshot.TotalOutputTokens)
	}
	if snapshot.TotalCost == 0 {
		t.Error("expected non-zero total cost")
	}
}

func TestAIMetrics_RecordError(t *testing.T) {
	m := NewAIMetrics()
	m.RecordCall(AICallMetric{
		Operation: "map_columns",
		Model:     "gpt-4o-mini",
		Latency:   100 * time.Millisecond,
		Error:     "transient",
	})

	snapshot := m.GetSnapshot()
	if snapshot.TotalErrors != 1 {
		t.Errorf("expected 1 error, got %d", snapshot.TotalErrors)
	}
	if snapshot.ErrorsByType["transient"] != 1 {
		t.Error("expected 1 transient error")
	}
}

func TestAIMetrics_CacheHitRate(t *testing.T) {
	m := NewAIMetrics()
	m.RecordCall(AICallMetric{Operation: "map_columns", CacheHit: true})
	m.RecordCall(AICallMetric{Operation: "map_columns", CacheHit: true})
	m.RecordCall(AICallMetric{Operation: "map_columns", CacheHit: false})

	snapshot := m.GetSnapshot()
	if snapshot.CacheHits != 2 {
		t.Errorf("expected 2 cache hits, got %d", snapshot.CacheHits)
	}
	// Hit rate should be 66.67%
	expectedRate := 66.67
	if snapshot.CacheHitRate < 60 || snapshot.CacheHitRate > 70 {
		t.Errorf("expected ~%.1f%% cache hit rate, got %.1f%%", expectedRate, snapshot.CacheHitRate)
	}
}

func TestAIMetrics_PerOperationStats(t *testing.T) {
	m := NewAIMetrics()
	m.RecordCall(AICallMetric{Operation: "map_columns", Latency: 200 * time.Millisecond, Cost: 0.001})
	m.RecordCall(AICallMetric{Operation: "map_columns", Latency: 300 * time.Millisecond, Cost: 0.002})
	m.RecordCall(AICallMetric{Operation: "suggestions", Latency: 100 * time.Millisecond, Cost: 0.001})

	snapshot := m.GetSnapshot()
	if len(snapshot.ByOperation) != 2 {
		t.Errorf("expected 2 operations, got %d", len(snapshot.ByOperation))
	}

	mapCols := snapshot.ByOperation["map_columns"]
	if mapCols.CallCount != 2 {
		t.Errorf("expected 2 map_columns calls, got %d", mapCols.CallCount)
	}
	// Average latency should be ~250ms
	avgMs := mapCols.AvgLatencyMs
	if avgMs < 200 || avgMs > 300 {
		t.Errorf("expected avg latency ~250ms, got %.0fms", avgMs)
	}
}

func TestAIMetrics_PrometheusFormat(t *testing.T) {
	m := NewAIMetrics()
	m.RecordCall(AICallMetric{
		Operation:    "map_columns",
		Model:        "gpt-4o-mini",
		Latency:      200 * time.Millisecond,
		InputTokens:  500,
		OutputTokens: 300,
		Cost:         0.00035,
		Confidence:   0.85,
	})

	output := m.PrometheusFormat()
	if output == "" {
		t.Error("expected non-empty Prometheus output")
	}

	// Check for expected metric names
	expectedMetrics := []string{
		"ai_calls_total",
		"ai_tokens_input_total",
		"ai_tokens_output_total",
		"ai_cost_usd_total",
		"ai_errors_total",
		"ai_latency_seconds",
		"ai_cache_hits_total",
	}
	for _, metric := range expectedMetrics {
		if !strings.Contains(output, metric) {
			t.Errorf("expected metric %q in output", metric)
		}
	}

	// Check for HELP and TYPE annotations
	if !strings.Contains(output, "# HELP") {
		t.Error("expected HELP annotations in Prometheus output")
	}
	if !strings.Contains(output, "# TYPE") {
		t.Error("expected TYPE annotations in Prometheus output")
	}
}

func TestAIMetrics_Reset(t *testing.T) {
	m := NewAIMetrics()
	m.RecordCall(AICallMetric{Operation: "test", Cost: 0.001})
	m.Reset()

	snapshot := m.GetSnapshot()
	if snapshot.TotalCalls != 0 {
		t.Errorf("expected 0 calls after reset, got %d", snapshot.TotalCalls)
	}
}

func TestAIMetrics_AvgConfidence(t *testing.T) {
	m := NewAIMetrics()
	m.RecordCall(AICallMetric{Operation: "map_columns", Confidence: 0.9})
	m.RecordCall(AICallMetric{Operation: "map_columns", Confidence: 0.8})
	m.RecordCall(AICallMetric{Operation: "map_columns", Confidence: 0.7})

	snapshot := m.GetSnapshot()
	// Average confidence should be 0.8
	if snapshot.AvgConfidence < 0.79 || snapshot.AvgConfidence > 0.81 {
		t.Errorf("expected avg confidence ~0.8, got %.2f", snapshot.AvgConfidence)
	}
}

func TestAIMetrics_ConcurrentSafety(t *testing.T) {
	m := NewAIMetrics()
	done := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func() {
			m.RecordCall(AICallMetric{
				Operation:    "test",
				Model:        "gpt-4o-mini",
				Latency:      100 * time.Millisecond,
				InputTokens:  100,
				OutputTokens: 50,
				Cost:         0.0001,
			})
			m.GetSnapshot()
			m.PrometheusFormat()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 100; i++ {
		<-done
	}
	snapshot := m.GetSnapshot()
	if snapshot.TotalCalls != 100 {
		t.Errorf("expected 100 calls, got %d", snapshot.TotalCalls)
	}
}
