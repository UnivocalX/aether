package v1

import (
	"fmt"

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

func HandleCreateAssetBatch(svc *data.Service, ctx *gin.Context) {
	var req AssetBatchPostRequest

	// Bind JSON payload
	if err := ctx.ShouldBindJSON(&req.AssetBatchPostPayload); err != nil {
		dto.HandleErrorResponse(
			ctx, 
			"failed to create assets batch", 
			fmt.Errorf("%w: %w", dto.ErrInvalidUri, err),
		)
		return
	}

	// Convert request to records
	assets, err := batchRequestToRecords(&req)
	if err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to create assets batch",
			fmt.Errorf("%w: %w", dto.ErrInvalidPayload, err),
		)
		return
	}

	if err := svc.CreateAssets(ctx.Request.Context(), assets...); err != nil {
		dto.HandleErrorResponse(ctx, "failed to create assets batch", err, )
		return 
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