package middleware

import (
	"errors"
	"fmt"
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

// RateLimitConfig holds rate limit parameters
type RateLimitConfig struct {
	Limit  int           // requests per window
	Window time.Duration // rate limit window
}

// RateLimit enforces a fixed-window, per-IP rate limit.
// Returns 429 RATE_LIMIT_EXCEEDED with Retry-After header.
//
// Example usage:
//   router.POST("/api/heavy", middleware.RateLimit(60, time.Minute), handler)
//
// Tracking:
// - Per client IP (uses ClientIP() which respects X-Forwarded-For with trusted proxies)
// - Fixed window resets every `window` duration
// - In-memory tracking (not distributed across instances)
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

			// Construct ErrRateLimit with retry info
			err := &ErrRateLimit{
				Err:        errors.New(fmt.Sprintf("rate limit exceeded: %d requests per %v", limit, window)),
				RetryAfter: retryAfter,
			}

			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, NewErrorPayload(http.StatusTooManyRequests,
				err.Error(),
				GetRequestID(c),
			).WithDetails(map[string]any{
				"limit":        limit,
				"window_secs":  int(window.Seconds()),
				"retry_after":  retryAfter,
			}))
			return
		}

		entry.count++
		hits[ip] = entry
		mu.Unlock()

		c.Next()
	}
}
