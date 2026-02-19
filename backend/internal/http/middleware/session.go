package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SessionID middleware ensures every request has a session_id
// Priority: Header > Query > Generate new
func SessionID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get from X-Session-ID header
		sessionID := c.GetHeader("X-Session-ID")
		if sessionID != "" {
			c.Set("session_id", sessionID)
			c.Next()
			return
		}

		// Try to get from query parameter
		sessionID = c.Query("session_id")
		if sessionID != "" {
			c.Set("session_id", sessionID)
			c.Next()
			return
		}

		// Generate new session ID
		sessionID = uuid.New().String()
		c.Set("session_id", sessionID)

		// Return it in response headers so client can use it
		c.Header("X-Session-ID", sessionID)

		c.Next()
	}
}
