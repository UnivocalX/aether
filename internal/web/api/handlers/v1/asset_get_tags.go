package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/UnivocalX/aether/internal/registry"
	"github.com/gin-gonic/gin"
)

type AssetTagsResponseData struct {
	Total int      `json:"total"`
	Tags  []string `json:"tags"`
}

func ListAssetTagsHandler(svc *data.Service, ctx *gin.Context) {
	var uri dto.AssetUri

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&uri); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to get asset tags",
			fmt.Errorf("%w, %w", dto.ErrInvalidUri, err),
		)
		return
	}

	tags, err := svc.GetAssetTags(ctx.Request.Context(), uri.AssetChecksum)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to get asset tags", err)
		return
	}

	// Success response
	data := NewAssetTagsResponseData(tags)
	response := dto.NewResponse(ctx, "got asset tags successfully").WithData(data)

	slog.InfoContext(ctx.Request.Context(), response.Message,
		"checksum", uri.AssetChecksum,
		"total", len(tags),
	)

	response.OK(ctx)
}

func NewAssetTagsResponseData(tags []*registry.Tag) *AssetTagsResponseData {
	tagsNames := make([]string, len(tags))

	for i, tag := range tags {
		tagsNames[i] = tag.Name
	}

	return &AssetTagsResponseData{
		Total: len(tagsNames),
		Tags:  tagsNames,
	}
}
