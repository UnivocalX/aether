package handlers

import (
	"time"

	"github.com/UnivocalX/aether/internal/api/models"
	"github.com/gin-gonic/gin"
)

// Success responses
func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(200, models.Response{
		Message: message,
		Data:    data,
		Meta:    buildMeta(c),
	})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(201, models.Response{
		Message: message,
		Data:    data,
		Meta:    buildMeta(c),
	})
}

// Error responses - Use HTTP status codes instead of success field
func BadRequest(c *gin.Context, message string) {
	c.JSON(400, models.Response{
		Message: message,
		Meta:    buildMeta(c),
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(404, models.Response{
		Message: message,
		Meta:    buildMeta(c),
	})
}

func InternalError(c *gin.Context, message string) {
	c.JSON(500, models.Response{
		Message: message,
		Meta:    buildMeta(c),
	})
}

// Helper to build meta information
func buildMeta(c *gin.Context) *models.Metadata {
	return &models.Metadata{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: c.GetString("request_id"),
	}
}
