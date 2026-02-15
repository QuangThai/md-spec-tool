package ai

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// BYOKServiceCacheConfig holds configuration for BYOK service caching
type BYOKServiceCacheConfig struct {
	TTL           time.Duration // How long to cache a service
	CleanupTicker time.Duration // How often to clean expired entries
	MaxEntries    int           // Maximum number of cached entries (0 = unlimited)
}

// byokCacheEntry holds a cached AI service with expiration
type byokCacheEntry struct {
	service Service
	expires time.Time
}

// BYOKServiceCache provides thread-safe caching of AI services keyed by API key hash
type BYOKServiceCache struct {
	mu           sync.RWMutex
	cache        map[string]*byokCacheEntry
	config       BYOKServiceCacheConfig
	cleanupDone  chan struct{}
	isRunning    bool
	newServiceFn func(apiKey string) (Service, error) // Function to create new services
}

// NewBYOKServiceCache creates a new BYOK service cache
func NewBYOKServiceCache(cfg BYOKServiceCacheConfig, newServiceFn func(apiKey string) (Service, error)) *BYOKServiceCache {
	if cfg.TTL == 0 {
		cfg.TTL = 5 * time.Minute
	}
	if cfg.CleanupTicker == 0 {
		cfg.CleanupTicker = 1 * time.Minute
	}
	if cfg.MaxEntries == 0 {
		cfg.MaxEntries = 1000 // Safe default
	}

	bc := &BYOKServiceCache{
		cache:        make(map[string]*byokCacheEntry),
		config:       cfg,
		cleanupDone:  make(chan struct{}),
		newServiceFn: newServiceFn,
	}

	bc.startCleanup()
	return bc
}

// GetOrCreate retrieves a cached service or creates a new one
func (bc *BYOKServiceCache) GetOrCreate(apiKey string) (Service, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("empty API key")
	}

	cacheKey := bc.hashKey(apiKey)

	// Try to get from cache
	bc.mu.RLock()
	if entry, exists := bc.cache[cacheKey]; exists && time.Now().Before(entry.expires) {
		bc.mu.RUnlock()
		return entry.service, nil
	}
	bc.mu.RUnlock()

	// Create new service
	service, err := bc.newServiceFn(apiKey)
	if err != nil {
		slog.Warn("BYOK: failed to create AI service", "error", err)
		return nil, err
	}

	// Store in cache
	bc.mu.Lock()
	if bc.config.MaxEntries > 0 && len(bc.cache) >= bc.config.MaxEntries {
		// Simple eviction: remove oldest entry
		bc.evictOne()
	}
	bc.cache[cacheKey] = &byokCacheEntry{
		service: service,
		expires: time.Now().Add(bc.config.TTL),
	}
	bc.mu.Unlock()

	return service, nil
}

// hashKey creates a hash of the API key for safe cache storage
func (bc *BYOKServiceCache) hashKey(apiKey string) string {
	h := sha256.Sum256([]byte(apiKey))
	return fmt.Sprintf("%x", h[:16]) // 128-bit hash
}

// evictOne removes one entry (oldest by expiry time).
func (bc *BYOKServiceCache) evictOne() {
	if len(bc.cache) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, entry := range bc.cache {
		if first || entry.expires.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.expires
			first = false
		}
	}

	delete(bc.cache, oldestKey)
}

// startCleanup begins the cleanup goroutine
func (bc *BYOKServiceCache) startCleanup() {
	if bc.isRunning {
		return
	}
	bc.isRunning = true

	go func() {
		ticker := time.NewTicker(bc.config.CleanupTicker)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				bc.mu.Lock()
				now := time.Now()
				for key, entry := range bc.cache {
					if now.After(entry.expires) {
						delete(bc.cache, key)
					}
				}
				bc.mu.Unlock()
			case <-bc.cleanupDone:
				return
			}
		}
	}()
}

// Close stops the cleanup goroutine and clears the cache
func (bc *BYOKServiceCache) Close() {
	if bc.isRunning {
		close(bc.cleanupDone)
		bc.isRunning = false
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.cache = make(map[string]*byokCacheEntry)
}

// Size returns the current number of cached entries
func (bc *BYOKServiceCache) Size() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return len(bc.cache)
}

// Clear removes all entries from the cache
func (bc *BYOKServiceCache) Clear() {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.cache = make(map[string]*byokCacheEntry)
}
