package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/config"
)

func CORS(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Check if origin is allowed
		isAllowed := false
		for _, allowedOrigin := range cfg.CORSOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}
		
		// Only set Access-Control-Allow-Origin if origin is in allowlist (deny by default)
		if isAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		// Always include Vary: Origin to prevent caching issues with different origins
		c.Writer.Header().Set("Vary", "Origin")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-OpenAI-API-Key, X-Session-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
