package ai

import (
	"errors"
	"fmt"
)

var (
	ErrAIUnavailable      = errors.New("ai_unavailable")       // Network, 5xx, timeout
	ErrAIRateLimited      = errors.New("ai_rate_limited")      // 429
	ErrAIInvalidOutput    = errors.New("ai_invalid_output")    // JSON parse error
	ErrAIValidationFailed = errors.New("ai_validation_failed") // Output validation failed
	ErrAIRefused          = errors.New("ai_refused")          // Model refused (safety/content policy)
	ErrAITruncated        = errors.New("ai_truncated")        // Response truncated (max tokens)
)

// AIError wraps domain errors with additional context
type AIError struct {
	Err        error
	Message    string
	RetryAfter int // seconds, for rate limiting
}

// Error implements the error interface
func (ae *AIError) Error() string {
	if ae.Message != "" {
		return fmt.Sprintf("%v: %s", ae.Err, ae.Message)
	}
	return ae.Err.Error()
}

// Unwrap returns the wrapped error for errors.Is/As chain traversal
func (ae *AIError) Unwrap() error {
	return ae.Err
}
