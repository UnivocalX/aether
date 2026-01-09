package v1

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/gin-gonic/gin"
)

type AssetPayload struct {
	Checksum string         `json:"checksum" binding:"required,len=64,hexadecimal"`
	Display  string         `json:"display" binding:"omitempty,max=120"`
	Extra    map[string]any `json:"extra" binding:"omitempty"`
}

type CreateAssetsBatchPayload struct {
	Assets []AssetPayload `json:"assets" binding:"required,max=1000,dive"`
}

type BatchAsset struct {
	ID         uint       `json:"id"`
	Checksum   string     `json:"checksum"`
	State      string     `json:"state"`
	IngressUrl string     `json:"ingress_url,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

type AssetsBatchResponseData struct {
	Assets []*BatchAsset
}

func CreateAssetsBatchHandler(svc *data.Service, ctx *gin.Context) {
	var payload CreateAssetsBatchPayload

	// Bind JSON payload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to execute batch",
			fmt.Errorf("%w, %w", dto.ErrInvalidPayload, err),
		)
		return
	}

	// Convert request to records
	assets, err := assetsBatchPayloadToRecords(&payload)
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
	data := NewCreateAssetsBatchResponseData(assets, ingressUrls)
	response := dto.NewResponse(ctx, "successfully executed batch").WithData(data)
	slog.InfoContext(ctx.Request.Context(), response.Message,
		"total", len(assets),
	)

	response.Created(ctx)
}

func NewBatchAsset(asset *registry.Asset, uploadURL *registry.PresignedUrl) *BatchAsset {
	response := &BatchAsset{
		ID:       asset.ID,
		Checksum: asset.Checksum,
		State:    string(asset.State),
	}

	if uploadURL != nil {
		response.IngressUrl = uploadURL.URL.Value()
		response.ExpiresAt = &uploadURL.ExpiresAt
	}

	return response
}

func NewCreateAssetsBatchResponseData(
	assets []*registry.Asset,
	urls []*registry.PresignedUrl,
) *AssetsBatchResponseData {

	// Build lookup map: checksum → presigned URL
	urlMap := make(map[string]*registry.PresignedUrl, len(urls))
	for _, u := range urls {
		urlMap[u.Checksum] = u
	}

	// Build response assets
	batchAssets := make([]*BatchAsset, len(assets))
	for i, a := range assets {
		uploadURL := urlMap[a.Checksum] // may be nil if no URL exists
		batchAssets[i] = NewBatchAsset(a, uploadURL)
	}

	return &AssetsBatchResponseData{
		Assets: batchAssets,
	}
}

func assetsBatchPayloadToRecords(payload *CreateAssetsBatchPayload) ([]*registry.Asset, error) {
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
