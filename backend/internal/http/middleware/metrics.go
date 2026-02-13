package middleware

import (
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics holds simple request metrics for observability.
// Lightweight alternative to Prometheus for basic monitoring.
type Metrics struct {
	totalRequests atomic.Uint64
	totalLatency  atomic.Uint64 // sum of request durations in milliseconds
}

var defaultMetrics = &Metrics{}

// MetricsMiddleware records request count and latency.
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Milliseconds()
		defaultMetrics.totalRequests.Add(1)
		defaultMetrics.totalLatency.Add(uint64(duration))
	}
}

// GetMetrics returns current metrics snapshot.
func GetMetrics() map[string]interface{} {
	requests := defaultMetrics.totalRequests.Load()
	latencySum := defaultMetrics.totalLatency.Load()
	avgMs := float64(0)
	if requests > 0 {
		avgMs = float64(latencySum) / float64(requests)
	}
	return map[string]interface{}{
		"total_requests": requests,
		"avg_latency_ms": avgMs,
	}
}
