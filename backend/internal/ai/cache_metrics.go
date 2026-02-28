package ai

import (
	"sync"
	"time"
)

// OperationStats holds per-operation cache statistics
type OperationStats struct {
	Hits         int64         `json:"hits"`
	Misses       int64         `json:"misses"`
	HitRate      float64       `json:"hit_rate"`
	TotalLatency time.Duration `json:"-"`
	AvgLatency   time.Duration `json:"avg_latency"`
}

// LevelStats holds per-cache-level statistics
type LevelStats struct {
	Hits         int64         `json:"hits"`
	AvgLatency   time.Duration `json:"avg_latency"`
	TotalLatency time.Duration `json:"-"`
}

// CacheMetricsSnapshot is a point-in-time view of all cache metrics
type CacheMetricsSnapshot struct {
	TotalHits      int64                     `json:"total_hits"`
	TotalMisses    int64                     `json:"total_misses"`
	HitRate        float64                   `json:"hit_rate"`
	AvgHitLatency  time.Duration             `json:"avg_hit_latency"`
	ByOperation    map[string]OperationStats `json:"by_operation"`
	ByLevel        map[string]LevelStats     `json:"by_level"`
	LayerSnapshots []CacheStats              `json:"layer_snapshots,omitempty"`
}

// CacheMetrics tracks cache performance across all levels and operations
type CacheMetrics struct {
	mu           sync.RWMutex
	totalHits    int64
	totalMisses  int64
	totalLatency time.Duration
	byOperation  map[string]*OperationStats
	byLevel      map[string]*LevelStats
	layers       map[string]CacheLayer // registered layers for snapshot
}

// NewCacheMetrics creates a new CacheMetrics instance
func NewCacheMetrics() *CacheMetrics {
	return &CacheMetrics{
		byOperation: make(map[string]*OperationStats),
		byLevel:     make(map[string]*LevelStats),
		layers:      make(map[string]CacheLayer),
	}
}

// RecordHit records a cache hit for the given operation and level with its lookup latency
func (m *CacheMetrics) RecordHit(operation, level string, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalHits++
	m.totalLatency += latency

	// Per operation
	op, ok := m.byOperation[operation]
	if !ok {
		op = &OperationStats{}
		m.byOperation[operation] = op
	}
	op.Hits++
	op.TotalLatency += latency

	// Per level
	lv, ok := m.byLevel[level]
	if !ok {
		lv = &LevelStats{}
		m.byLevel[level] = lv
	}
	lv.Hits++
	lv.TotalLatency += latency
}

// RecordMiss records a cache miss for the given operation
func (m *CacheMetrics) RecordMiss(operation string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalMisses++

	op, ok := m.byOperation[operation]
	if !ok {
		op = &OperationStats{}
		m.byOperation[operation] = op
	}
	op.Misses++
}

// RegisterLayer registers a CacheLayer so its Stats() are included in snapshots
func (m *CacheMetrics) RegisterLayer(name string, layer CacheLayer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.layers[name] = layer
}

// GetStats returns a point-in-time snapshot of all cache metrics
func (m *CacheMetrics) GetStats() CacheMetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := m.totalHits + m.totalMisses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(m.totalHits) / float64(total) * 100
	}

	avgLatency := time.Duration(0)
	if m.totalHits > 0 {
		avgLatency = m.totalLatency / time.Duration(m.totalHits)
	}

	// Copy operation stats
	byOp := make(map[string]OperationStats, len(m.byOperation))
	for name, op := range m.byOperation {
		opCopy := *op
		opTotal := op.Hits + op.Misses
		if opTotal > 0 {
			opCopy.HitRate = float64(op.Hits) / float64(opTotal) * 100
		}
		if op.Hits > 0 {
			opCopy.AvgLatency = op.TotalLatency / time.Duration(op.Hits)
		}
		byOp[name] = opCopy
	}

	// Copy level stats
	byLv := make(map[string]LevelStats, len(m.byLevel))
	for name, lv := range m.byLevel {
		lvCopy := *lv
		if lv.Hits > 0 {
			lvCopy.AvgLatency = lv.TotalLatency / time.Duration(lv.Hits)
		}
		byLv[name] = lvCopy
	}

	// Layer snapshots (collect under read lock — CacheLayer.Stats() must be safe to call)
	var snapshots []CacheStats
	for _, layer := range m.layers {
		snapshots = append(snapshots, layer.Stats())
	}

	return CacheMetricsSnapshot{
		TotalHits:      m.totalHits,
		TotalMisses:    m.totalMisses,
		HitRate:        hitRate,
		AvgHitLatency:  avgLatency,
		ByOperation:    byOp,
		ByLevel:        byLv,
		LayerSnapshots: snapshots,
	}
}

// Reset clears all accumulated metrics (operation and level maps, totals)
func (m *CacheMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalHits = 0
	m.totalMisses = 0
	m.totalLatency = 0
	m.byOperation = make(map[string]*OperationStats)
	m.byLevel = make(map[string]*LevelStats)
}
