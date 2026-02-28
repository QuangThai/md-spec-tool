package ai

import (
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker
type CircuitState string

const (
	CircuitStateClosed   CircuitState = "closed"
	CircuitStateOpen     CircuitState = "open"
	CircuitStateHalfOpen CircuitState = "half_open"
)

// CircuitBreakerConfig configures the circuit breaker
type CircuitBreakerConfig struct {
	FailureThreshold int           // consecutive failures before opening
	ResetTimeout     time.Duration // initial cooldown before transitioning to half-open
	HalfOpenMax      int           // max probe requests allowed in half-open
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		ResetTimeout:     30 * time.Second,
		HalfOpenMax:      1,
	}
}

// CircuitBreaker implements the circuit breaker pattern for AI calls with exponential backoff
type CircuitBreaker struct {
	mu              sync.Mutex
	config          CircuitBreakerConfig
	state           CircuitState
	failures        int
	lastFailureAt   time.Time
	halfOpenCount   int
	consecutiveOpen int // count how many times we've re-opened after half-open probe failure
}

// NewCircuitBreaker creates a new circuit breaker in closed state
func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: cfg,
		state:  CircuitStateClosed,
	}
}

// GetExponentialBackoffDuration calculates the backoff duration with exponential growth.
// Each consecutive re-open doubles the wait time, capped at 5 minutes.
func (cb *CircuitBreaker) GetExponentialBackoffDuration() time.Duration {
	base := cb.config.ResetTimeout
	// Apply exponential backoff: 2^consecutiveOpen * base, capped at 5min
	multiplier := 1 << uint(cb.consecutiveOpen)
	backoff := time.Duration(multiplier) * base
	const maxBackoff = 5 * time.Minute
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	return backoff
}

// State returns the current circuit breaker state.
// Automatically transitions Open -> HalfOpen when exponential backoff has elapsed.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.checkAndTransition()
	return cb.state
}

// checkAndTransition checks if Open -> HalfOpen transition should occur (helper for State and Allow)
func (cb *CircuitBreaker) checkAndTransition() {
	if cb.state == CircuitStateOpen && time.Since(cb.lastFailureAt) > cb.GetExponentialBackoffDuration() {
		cb.state = CircuitStateHalfOpen
		cb.halfOpenCount = 0
	}
}

// Allow returns true if a request should be allowed through.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	// Check for Open -> HalfOpen transition with exponential backoff
	cb.checkAndTransition()
	
	switch cb.state {
	case CircuitStateClosed:
		return true
	case CircuitStateOpen:
		return false
	case CircuitStateHalfOpen:
		if cb.halfOpenCount < cb.config.HalfOpenMax {
			cb.halfOpenCount++
			return true
		}
		return false
	}
	return false
}

// RecordSuccess records a successful call. Transitions HalfOpen -> Closed and resets all failure counts.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.consecutiveOpen = 0 // reset exponential backoff counter on success
	cb.state = CircuitStateClosed
	cb.halfOpenCount = 0
}

// RecordFailure records a failed call. May transition Closed -> Open or HalfOpen -> Open.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailureAt = time.Now()
	if cb.state == CircuitStateHalfOpen {
		cb.state = CircuitStateOpen
		cb.consecutiveOpen++ // increment exponential backoff counter when re-opening from half-open
		return
	}
	if cb.failures >= cb.config.FailureThreshold {
		cb.state = CircuitStateOpen
		// Don't increment consecutiveOpen here; only on half-open probe failures
		// Incrementing only on half-open re-opens prevents aggressive backoff on initial burst failures
	}
}
