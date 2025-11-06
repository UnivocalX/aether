package handlers

import (
	"time"

	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/UnivocalX/aether/internal/api/v1/schemas"
	"github.com/gin-gonic/gin"
)

// Success responses
func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(200, schemas.Response{
		Message: message,
		Data:    data,
		Meta:    buildMeta(c),
	})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(201, schemas.Response{
		Message: message,
		Data:    data,
		Meta:    buildMeta(c),
	})
}

func MultiStatus(c *gin.Context, message string, data interface{}) {
	c.JSON(207, schemas.Response{
		Message: message,
		Data:    data,
		Meta:    buildMeta(c),
	})
}

// Error responses - Use HTTP status codes instead of success field
func BadRequest(c *gin.Context, message string) {
	c.JSON(400, schemas.Response{
		Message: message,
		Meta:    buildMeta(c),
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(401, schemas.Response{
		Message: message,
		Meta:    buildMeta(c),
	})
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(403, schemas.Response{
		Message: message,
		Meta:    buildMeta(c),
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(404, schemas.Response{
		Message: message,
		Meta:    buildMeta(c),
	})
}

func InternalError(c *gin.Context, message string) {
	c.JSON(500, schemas.Response{
		Message: message,
		Meta:    buildMeta(c),
	})
}

func Conflict(c *gin.Context, message string) {
    c.JSON(409, schemas.Response{
        Message: message,
        Meta:    buildMeta(c),
    })
}

// Helper to build meta information
func buildMeta(c *gin.Context) *schemas.ResponseMetadata {
	return &schemas.ResponseMetadata{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: c.GetString("request_id"),
	}
}

type RegistryHandler struct {
	registry *registry.Engine
}

func NewRegistryHandler(reg *registry.Engine) *RegistryHandler {
	return &RegistryHandler{registry: reg}
}
