package dto

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/UnivocalX/aether/internal/registry"
	dataService "github.com/UnivocalX/aether/pkg/web/services/data"
	"github.com/gin-gonic/gin"
)

// Standard API response
type Response struct {
	Msg  string            `json:"message"`
	Meta *ResponseMetadata `json:"meta,omitempty"`
}

// Error response wrapper
type ErrorDetails struct {
	Msg string          `json:"message"`
	Details *map[string]any `json:"details,omitempty"`
}

type ErrorResponse struct {
	Msg string            `json:"message"`
	Err   *ErrorDetails     `json:"error"`
	Meta    *ResponseMetadata `json:"meta,omitempty"`
}

// Metadata attached to every response
type ResponseMetadata struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id,omitempty"`
	Path      string `json:"path,omitempty"`
}

func NewResponse(c *gin.Context, msg string) *Response {
	return &Response{
		Msg:  msg,
		Meta: buildMeta(c),
	}
}

func NewErrorResponse(c *gin.Context, msg string, err error) *ErrorResponse {
	return &ErrorResponse{
		Msg: msg,
		Err: &ErrorDetails{
			Msg: err.Error(),
		},
		Meta: buildMeta(c),
	}
}

func buildMeta(c *gin.Context) *ResponseMetadata {
	return &ResponseMetadata{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: c.GetString("requestId"),
		Path:      c.Request.URL.Path,
	}
}

func (r *Response) OK(c *gin.Context)                   { c.JSON(http.StatusOK, r) }
func (r *Response) Created(c *gin.Context)              { c.JSON(http.StatusCreated, r) }
func (r *Response) NoContent(c *gin.Context)            { c.Status(http.StatusNoContent) }
func (r *Response) MultiStatus(c *gin.Context)          { c.JSON(http.StatusMultiStatus, r) }
func (r *ErrorResponse) BadRequest(c *gin.Context)      { c.JSON(http.StatusBadRequest, r) }
func (r *ErrorResponse) Unauthorized(c *gin.Context)    { c.JSON(http.StatusUnauthorized, r) }
func (r *ErrorResponse) Forbidden(c *gin.Context)       { c.JSON(http.StatusForbidden, r) }
func (r *ErrorResponse) NotFound(c *gin.Context)        { c.JSON(http.StatusNotFound, r) }
func (r *ErrorResponse) ContentTooLarge(c *gin.Context) { c.JSON(http.StatusRequestEntityTooLarge, r) }
func (r *ErrorResponse) Conflict(c *gin.Context)        { c.JSON(http.StatusConflict, r) }
func (r *ErrorResponse) InternalError(c *gin.Context)   { c.JSON(http.StatusInternalServerError, r) }

func HandleErrorResponse(ctx *gin.Context, msg string, err error) {
	response := NewErrorResponse(ctx, msg, err)

	slog.ErrorContext(
		ctx.Request.Context(),
		err.Error(),
	)

	var assetsExistError dataService.AssetsExistsError
	var maxBytesError *http.MaxBytesError

	switch {
	case errors.As(err, &maxBytesError):
		response.Err.Msg = maxBytesError.Error()
		response.Err.Details = &map[string]any{
			"max_bytes": maxBytesError.Limit,
		}
		response.ContentTooLarge(ctx)

	case errors.Is(err, ErrInvalidUri),
		errors.Is(err, ErrInvalidPayload),
		errors.Is(err, registry.ErrValidation):
		response.BadRequest(ctx)

	case errors.Is(err, dataService.ErrAssetNotFound),
		errors.Is(err, dataService.ErrTagNotFound):
		response.NotFound(ctx)

	case errors.As(err, &assetsExistError):
		response.Err.Details = &map[string]any{
			"checksums": assetsExistError.Checksums,
		}
		response.Conflict(ctx)

	case errors.Is(err, dataService.ErrAssetAlreadyExists),
		errors.Is(err, dataService.ErrTagAlreadyExists),
		errors.Is(err, dataService.ErrDatasetAlreadyExists):
		response.Conflict(ctx)

	default:
		response.InternalError(ctx)
	}
}
