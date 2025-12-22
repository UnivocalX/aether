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

type AssetPutTagUriParams struct {
	AssetUriParams
	TagUriParams
}

type AssetPutTaggingRequest struct {
	AssetPutTagUriParams
}

func HandleTaggingAsset(svc *data.Service, ctx *gin.Context) {
	var req AssetPutTaggingRequest

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&req.AssetPutTagUriParams); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid URI parameters", "error", err.Error())
		dto.BadRequest(ctx, "Invalid URI parameters")
		return
	}

	if err := svc.AddTagToAsset(ctx.Request.Context(), req.SHA256, req.Name); err != nil {
		handleTaggingAssetError(ctx, err, req.SHA256, req.Name)
		return
	}

	// Success response
	slog.InfoContext(ctx.Request.Context(), "tag asset successfully",
		"tagName", req.Name,
		"assetSha256", req.SHA256,
	)

	dto.NoContent(ctx, "tag asset successfully")
}

func handleTaggingAssetError(ctx *gin.Context, err error, sha256 string, name string) {
	switch {
	case errors.Is(err, registry.ErrValidation):
		dto.BadRequest(ctx, err.Error())

	case errors.Is(err, data.ErrAssetNotFound):
		dto.NotFound(ctx, fmt.Sprintf("asset %s does not exist", sha256))

	case errors.Is(err, data.ErrTagNotFound):
		dto.NotFound(ctx, fmt.Sprintf("tag %s does not exist", name))

	default:
		slog.ErrorContext(ctx.Request.Context(), "Failed to tag asset",
			"error", err.Error(),
			"tag", name,
			"sha256", sha256,
		)
		dto.InternalError(ctx, "Failed to tag asset")
	}
}
