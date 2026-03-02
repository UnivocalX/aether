package v1

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/UnivocalX/aether/internal/registry"
	"github.com/UnivocalX/aether/pkg/web/api/dto"
	"github.com/UnivocalX/aether/pkg/web/services/data"
	"github.com/gin-gonic/gin"
)

type GetAssetResponse struct {
	dto.Response
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
	response := newGetAssetResponse(ctx, asset)
	response.OK(ctx)
}

func newGetAssetResponse(ctx *gin.Context, asset *registry.Asset) GetAssetResponse {
	response := GetAssetResponse{
		Response:  *dto.NewResponse(ctx, "got asset successfully"),
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

	slog.InfoContext(ctx.Request.Context(), response.Msg,
		"checksum", asset.Checksum,
	)

	return response
}
