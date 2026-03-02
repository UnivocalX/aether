package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/registry"
	"github.com/UnivocalX/aether/pkg/web/api/dto"
	"github.com/UnivocalX/aether/pkg/web/services/data"
	"github.com/gin-gonic/gin"
)

type AssetTagsResponse struct {
	dto.Response
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
	response := newAssetTagsResponse(ctx, tags)
	response.OK(ctx)
}

func newAssetTagsResponse(ctx *gin.Context, tags []*registry.Tag) AssetTagsResponse {
	tagsNames := make([]string, len(tags))

	for i, tag := range tags {
		tagsNames[i] = tag.Name
	}

	response := AssetTagsResponse{
		Response: *dto.NewResponse(ctx, "got asset tags successfully"),
		Total: len(tagsNames),
		Tags:  tagsNames,
	}
	slog.InfoContext(ctx.Request.Context(), response.Msg,
		"total", len(tags),
	)

	return response
}
