// api/v1/asset_handler.go
package v1

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/UnivocalX/aether/internal/api/dto"
	"github.com/UnivocalX/aether/internal/api/services/data"
	"github.com/UnivocalX/aether/pkg/registry"

	"github.com/gin-gonic/gin"
)

type AssetPostUriParams struct {
	SHA256 string `uri:"sha256" binding:"required,len=64,hexadecimal"`
}

type AssetPostPayload struct {
	Display string                 `json:"display" binding:"omitempty,max=500"`
	Tags    []uint                 `json:"tags" binding:"omitempty,dive,gt=0"`
	Extra   map[string]interface{} `json:"extra" binding:"omitempty"`
}

type AssetPostRequest struct {
	AssetPostUriParams
	AssetPostPayload
}

type AssetPostResponse struct {
	ID        uint       `json:"id"`
	Checksum  string     `json:"checksum"`
	State     string     `json:"state"`
	UploadURL string     `json:"upload_url,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// HandleCreateAsset handles the HTTP Asset post request/response cycle
func HandleCreateAsset(svc *data.Service, ctx *gin.Context) {
	var req AssetPostRequest

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&req.AssetPostUriParams); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid URI parameters", "error", err.Error())
		dto.BadRequest(ctx, "Invalid SHA256 in URI")
		return
	}

	// Bind JSON payload
	if err := ctx.ShouldBindJSON(&req.AssetPostPayload); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid JSON payload", "error", err.Error())
		dto.BadRequest(ctx, err.Error())
		return
	}

	// Execute business logic
	result := svc.CreateAsset(ctx.Request.Context(), data.CreateAssetParams{
		SHA256:  req.SHA256,
		Display: req.Display,
		Tags:    req.Tags,
		Extra:   req.Extra,
	})

	// Handle errors
	if result.Err != nil {		
		handleCreateAssetError(ctx, result.Err, req.SHA256)
		return
	}

	// Build API response for successful creation
	response := buildAssetPostResponse(result)

	// Success response
	slog.InfoContext(ctx.Request.Context(), "asset created successfully",
		"sha256", result.Asset.Checksum,
		"assetId", result.Asset.ID,
	)

	dto.Created(ctx, "Successfully created asset", response)
}

// mapToAssetPostResponse converts service result to API response
func buildAssetPostResponse(result *data.CreateAssetResult) *AssetPostResponse {
	response := &AssetPostResponse{
		ID:       result.Asset.ID,
		Checksum: result.Asset.Checksum,
		State:    string(result.Asset.State),
	}

	if result.UploadURL != nil {
		response.UploadURL = result.UploadURL.URL.Value()
		response.ExpiresAt = &result.UploadURL.ExpiresAt
	}

	return response
}

// handleCreateAssetError maps business errors to HTTP responses
func handleCreateAssetError(ctx *gin.Context, err error, sha256 string) {
	switch {
	case errors.Is(err, registry.ErrValidation):
		dto.BadRequest(ctx, err.Error())

	case errors.Is(err, data.ErrTagNotFound):
		dto.NotFound(ctx, "One or more tags not found")

	case errors.Is(err, data.ErrAssetAlreadyExists):
		dto.Conflict(ctx, fmt.Sprintf("Asset %s already exists.", sha256))

	default:
		slog.ErrorContext(ctx.Request.Context(), "Failed to create asset",
			"error", err.Error(),
			"sha256", sha256,
		)
		dto.InternalError(ctx, "Failed to create asset")
	}
}
