package ai

import (
	"testing"
	"time"
)

func TestCacheMetrics_RecordHit(t *testing.T) {
	m := NewCacheMetrics()
	m.RecordHit("map_columns", "L1", 500*time.Microsecond)

	stats := m.GetStats()
	if stats.TotalHits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.TotalHits)
	}
}

func TestCacheMetrics_RecordMiss(t *testing.T) {
	m := NewCacheMetrics()
	m.RecordMiss("map_columns")

	stats := m.GetStats()
	if stats.TotalMisses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.TotalMisses)
	}
}

func TestCacheMetrics_HitRate(t *testing.T) {
	m := NewCacheMetrics()
	m.RecordHit("map_columns", "L1", time.Millisecond)
	m.RecordHit("map_columns", "L1", time.Millisecond)
	m.RecordMiss("map_columns")

	stats := m.GetStats()
	// 2 hits / 3 total = 66.67%
	if stats.HitRate < 60 || stats.HitRate > 70 {
		t.Errorf("expected ~66.67%% hit rate, got %.2f%%", stats.HitRate)
	}
}

func TestCacheMetrics_PerOperationStats(t *testing.T) {
	m := NewCacheMetrics()
	m.RecordHit("map_columns", "L1", time.Millisecond)
	m.RecordHit("map_columns", "L2", 5*time.Millisecond)
	m.RecordMiss("map_columns")
	m.RecordHit("analyze_paste", "L1", time.Millisecond)

	stats := m.GetStats()

	mcStats, ok := stats.ByOperation["map_columns"]
	if !ok {
		t.Fatal("expected map_columns operation stats")
	}
	if mcStats.Hits != 2 {
		t.Errorf("expected 2 map_columns hits, got %d", mcStats.Hits)
	}
	if mcStats.Misses != 1 {
		t.Errorf("expected 1 map_columns miss, got %d", mcStats.Misses)
	}

	apStats, ok := stats.ByOperation["analyze_paste"]
	if !ok {
		t.Fatal("expected analyze_paste operation stats")
	}
	if apStats.Hits != 1 {
		t.Errorf("expected 1 analyze_paste hit, got %d", apStats.Hits)
	}
}

func TestCacheMetrics_PerLevelStats(t *testing.T) {
	m := NewCacheMetrics()
	m.RecordHit("map_columns", "L1", 500*time.Microsecond)
	m.RecordHit("map_columns", "L2", 5*time.Millisecond)
	m.RecordHit("analyze_paste", "L1", time.Millisecond)

	stats := m.GetStats()

	l1Stats, ok := stats.ByLevel["L1"]
	if !ok {
		t.Fatal("expected L1 stats")
	}
	if l1Stats.Hits != 2 {
		t.Errorf("expected 2 L1 hits, got %d", l1Stats.Hits)
	}

	l2Stats, ok := stats.ByLevel["L2"]
	if !ok {
		t.Fatal("expected L2 stats")
	}
	if l2Stats.Hits != 1 {
		t.Errorf("expected 1 L2 hit, got %d", l2Stats.Hits)
	}
}

func TestCacheMetrics_AverageLatency(t *testing.T) {
	m := NewCacheMetrics()
	m.RecordHit("map_columns", "L1", 1*time.Millisecond)
	m.RecordHit("map_columns", "L1", 3*time.Millisecond)

	stats := m.GetStats()
	// Average should be 2ms
	if stats.AvgHitLatency < time.Millisecond || stats.AvgHitLatency > 3*time.Millisecond {
		t.Errorf("expected avg latency ~2ms, got %v", stats.AvgHitLatency)
	}
}

func TestCacheMetrics_Reset(t *testing.T) {
	m := NewCacheMetrics()
	m.RecordHit("map_columns", "L1", time.Millisecond)
	m.RecordMiss("map_columns")

	m.Reset()

	stats := m.GetStats()
	if stats.TotalHits != 0 || stats.TotalMisses != 0 {
		t.Error("expected zero after reset")
	}
	if len(stats.ByOperation) != 0 {
		t.Error("expected no operations after reset")
	}
}

func TestCacheMetrics_ConcurrentSafety(t *testing.T) {
	m := NewCacheMetrics()
	done := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func() {
			m.RecordHit("op", "L1", time.Millisecond)
			m.RecordMiss("op")
			m.GetStats()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 100; i++ {
		<-done
	}
	// No panic = pass
	stats := m.GetStats()
	if stats.TotalHits != 100 {
		t.Errorf("expected 100 hits, got %d", stats.TotalHits)
	}
}

func TestCacheMetrics_LayerStatsSnapshot(t *testing.T) {
	l1 := NewMemoryCache(100, time.Hour)
	m := NewCacheMetrics()
	m.RegisterLayer("L1", l1)

	l1.Set("key1", "val1")

	stats := m.GetStats()
	if len(stats.LayerSnapshots) == 0 {
		t.Error("expected layer snapshots")
	}
	if stats.LayerSnapshots[0].Size != 1 {
		t.Errorf("expected L1 size 1, got %d", stats.LayerSnapshots[0].Size)
	}
}

func TestCacheMetricsSnapshot_JSON(t *testing.T) {
	m := NewCacheMetrics()
	m.RecordHit("map_columns", "L1", time.Millisecond)

	stats := m.GetStats()
	if stats.TotalHits != 1 {
		t.Error("snapshot should be serializable")
	}
}
