package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/gin-gonic/gin"
)

type AssetBatchPostPayload struct {
	assets []struct {
		Checksum string         `json:"checksum" binding:"required,len=64,hexadecimal"`
		Display  string         `json:"display" binding:"omitempty,max=120"`
		Extra    map[string]any `json:"extra" binding:"omitempty"`
	}
}

type AssetBatchPostRequest struct {
	AssetBatchPostPayload
}

type AssetBatchResponseData struct {
	Assets []*AssetPostResponseData
}

func HandleCreateAssetBatch(svc *data.Service, ctx *gin.Context) {
	var req AssetBatchPostRequest

	// Bind JSON payload
	if err := ctx.ShouldBindJSON(&req.AssetBatchPostPayload); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to execute batch",
			fmt.Errorf("%w: %w", dto.ErrInvalidUri, err),
		)
		return
	}

	// Convert request to records
	assets, err := batchRequestToRecords(&req)
	if err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to execute batch",
			fmt.Errorf("%w: %w", dto.ErrInvalidPayload, err),
		)
		return
	}

	ingressUrls, err := svc.CreateAssets(ctx.Request.Context(), assets...)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to execute batch", err)
		return
	}

	// Success response
	data := NewAssetBatchResponseData(assets, ingressUrls)
	response := dto.NewResponse(ctx, "successfully executed batch").WithData(data)
	slog.InfoContext(ctx.Request.Context(), response.Message,
		"total", len(assets),
	)

	response.Created(ctx)
}

func NewAssetBatchResponseData(
	assets []*registry.Asset,
	urls []*registry.PresignedUrl,
) *AssetBatchResponseData {

	// Build lookup map: checksum → presigned URL
	urlMap := make(map[string]*registry.PresignedUrl, len(urls))
	for _, u := range urls {
		urlMap[u.Checksum] = u
	}

	// Build response assets
	dataAssets := make([]*AssetPostResponseData, len(assets))
	for i, a := range assets {
		uploadURL := urlMap[a.Checksum] // may be nil if no URL exists
		dataAssets[i] = NewAssetPostResponseData(a, uploadURL)
	}

	return &AssetBatchResponseData{
		Assets: dataAssets,
	}
}

func batchRequestToRecords(req *AssetBatchPostRequest) ([]*registry.Asset, error) {
	records := make([]*registry.Asset, len(req.assets))

	for i, asset := range req.assets {
		record := &registry.Asset{
			Checksum: asset.Checksum,
			Display:  asset.Display,
		}

		if err := record.SetExtra(asset.Extra); err != nil {
			return nil, err
		}

		records[i] = record
	}

	return records, nil
}
