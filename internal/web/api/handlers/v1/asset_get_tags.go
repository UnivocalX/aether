package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/gin-gonic/gin"
)

type AssetTagsGetRequest struct {
	AssetUriParams
}

type AssetTagsGetResponseData struct {
	Total int      `json:"total"`
	Tags  []string `json:"tags"`
}

func HandleGetAssetTags(svc *data.Service, ctx *gin.Context) {
	var req AssetTagsGetRequest

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&req.AssetUriParams); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to get asset tags",
			fmt.Errorf("%w: %w", dto.ErrInvalidUri, err),
		)
		return
	}

	tags, err := svc.GetAssetTags(ctx.Request.Context(), req.Checksum)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to get asset tags", err)
		return
	}

	// Success response
	data := NewGetAssetTagsResponseData(tags)
	response := dto.NewResponse(ctx, "got asset tags successfully").WithData(data)

	slog.InfoContext(ctx.Request.Context(), response.Message,
		"checksum", req.Checksum,
		"total", len(tags),
	)
	
	response.OK(ctx)
}

func NewGetAssetTagsResponseData(tags []*registry.Tag) *AssetTagsGetResponseData {
	tagsNames := make([]string, len(tags))

	for i, tag := range tags {
		tagsNames[i] = tag.Name
	}

	return &AssetTagsGetResponseData{
		Total: len(tagsNames),
		Tags:  tagsNames,
	}
}