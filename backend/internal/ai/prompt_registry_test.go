package ai

import (
	"testing"
)

func TestPromptRegistry_RegisterAndGet(t *testing.T) {
	reg := NewPromptRegistry()
	reg.Register(PromptEntry{
		ID:      "column_mapping",
		Version: "v3",
		Content: "You are an expert at analyzing spreadsheet headers...",
	})

	entry, ok := reg.Get("column_mapping")
	if !ok {
		t.Fatal("expected to find registered prompt")
	}
	if entry.Version != "v3" {
		t.Errorf("expected version v3, got %s", entry.Version)
	}
	if entry.Hash == "" {
		t.Error("expected non-empty content hash")
	}
}

func TestPromptRegistry_HashChangesWithContent(t *testing.T) {
	reg := NewPromptRegistry()
	reg.Register(PromptEntry{
		ID:      "test_prompt",
		Version: "v1",
		Content: "Original content",
	})
	entry1, _ := reg.Get("test_prompt")
	hash1 := entry1.Hash

	reg.Register(PromptEntry{
		ID:      "test_prompt",
		Version: "v1",
		Content: "Modified content",
	})
	entry2, _ := reg.Get("test_prompt")
	hash2 := entry2.Hash

	if hash1 == hash2 {
		t.Error("hash should change when content changes")
	}
}

func TestPromptRegistry_SameContentSameHash(t *testing.T) {
	reg := NewPromptRegistry()
	content := "Same content"
	reg.Register(PromptEntry{
		ID:      "prompt_a",
		Version: "v1",
		Content: content,
	})
	reg.Register(PromptEntry{
		ID:      "prompt_b",
		Version: "v1",
		Content: content,
	})

	a, _ := reg.Get("prompt_a")
	b, _ := reg.Get("prompt_b")
	if a.Hash != b.Hash {
		t.Error("same content should produce same hash")
	}
}

func TestPromptRegistry_GetMissing(t *testing.T) {
	reg := NewPromptRegistry()
	_, ok := reg.Get("nonexistent")
	if ok {
		t.Error("expected not found for unregistered prompt")
	}
}

func TestPromptRegistry_ListAll(t *testing.T) {
	reg := NewPromptRegistry()
	reg.Register(PromptEntry{ID: "a", Version: "v1", Content: "A"})
	reg.Register(PromptEntry{ID: "b", Version: "v2", Content: "B"})

	entries := reg.List()
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestPromptRegistry_VersionOverride(t *testing.T) {
	reg := NewPromptRegistry()
	reg.Register(PromptEntry{ID: "test", Version: "v1", Content: "V1 content"})
	reg.Register(PromptEntry{ID: "test", Version: "v2", Content: "V2 content"})

	// Override to use v1
	reg.SetVersionOverride("test", "v1")
	entry, ok := reg.Get("test")
	if !ok {
		t.Fatal("expected to find prompt")
	}
	if entry.Version != "v1" {
		t.Errorf("expected version v1 after override, got %s", entry.Version)
	}
	if entry.Content != "V1 content" {
		t.Error("expected V1 content after override")
	}
}

func TestPromptRegistry_CacheKeyIncludesHash(t *testing.T) {
	reg := NewPromptRegistry()
	reg.Register(PromptEntry{ID: "test", Version: "v1", Content: "My prompt"})

	entry, _ := reg.Get("test")
	cacheVersion := entry.CacheVersion()
	if cacheVersion == "" {
		t.Error("cache version should not be empty")
	}
	// Cache version should include hash prefix for uniqueness
	if len(cacheVersion) < 10 {
		t.Error("cache version should be substantive")
	}
}

func TestPromptRegistry_ConcurrentSafety(t *testing.T) {
	reg := NewPromptRegistry()
	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func(n int) {
			reg.Register(PromptEntry{
				ID:      "concurrent",
				Version: "v1",
				Content: "Content",
			})
			reg.Get("concurrent")
			reg.List()
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 50; i++ {
		<-done
	}
}
