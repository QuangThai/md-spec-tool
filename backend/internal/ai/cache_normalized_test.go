package ai

import (
	"testing"
	"time"
)

// --- Interface compliance ---

func TestNormalizedCache_ImplementsCacheLayer(t *testing.T) {
	inner := NewMemoryCache(100, time.Hour)
	var _ CacheLayer = NewNormalizedCache(inner)
}

// --- NormalizeHeaders ---

func TestNormalizeHeaders_SortsAlphabetically(t *testing.T) {
	headers := []string{"Status", "ID", "Title"}
	normalized := NormalizeHeaders(headers)
	expected := []string{"id", "status", "title"}
	if len(normalized) != len(expected) {
		t.Fatalf("expected %d headers, got %d", len(expected), len(normalized))
	}
	for i, v := range normalized {
		if v != expected[i] {
			t.Errorf("index %d: expected %q, got %q", i, expected[i], v)
		}
	}
}

func TestNormalizeHeaders_TrimsAndLowercases(t *testing.T) {
	headers := []string{"  ID  ", "\tTitle\n", "  STATUS  "}
	normalized := NormalizeHeaders(headers)
	// sorted + trimmed + lowercased
	expected := []string{"id", "status", "title"}
	if len(normalized) != len(expected) {
		t.Fatalf("expected %d headers, got %d", len(expected), len(normalized))
	}
	for i, v := range normalized {
		if v != expected[i] {
			t.Errorf("index %d: expected %q, got %q", i, expected[i], v)
		}
	}
}

func TestNormalizeHeaders_EmptyInput(t *testing.T) {
	if result := NormalizeHeaders(nil); len(result) != 0 {
		t.Error("expected empty result for nil input")
	}
	if result := NormalizeHeaders([]string{}); len(result) != 0 {
		t.Error("expected empty result for empty slice")
	}
}

func TestNormalizeHeaders_DoesNotMutateInput(t *testing.T) {
	original := []string{"Status", "ID", "Title"}
	input := make([]string, len(original))
	copy(input, original)
	NormalizeHeaders(input)
	for i, v := range input {
		if v != original[i] {
			t.Errorf("input mutated at index %d: expected %q, got %q", i, original[i], v)
		}
	}
}

// --- NormalizePayloadHash ---

func TestNormalizePayloadHash_SameHeadersDifferentOrder(t *testing.T) {
	req1 := MapColumnsRequest{Headers: []string{"ID", "Title", "Status"}}
	req2 := MapColumnsRequest{Headers: []string{"Status", "ID", "Title"}}

	hash1, err := NormalizePayloadHash(req1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hash2, err := NormalizePayloadHash(req2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("same headers in different order should produce same hash\nhash1=%s\nhash2=%s", hash1, hash2)
	}
}

func TestNormalizePayloadHash_DifferentHeaders(t *testing.T) {
	req1 := MapColumnsRequest{Headers: []string{"ID", "Title"}}
	req2 := MapColumnsRequest{Headers: []string{"ID", "Status"}}

	hash1, _ := NormalizePayloadHash(req1)
	hash2, _ := NormalizePayloadHash(req2)

	if hash1 == hash2 {
		t.Error("different headers should produce different hashes")
	}
}

func TestNormalizePayloadHash_CaseInsensitive(t *testing.T) {
	req1 := MapColumnsRequest{Headers: []string{"ID", "Title"}}
	req2 := MapColumnsRequest{Headers: []string{"id", "title"}}

	hash1, _ := NormalizePayloadHash(req1)
	hash2, _ := NormalizePayloadHash(req2)

	if hash1 != hash2 {
		t.Errorf("headers differing only by case should produce same hash\nhash1=%s\nhash2=%s", hash1, hash2)
	}
}

func TestNormalizePayloadHash_WhitespaceTrimmed(t *testing.T) {
	req1 := MapColumnsRequest{Headers: []string{"ID", "Title"}}
	req2 := MapColumnsRequest{Headers: []string{"  ID  ", "\tTitle\n"}}

	hash1, _ := NormalizePayloadHash(req1)
	hash2, _ := NormalizePayloadHash(req2)

	if hash1 != hash2 {
		t.Errorf("headers with surrounding whitespace should produce same hash\nhash1=%s\nhash2=%s", hash1, hash2)
	}
}

func TestNormalizePayloadHash_EmptyRequest(t *testing.T) {
	req := MapColumnsRequest{}
	hash, err := NormalizePayloadHash(req)
	if err != nil {
		t.Fatalf("unexpected error for empty request: %v", err)
	}
	if hash == "" {
		t.Error("expected non-empty hash for empty request")
	}
}

// --- NormalizedCache behavior ---

func TestNormalizedCache_HitOnReorderedHeaders(t *testing.T) {
	inner := NewMemoryCache(100, time.Hour)
	nc := NewNormalizedCache(inner)

	req1 := MapColumnsRequest{Headers: []string{"ID", "Title", "Status"}}
	req2 := MapColumnsRequest{Headers: []string{"Status", "ID", "Title"}}

	hash1, err := NormalizePayloadHash(req1)
	if err != nil {
		t.Fatalf("hash1 error: %v", err)
	}
	hash2, err := NormalizePayloadHash(req2)
	if err != nil {
		t.Fatalf("hash2 error: %v", err)
	}

	// Normalized hashes must be equal — the core invariant
	if hash1 != hash2 {
		t.Fatalf("NormalizePayloadHash must return same hash for same headers in different order")
	}

	key1 := "map_columns:gpt-4o-mini:v3:v2:" + hash1
	key2 := "map_columns:gpt-4o-mini:v3:v2:" + hash2

	nc.Set(key1, "cached_result")
	val, ok := nc.Get(key2)
	if !ok {
		t.Fatal("expected cache hit: key1 and key2 should be identical after normalization")
	}
	if val != "cached_result" {
		t.Errorf("expected 'cached_result', got %v", val)
	}
}

func TestNormalizedCache_GetMiss(t *testing.T) {
	inner := NewMemoryCache(100, time.Hour)
	nc := NewNormalizedCache(inner)
	_, ok := nc.Get("nonexistent")
	if ok {
		t.Error("expected miss for nonexistent key")
	}
}

func TestNormalizedCache_SetAndGet(t *testing.T) {
	inner := NewMemoryCache(100, time.Hour)
	nc := NewNormalizedCache(inner)
	nc.Set("key1", "value1")
	val, ok := nc.Get("key1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got %v", val)
	}
}

func TestNormalizedCache_Stats(t *testing.T) {
	inner := NewMemoryCache(100, time.Hour)
	nc := NewNormalizedCache(inner)

	nc.Set("key1", "val1")
	nc.Get("key1") // hit
	nc.Get("key2") // miss

	stats := nc.Stats()
	if stats.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
	if stats.Level != "L3" {
		t.Errorf("expected level 'L3', got %q", stats.Level)
	}
}

func TestNormalizedCache_Clear(t *testing.T) {
	inner := NewMemoryCache(100, time.Hour)
	nc := NewNormalizedCache(inner)
	nc.Set("key1", "val1")
	nc.Clear()
	_, ok := nc.Get("key1")
	if ok {
		t.Error("expected miss after Clear()")
	}
}

func TestNormalizedCache_StatsSizeFromInner(t *testing.T) {
	inner := NewMemoryCache(100, time.Hour)
	nc := NewNormalizedCache(inner)
	nc.Set("a", 1)
	nc.Set("b", 2)
	stats := nc.Stats()
	if stats.Size < 2 {
		t.Errorf("expected size >= 2, got %d", stats.Size)
	}
}
