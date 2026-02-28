package ai

import (
	"testing"
	"time"
)

func TestCacheKey_UsesPromptHash(t *testing.T) {
	reg := DefaultPromptRegistry()
	entry, _ := reg.Get(PromptIDColumnMapping)

	// Cache version should include hash prefix
	cacheVersion := entry.CacheVersion()
	if len(cacheVersion) < 10 {
		t.Errorf("expected cache version with hash, got %q", cacheVersion)
	}

	// Make a cache key with this version
	key, err := MakeCacheKey(CacheKeyScopeMapColumns, "gpt-4o-mini", cacheVersion, SchemaVersionColumnMapping, "test_payload")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == "" {
		t.Error("expected non-empty cache key")
	}
}

func TestCacheKey_InvalidatesOnPromptChange(t *testing.T) {
	reg := NewPromptRegistry()
	reg.Register(PromptEntry{
		ID:      "test_prompt",
		Version: "v1",
		Content: "Original prompt content",
	})

	entry1, _ := reg.Get("test_prompt")
	key1, _ := MakeCacheKey("test_op", "gpt-4o-mini", entry1.CacheVersion(), "v1", "payload")

	// Simulate prompt content change
	reg.Register(PromptEntry{
		ID:      "test_prompt",
		Version: "v1",
		Content: "Modified prompt content",
	})

	entry2, _ := reg.Get("test_prompt")
	key2, _ := MakeCacheKey("test_op", "gpt-4o-mini", entry2.CacheVersion(), "v1", "payload")

	if key1 == key2 {
		t.Error("cache key should change when prompt content changes")
	}
}

func TestCacheKey_StableForSameContent(t *testing.T) {
	reg := DefaultPromptRegistry()
	entry, _ := reg.Get(PromptIDColumnMapping)
	v1 := entry.CacheVersion()
	v2 := entry.CacheVersion()
	if v1 != v2 {
		t.Error("same content should produce stable cache version")
	}
}

func TestServiceImpl_HasPromptRegistry(t *testing.T) {
	// Verify ServiceImpl can accept a PromptRegistry
	reg := DefaultPromptRegistry()
	cache := NewMultiLevelCache(NewMemoryCache(100, time.Hour))

	svc := &ServiceImpl{
		model:          "gpt-4o-mini",
		promptProfile:  PromptProfileStaticV3,
		cache:          cache,
		promptRegistry: reg,
	}

	if svc.promptRegistry == nil {
		t.Error("expected non-nil prompt registry")
	}
}

func TestServiceImpl_GetPromptInfo(t *testing.T) {
	reg := DefaultPromptRegistry()
	cache := NewMultiLevelCache(NewMemoryCache(100, time.Hour))

	svc := &ServiceImpl{
		model:          "gpt-4o-mini",
		promptProfile:  PromptProfileStaticV3,
		cache:          cache,
		promptRegistry: reg,
	}

	info := svc.GetPromptInfo()
	if len(info) == 0 {
		t.Error("expected non-empty prompt info")
	}

	// Should contain info for each registered prompt
	foundMapping := false
	for _, p := range info {
		if p.ID == PromptIDColumnMapping {
			foundMapping = true
			if p.Hash == "" {
				t.Error("expected hash in prompt info")
			}
			if p.CacheVersion == "" {
				t.Error("expected cache version in prompt info")
			}
		}
	}
	if !foundMapping {
		t.Error("expected column_mapping in prompt info")
	}
}

func TestMakeCacheKeyWithRegistry_DifferentFromLegacy(t *testing.T) {
	// Legacy key with simple "v3"
	legacyKey, _ := MakeCacheKey(CacheKeyScopeMapColumns, "gpt-4o-mini", "v3", SchemaVersionColumnMapping, "payload")

	// New key with registry hash-based version
	reg := DefaultPromptRegistry()
	entry, _ := reg.Get(PromptIDColumnMapping)
	newKey, _ := MakeCacheKey(CacheKeyScopeMapColumns, "gpt-4o-mini", entry.CacheVersion(), SchemaVersionColumnMapping, "payload")

	if legacyKey == newKey {
		t.Error("registry-based cache key should differ from legacy string-only key")
	}
}
