package ai

import "testing"

func TestBYOKServiceCache_RespectsMaxEntriesWithoutExpiredItems(t *testing.T) {
	cache := NewBYOKServiceCache(
		BYOKServiceCacheConfig{
			MaxEntries: 2,
		},
		func(apiKey string) (Service, error) {
			return nil, nil
		},
	)
	defer cache.Close()

	if _, err := cache.GetOrCreate("key-1"); err != nil {
		t.Fatalf("GetOrCreate key-1 failed: %v", err)
	}
	if _, err := cache.GetOrCreate("key-2"); err != nil {
		t.Fatalf("GetOrCreate key-2 failed: %v", err)
	}
	if _, err := cache.GetOrCreate("key-3"); err != nil {
		t.Fatalf("GetOrCreate key-3 failed: %v", err)
	}

	if got := cache.Size(); got != 2 {
		t.Fatalf("expected cache size 2, got %d", got)
	}
}
