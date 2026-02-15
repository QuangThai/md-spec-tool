package ai

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheKey structure: op:model:vP:vS:hash
type CacheKey struct {
	Operation     string `json:"op"` // "map_columns" | "analyze_paste"
	Model         string `json:"model"`
	PromptVersion string `json:"pv"`   // From prompts.go
	SchemaVersion string `json:"sv"`   // From schemas.go
	PayloadHash   string `json:"hash"` // SHA256 of canonical payload
}

func (k CacheKey) String() string {
	return fmt.Sprintf("%s:%s:%s:%s:%s",
		k.Operation, k.Model, k.PromptVersion, k.SchemaVersion, k.PayloadHash)
}

// CacheEntry holds cached AI response with expiration
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// Cache implements LRU cache for AI responses
type Cache struct {
	mu       sync.RWMutex
	entries  map[string]*CacheEntry
	maxSize  int
	ttl      time.Duration
	hitCount map[string]int // Track usage for eviction
	order    []string       // Track insertion order for LRU
}

// NewCache creates a new cache with size and TTL constraints
func NewCache(maxSize int, ttl time.Duration) *Cache {
	return &Cache{
		entries:  make(map[string]*CacheEntry),
		maxSize:  maxSize,
		ttl:      ttl,
		hitCount: make(map[string]int),
		order:    make([]string, 0, maxSize),
	}
}

// Get retrieves a cached value if it exists and hasn't expired
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	c.hitCount[key]++
	return entry.Value, true
}

// Set stores a value in the cache
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict least used entry if at capacity
	if len(c.entries) >= c.maxSize && c.entries[key] == nil {
		c.evictLRU()
	}

	c.entries[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}

	c.hitCount[key] = 0
	c.order = append(c.order, key)
}

// evictLRU removes the least recently used entry
func (c *Cache) evictLRU() {
	if len(c.entries) == 0 {
		return
	}

	// Find key with lowest hit count
	var minKey string
	var minCount int = 1<<31 - 1

	for _, key := range c.order {
		if count, ok := c.hitCount[key]; ok && count < minCount {
			minKey = key
			minCount = count
		}
	}

	if minKey != "" {
		delete(c.entries, minKey)
		delete(c.hitCount, minKey)
	}

	// Clean up order slice
	newOrder := make([]string, 0, len(c.order))
	for _, k := range c.order {
		if _, ok := c.entries[k]; ok {
			newOrder = append(newOrder, k)
		}
	}
	c.order = newOrder
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	c.hitCount = make(map[string]int)
	c.order = make([]string, 0, c.maxSize)
}

// MakePayloadHash creates SHA256 hash of canonical JSON payload
func MakePayloadHash(payload interface{}) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:]), nil // Full 32 bytes for collision-free hashing
}

// MakeCacheKey constructs a cache key from operation parameters
func MakeCacheKey(operation, model, promptVersion, schemaVersion string, payload interface{}) (string, error) {
	hash, err := MakePayloadHash(payload)
	if err != nil {
		return "", err
	}

	key := CacheKey{
		Operation:     operation,
		Model:         model,
		PromptVersion: promptVersion,
		SchemaVersion: schemaVersion,
		PayloadHash:   hash,
	}

	return key.String(), nil
}
