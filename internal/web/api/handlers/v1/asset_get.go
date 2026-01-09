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

type DetailedAssetResponseData struct {
	ID        uint            `json:"id"`
	Checksum  string          `json:"checksum"`
	Display   string          `json:"display"`
	Extra     json.RawMessage `json:"extra,omitempty"`
	MimeType  string          `json:"mime_type"`
	SizeBytes int64           `json:"size_bytes"`
	State     registry.Status `json:"state"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

func GetAssetHandler(svc *data.Service, ctx *gin.Context) {
	var uri dto.AssetUri

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&uri); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to get asset",
			fmt.Errorf("%w, %w", dto.ErrInvalidUri, err),
		)
		return
	}

	asset, err := svc.GetAsset(ctx.Request.Context(), uri.AssetChecksum)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to get asset", err)
		return
	}

	// Success response
	data := NewDetailedAssetResponseData(asset)
	response := dto.NewResponse(ctx, "got asset successfully").WithData(data)

	slog.InfoContext(ctx.Request.Context(), response.Message,
		"checksum", asset.Checksum,
	)

	response.OK(ctx)
}

func NewDetailedAssetResponseData(asset *registry.Asset) *DetailedAssetResponseData {
	return &DetailedAssetResponseData{
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
