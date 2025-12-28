package v1

import (
	"errors"
	"log/slog"

	"github.com/UnivocalX/aether/internal/api/dto"
	"github.com/UnivocalX/aether/internal/api/services/data"
	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/gin-gonic/gin"
)


type TagAssetsGetRequest struct {
	TagUriParams
}

type TagAssetsGetResponse struct {
	Total int `json:"total"`
	Assets []string `json:"assets"`
}

func NewTagAssetsGetResponse(assets []*registry.Asset) *TagAssetsGetResponse {
	assetsChecksums := make([]string, len(assets))

	for i, asset := range assets {
		assetsChecksums[i] = asset.Checksum
	}

	return &TagAssetsGetResponse{
		Total: len(assetsChecksums),
		Assets: assetsChecksums,
	}
}

func HandleGetTagAssets(svc *data.Service, ctx *gin.Context) {
	var req TagAssetsGetRequest

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&req.TagUriParams); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid URI parameters", "error", err.Error())
		dto.BadRequest(ctx, "Invalid SHA256 in URI")
		return
	}

	assets, err := svc.GetTagAssets(ctx.Request.Context(), req.Name)
	if err != nil {
		handleGetTagAssetsError(ctx, err, req.Name)
		return
	}

	// Success response
	slog.InfoContext(ctx.Request.Context(), "got tags assets successfully",
		"name", req.Name,
		"total", len(assets),
	)

	response := NewTagAssetsGetResponse(assets)
	dto.OK(ctx, "got tag assets successfully", response)
}

func handleGetTagAssetsError(ctx *gin.Context, err error, name string) {
	switch {
	case errors.Is(err, registry.ErrValidation):
		dto.BadRequest(ctx, err.Error())

	case errors.Is(err, data.ErrTagNotFound):
		dto.NotFound(ctx, err.Error())

	default:
		slog.ErrorContext(ctx.Request.Context(), "Failed to get tag assets",
			"error", err.Error(),
			"name", name,
		)
		dto.InternalError(ctx, "Failed to get tag assets")
	}
}