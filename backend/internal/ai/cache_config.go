package ai

import (
	"fmt"
	"log/slog"
	"time"
)

// CacheConfig holds configuration for the multi-level cache stack
type CacheConfig struct {
	// L1 - In-Memory LRU
	L1MaxSize int           // Maximum entries in memory cache. Default: 1000
	L1TTL     time.Duration // TTL for memory cache entries. Default: 1h

	// L2 - Persistent SQLite
	EnableL2  bool          // Enable persistent cache (disabled by default)
	L2DBPath  string        // Path to SQLite database. Default: ".cache/ai_cache.db"
	L2MaxSize int           // Maximum entries in persistent cache. Default: 10000
	L2TTL     time.Duration // TTL for persistent entries. Default: 24h

	// Metrics
	EnableMetrics bool // Enable cache metrics tracking. Default: true
}

// DefaultCacheConfig returns sensible defaults for the cache stack
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		L1MaxSize:     1000,
		L1TTL:         1 * time.Hour,
		EnableL2:      false, // Opt-in: requires disk I/O
		L2DBPath:      ".cache/ai_cache.db",
		L2MaxSize:     10000,
		L2TTL:         24 * time.Hour,
		EnableMetrics: true,
	}
}

// BuildCacheStack constructs the multi-level cache based on configuration.
// Returns the top-level CacheLayer, a cleanup function (closes persistent DB),
// and any error encountered during setup.
//
// Stack structure (when all enabled):
//
//	Request → L1 Memory (< 1ms, 1000 entries, 1h TTL)
//	              ↓ miss
//	          L2 Persistent (< 5ms, 10000 entries, 24h TTL)
//	              ↓ miss
//	          OpenAI API call → populate L1 + L2
func BuildCacheStack(cfg CacheConfig) (CacheLayer, func(), error) {
	var layers []CacheLayer
	var cleanups []func()

	// L1: always enabled
	l1 := NewMemoryCache(cfg.L1MaxSize, cfg.L1TTL)
	layers = append(layers, l1)

	// L2: persistent SQLite (opt-in)
	if cfg.EnableL2 {
		l2, err := NewPersistentCache(PersistentCacheConfig{
			DBPath:  cfg.L2DBPath,
			MaxSize: cfg.L2MaxSize,
			TTL:     cfg.L2TTL,
		})
		if err != nil {
			// L2 failure is non-fatal: log and continue with L1 only
			slog.Warn("persistent cache (L2) disabled: failed to initialize",
				"error", err,
				"db_path", cfg.L2DBPath,
			)
		} else {
			layers = append(layers, l2)
			cleanups = append(cleanups, func() {
				if cerr := l2.Close(); cerr != nil {
					slog.Warn("persistent cache close error", "error", cerr)
				}
			})
		}
	}

	multi := NewMultiLevelCache(layers...)

	cleanup := func() {
		for _, fn := range cleanups {
			fn()
		}
	}

	if len(layers) > 1 {
		slog.Info("cache stack initialized",
			"levels", len(layers),
			"l1_max", cfg.L1MaxSize,
			"l2_enabled", cfg.EnableL2,
		)
	}

	return multi, cleanup, nil
}

// CacheConfigFromServiceConfig derives a CacheConfig from the existing service Config,
// preserving backward compatibility.
func CacheConfigFromServiceConfig(svcCfg Config) CacheConfig {
	cfg := DefaultCacheConfig()
	if svcCfg.MaxCacheSize > 0 {
		cfg.L1MaxSize = svcCfg.MaxCacheSize
	}
	if svcCfg.CacheTTL > 0 {
		cfg.L1TTL = svcCfg.CacheTTL
	}
	// L2 is enabled via AI_CACHE_ENABLE_L2 env var or explicit config
	// For now, derived from service config — caller can override
	return cfg
}

// WrapWithNormalized wraps a cache layer with L3 normalized key behaviour.
// Use this when building cache keys for MapColumns requests where header
// order should not affect cache hits.
func WrapWithNormalized(inner CacheLayer) *NormalizedCache {
	return NewNormalizedCache(inner)
}

// AttachMetrics creates a CacheMetrics instance and registers all layers.
// Returns the metrics tracker for the service to use.
func AttachMetrics(multi *MultiLevelCache) *CacheMetrics {
	metrics := NewCacheMetrics()
	for i, layer := range multi.layers {
		name := fmt.Sprintf("L%d", i+1)
		metrics.RegisterLayer(name, layer)
	}
	return metrics
}
