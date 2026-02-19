package middleware

import (
	"bytes"
	"errors"
	"io"
	"log/slog"

	"github.com/gin-gonic/gin"
)

// ValidationError represents a validation error with details
type ValidationError struct {
	Err            error
	Reason         string                 // human-readable reason
	FieldDetails   map[string]any         // field-specific errors
}

func (e *ValidationError) Error() string { return e.Err.Error() }
func (e *ValidationError) Details() map[string]any { return e.FieldDetails }
func (e *ValidationError) Unwrap() error { return e.Err }

// RequestBodyValidator validates request body size and content type
func RequestBodyValidator(maxBodyBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip GET, HEAD, DELETE (no body)
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "DELETE" {
			c.Next()
			return
		}

		// Check Content-Length if available
		if c.Request.ContentLength > 0 && c.Request.ContentLength > maxBodyBytes {
			err := &ErrRequestTooLarge{
				Err: errors.New("request body exceeds maximum size limit"),
			}
			requestID := GetRequestID(c)
			slog.With("request_id", requestID).Warn("request_too_large",
				"content_length", c.Request.ContentLength,
				"max_bytes", maxBodyBytes,
			)
			c.Error(err)
			c.AbortWithStatusJSON(413, NewErrorPayload(413,
				err.Error(),
				requestID,
			).WithDetails(map[string]any{
				"max_bytes": maxBodyBytes,
				"received":  c.Request.ContentLength,
			}))
			return
		}

		// Read and validate body
		body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodyBytes+1))
		if err != nil {
			err := &ErrBadRequest{Err: err}
			requestID := GetRequestID(c)
			slog.With("request_id", requestID).Warn("failed_to_read_body", "error", err)
			c.Error(err)
			c.AbortWithStatusJSON(400, NewErrorPayload(400, err.Error(), requestID))
			return
		}

		// Check if body exceeded limit (read one extra byte)
		if int64(len(body)) > maxBodyBytes {
			err := &ErrRequestTooLarge{
				Err: errors.New("request body exceeds maximum size limit"),
			}
			requestID := GetRequestID(c)
			slog.With("request_id", requestID).Warn("request_too_large",
				"body_bytes", len(body),
				"max_bytes", maxBodyBytes,
			)
			c.Error(err)
			c.AbortWithStatusJSON(413, NewErrorPayload(413, err.Error(), requestID))
			return
		}

		// Restore body for handler
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		c.Next()
	}
}

// FileValidator validates uploaded file properties
type FileValidator struct {
	MaxBytes     int64
	AllowedTypes map[string]bool // e.g., {"application/vnd.ms-excel": true}
}

// ValidateFile checks file size and content type
// Returns ValidationError if validation fails
func (v *FileValidator) ValidateFile(filename string, contentType string, fileSize int64) error {
	details := make(map[string]any)

	// Check file size
	if fileSize > v.MaxBytes {
		details["size"] = "file exceeds maximum size"
		return &ValidationError{
			Err:            errors.New("file size validation failed"),
			Reason:         "file too large",
			FieldDetails:   details,
		}
	}

	// Check content type if whitelist provided
	if len(v.AllowedTypes) > 0 && !v.AllowedTypes[contentType] {
		details["content_type"] = "file type not allowed: " + contentType
		return &ValidationError{
			Err:            errors.New("file content type validation failed"),
			Reason:         "invalid file type",
			FieldDetails:   details,
		}
	}

	return nil
}

// AllowedExcelMimeTypes defines valid Excel file MIME types
var AllowedExcelMimeTypes = map[string]bool{
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true, // .xlsx
	"application/vnd.ms-excel": true, // .xls
	"application/octet-stream": true, // fallback for Excel files
}

// AllowedImageMimeTypes defines valid image MIME types
var AllowedImageMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// JSONBodyValidator ensures request body is valid JSON
func JSONBodyValidator(maxBodyBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip non-JSON endpoints
		if c.ContentType() != "application/json" && c.ContentType() != "" {
			c.Next()
			return
		}

		// Skip GET, HEAD, DELETE
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "DELETE" {
			c.Next()
			return
		}

		// Apply body size check
		if c.Request.ContentLength > maxBodyBytes {
			err := &ErrRequestTooLarge{
				Err: errors.New("request body exceeds maximum size limit"),
			}
			requestID := GetRequestID(c)
			slog.With("request_id", requestID).Warn("json_body_too_large",
				"max_bytes", maxBodyBytes,
				"received", c.Request.ContentLength,
			)
			c.Error(err)
			c.AbortWithStatusJSON(413, NewErrorPayload(413, err.Error(), requestID))
			return
		}

		c.Next()
	}
}
