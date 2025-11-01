package handlers

import (
	"context"
	"log/slog"
	
	"github.com/UnivocalX/aether/internal/api/v1/schemas"
	"github.com/gin-gonic/gin"
)

func (handler *RegistryHandler) ListTags(ctx *gin.Context) {
	var req schemas.ListTagsRequest

	// Bind and validate request
	if err := ctx.ShouldBindQuery(&req); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid query parameter", "error", err)
		BadRequest(ctx, "Invalid query parameters")
		return
	}

	// Execute business logic
	response, err := handler.listTags(ctx.Request.Context(), &req)
	if err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Failed to list tags", "error", err)
		InternalError(ctx, "Failed to list tags")
		return
	}

	// Success response
	slog.InfoContext(ctx.Request.Context(), "Tags listed successfully", "tagCount", len(response.Tags))
	OK(ctx, "Tags listed successfully", response)
}

// listTags contains the core business logic (no HTTP concerns)
func (handler *RegistryHandler) listTags(ctx context.Context, req *schemas.ListTagsRequest) (*schemas.ListTagsResponse, error) {
	if req.Limit == 0 {
		req.Limit = 50
	}

	tags, nextCursor, hasMore, err := handler.registry.ListTags(ctx, req.Cursor, req.Limit)
	if err != nil {
		return nil, err
	}

	return &schemas.ListTagsResponse{
		Tags:       tags,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}