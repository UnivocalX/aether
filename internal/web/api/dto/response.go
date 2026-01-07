package dto

import (
	"errors"
	"log/slog"
	"time"

	dataService "github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/gin-gonic/gin"
)

// Standard API Response - Minimalist version
type Response struct {
	Message string
	Data    any
	Error   error
	Meta    *ResponseMetadata `json:"meta,omitempty"`
}

type ResponseMetadata struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id,omitempty"`
}

func NewResponse(c *gin.Context, msg string) *Response {
	return &Response{
		Message: msg,
		Meta:    buildMeta(c),
	}
}

// Helper to build meta information
func buildMeta(c *gin.Context) *ResponseMetadata {
	return &ResponseMetadata{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: c.GetString("requestId"),
	}
}

// Success responses
func (response *Response) OK(c *gin.Context) {
	c.JSON(200, response)
}

func (response *Response) Created(c *gin.Context) {
	c.JSON(201, response)
}

func (response *Response) NoContent(c *gin.Context) {
	c.JSON(204, response)
}

func (response *Response) MultiStatus(c *gin.Context) {
	c.JSON(207, response)
}

// Error responses - Use HTTP status codes instead of success field
func (response *Response) BadRequest(c *gin.Context) {
	c.JSON(400, response)
}

func (response *Response) Unauthorized(c *gin.Context) {
	c.JSON(401, response)
}

func (response *Response) Forbidden(c *gin.Context) {
	c.JSON(403, response)
}

func (response *Response) NotFound(c *gin.Context) {
	c.JSON(404, response)
}

func (response *Response) InternalError(c *gin.Context) {
	c.JSON(500, response)
}

func (response *Response) Conflict(c *gin.Context) {
	c.JSON(409, response)
}

func HandleErrorResponse(ctx *gin.Context, msg string, err error, data ...any) {
	response := NewResponse(ctx, msg)
	response.Error = err
	response.Data = data

	slog.ErrorContext(ctx.Request.Context(), response.Message, "error", response.Error)

	switch {
	case errors.Is(response.Error, ErrInvalidUri):
		response.BadRequest(ctx)

	case errors.Is(response.Error, ErrInvalidPayload):
		response.BadRequest(ctx)

	case errors.Is(response.Error, registry.ErrValidation):
		response.BadRequest(ctx)

	case errors.Is(response.Error, dataService.ErrAssetNotFound):
		response.NotFound(ctx)

	case errors.Is(response.Error, dataService.ErrTagNotFound):
		response.NotFound(ctx)

	case errors.Is(err, dataService.ErrAssetIsReady):
		response.Conflict(ctx)

	case errors.Is(err, dataService.ErrAssetAlreadyExists):
		response.Conflict(ctx)

	case errors.Is(err, dataService.ErrTagAlreadyExists):
		response.Conflict(ctx)

	case errors.As(err, &dataService.ErrAssetsExists{}):
		response.Conflict(ctx)

	case errors.Is(err, dataService.ErrDatasetAlreadyExists):
		response.Conflict(ctx)

	default:
		response.InternalError(ctx)
	}
}
