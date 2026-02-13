package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/http/middleware"
)

func HealthHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "md-spec-tool",
	})
}

// MetricsHandler returns basic request metrics (count, avg latency) for observability.
func MetricsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, middleware.GetMetrics())
}
