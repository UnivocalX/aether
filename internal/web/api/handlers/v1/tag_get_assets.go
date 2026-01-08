package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/gin-gonic/gin"
)

type TagAssetsGetPayload struct {
	Limit  uint `json:"limit" binding:"omitempty,min=1,max=1000"`
	Offset uint `json:"offset" binding:"omitempty,min=0"`
}

type TagAssetsGetRequest struct {
	TagUriParams
	TagAssetsGetPayload
}

type TagAssetsGetResponseData struct {
	Total      int      `json:"total"`
	NextOffset *uint    `json:"next_offset,omitempty"`
	Assets     []string `json:"assets"`
}

func NewTagAssetsGetResponseData(assets []*registry.Asset, limit uint, offset uint) *TagAssetsGetResponseData {
	assetsChecksums := make([]string, len(assets))

	for i, asset := range assets {
		assetsChecksums[i] = asset.Checksum
	}

	response := &TagAssetsGetResponseData{
		Total:  len(assetsChecksums),
		Assets: assetsChecksums,
	}

	// Only include NextOffset if we got a full page (meaning there might be more)
	if len(assets) == int(limit) {
		NextOffset := offset + uint(len(assets))
		response.NextOffset = &NextOffset
	}

	return response
}

func HandleGetTagAssets(svc *data.Service, ctx *gin.Context) {
	var req TagAssetsGetRequest

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&req.TagUriParams); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to get tag assets",
			fmt.Errorf("%w: %w", dto.ErrInvalidUri, err),
		)
		return
	}

	// Bind JSON payload
	if err := ctx.ShouldBindJSON(&req.TagAssetsGetPayload); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to get tag assets",
			fmt.Errorf("%w: %w", dto.ErrInvalidPayload, err),
		)
		return
	}

	assets, err := svc.GetTagAssets(
		ctx.Request.Context(),
		data.GetTagAssetsParams{
			Name:   req.Name,
			Limit:  req.Limit,
			Offset: req.Offset,
		},
	)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to get tag assets", err)
		return
	}

	// Success response
	data := NewTagAssetsGetResponseData(assets, req.Limit, req.Offset)
	response := dto.NewResponse(ctx, "got tags assets successfully").WithData(data)

	slog.InfoContext(ctx.Request.Context(), response.Message,
		"name", req.Name,
		"total", len(assets),
	)
	response.OK(ctx)
}
