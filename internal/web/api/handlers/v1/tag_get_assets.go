package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/gin-gonic/gin"
)

type ListTagAssetsPayload struct {
	Limit  uint `json:"limit" binding:"omitempty,min=1,max=1000"`
	Offset uint `json:"offset" binding:"omitempty,min=0"`
}

type ListTagAssetsResponseData struct {
	Total      int      `json:"total"`
	NextOffset *uint    `json:"next_offset,omitempty"`
	Assets     []string `json:"assets"`
}

func ListTagAssetsHandler(svc *data.Service, ctx *gin.Context) {
	var uri dto.TagUri
	var payload ListTagAssetsPayload

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&uri); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to get tag assets",
			fmt.Errorf("%w, %w", dto.ErrInvalidUri, err),
		)
		return
	}

	// Bind JSON payload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to get tag assets",
			fmt.Errorf("%w, %w", dto.ErrInvalidPayload, err),
		)
		return
	}

	assets, err := svc.GetTagAssets(
		ctx.Request.Context(),
		data.GetTagAssetsParams{
			Name:   uri.TagName,
			Limit:  payload.Limit,
			Offset: payload.Offset,
		},
	)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to get tag assets", err)
		return
	}

	// Success response
	data := NewListTagAssetsResponseData(assets, payload.Limit, payload.Offset)
	response := dto.NewResponse(ctx, "got tags assets successfully").WithData(data)

	slog.InfoContext(ctx.Request.Context(), response.Message,
		"name", uri.TagName,
		"total", len(assets),
	)
	response.OK(ctx)
}

func NewListTagAssetsResponseData(assets []*registry.Asset, limit uint, offset uint) *ListTagAssetsResponseData {
	assetsChecksums := make([]string, len(assets))

	for i, asset := range assets {
		assetsChecksums[i] = asset.Checksum
	}

	response := &ListTagAssetsResponseData{
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
