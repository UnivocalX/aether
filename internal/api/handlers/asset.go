package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/UnivocalX/aether/pkg/registry/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/UnivocalX/aether/internal/api/schemas"
)

// Custom errors for better error handling
var (
	ErrAssetAlreadyReady = errors.New("asset already exists and is ready")
)

type RegistryHandler struct {
	registry *registry.Engine
}

func NewRegistryHandler(reg *registry.Engine) *RegistryHandler {
	return &RegistryHandler{registry: reg}
}

// CreateAsset handles the HTTP request/response cycle
func (handler *RegistryHandler) CreateAsset(ctx *gin.Context) {
	var req schemas.CreateAssetRequest

	// Bind and validate request
	if err := ctx.ShouldBindUri(req); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid URI parameters", "error", err)
		BadRequest(ctx, "Invalid SHA256 in path parameter")
		return
	}

	if err := ctx.ShouldBindJSON(req); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid JSON payload", "error", err)
		BadRequest(ctx, "Invalid request payload")
		return
	}

	// Execute business logic
	response, err := handler.createAsset(ctx.Request.Context(), req)
	if err != nil {
		handler.handleCreateAssetError(ctx, err, req.SHA256)
		return
	}

	// Success response
	slog.InfoContext(ctx.Request.Context(), "Asset created successfully",
		"sha256", req.SHA256,
		"asset_id", response.AssetID,
	)
	Created(ctx, "Asset created successfully", response)
}

// createAsset contains the core business logic (no HTTP concerns)
func (handler *RegistryHandler) createAsset(ctx context.Context, req schemas.CreateAssetRequest) (*schemas.CreateAssetResponse, error) {
	// Check if asset exists
	existingAsset, err := handler.registry.GetAssetRecord(ctx, req.SHA256)
	if err != nil {
		return nil, fmt.Errorf("checking asset existence: %w", err)
	}

	// Handle existing asset
	if existingAsset != nil {
		return handler.handleExistingAsset(ctx, req, existingAsset)
	}

	// Create new asset
	return handler.handleNewAsset(ctx, req)
}

// handleExistingAsset processes re-upload attempts
func (handler *RegistryHandler) handleExistingAsset(ctx context.Context, req schemas.CreateAssetRequest, asset *models.Asset) (*schemas.CreateAssetResponse, error) {
	// Don't allow re-upload of ready assets
	if asset.State == models.StatusReady {
		slog.WarnContext(ctx, "Attempt to recreate ready asset",
			"sha256", req.SHA256,
			"asset_id", asset.ID,
		)
		return nil, ErrAssetAlreadyReady
	}

	// Associate tags and generate URL in transaction
	err := handler.registry.DB.Transaction(func(tx *gorm.DB) error {
		return handler.associateTags(ctx, asset.ID, req.Tags, tx)
	})
	if err != nil {
		return nil, err
	}

	return handler.buildAssetResponse(ctx, req.SHA256, asset.ID)
}

// handleNewAsset creates a new asset with tags
func (handler *RegistryHandler) handleNewAsset(ctx context.Context, req schemas.CreateAssetRequest) (*schemas.CreateAssetResponse, error) {
	// Generate presigned URL FIRST (before transaction)
	url, err := handler.registry.PutURL(ctx, req.SHA256)
	if err != nil {
		return nil, fmt.Errorf("generating presigned URL: %w", err)
	}

	var assetID uint

	// Create asset and associate tags in transaction
	err = handler.registry.DB.Transaction(func(tx *gorm.DB) error {
		// Create asset record
		asset, err := handler.registry.CreateAssetRecord(ctx, req.SHA256, req.Display)
		if err != nil {
			return fmt.Errorf("creating asset record: %w", err)
		}
		assetID = asset.ID

		// Associate tags
		if err := handler.associateTags(ctx, assetID, req.Tags, tx); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &schemas.CreateAssetResponse{
		SHA256:       req.SHA256,
		PresignedURL: url,
		Expiry:       handler.registry.Config.Storage.TTL.String(),
		AssetID:      assetID,
	}, nil
}

// associateTags links tags to an asset (pure data operation)
func (handler *RegistryHandler) associateTags(ctx context.Context, assetID uint, tagIDs []uint, tx *gorm.DB) error {
	if len(tagIDs) == 0 {
		return nil
	}

	slog.InfoContext(ctx, "Associating tags with asset",
		"asset_id", assetID,
		"tag_count", len(tagIDs),
	)

	for _, tagID := range tagIDs {
		if err := handler.registry.AssociateTagWithAsset(assetID, tagID); err != nil {
			return fmt.Errorf("associating tag %d: %w", tagID, err)
		}
	}

	return nil
}

// buildAssetResponse generates the final response with presigned URL
func (handler *RegistryHandler) buildAssetResponse(ctx context.Context, sha256 string, assetID uint) (*schemas.CreateAssetResponse, error) {
	url, err := handler.registry.PutURL(ctx, sha256)
	if err != nil {
		return nil, fmt.Errorf("generating presigned URL: %w", err)
	}

	return &schemas.CreateAssetResponse{
		SHA256:       sha256,
		PresignedURL: url,
		Expiry:       handler.registry.Config.Storage.TTL.String(),
		AssetID:      assetID,
	}, nil
}

// handleCreateAssetError maps business errors to HTTP responses
func (handler *RegistryHandler) handleCreateAssetError(ctx *gin.Context, err error, sha256 string) {
	switch {
	case errors.Is(err, ErrAssetAlreadyReady):
		Conflict(ctx, "Asset already exists and is ready")

	case errors.Is(err, registry.ErrTagNotFound):
		NotFound(ctx, "One or more tags not found")

	case errors.Is(err, models.ErrValidation):
		BadRequest(ctx, err.Error())

	default:
		slog.ErrorContext(ctx.Request.Context(), "Failed to create asset",
			"error", err,
			"sha256", sha256,
		)
		InternalError(ctx, "Failed to create asset")
	}
}
