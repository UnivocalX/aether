package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/UnivocalX/aether/internal/api/dto"
	"github.com/UnivocalX/aether/internal/api/services/data"
	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/gin-gonic/gin"
)

type AssetGetRequest struct {
	AssetUriParams
}

type AssetGetResponse struct {
	ID        uint            `json:"id"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
	Checksum  string          `json:"checksum"`
	Display   string          `json:"display"`
	Extra     json.RawMessage `json:"extra,omitempty"`
	MimeType  string          `json:"mime_type"`
	SizeBytes int64           `json:"size_bytes"`
	State     registry.Status `json:"state"`
}


func NewAssetGetResponse(asset *registry.Asset) *AssetGetResponse {
    return &AssetGetResponse{
        ID:        asset.ID,
        CreatedAt: asset.CreatedAt.Format(time.RFC3339),
        UpdatedAt: asset.UpdatedAt.Format(time.RFC3339),
        Checksum:  asset.Checksum,
        Display:   asset.Display,
        Extra:     json.RawMessage(asset.Extra),
        MimeType:  asset.MimeType,
        SizeBytes: asset.SizeBytes,
        State:     asset.State,
    }
}

func HandleGetAsset(svc *data.Service, ctx *gin.Context) {
	var req AssetGetRequest

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&req.AssetUriParams); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid URI parameters", "error", err.Error())
		dto.BadRequest(ctx, "Invalid SHA256 in URI")
		return
	}

	asset, err := svc.GetAsset(ctx.Request.Context(), req.SHA256)
	if err != nil {
		handleGetAssetError(ctx, err, req.SHA256)
		return
	}

	// Success response
	slog.InfoContext(ctx.Request.Context(), "got asset successfully",
		"assetSha256", asset.Checksum,
	)

	response := NewAssetGetResponse(asset)
	dto.OK(ctx, "got asset successfully", response)
}

func handleGetAssetError(ctx *gin.Context, err error, sha256 string) {
	switch {
	case errors.Is(err, registry.ErrValidation):
		dto.BadRequest(ctx, err.Error())

	case errors.Is(err, data.ErrAssetNotFound):
		dto.NotFound(ctx, fmt.Sprintf("asset %s does not exist", sha256))

	default:
		slog.ErrorContext(ctx.Request.Context(), "Failed to get asset",
			"error", err.Error(),
			"sha256", sha256,
		)
		dto.InternalError(ctx, "Failed to get asset")
	}
}
