package handlers

import (
	"context"
	"github.com/UnivocalX/aether/pkg/registry"
	"log/slog"

	"github.com/UnivocalX/aether/internal/api/v1/schemas"
	"github.com/gin-gonic/gin"
)

// GetTag handles the HTTP request/response cycle
func (handler *RegistryHandler) GetTag(ctx *gin.Context) {
	var req schemas.GetTagRequest

	// Bind and validate request
	if err := ctx.ShouldBindUri(req); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid URI parameter", "error", err)
		BadRequest(ctx, "Invalid name in path parameter")
		return
	}

	// Execute business logic
	response, err := handler.getTag(ctx.Request.Context(), req)
	if err != nil {
		handler.handleGetTagError(ctx, err, req.Name)
		return
	}

	// Success response
	slog.InfoContext(ctx.Request.Context(), "Tag retrieved successfully",
		"tagName", response.Tag.Name,
		"tagID", response.Tag.ID,
	)
	OK(ctx, "Tag retrieved successfully", response)
}

// getTag contains the core business logic (no HTTP concerns)
func (handler *RegistryHandler) getTag(ctx context.Context, req schemas.GetTagRequest) (*schemas.GetTagResponse, error) {

	// Retrieve tag
	tag, err := handler.registry.GetTagRecord(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	return &schemas.GetTagResponse{
		Tag: tag,
	}, nil
}

func (handler *RegistryHandler) handleGetTagError(ctx *gin.Context, err error, name string) {
	if err == registry.ErrTagNotFound {
		slog.InfoContext(ctx.Request.Context(), "Tag not found", "tagName", name)
		NotFound(ctx, "Tag not found")
		return
	}

	slog.ErrorContext(ctx.Request.Context(), "Failed to get tag",
		"tagName", name,
		"error", err,
	)
	InternalError(ctx, "Failed to get tag")
}
