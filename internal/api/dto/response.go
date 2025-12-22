package dto

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Standard API Response - Minimalist version
type Response struct {
	Message string            `json:"message"`
	Data    interface{}       `json:"data,omitempty"`
	Meta    *ResponseMetadata `json:"meta,omitempty"`
}

type ResponseMetadata struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id,omitempty"`
}

// Success responses
func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(200, Response{
		Message: message,
		Data:    data,
		Meta:    buildMeta(c),
	})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(201, Response{
		Message: message,
		Data:    data,
		Meta:    buildMeta(c),
	})
}

func NoContent(c *gin.Context, message string) {
	c.JSON(204, Response{
		Message: message,
		Meta:    buildMeta(c),
	})
}

func MultiStatus(c *gin.Context, message string, data interface{}) {
	c.JSON(207, Response{
		Message: message,
		Data:    data,
		Meta:    buildMeta(c),
	})
}

// Error responses - Use HTTP status codes instead of success field
func BadRequest(c *gin.Context, message string) {
	c.JSON(400, Response{
		Message: message,
		Meta:    buildMeta(c),
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(401, Response{
		Message: message,
		Meta:    buildMeta(c),
	})
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(403, Response{
		Message: message,
		Meta:    buildMeta(c),
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(404, Response{
		Message: message,
		Meta:    buildMeta(c),
	})
}

func InternalError(c *gin.Context, message string) {
	c.JSON(500, Response{
		Message: message,
		Meta:    buildMeta(c),
	})
}

func Conflict(c *gin.Context, message string) {
	c.JSON(409, Response{
		Message: message,
		Meta:    buildMeta(c),
	})
}

// Helper to build meta information
func buildMeta(c *gin.Context) *ResponseMetadata {
	return &ResponseMetadata{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: c.GetString("requestId"),
	}
}
