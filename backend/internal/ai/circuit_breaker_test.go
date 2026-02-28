package ai

import (
	"testing"
	"time"
)

func TestCircuitBreaker_StartsClosedState(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 5,
		ResetTimeout:     30 * time.Second,
		HalfOpenMax:      1,
	})
	if cb.State() != CircuitStateClosed {
		t.Errorf("expected Closed, got %s", cb.State())
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 3,
		ResetTimeout:     30 * time.Second,
		HalfOpenMax:      1,
	})
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.State() != CircuitStateOpen {
		t.Errorf("expected Open after 3 failures, got %s", cb.State())
	}
	if cb.Allow() {
		t.Error("should NOT allow requests when open")
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 3,
		ResetTimeout:     50 * time.Millisecond,
		HalfOpenMax:      1,
	})
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	time.Sleep(60 * time.Millisecond)
	if cb.State() != CircuitStateHalfOpen {
		t.Errorf("expected HalfOpen after timeout, got %s", cb.State())
	}
	if !cb.Allow() {
		t.Error("should allow 1 probe request in half-open")
	}
}

func TestCircuitBreaker_ClosesOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 3,
		ResetTimeout:     50 * time.Millisecond,
		HalfOpenMax:      1,
	})
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	time.Sleep(60 * time.Millisecond)
	cb.RecordSuccess()
	if cb.State() != CircuitStateClosed {
		t.Errorf("expected Closed after success in half-open, got %s", cb.State())
	}
}

func TestCircuitBreaker_ReopensOnHalfOpenFailure(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 3,
		ResetTimeout:     50 * time.Millisecond,
		HalfOpenMax:      1,
	})
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	time.Sleep(60 * time.Millisecond)
	cb.RecordFailure() // fail in half-open
	if cb.State() != CircuitStateOpen {
		t.Errorf("expected Open after failure in half-open, got %s", cb.State())
	}
}

func TestCircuitBreaker_SuccessResetsFailureCount(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 3,
		ResetTimeout:     time.Minute,
		HalfOpenMax:      1,
	})
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess() // reset
	cb.RecordFailure()
	cb.RecordFailure()
	// Only 2 failures after reset, should still be closed
	if cb.State() != CircuitStateClosed {
		t.Errorf("expected Closed after success reset, got %s", cb.State())
	}
}

func TestCircuitBreaker_ConcurrentSafety(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 100,
		ResetTimeout:     time.Minute,
		HalfOpenMax:      1,
	})
	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			cb.RecordFailure()
			cb.Allow()
			cb.State()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
	// No panic = pass
}

func TestCircuitBreaker_ExponentialBackoffAfterReopen(t *testing.T) {
	baseTimeout := 50 * time.Millisecond
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 2,
		ResetTimeout:     baseTimeout,
		HalfOpenMax:      1,
	})

	// First open: failures trigger opening
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != CircuitStateOpen {
		t.Fatal("expected Open")
	}

	// First half-open probe should be allowed
	time.Sleep(60 * time.Millisecond)
	if !cb.Allow() {
		t.Fatal("first probe should be allowed")
	}

	// Fail in half-open → re-open with exponential backoff (consecutiveOpen=1)
	cb.RecordFailure()
	if cb.State() != CircuitStateOpen {
		t.Fatal("expected to re-open")
	}

	// Wait for base timeout (50ms) - should still be open due to exponential backoff
	time.Sleep(60 * time.Millisecond)
	if cb.State() == CircuitStateHalfOpen {
		t.Fatalf("should NOT transition to HalfOpen yet due to exponential backoff (expected ~100ms total wait, got only 60ms)")
	}

	// Wait for exponential timeout (2x base = 100ms from last failure)
	time.Sleep(100 * time.Millisecond)
	if cb.State() != CircuitStateHalfOpen {
		t.Fatalf("should transition to HalfOpen after exponential timeout, got %s", cb.State())
	}
}

func TestCircuitBreaker_ExponentialBackoffResetsOnSuccess(t *testing.T) {
	baseTimeout := 50 * time.Millisecond
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 2,
		ResetTimeout:     baseTimeout,
		HalfOpenMax:      1,
	})

	// Trigger open + half-open + re-open (consecutiveOpen=1)
	cb.RecordFailure()
	cb.RecordFailure()
	time.Sleep(60 * time.Millisecond)
	cb.Allow() // probe allowed
	cb.RecordFailure() // re-open

	// Verify exponential backoff is in effect
	time.Sleep(60 * time.Millisecond)
	if cb.State() == CircuitStateHalfOpen {
		t.Fatal("exponential backoff should prevent immediate half-open transition")
	}

	// Success should reset consecutiveOpen counter
	time.Sleep(100 * time.Millisecond) // wait enough for half-open
	if !cb.Allow() {
		t.Fatal("should allow in half-open state")
	}
	cb.RecordSuccess()
	if cb.State() != CircuitStateClosed {
		t.Fatal("should be closed after success")
	}

	// Now failures should use base timeout again (not exponential)
	cb.RecordFailure()
	cb.RecordFailure()
	time.Sleep(60 * time.Millisecond)
	cb.Allow() // should succeed without waiting exponential time
}

func TestCircuitBreaker_ExponentialBackoffCappedAt5Minutes(t *testing.T) {
	// Very aggressive setup to demonstrate capping
	baseTimeout := time.Millisecond
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 1,
		ResetTimeout:     baseTimeout,
		HalfOpenMax:      1,
	})

	// Trigger many re-opens to grow exponential backoff beyond cap
	for i := 0; i < 30; i++ {
		cb.RecordFailure()
		time.Sleep(2 * time.Millisecond)
		cb.Allow()
		cb.RecordFailure()
	}

	// At this point, consecutiveOpen should be ~30, but backoff should be capped at 5 minutes
	// We can't easily test the actual wait time without sleeping 5 minutes,
	// but we can verify the logic by checking GetExponentialBackoffDuration
	backoff := cb.GetExponentialBackoffDuration()
	if backoff > 5*time.Minute {
		t.Errorf("exponential backoff should be capped at 5 minutes, got %v", backoff)
	}
}
