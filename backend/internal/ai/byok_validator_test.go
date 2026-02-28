package ai

import (
	"strings"
	"testing"
	"time"
)

// ── Format validation ─────────────────────────────────────────────────────────

func TestValidateBYOKKey_ValidFormat(t *testing.T) {
	// A realistic-length key that meets all requirements
	key := "sk-" + strings.Repeat("a", 48) // 51 chars total
	result := ValidateBYOKKey(key)
	if !result.Valid {
		t.Errorf("expected valid key, got reason: %q", result.Reason)
	}
	if result.Reason != "" {
		t.Errorf("expected empty reason for valid key, got %q", result.Reason)
	}
}

func TestValidateBYOKKey_Empty(t *testing.T) {
	result := ValidateBYOKKey("")
	if result.Valid {
		t.Error("empty key should be invalid")
	}
	if result.Reason == "" {
		t.Error("expected a non-empty reason for empty key")
	}
}

func TestValidateBYOKKey_TooShort(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"just_prefix", "sk-"},
		{"short_key", "sk-abc"},
		{"39_chars", "sk-" + strings.Repeat("x", 36)}, // 3+36 = 39 chars
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateBYOKKey(tc.key)
			if result.Valid {
				t.Errorf("key %q (%d chars) should be invalid (too short)", tc.key, len(tc.key))
			}
		})
	}
}

func TestValidateBYOKKey_MissingPrefix(t *testing.T) {
	keys := []string{
		"openai-" + strings.Repeat("x", 40),
		strings.Repeat("x", 50),
		"Bearer " + strings.Repeat("x", 50),
		"SK-" + strings.Repeat("x", 48), // wrong case
	}
	for _, key := range keys {
		t.Run(key[:min(len(key), 20)], func(t *testing.T) {
			result := ValidateBYOKKey(key)
			if result.Valid {
				t.Errorf("key without 'sk-' prefix should be invalid: %q", key[:20])
			}
		})
	}
}

func TestValidateBYOKKey_TestKey(t *testing.T) {
	// Keys starting with sk-test- must be rejected regardless of length
	key := "sk-test-" + strings.Repeat("x", 50)
	result := ValidateBYOKKey(key)
	if result.Valid {
		t.Error("test key should be rejected")
	}
	if result.Reason == "" {
		t.Error("expected non-empty reason for test key")
	}
}

func TestValidateBYOKKey_ProjectKey(t *testing.T) {
	// sk-proj-... is the format for OpenAI project-scoped API keys — must be accepted
	key := "sk-proj-" + strings.Repeat("a", 48) // 56 chars
	result := ValidateBYOKKey(key)
	if !result.Valid {
		t.Errorf("project-scoped key should be valid, got reason: %q", result.Reason)
	}
}

// ── Rate limiter ─────────────────────────────────────────────────────────────

func TestBYOKRateLimiter_AllowsNormalRate(t *testing.T) {
	limiter := NewBYOKRateLimiter(3, 1*time.Minute)
	ip := "192.168.1.1"

	for i := 1; i <= 3; i++ {
		if !limiter.Allow(ip) {
			t.Errorf("attempt %d should be allowed (limit is 3)", i)
		}
	}
}

func TestBYOKRateLimiter_BlocksExcessiveAttempts(t *testing.T) {
	limiter := NewBYOKRateLimiter(3, 1*time.Minute)
	ip := "10.0.0.1"

	// Consume all 3 slots
	for i := 0; i < 3; i++ {
		if !limiter.Allow(ip) {
			t.Fatalf("attempt %d should be allowed before limit is reached", i+1)
		}
	}

	// 4th attempt must be blocked
	if limiter.Allow(ip) {
		t.Error("4th attempt should be blocked (limit exceeded)")
	}

	// 5th attempt also blocked
	if limiter.Allow(ip) {
		t.Error("5th attempt should also be blocked")
	}
}

func TestBYOKRateLimiter_IsolatesClients(t *testing.T) {
	limiter := NewBYOKRateLimiter(2, 1*time.Minute)

	// Exhaust ip1
	limiter.Allow("1.1.1.1")
	limiter.Allow("1.1.1.1")
	if limiter.Allow("1.1.1.1") {
		t.Error("1.1.1.1 should be blocked after 2 attempts")
	}

	// ip2 should still have a clean slate
	if !limiter.Allow("2.2.2.2") {
		t.Error("2.2.2.2 should not be affected by 1.1.1.1's exhaustion")
	}
}

func TestBYOKRateLimiter_ResetsAfterWindow(t *testing.T) {
	// Use a very short window so we can test expiry without long sleeps.
	window := 50 * time.Millisecond
	limiter := NewBYOKRateLimiter(2, window)
	ip := "172.16.0.1"

	limiter.Allow(ip)
	limiter.Allow(ip)
	if limiter.Allow(ip) {
		t.Fatal("should be blocked after 2 attempts")
	}

	// Wait for window to expire
	time.Sleep(window + 10*time.Millisecond)

	// Now the window has slid; old attempts are stale
	if !limiter.Allow(ip) {
		t.Error("should be allowed after window expiry")
	}
}
