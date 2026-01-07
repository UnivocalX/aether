package v1

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/gin-gonic/gin"
)

type AssetGetRequest struct {
	AssetUriParams
}

type AssetGetResponseData struct {
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

func HandleGetAsset(svc *data.Service, ctx *gin.Context) {
	var req AssetGetRequest

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&req.AssetUriParams); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to get asset",
			fmt.Errorf("%w: %w", dto.ErrInvalidUri, err),
		)
		return
	}

	asset, err := svc.GetAsset(ctx.Request.Context(), req.Checksum)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to get asset", err)
		return
	}

	// Success response
	response := dto.NewResponse(ctx, "got asset successfully")
	response.Data = NewAssetGetResponseData(asset)

	slog.InfoContext(ctx.Request.Context(), response.Message,
		"checksum", asset.Checksum,
	)

	response.OK(ctx)
}

func NewAssetGetResponseData(asset *registry.Asset) *AssetGetResponseData {
	return &AssetGetResponseData{
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