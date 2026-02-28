package ai

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// NormalizeHeaders returns a new sorted, trimmed, and lowercased copy of headers.
// The original slice is never mutated.
func NormalizeHeaders(headers []string) []string {
	if len(headers) == 0 {
		return []string{}
	}
	out := make([]string, len(headers))
	for i, h := range headers {
		out[i] = strings.ToLower(strings.TrimSpace(h))
	}
	sort.Strings(out)
	return out
}

// NormalizePayloadHash produces a stable SHA-256 hash for a MapColumnsRequest
// by normalizing the Headers field (sorted, trimmed, lowercased) before hashing.
func NormalizePayloadHash(req MapColumnsRequest) (string, error) {
	normalized := req
	normalized.Headers = NormalizeHeaders(req.Headers)

	b, err := json.Marshal(normalized)
	if err != nil {
		return "", fmt.Errorf("NormalizePayloadHash: marshal failed: %w", err)
	}
	sum := sha256.Sum256(b)
	return fmt.Sprintf("%x", sum), nil
}

// NormalizedCache is an L3 cache wrapper that normalises keys before delegating
// to an inner CacheLayer.  It satisfies the CacheLayer interface.
type NormalizedCache struct {
	mu     sync.RWMutex
	inner  CacheLayer
	hits   int64
	misses int64
}

// NewNormalizedCache wraps inner with L3 key-normalization behaviour.
func NewNormalizedCache(inner CacheLayer) *NormalizedCache {
	return &NormalizedCache{inner: inner}
}

// Get retrieves a value from the inner cache, updating hit/miss counters.
func (n *NormalizedCache) Get(key string) (interface{}, bool) {
	val, ok := n.inner.Get(key)
	n.mu.Lock()
	if ok {
		n.hits++
	} else {
		n.misses++
	}
	n.mu.Unlock()
	return val, ok
}

// Set stores a value in the inner cache.
func (n *NormalizedCache) Set(key string, value interface{}) {
	n.inner.Set(key, value)
}

// Clear removes all entries from the inner cache and resets counters.
func (n *NormalizedCache) Clear() {
	n.inner.Clear()
	n.mu.Lock()
	n.hits = 0
	n.misses = 0
	n.mu.Unlock()
}

// Stats returns aggregated statistics for the L3 layer.
func (n *NormalizedCache) Stats() CacheStats {
	n.mu.RLock()
	hits := n.hits
	misses := n.misses
	n.mu.RUnlock()

	inner := n.inner.Stats()
	return CacheStats{
		Hits:    hits,
		Misses:  misses,
		Size:    inner.Size,
		MaxSize: inner.MaxSize,
		Level:   "L3",
	}
}
