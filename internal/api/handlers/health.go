package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck returns a gin.HandlerFunc.
func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
