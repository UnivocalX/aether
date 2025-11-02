package handlers

import (
	"context"
	"log/slog"

	"github.com/UnivocalX/aether/internal/api/v1/schemas"
	"github.com/UnivocalX/aether/pkg/registry/models"
	"github.com/gin-gonic/gin"
)

func (handler *RegistryHandler) ListAssets(ctx *gin.Context) {
	var req schemas.ListAssetsRequest

	// Bind and validate request
	if err := ctx.ShouldBindQuery(&req); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid query parameter", "error", err)
		BadRequest(ctx, "Invalid query parameters")
		return
	}

	// Execute business logic
	response, err := handler.listAssets(ctx.Request.Context(), &req.SearchAssetsOptions)
	if err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Failed to list assets", "error", err)
		InternalError(ctx, "Failed to list assets")
		return
	}

	// Success response
	slog.InfoContext(ctx.Request.Context(), "Assets listed successfully", "assetCount", len(response.Assets))
	OK(ctx, "Assets listed successfully", response)
}

// listAssets contains the core business logic
func (handler *RegistryHandler) listAssets(ctx context.Context, opts *models.SearchAssetsOptions) (*schemas.ListAssetsResponse, error) {
	assets, nextCursor, hasMore, err := handler.registry.ListAssets(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &schemas.ListAssetsResponse{
		PaginatedResponse: schemas.PaginatedResponse{
			NextCursor: nextCursor,
			HasMore:    hasMore,
		},
		Assets: assets,
	}, nil
}