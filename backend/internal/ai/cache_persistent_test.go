package ai

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPersistentCache_ImplementsCacheLayer(t *testing.T) {
	dir := t.TempDir()
	pc, err := NewPersistentCache(PersistentCacheConfig{
		DBPath:  filepath.Join(dir, "test.db"),
		MaxSize: 100,
		TTL:     time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to create persistent cache: %v", err)
	}
	defer pc.Close()

	var _ CacheLayer = pc
}

func TestPersistentCache_GetSet(t *testing.T) {
	dir := t.TempDir()
	pc, err := NewPersistentCache(PersistentCacheConfig{
		DBPath:  filepath.Join(dir, "test.db"),
		MaxSize: 100,
		TTL:     time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}
	defer pc.Close()

	// Store a JSON-serializable value
	original := map[string]string{"key": "value"}
	data, _ := json.Marshal(original)
	pc.Set("test_key", json.RawMessage(data))

	val, ok := pc.Get("test_key")
	if !ok {
		t.Fatal("expected cache hit")
	}

	// Value should be json.RawMessage
	raw, ok := val.(json.RawMessage)
	if !ok {
		// Could also be string or []byte depending on implementation
		rawBytes, ok2 := val.([]byte)
		if !ok2 {
			t.Fatalf("expected json.RawMessage or []byte, got %T", val)
		}
		raw = json.RawMessage(rawBytes)
	}

	var result map[string]string
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("expected 'value', got %q", result["key"])
	}
}

func TestPersistentCache_SurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Session 1: write data
	pc1, err := NewPersistentCache(PersistentCacheConfig{
		DBPath:  dbPath,
		MaxSize: 100,
		TTL:     time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}
	data, _ := json.Marshal("persistent_value")
	pc1.Set("survive_key", json.RawMessage(data))
	pc1.Close()

	// Session 2: read data back
	pc2, err := NewPersistentCache(PersistentCacheConfig{
		DBPath:  dbPath,
		MaxSize: 100,
		TTL:     time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to reopen: %v", err)
	}
	defer pc2.Close()

	val, ok := pc2.Get("survive_key")
	if !ok {
		t.Fatal("expected cache hit after restart")
	}

	raw := extractRawJSON(t, val)
	var result string
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result != "persistent_value" {
		t.Errorf("expected 'persistent_value', got %q", result)
	}
}

func TestPersistentCache_TTLExpiration(t *testing.T) {
	dir := t.TempDir()
	pc, err := NewPersistentCache(PersistentCacheConfig{
		DBPath:  filepath.Join(dir, "test.db"),
		MaxSize: 100,
		TTL:     100 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}
	defer pc.Close()

	data, _ := json.Marshal("expiring")
	pc.Set("ttl_key", json.RawMessage(data))

	time.Sleep(150 * time.Millisecond)

	_, ok := pc.Get("ttl_key")
	if ok {
		t.Error("expected cache miss after TTL expiration")
	}
}

func TestPersistentCache_Eviction(t *testing.T) {
	dir := t.TempDir()
	pc, err := NewPersistentCache(PersistentCacheConfig{
		DBPath:  filepath.Join(dir, "test.db"),
		MaxSize: 3,
		TTL:     time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}
	defer pc.Close()

	for i := 0; i < 5; i++ {
		data, _ := json.Marshal(i)
		pc.Set(string(rune('a'+i)), json.RawMessage(data))
	}

	stats := pc.Stats()
	if stats.Size > 3 {
		t.Errorf("expected max 3 entries, got %d", stats.Size)
	}
}

func TestPersistentCache_Stats(t *testing.T) {
	dir := t.TempDir()
	pc, err := NewPersistentCache(PersistentCacheConfig{
		DBPath:  filepath.Join(dir, "test.db"),
		MaxSize: 100,
		TTL:     time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}
	defer pc.Close()

	data, _ := json.Marshal("val")
	pc.Set("key1", json.RawMessage(data))
	pc.Get("key1") // hit
	pc.Get("key2") // miss

	stats := pc.Stats()
	if stats.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
	if stats.Level != "L2" {
		t.Errorf("expected level L2, got %s", stats.Level)
	}
}

func TestPersistentCache_Clear(t *testing.T) {
	dir := t.TempDir()
	pc, err := NewPersistentCache(PersistentCacheConfig{
		DBPath:  filepath.Join(dir, "test.db"),
		MaxSize: 100,
		TTL:     time.Hour,
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}
	defer pc.Close()

	data, _ := json.Marshal("val")
	pc.Set("key1", json.RawMessage(data))
	pc.Clear()

	stats := pc.Stats()
	if stats.Size != 0 {
		t.Errorf("expected 0 after clear, got %d", stats.Size)
	}
}

func TestPersistentCache_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "subdir", "nested", "test.db")

	pc, err := NewPersistentCache(PersistentCacheConfig{
		DBPath:  dbPath,
		MaxSize: 100,
		TTL:     time.Hour,
	})
	if err != nil {
		t.Fatalf("should create nested directories: %v", err)
	}
	defer pc.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected database file to exist")
	}
}

// Helper to extract json.RawMessage from interface{}
func extractRawJSON(t *testing.T, val interface{}) json.RawMessage {
	t.Helper()
	switch v := val.(type) {
	case json.RawMessage:
		return v
	case []byte:
		return json.RawMessage(v)
	case string:
		return json.RawMessage(v)
	default:
		t.Fatalf("unexpected type %T", val)
		return nil
	}
}
