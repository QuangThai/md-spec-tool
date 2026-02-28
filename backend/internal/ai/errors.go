package ai

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrAIUnavailable      = errors.New("ai_unavailable")       // Network, 5xx, timeout
	ErrAIRateLimited      = errors.New("ai_rate_limited")      // 429
	ErrAIInvalidOutput    = errors.New("ai_invalid_output")    // JSON parse error
	ErrAIValidationFailed = errors.New("ai_validation_failed") // Output validation failed
	ErrAIRefused          = errors.New("ai_refused")           // Model refused (safety/content policy)
	ErrAITruncated        = errors.New("ai_truncated")         // Response truncated (max tokens)
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

// ErrAIContentFiltered is returned when the AI's content filter blocks a response
var ErrAIContentFiltered = errors.New("ai_content_filtered")

// ErrorCategory classifies errors for retry/escalation decisions
type ErrorCategory string

const (
	ErrorCategoryTransient ErrorCategory = "transient" // retry with backoff
	ErrorCategoryPermanent ErrorCategory = "permanent" // fail fast, don't retry
	ErrorCategoryContent   ErrorCategory = "content"   // AI-specific content issue
)

// ClassifiedError wraps an error with classification metadata
type ClassifiedError struct {
	Original    error
	Category    ErrorCategory
	ShouldRetry bool
	StatusCode  int
	Message     string
}

func (e *ClassifiedError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Category, e.Message, e.Original)
}

func (e *ClassifiedError) Unwrap() error {
	return e.Original
}

// ClassifyError categorizes OpenAI API errors into transient/permanent/content
func ClassifyError(statusCode int, err error) *ClassifiedError {
	// 1. Check content errors first (sentinel errors - these are highest priority)
	switch {
	case errors.Is(err, ErrAIRefused):
		return &ClassifiedError{Original: err, Category: ErrorCategoryContent, ShouldRetry: false, StatusCode: statusCode, Message: "model refused request"}
	case errors.Is(err, ErrAITruncated):
		return &ClassifiedError{Original: err, Category: ErrorCategoryContent, ShouldRetry: true, StatusCode: statusCode, Message: "response truncated, should retry with higher max_tokens"}
	case errors.Is(err, ErrAIContentFiltered):
		return &ClassifiedError{Original: err, Category: ErrorCategoryContent, ShouldRetry: false, StatusCode: statusCode, Message: "content filtered"}
	}

	// 2. Check internal validation/output errors - these are permanent
	switch {
	case errors.Is(err, ErrAIInvalidOutput):
		return &ClassifiedError{Original: err, Category: ErrorCategoryPermanent, ShouldRetry: false, StatusCode: statusCode, Message: "invalid JSON output from model"}
	case errors.Is(err, ErrAIValidationFailed):
		return &ClassifiedError{Original: err, Category: ErrorCategoryPermanent, ShouldRetry: false, StatusCode: statusCode, Message: "AI output validation failed"}
	}

	// 3. Check rate limit and unavailable sentinels
	switch {
	case errors.Is(err, ErrAIRateLimited):
		return &ClassifiedError{Original: err, Category: ErrorCategoryTransient, ShouldRetry: true, StatusCode: 429, Message: "rate limited"}
	case errors.Is(err, ErrAIUnavailable):
		return &ClassifiedError{Original: err, Category: ErrorCategoryTransient, ShouldRetry: true, StatusCode: statusCode, Message: "AI service unavailable"}
	}

	// 4. Check transient by status code
	switch {
	case statusCode == 429:
		return &ClassifiedError{Original: err, Category: ErrorCategoryTransient, ShouldRetry: true, StatusCode: statusCode, Message: "rate limited"}
	case statusCode >= 500:
		return &ClassifiedError{Original: err, Category: ErrorCategoryTransient, ShouldRetry: true, StatusCode: statusCode, Message: "server error"}
	case statusCode == 408 || statusCode == 504:
		// Timeout-like statuses are transient
		return &ClassifiedError{Original: err, Category: ErrorCategoryTransient, ShouldRetry: true, StatusCode: statusCode, Message: "timeout"}
	}

	// 5. Check transient by error type (timeout, network)
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return &ClassifiedError{Original: err, Category: ErrorCategoryTransient, ShouldRetry: true, StatusCode: 0, Message: "timeout or cancelled"}
	}

	// 6. Remaining HTTP client errors (4xx except 429) → permanent
	if statusCode >= 400 && statusCode < 500 && statusCode != 429 {
		return &ClassifiedError{Original: err, Category: ErrorCategoryPermanent, ShouldRetry: false, StatusCode: statusCode, Message: "client error"}
	}

	// 7. Unknown → transient (safer to retry, but bounded by outer retry logic)
	return &ClassifiedError{Original: err, Category: ErrorCategoryTransient, ShouldRetry: true, StatusCode: statusCode, Message: "unknown error"}
}
