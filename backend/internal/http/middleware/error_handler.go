package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrBadRequest wraps an error with 400 status
type ErrBadRequest struct{ Err error }

func (e *ErrBadRequest) Error() string { return e.Err.Error() }
func (e *ErrBadRequest) Unwrap() error { return e.Err }

// ErrNotFound wraps an error with 404 status
type ErrNotFound struct{ Err error }

func (e *ErrNotFound) Error() string { return e.Err.Error() }
func (e *ErrNotFound) Unwrap() error { return e.Err }

// ErrRequestTooLarge wraps an error with 413 status
type ErrRequestTooLarge struct{ Err error }

func (e *ErrRequestTooLarge) Error() string { return e.Err.Error() }
func (e *ErrRequestTooLarge) Unwrap() error { return e.Err }

// ErrorPayload is the structured JSON error response
type ErrorPayload struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// ErrorHandler returns middleware that centralizes error handling.
// Handlers should call c.Error(err) and return without writing a response;
// this middleware maps errors to status codes and returns consistent JSON.
// Skips when the handler has already written a response.
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

		slog.Debug("error handler", "status", status, "error", msg)
		c.JSON(status, ErrorPayload{Error: msg})
	}
}

func statusForError(err error) int {
	switch {
	case errors.As(err, new(*ErrBadRequest)):
		return http.StatusBadRequest
	case errors.As(err, new(*ErrNotFound)):
		return http.StatusNotFound
	case errors.As(err, new(*ErrRequestTooLarge)):
		return http.StatusRequestEntityTooLarge
	default:
		return http.StatusInternalServerError
	}
}
