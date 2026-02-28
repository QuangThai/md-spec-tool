package ai

import (
	"testing"
	"time"
)

// --- CacheLayer interface tests ---

func TestMemoryCache_ImplementsCacheLayer(t *testing.T) {
	var _ CacheLayer = NewMemoryCache(100, time.Hour)
}

func TestMemoryCache_GetSet(t *testing.T) {
	mc := NewMemoryCache(100, time.Hour)
	mc.Set("key1", "value1")

	val, ok := mc.Get("key1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestMemoryCache_TTLExpiration(t *testing.T) {
	mc := NewMemoryCache(100, 50*time.Millisecond)
	mc.Set("key1", "value1")

	time.Sleep(60 * time.Millisecond)

	_, ok := mc.Get("key1")
	if ok {
		t.Error("expected cache miss after TTL expiration")
	}
}

func TestMemoryCache_Eviction(t *testing.T) {
	mc := NewMemoryCache(2, time.Hour)
	mc.Set("key1", "val1")
	mc.Set("key2", "val2")
	mc.Set("key3", "val3") // Should evict key1

	_, ok := mc.Get("key1")
	if ok {
		t.Error("expected key1 to be evicted")
	}

	val, ok := mc.Get("key3")
	if !ok || val != "val3" {
		t.Error("expected key3 to be present")
	}
}

func TestMemoryCache_Stats(t *testing.T) {
	mc := NewMemoryCache(100, time.Hour)
	mc.Set("key1", "val1")
	mc.Get("key1") // hit
	mc.Get("key2") // miss

	stats := mc.Stats()
	if stats.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
	if stats.Size != 1 {
		t.Errorf("expected size 1, got %d", stats.Size)
	}
}

// --- MultiLevelCache tests ---

func TestMultiLevelCache_L1Hit(t *testing.T) {
	l1 := NewMemoryCache(100, time.Hour)
	mlc := NewMultiLevelCache(l1)

	mlc.Set("key1", "value1")
	val, ok := mlc.Get("key1")
	if !ok {
		t.Fatal("expected cache hit from L1")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestMultiLevelCache_L2Fallback(t *testing.T) {
	l1 := NewMemoryCache(100, time.Hour)
	l2 := NewMemoryCache(100, time.Hour)
	mlc := NewMultiLevelCache(l1, l2)

	// Put in L2 only
	l2.Set("key1", "value1")

	// Get should find in L2 and backfill L1
	val, ok := mlc.Get("key1")
	if !ok {
		t.Fatal("expected cache hit from L2")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}

	// Should now be in L1 (backfill)
	val2, ok2 := l1.Get("key1")
	if !ok2 {
		t.Error("expected L1 backfill after L2 hit")
	}
	if val2 != "value1" {
		t.Errorf("expected value1 in L1, got %v", val2)
	}
}

func TestMultiLevelCache_SetPopulatesAllLayers(t *testing.T) {
	l1 := NewMemoryCache(100, time.Hour)
	l2 := NewMemoryCache(100, time.Hour)
	mlc := NewMultiLevelCache(l1, l2)

	mlc.Set("key1", "value1")

	// Both layers should have the value
	if _, ok := l1.Get("key1"); !ok {
		t.Error("expected key in L1")
	}
	if _, ok := l2.Get("key1"); !ok {
		t.Error("expected key in L2")
	}
}

func TestMultiLevelCache_ClearAll(t *testing.T) {
	l1 := NewMemoryCache(100, time.Hour)
	l2 := NewMemoryCache(100, time.Hour)
	mlc := NewMultiLevelCache(l1, l2)

	mlc.Set("key1", "value1")
	mlc.Clear()

	if _, ok := l1.Get("key1"); ok {
		t.Error("expected L1 cleared")
	}
	if _, ok := l2.Get("key1"); ok {
		t.Error("expected L2 cleared")
	}
}

func TestMultiLevelCache_StatsAggregated(t *testing.T) {
	l1 := NewMemoryCache(100, time.Hour)
	l2 := NewMemoryCache(100, time.Hour)
	mlc := NewMultiLevelCache(l1, l2)

	mlc.Set("key1", "value1")
	mlc.Get("key1") // L1 hit

	stats := mlc.Stats()
	if stats.Hits < 1 {
		t.Errorf("expected at least 1 hit, got %d", stats.Hits)
	}
}

func TestMultiLevelCache_ImplementsCacheLayer(t *testing.T) {
	l1 := NewMemoryCache(100, time.Hour)
	var _ CacheLayer = NewMultiLevelCache(l1)
}

func TestMultiLevelCache_EmptyLayers(t *testing.T) {
	mlc := NewMultiLevelCache()
	_, ok := mlc.Get("anything")
	if ok {
		t.Error("expected miss with no layers")
	}
	// Set and Clear should not panic
	mlc.Set("key", "val")
	mlc.Clear()
}
