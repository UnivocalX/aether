package handlers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/pkg/registry/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/UnivocalX/aether/internal/api/v1/schemas"
)

// BatchCreateAsset handles the HTTP request/response cycle for batch creation
func (handler *RegistryHandler) BatchCreateAsset(ctx *gin.Context) {
	var req schemas.BatchCreateAssetRequest

	// Bind and validate request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid JSON payload for batch create", "error", err)
		BadRequest(ctx, "Invalid request payload")
		return
	}

	// Execute business logic
	responses, err := handler.createAssets(ctx.Request.Context(), req)
	if err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Failed to process batch create",
			"error", err, "asset_count", len(req.Assets))
		InternalError(ctx, "Failed to process batch request")
		return
	}

	// Determine overall status
	statusCode, message := determineBatchStatus(responses)

	slog.InfoContext(ctx.Request.Context(), "Batch asset creation completed",
		"totalAssets", len(req.Assets),
		"successful", countSuccessful(responses),
		"failed", countFailed(responses),
	)

	// Use your existing response helper with the appropriate status code
	ctx.JSON(statusCode, gin.H{
		"statusCode": statusCode,
		"msg":        message,
		"data":       responses,
	})
}

// createAssets contains the core business logic for batch creation
func (handler *RegistryHandler) createAssets(ctx context.Context, req schemas.BatchCreateAssetRequest) ([]schemas.BatchCreateAssetResponse, error) {
	responses := make([]schemas.BatchCreateAssetResponse, len(req.Assets))

	for i, assetReq := range req.Assets {
		response, err := handler.processSingleAssetInBatch(ctx, assetReq, req.Tags)
		if err != nil {
			responses[i] = schemas.BatchCreateAssetResponse{
				CreateAssetResponse: schemas.CreateAssetResponse{
					SHA256: assetReq.SHA256,
					// AssetID, PresignedURL, and Expiry will be zero values
				},
				Error: err.Error(),
			}
		} else {
			responses[i] = *response
		}
	}

	return responses, nil
}

// processSingleAssetInBatch processes a single asset within the batch context
func (handler *RegistryHandler) processSingleAssetInBatch(ctx context.Context, assetReq schemas.CreateAssetRequest, globalTags []uint) (*schemas.BatchCreateAssetResponse, error) {
	// Merge global tags with per-asset tags
	allTags := mergeTags(globalTags, assetReq.Tags)

	// Check if asset exists
	existingAsset, err := handler.registry.GetAssetRecord(ctx, assetReq.SHA256)
	if err != nil {
		return nil, fmt.Errorf("checking asset existence: %w", err)
	}

	// Handle existing asset
	if existingAsset != nil {
		return handler.handleExistingAssetInBatch(ctx, assetReq, allTags, existingAsset)
	}

	// Create new asset
	return handler.handleNewAssetInBatch(ctx, assetReq, allTags)
}

// handleExistingAssetInBatch processes re-upload attempts in batch context
func (handler *RegistryHandler) handleExistingAssetInBatch(ctx context.Context, assetReq schemas.CreateAssetRequest, allTags []uint, asset *models.Asset) (*schemas.BatchCreateAssetResponse, error) {
	// Don't allow re-upload of ready assets
	if asset.State == models.StatusReady {
		slog.WarnContext(ctx, "Attempt to recreate ready asset in batch",
			"sha256", assetReq.SHA256,
			"asset_id", asset.ID,
		)
		return nil, ErrAssetAlreadyReady
	}

	// Associate tags in transaction
	err := handler.registry.DB.Transaction(func(tx *gorm.DB) error {
		return handler.associateTags(ctx, asset.ID, allTags, tx)
	})
	if err != nil {
		return nil, err
	}

	return handler.buildBatchAssetResponse(ctx, assetReq.SHA256, asset.ID)
}

// handleNewAssetInBatch creates a new asset with tags in batch context
func (handler *RegistryHandler) handleNewAssetInBatch(ctx context.Context, assetReq schemas.CreateAssetRequest, allTags []uint) (*schemas.BatchCreateAssetResponse, error) {
	// Generate presigned URL FIRST (before transaction)
	url, err := handler.registry.PutURL(ctx, assetReq.SHA256)
	if err != nil {
		return nil, fmt.Errorf("generating presigned URL: %w", err)
	}

	var assetID uint

	// Create asset and associate tags in transaction
	err = handler.registry.DB.Transaction(func(tx *gorm.DB) error {
		// Create asset record
		asset, err := handler.registry.CreateAssetRecord(ctx, assetReq.SHA256, assetReq.Display, assetReq.Extra)
		if err != nil {
			return fmt.Errorf("creating asset record: %w", err)
		}
		assetID = asset.ID

		// Associate tags
		if err := handler.associateTags(ctx, assetID, allTags, tx); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &schemas.BatchCreateAssetResponse{
		CreateAssetResponse: schemas.CreateAssetResponse{
			SHA256:       assetReq.SHA256,
			PresignedURL: url.Value(),
			Expiry:       handler.registry.Config.Storage.TTL.String(),
			AssetID:      assetID,
		},
	}, nil
}

// buildBatchAssetResponse generates the final batch response with presigned URL
func (handler *RegistryHandler) buildBatchAssetResponse(ctx context.Context, sha256 string, assetID uint) (*schemas.BatchCreateAssetResponse, error) {
	url, err := handler.registry.PutURL(ctx, sha256)
	if err != nil {
		return nil, fmt.Errorf("generating presigned URL: %w", err)
	}

	return &schemas.BatchCreateAssetResponse{
		CreateAssetResponse: schemas.CreateAssetResponse{
			SHA256:       sha256,
			PresignedURL: url.Value(),
			Expiry:       handler.registry.Config.Storage.TTL.String(),
			AssetID:      assetID,
		},
	}, nil
}

// Helper functions
func mergeTags(globalTags, assetTags []uint) []uint {
	tagSet := make(map[uint]bool)

	for _, tag := range globalTags {
		tagSet[tag] = true
	}

	for _, tag := range assetTags {
		tagSet[tag] = true
	}

	result := make([]uint, 0, len(tagSet))
	for tag := range tagSet {
		result = append(result, tag)
	}

	return result
}

func determineBatchStatus(responses []schemas.BatchCreateAssetResponse) (int, string) {
	total := len(responses)
	successful := 0
	failed := 0

	for _, response := range responses {
		if response.Error == "" {
			successful++
		} else {
			failed++
		}
	}

	if failed == 0 {
		return 201, fmt.Sprintf("All %d assets created successfully", total)
	} else if successful == 0 {
		return 400, fmt.Sprintf("All %d assets failed to create", total)
	} else {
		return 207, fmt.Sprintf("Created %d assets, %d failed", successful, failed)
	}
}

func countSuccessful(responses []schemas.BatchCreateAssetResponse) int {
	count := 0
	for _, response := range responses {
		if response.Error == "" {
			count++
		}
	}
	return count
}

func countFailed(responses []schemas.BatchCreateAssetResponse) int {
	count := 0
	for _, response := range responses {
		if response.Error != "" {
			count++
		}
	}
	return count
}
