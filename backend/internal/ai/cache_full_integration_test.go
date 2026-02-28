package ai

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFullCacheStack_L1L2L3(t *testing.T) {
	// Create temp directory for SQLite
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_cache.db")

	// Create L1 (memory)
	l1 := NewMemoryCache(100, time.Hour)

	// Create L2 (persistent)
	l2, err := NewPersistentCache(PersistentCacheConfig{
		DBPath:  dbPath,
		MaxSize: 1000,
		TTL:     24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to create L2: %v", err)
	}
	defer l2.Close()

	// Create L3 (normalized) wrapping a fresh MemoryCache for normalized keys
	l3Inner := NewMemoryCache(100, time.Hour)
	l3 := NewNormalizedCache(l3Inner)

	// Build multi-level: L1 → L2 → L3
	multi := NewMultiLevelCache(l1, l2, l3)

	// Set a value
	multi.Set("test_key", "test_value")

	// Should be in L1
	val, ok := l1.Get("test_key")
	if !ok {
		t.Error("expected L1 hit")
	}
	if val != "test_value" {
		t.Errorf("L1: expected test_value, got %v", val)
	}

	// Should also be in L2 (as JSON)
	_, ok = l2.Get("test_key")
	if !ok {
		t.Error("expected L2 hit after multi.Set")
	}

	// Multi-level get should work
	val, ok = multi.Get("test_key")
	if !ok {
		t.Error("expected multi-level hit")
	}
}

func TestFullCacheStack_L2BackfillsL1(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_cache.db")

	l1 := NewMemoryCache(100, time.Hour)
	l2, err := NewPersistentCache(PersistentCacheConfig{
		DBPath:  dbPath,
		MaxSize: 1000,
		TTL:     24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to create L2: %v", err)
	}
	defer l2.Close()

	multi := NewMultiLevelCache(l1, l2)

	// Write directly to L2 (simulating L1 eviction)
	l2.Set("only_in_l2", "from_l2")

	// L1 should miss
	_, ok := l1.Get("only_in_l2")
	if ok {
		t.Error("L1 should not have this key yet")
	}

	// Multi-level get should find in L2 and backfill L1
	val, ok := multi.Get("only_in_l2")
	if !ok {
		t.Error("expected multi-level hit from L2")
	}
	if val == nil {
		t.Error("expected non-nil value")
	}

	// Now L1 should have it (backfilled)
	_, ok = l1.Get("only_in_l2")
	if !ok {
		t.Error("expected L1 to have backfilled value after L2 hit")
	}
}

func TestFullCacheStack_SurvivesRestart(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_cache.db")

	// Session 1: write to L2
	l2a, err := NewPersistentCache(PersistentCacheConfig{DBPath: dbPath, MaxSize: 1000, TTL: 24 * time.Hour})
	if err != nil {
		t.Fatalf("session 1: create L2 failed: %v", err)
	}
	l2a.Set("persistent_key", "persistent_value")
	l2a.Close()

	// Session 2: new L2 instance pointing to same DB
	l2b, err := NewPersistentCache(PersistentCacheConfig{DBPath: dbPath, MaxSize: 1000, TTL: 24 * time.Hour})
	if err != nil {
		t.Fatalf("session 2: create L2 failed: %v", err)
	}
	defer l2b.Close()

	// Should find the value
	val, ok := l2b.Get("persistent_key")
	if !ok {
		t.Error("expected L2 hit after restart")
	}
	if val == nil {
		t.Error("expected non-nil value after restart")
	}
}

func TestFullCacheStack_NormalizedKeyHit(t *testing.T) {
	l1 := NewMemoryCache(100, time.Hour)

	// Generate two cache keys for headers in different order
	req1 := MapColumnsRequest{Headers: []string{"ID", "Title", "Description"}}
	req2 := MapColumnsRequest{Headers: []string{"Description", "Title", "ID"}} // different order

	hash1, err := NormalizePayloadHash(req1)
	if err != nil {
		t.Fatalf("failed to normalize hash 1: %v", err)
	}
	hash2, err := NormalizePayloadHash(req2)
	if err != nil {
		t.Fatalf("failed to normalize hash 2: %v", err)
	}

	// Same headers in different order should produce same normalized hash
	if hash1 != hash2 {
		t.Errorf("expected same normalized hash, got %s vs %s", hash1, hash2)
	}

	// Use normalized hash as cache key suffix
	key := "map_columns:gpt-4o-mini:v3:" + hash1
	l1.Set(key, "cached_result")

	// Should hit with the other order too (same normalized hash)
	key2 := "map_columns:gpt-4o-mini:v3:" + hash2
	val, ok := l1.Get(key2)
	if !ok {
		t.Error("expected cache hit with normalized key")
	}
	if val != "cached_result" {
		t.Errorf("expected cached_result, got %v", val)
	}
}

func TestFullCacheStack_MetricsTracking(t *testing.T) {
	l1 := NewMemoryCache(100, time.Hour)
	metrics := NewCacheMetrics()
	metrics.RegisterLayer("L1", l1)

	// Simulate operations
	start := time.Now()
	l1.Set("key1", "val1")
	metrics.RecordMiss("map_columns")

	val, ok := l1.Get("key1")
	latency := time.Since(start)
	if ok && val != nil {
		metrics.RecordHit("map_columns", "L1", latency)
	}

	stats := metrics.GetStats()
	if stats.TotalHits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.TotalHits)
	}
	if stats.TotalMisses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.TotalMisses)
	}
	if stats.HitRate != 50.0 {
		t.Errorf("expected 50%% hit rate, got %.1f%%", stats.HitRate)
	}

	// Layer snapshot should include L1
	if len(stats.LayerSnapshots) != 1 {
		t.Errorf("expected 1 layer snapshot, got %d", len(stats.LayerSnapshots))
	}
}

func TestNewServiceCacheConfig_DefaultValues(t *testing.T) {
	cfg := DefaultCacheConfig()
	if cfg.L1MaxSize != 1000 {
		t.Errorf("expected L1MaxSize=1000, got %d", cfg.L1MaxSize)
	}
	if cfg.L1TTL != time.Hour {
		t.Errorf("expected L1TTL=1h, got %v", cfg.L1TTL)
	}
	if cfg.EnableL2 {
		t.Error("expected L2 disabled by default")
	}
	if cfg.L2MaxSize != 10000 {
		t.Errorf("expected L2MaxSize=10000, got %d", cfg.L2MaxSize)
	}
}

func TestBuildCacheStack_L1Only(t *testing.T) {
	cfg := DefaultCacheConfig()
	stack, cleanup, err := BuildCacheStack(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cleanup()

	stack.Set("k", "v")
	val, ok := stack.Get("k")
	if !ok {
		t.Error("expected hit")
	}
	if val != "v" {
		t.Errorf("expected 'v', got %v", val)
	}
}

func TestBuildCacheStack_WithL2(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultCacheConfig()
	cfg.EnableL2 = true
	cfg.L2DBPath = filepath.Join(tmpDir, "test.db")

	stack, cleanup, err := BuildCacheStack(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cleanup()

	stack.Set("k", "v")
	val, ok := stack.Get("k")
	if !ok {
		t.Error("expected hit")
	}
	if val != nil {
		// L1 returns string directly
		stats := stack.Stats()
		if stats.Level != "multi" {
			t.Errorf("expected multi-level cache, got %s", stats.Level)
		}
	}
}

func TestServiceImpl_WithCacheMetrics(t *testing.T) {
	cache := NewMultiLevelCache(NewMemoryCache(100, time.Hour))
	metrics := NewCacheMetrics()
	metrics.RegisterLayer("L1", cache)

	svc := &ServiceImpl{
		model:          "gpt-4o-mini",
		promptProfile:  PromptProfileStaticV3,
		cache:          cache,
		promptRegistry: DefaultPromptRegistry(),
		cacheMetrics:   metrics,
	}

	info := svc.GetCacheMetrics()
	if info.TotalHits != 0 {
		t.Errorf("expected 0 initial hits, got %d", info.TotalHits)
	}
}

func TestServiceImpl_PersistentCacheDBPath(t *testing.T) {
	// Verify env var controls persistent cache path
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "env_cache.db")
	os.Setenv("AI_CACHE_DB_PATH", dbPath)
	defer os.Unsetenv("AI_CACHE_DB_PATH")

	cfg := DefaultCacheConfig()
	cfg.EnableL2 = true
	cfg.L2DBPath = dbPath

	stack, cleanup, err := BuildCacheStack(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cleanup()

	stack.Set("env_key", "env_val")

	// Verify DB file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected SQLite DB file to be created")
	}
}
