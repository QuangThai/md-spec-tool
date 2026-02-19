package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

const RequestIDHeader = "X-Request-ID"

type contextKey struct{}

var RequestIDContextKey = contextKey{}

// RequestID generates and injects a unique request ID for traceability
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Writer.Header().Set(RequestIDHeader, requestID)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), RequestIDContextKey, requestID))

		startedAt := time.Now()
		logger := slog.With("request_id", requestID)
		logger.Info("request started",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)

		c.Next()

		logger.Info("request completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(startedAt).Milliseconds(),
		)
	}
}

// generateRequestID creates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%x", time.Now().UnixNano(), time.Now().Unix())
}
