package ai

import (
	"log/slog"
	"strings"
	"sync"
	"time"
)

// BYOKValidationResult holds the outcome of an API key format validation.
type BYOKValidationResult struct {
	Valid  bool
	Reason string // empty when Valid is true
}

// ValidateBYOKKey checks that key conforms to the expected OpenAI API key
// format: must start with "sk-", be at least 40 characters, and must not be
// an obviously invalid/test key.
//
// This is a lightweight structural check — it does not make a network request.
func ValidateBYOKKey(key string) BYOKValidationResult {
	if key == "" {
		return BYOKValidationResult{Valid: false, Reason: "API key is empty"}
	}

	if !strings.HasPrefix(key, "sk-") {
		return BYOKValidationResult{Valid: false, Reason: "API key must start with 'sk-'"}
	}

	if len(key) < 40 {
		return BYOKValidationResult{Valid: false, Reason: "API key is too short (minimum 40 characters)"}
	}

	// Reject keys that are explicitly marked as test keys.
	if strings.HasPrefix(key, "sk-test-") {
		slog.Warn("BYOK: test API key rejected", "key_prefix", key[:min(len(key), 12)])
		return BYOKValidationResult{Valid: false, Reason: "test API keys are not accepted"}
	}

	// sk-proj-... (project-scoped keys) are valid as long as they pass the
	// length check above; no additional restrictions apply.
	return BYOKValidationResult{Valid: true}
}

// ── Rate limiter ──────────────────────────────────────────────────────────────

// BYOKRateLimiter enforces a sliding-window rate limit on BYOK key validation
// attempts per client IP. This prevents brute-force enumeration of API keys.
type BYOKRateLimiter struct {
	mu          sync.Mutex
	attempts    map[string][]time.Time
	maxAttempts int
	window      time.Duration
}

// NewBYOKRateLimiter creates a rate limiter allowing at most maxAttempts
// validation attempts per clientIP within the given window duration.
func NewBYOKRateLimiter(maxAttempts int, window time.Duration) *BYOKRateLimiter {
	return &BYOKRateLimiter{
		attempts:    make(map[string][]time.Time),
		maxAttempts: maxAttempts,
		window:      window,
	}
}

// Allow returns true if clientIP has not exceeded the rate limit and records
// the attempt. Returns false (without recording) when the limit is exceeded.
func (r *BYOKRateLimiter) Allow(clientIP string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	// Slide the window: discard attempts older than cutoff.
	prev := r.attempts[clientIP]
	valid := prev[:0] // reuse the backing array
	for _, t := range prev {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= r.maxAttempts {
		r.attempts[clientIP] = valid
		slog.Warn("BYOK: rate limit exceeded", "client_ip", clientIP, "attempts", len(valid))
		return false
	}

	r.attempts[clientIP] = append(valid, now)
	return true
}
