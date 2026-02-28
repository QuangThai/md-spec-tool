package ai

import (
	"sync"
	"time"
)

// CacheStats holds cache performance metrics
type CacheStats struct {
	Hits    int64  `json:"hits"`
	Misses  int64  `json:"misses"`
	Size    int    `json:"size"`
	MaxSize int    `json:"max_size"`
	Level   string `json:"level"` // "L1", "L2", "L3", "multi"
}

// HitRate returns the cache hit rate as a percentage
func (s CacheStats) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total) * 100
}

// CacheLayer is the interface for all cache levels
type CacheLayer interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Clear()
	Stats() CacheStats
}

// MemoryCache is an L1 in-memory LFU cache implementing CacheLayer.
// It mirrors the logic of the existing Cache struct but also tracks Stats().
type MemoryCache struct {
	mu       sync.RWMutex
	entries  map[string]*CacheEntry
	maxSize  int
	ttl      time.Duration
	hitCount map[string]int
	order    []string
	hits     int64
	misses   int64
}

// NewMemoryCache creates a new in-memory LFU cache
func NewMemoryCache(maxSize int, ttl time.Duration) *MemoryCache {
	return &MemoryCache{
		entries:  make(map[string]*CacheEntry),
		maxSize:  maxSize,
		ttl:      ttl,
		hitCount: make(map[string]int),
		order:    make([]string, 0, maxSize),
	}
}

// Get retrieves a cached value if it exists and hasn't expired.
func (m *MemoryCache) Get(key string) (interface{}, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.entries[key]
	if !ok {
		m.misses++
		return nil, false
	}
	if time.Now().After(entry.ExpiresAt) {
		m.misses++
		return nil, false
	}
	m.hitCount[key]++
	m.hits++
	return entry.Value, true
}

// Set stores a value in the cache, evicting the LFU entry if at capacity.
func (m *MemoryCache) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.entries) >= m.maxSize && m.entries[key] == nil {
		m.evictLFU()
	}

	m.entries[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(m.ttl),
	}
	m.hitCount[key] = 0
	m.order = append(m.order, key)
}

// evictLFU removes the entry with the lowest hit count (least frequently used).
func (m *MemoryCache) evictLFU() {
	if len(m.entries) == 0 {
		return
	}

	var minKey string
	minCount := int(^uint(0) >> 1) // max int

	for _, key := range m.order {
		if count, ok := m.hitCount[key]; ok && count < minCount {
			minKey = key
			minCount = count
		}
	}

	if minKey != "" {
		delete(m.entries, minKey)
		delete(m.hitCount, minKey)
	}

	newOrder := make([]string, 0, len(m.order))
	for _, k := range m.order {
		if _, ok := m.entries[k]; ok {
			newOrder = append(newOrder, k)
		}
	}
	m.order = newOrder
}

// Clear removes all entries from the cache.
func (m *MemoryCache) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = make(map[string]*CacheEntry)
	m.hitCount = make(map[string]int)
	m.order = make([]string, 0, m.maxSize)
}

// Stats returns current performance metrics for this cache layer.
func (m *MemoryCache) Stats() CacheStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return CacheStats{
		Hits:    m.hits,
		Misses:  m.misses,
		Size:    len(m.entries),
		MaxSize: m.maxSize,
		Level:   "L1",
	}
}

// MultiLevelCache chains multiple CacheLayer instances.
//
// Get: checks layers in order (L1 first), backfills upper layers on a lower-level hit.
// Set: populates all layers.
// Clear: clears all layers.
type MultiLevelCache struct {
	layers []CacheLayer
	mu     sync.RWMutex
	hits   int64
	misses int64
}

// NewMultiLevelCache creates a cache that checks layers in order (first layer = L1).
func NewMultiLevelCache(layers ...CacheLayer) *MultiLevelCache {
	return &MultiLevelCache{
		layers: layers,
	}
}

// Get checks each layer in order and backfills upper layers on a lower-level hit.
func (c *MultiLevelCache) Get(key string) (interface{}, bool) {
	for i, layer := range c.layers {
		if val, ok := layer.Get(key); ok {
			// Backfill upper (closer-to-L1) layers
			for j := 0; j < i; j++ {
				c.layers[j].Set(key, val)
			}
			c.mu.Lock()
			c.hits++
			c.mu.Unlock()
			return val, true
		}
	}
	c.mu.Lock()
	c.misses++
	c.mu.Unlock()
	return nil, false
}

// Set writes the value to every layer.
func (c *MultiLevelCache) Set(key string, value interface{}) {
	for _, layer := range c.layers {
		layer.Set(key, value)
	}
}

// Clear clears every layer.
func (c *MultiLevelCache) Clear() {
	for _, layer := range c.layers {
		layer.Clear()
	}
}

// Stats returns aggregated metrics across all layers.
func (c *MultiLevelCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalSize := 0
	for _, layer := range c.layers {
		totalSize += layer.Stats().Size
	}

	return CacheStats{
		Hits:   c.hits,
		Misses: c.misses,
		Size:   totalSize,
		Level:  "multi",
	}
}
