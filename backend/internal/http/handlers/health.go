package handlers

import (
	"github.com/gin-gonic/gin"
)

func HealthHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "md-spec-tool",
	})
}
