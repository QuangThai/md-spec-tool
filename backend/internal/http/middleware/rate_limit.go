package middleware

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	count       int
	windowStart time.Time
}

// RateLimit enforces a fixed-window, per-IP rate limit.
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	hits := make(map[string]rateLimitEntry)
	
	// Start periodic cleanup goroutine for expired entries
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for ip, entry := range hits {
				if now.Sub(entry.windowStart) >= window {
					delete(hits, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		entry := hits[ip]
		if entry.windowStart.IsZero() || now.Sub(entry.windowStart) >= window {
			entry.windowStart = now
			entry.count = 0
		}

		if entry.count >= limit {
			remaining := window - now.Sub(entry.windowStart)
			mu.Unlock()

			retryAfter := int(math.Ceil(remaining.Seconds()))
			if retryAfter < 0 {
				retryAfter = 0
			}
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		entry.count++
		hits[ip] = entry
		mu.Unlock()

		c.Next()
	}
}
