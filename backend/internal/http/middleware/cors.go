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
		
		if isAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(cfg.CORSOrigins) > 0 {
			// Fallback to first allowed origin if no match
			c.Writer.Header().Set("Access-Control-Allow-Origin", cfg.CORSOrigins[0])
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
