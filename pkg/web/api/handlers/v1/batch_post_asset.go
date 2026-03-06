package v1

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/UnivocalX/aether/internal/registry"
	"github.com/UnivocalX/aether/pkg/web/api/dto"
	"github.com/UnivocalX/aether/pkg/web/services/data"
	"github.com/gin-gonic/gin"
)

type CreateAssetsBatchRequest struct {
	Assets []AssetPayload `json:"assets" binding:"required,max=1000,dive"`
}

type AssetPayload struct {
	Checksum string         `json:"checksum" binding:"required,len=64,hexadecimal"`
	Display  string         `json:"display" binding:"omitempty,max=120"`
	Extra    map[string]any `json:"extra" binding:"omitempty"`
}

type AssetsBatchResponse struct {
	dto.Response
	Assets []*BatchAssetDetails
}

type BatchAssetDetails struct {
	ID         uint            `json:"id"`
	Checksum   string          `json:"checksum"`
	State      string          `json:"state"`
	IngressUrl registry.Secret `json:"ingress_url,omitempty"`
	ExpiresAt  *time.Time      `json:"expires_at,omitempty"`
}

func CreateAssetsBatchHandler(svc *data.Service, ctx *gin.Context) {
	var request CreateAssetsBatchRequest

	// Bind JSON payload
	if err := ctx.ShouldBindJSON(&request); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to execute batch",
			fmt.Errorf("%w, %w", dto.ErrInvalidPayload, err),
		)
		return
	}

	// Convert request to records
	assets, err := assetsBatchRequest2Records(&request)
	if err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to execute batch",
			fmt.Errorf("%w, %w", dto.ErrInvalidPayload, err),
		)
		return
	}

	ingressUrls, err := svc.CreateAssets(ctx.Request.Context(), assets...)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to execute batch", err)
		return
	}

	// Success response
	response := newAssetsBatchResponse(ctx, assets, ingressUrls)
	response.Created(ctx)
}

func assetsBatchRequest2Records(payload *CreateAssetsBatchRequest) ([]*registry.Asset, error) {
	records := make([]*registry.Asset, len(payload.Assets))

	for i, asset := range payload.Assets {
		record := &registry.Asset{
			Checksum: asset.Checksum,
			Display:  asset.Display,
		}

		if len(asset.Extra) > 0 {
			if err := record.SetExtra(asset.Extra); err != nil {
				return nil, err
			}
		}

		records[i] = record
	}

	return records, nil
}

func newAssetsBatchResponse(
	ctx *gin.Context,
	assets []*registry.Asset,
	urls []*registry.PresignedUrl,
) AssetsBatchResponse {

	// Build lookup map: checksum → presigned URL
	urlMap := make(map[string]*registry.PresignedUrl, len(urls))
	for _, u := range urls {
		urlMap[u.Checksum] = u
	}

	// Build response assets
	batchAssets := make([]*BatchAssetDetails, len(assets))
	for i, a := range assets {
		uploadURL := urlMap[a.Checksum] // may be nil if no URL exists
		batchAssets[i] = &BatchAssetDetails{
			ID:         a.ID,
			Checksum:   a.Checksum,
			State:      string(a.State),
			IngressUrl: uploadURL.URL,
			ExpiresAt:  &uploadURL.ExpiresAt,
		}
	}

	response := AssetsBatchResponse{
		Response: *dto.NewResponse(ctx, "successfully executed batch"),
		Assets:   batchAssets,
	}
	slog.InfoContext(ctx.Request.Context(), response.Msg,
		"total", len(assets),
	)
	return response
}
