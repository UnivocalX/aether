package v1

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/api/dto"
	"github.com/UnivocalX/aether/internal/api/services/data"
	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/gin-gonic/gin"
)

type AssetTagsGetRequest struct {
	AssetUriParams
}

func NewGetAssetTagsResponse(tags []*registry.Tag) []string {
	tagsNames := make([]string, len(tags))

	for i, tag := range tags {
		tagsNames[i] = tag.Name
	}

	return tagsNames
}

func HandleGetAssetTags(svc *data.Service, ctx *gin.Context) {
	var req AssetTagsGetRequest

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&req.AssetUriParams); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid URI parameters", "error", err.Error())
		dto.BadRequest(ctx, "Invalid SHA256 in URI")
		return
	}

	tags, err := svc.GetAssetTags(ctx.Request.Context(), req.SHA256)
	if err != nil {
		handleGetAssetTagsError(ctx, err, req.SHA256)
		return
	}

	// Success response
	slog.InfoContext(ctx.Request.Context(), "got asset tags successfully",
		"assetSha256", req.SHA256,
		"totalTags", len(tags),
	)

	response := NewGetAssetTagsResponse(tags)
	dto.OK(ctx, "got asset tags successfully", response)
}

func handleGetAssetTagsError(ctx *gin.Context, err error, sha256 string) {
	switch {
	case errors.Is(err, registry.ErrValidation):
		dto.BadRequest(ctx, err.Error())

	case errors.Is(err, data.ErrAssetNotFound):
		dto.NotFound(ctx, fmt.Sprintf("asset %s does not exist", sha256))

	default:
		slog.ErrorContext(ctx.Request.Context(), "Failed to get asset tags",
			"error", err.Error(),
			"sha256", sha256,
		)
		dto.InternalError(ctx, "Failed to get asset tags")
	}
}
