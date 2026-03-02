package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/registry"
	"github.com/UnivocalX/aether/pkg/web/api/dto"
	"github.com/UnivocalX/aether/pkg/web/services/data"
	"github.com/gin-gonic/gin"
)

type ListTagAssetsRequest struct {
	Limit  uint `json:"limit" binding:"omitempty,min=1,max=1000"`
	Offset uint `json:"offset" binding:"omitempty,min=0"`
}

type ListTagAssetsResponse struct {
	dto.Response
	Total      int      `json:"total"`
	NextOffset *uint    `json:"next_offset,omitempty"`
	Assets     []string `json:"assets"`
}

func ListTagAssetsHandler(svc *data.Service, ctx *gin.Context) {
	var uri dto.TagUri
	var request ListTagAssetsRequest

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
	if err := ctx.ShouldBindJSON(&request); err != nil {
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
			Limit:  request.Limit,
			Offset: request.Offset,
		},
	)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to get tag assets", err)
		return
	}

	// Success response
	response := newListTagAssetsResponse(ctx, assets, request.Limit, request.Offset)
	response.OK(ctx)
}

func newListTagAssetsResponse(ctx *gin.Context, assets []*registry.Asset, limit uint, offset uint) ListTagAssetsResponse {
	assetsChecksums := make([]string, len(assets))

	for i, asset := range assets {
		assetsChecksums[i] = asset.Checksum
	}

	response := ListTagAssetsResponse{
		Response: *dto.NewResponse(ctx, "got tags assets successfully"),
		Total:    len(assetsChecksums),
		Assets:   assetsChecksums,
	}

	// Only include NextOffset if we got a full page (meaning there might be more)
	if len(assets) == int(limit) {
		NextOffset := offset + uint(len(assets))
		response.NextOffset = &NextOffset
	}

	slog.InfoContext(ctx.Request.Context(), response.Msg,
		"total", len(assets),
	)
	return response
}
