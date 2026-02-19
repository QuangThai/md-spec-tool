package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// QuotaUsage represents quota usage details
type QuotaUsage struct {
	TokensUsedToday  int64
	DailyConversions int
}

type QuotaChecker interface {
	ValidateQuotaAvailable(ctx context.Context, sessionID string) (bool, error)
	GetUsageDetails(ctx context.Context, sessionID string) (*QuotaUsage, error)
}

// QuotaMiddleware checks quota before allowing request to proceed
// Should be applied to expensive endpoints (preview, convert, ai/suggest)
// Injects X-Quota-Remaining header into response
func QuotaMiddleware(checker QuotaChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.GetString("session_id")
		if sessionID == "" {
			// No session ID, allow through (will be set by session middleware)
			c.Next()
			return
		}

		available, err := checker.ValidateQuotaAvailable(c.Request.Context(), sessionID)
		if err != nil {
			slog.Warn("quota check failed",
				"session_id", sessionID,
				"error", err,
			)
			// On error, allow through (fail open)
			c.Next()
			return
		}

		// Inject remaining quota into response header
		// This allows client to display quota status
		if usage, err := checker.GetUsageDetails(c.Request.Context(), sessionID); err == nil && usage != nil {
			remaining := 100000 - usage.TokensUsedToday
			if remaining < 0 {
				remaining = 0
			}
			c.Header("X-Quota-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("X-Quota-Used", fmt.Sprintf("%d", usage.TokensUsedToday))
			c.Header("X-Quota-Daily-Conversions", fmt.Sprintf("%d", usage.DailyConversions))
		}

		if !available {
			slog.Info("quota exceeded",
				"session_id", sessionID,
			)

			c.AbortWithStatusJSON(http.StatusTooManyRequests, ErrorPayload{
				Error:     "daily quota exceeded",
				Code:      "QUOTA_EXCEEDED",
				RequestID: GetRequestID(c),
			})
			return
		}

		c.Next()
	}
}
