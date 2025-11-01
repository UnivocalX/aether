package handlers

import (
	"context"
	"github.com/UnivocalX/aether/pkg/registry"
	"log/slog"

	"github.com/UnivocalX/aether/internal/api/v1/schemas"
	"github.com/gin-gonic/gin"
)

// GetAsset handles the HTTP request/response cycle
func (handler *RegistryHandler) GetAsset(ctx *gin.Context) {
	var req schemas.GetAssetRequest

	// Bind and validate request
	if err := ctx.ShouldBindUri(&req); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid URI parameter", "error", err)
		BadRequest(ctx, "Invalid name in path parameter")
		return
	}

	// Execute business logic
	response, err := handler.getAsset(ctx.Request.Context(), req)
	if err != nil {
		handler.handleGetAssetError(ctx, err, req.SHA256)
		return
	}

	// Success response
	slog.InfoContext(ctx.Request.Context(), "Asset retrieved successfully",
		"assetSHA256", response.Asset.Checksum,
		"assetID", response.Asset.ID,
	)
	OK(ctx, "Asset retrieved successfully", response)
}

// getAsset contains the core business logic (no HTTP concerns)
func (handler *RegistryHandler) getAsset(ctx context.Context, req schemas.GetAssetRequest) (*schemas.GetAssetResponse, error) {

	// Retrieve asset
	asset, err := handler.registry.GetAssetRecord(ctx, req.SHA256)
	if err != nil {
		return nil, err
	}

	return &schemas.GetAssetResponse{
		Asset: asset,
	}, nil
}

func (handler *RegistryHandler) handleGetAssetError(ctx *gin.Context, err error, name string) {
	if err == registry.ErrAssetNotFound {
		slog.InfoContext(ctx.Request.Context(), "Asset not found", "assetName", name)
		NotFound(ctx, "Asset not found")
		return
	}

	slog.ErrorContext(ctx.Request.Context(), "Failed to get asset", "error", err, "assetName", name)
	InternalError(ctx, "Failed to get asset")
}