package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetRequestID retrieves the request ID from context
func GetRequestID(c *gin.Context) string {
	if id, ok := c.Request.Context().Value(RequestIDContextKey).(string); ok {
		return id
	}
	return ""
}

// =====================
// Error Types (Taxonomy)
// =====================

// ErrBadRequest wraps an error with 400 status
type ErrBadRequest struct{ Err error }

func (e *ErrBadRequest) Error() string { return e.Err.Error() }
func (e *ErrBadRequest) Unwrap() error { return e.Err }

// ErrUnauthorized wraps an error with 401 status
type ErrUnauthorized struct{ Err error }

func (e *ErrUnauthorized) Error() string { return e.Err.Error() }
func (e *ErrUnauthorized) Unwrap() error { return e.Err }

// ErrForbidden wraps an error with 403 status (permission/quota denied)
type ErrForbidden struct{ Err error }

func (e *ErrForbidden) Error() string { return e.Err.Error() }
func (e *ErrForbidden) Unwrap() error { return e.Err }

// ErrNotFound wraps an error with 404 status
type ErrNotFound struct{ Err error }

func (e *ErrNotFound) Error() string { return e.Err.Error() }
func (e *ErrNotFound) Unwrap() error { return e.Err }

// ErrRequestTooLarge wraps an error with 413 status
type ErrRequestTooLarge struct{ Err error }

func (e *ErrRequestTooLarge) Error() string { return e.Err.Error() }
func (e *ErrRequestTooLarge) Unwrap() error { return e.Err }

// ErrRateLimit wraps an error with 429 status
// Includes RetryAfter for client backoff guidance
type ErrRateLimit struct {
	Err        error
	RetryAfter int // seconds
}

func (e *ErrRateLimit) Error() string { return e.Err.Error() }
func (e *ErrRateLimit) Unwrap() error { return e.Err }

// ErrServiceUnavailable wraps an error with 503 status
type ErrServiceUnavailable struct{ Err error }

func (e *ErrServiceUnavailable) Error() string { return e.Err.Error() }
func (e *ErrServiceUnavailable) Unwrap() error { return e.Err }

// =====================
// Standard Error Response
// =====================

// ErrorPayload is the structured JSON error response.
// Follows API error contract: https://docs.internal/ERROR-CONTRACT.md
type ErrorPayload struct {
	Error            string         `json:"error"`
	Code             string         `json:"code,omitempty"`
	ValidationReason string         `json:"validation_reason,omitempty"`
	Details          map[string]any `json:"details,omitempty"`
	RequestID        string         `json:"request_id,omitempty"`
}

// NewErrorPayload constructs a standardized error response
func NewErrorPayload(status int, msg string, requestID string) ErrorPayload {
	return ErrorPayload{
		Error:     msg,
		Code:      codeForStatus(status),
		RequestID: requestID,
	}
}

// WithDetails adds detail fields to error payload
func (e ErrorPayload) WithDetails(details map[string]any) ErrorPayload {
	e.Details = details
	return e
}

// WithValidationReason adds validation context
func (e ErrorPayload) WithValidationReason(reason string) ErrorPayload {
	e.ValidationReason = reason
	return e
}

// =====================
// Error Handler Middleware
// =====================

// ErrorHandler returns middleware that centralizes error handling.
// Handlers should call c.Error(err) and return without writing a response;
// this middleware maps errors to status codes and returns consistent JSON.
// Skips when the handler has already written a response.
//
// Maps Go errors to HTTP status + error codes:
// - ErrBadRequest → 400 BAD_REQUEST
// - ErrUnauthorized → 401 UNAUTHORIZED
// - ErrForbidden → 403 FORBIDDEN
// - ErrNotFound → 404 NOT_FOUND
// - ErrRequestTooLarge → 413 REQUEST_TOO_LARGE
// - ErrRateLimit → 429 RATE_LIMIT_EXCEEDED
// - ErrServiceUnavailable → 503 SERVICE_UNAVAILABLE
// - all others → 500 INTERNAL_ERROR
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Written() {
			return
		}
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		status := statusForError(err)
		msg := err.Error()
		requestID := GetRequestID(c)

		// Log error with context
		logLevel := slog.LevelInfo
		if status >= 500 {
			logLevel = slog.LevelError
		}
		slog.Log(c.Request.Context(), logLevel,
			"api_error",
			"request_id", requestID,
			"status", status,
			"error", msg,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
		)

		// Build response
		payload := NewErrorPayload(status, msg, requestID)

		// Extract details if available (for validation errors, etc.)
		if errWithDetails, ok := err.(interface{ Details() map[string]any }); ok {
			payload.Details = errWithDetails.Details()
		}

		// Set Retry-After header for rate limit responses
		if rl, ok := err.(*ErrRateLimit); ok && rl.RetryAfter > 0 {
			c.Header("Retry-After", string(rune(rl.RetryAfter)))
		}

		c.JSON(status, payload)
	}
}

// =====================
// Helper Functions
// =====================

func codeForStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusRequestEntityTooLarge:
		return "REQUEST_TOO_LARGE"
	case http.StatusTooManyRequests:
		return "RATE_LIMIT_EXCEEDED"
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	default:
		return "INTERNAL_ERROR"
	}
}

func statusForError(err error) int {
	switch {
	case errors.As(err, new(*ErrBadRequest)):
		return http.StatusBadRequest
	case errors.As(err, new(*ErrUnauthorized)):
		return http.StatusUnauthorized
	case errors.As(err, new(*ErrForbidden)):
		return http.StatusForbidden
	case errors.As(err, new(*ErrNotFound)):
		return http.StatusNotFound
	case errors.As(err, new(*ErrRequestTooLarge)):
		return http.StatusRequestEntityTooLarge
	case errors.As(err, new(*ErrRateLimit)):
		return http.StatusTooManyRequests
	case errors.As(err, new(*ErrServiceUnavailable)):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
