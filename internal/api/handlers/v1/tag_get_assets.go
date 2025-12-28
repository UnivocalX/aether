package v1

import (
	"errors"
	"log/slog"

	"github.com/UnivocalX/aether/internal/api/dto"
	"github.com/UnivocalX/aether/internal/api/services/data"
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

type TagAssetsGetResponse struct {
	Total      int      `json:"total"`
	NextOffset *uint    `json:"next_offset,omitempty"`
	Assets     []string `json:"assets"`
}

func NewTagAssetsGetResponse(assets []*registry.Asset, limit uint, offset uint) *TagAssetsGetResponse {
	assetsChecksums := make([]string, len(assets))

	for i, asset := range assets {
		assetsChecksums[i] = asset.Checksum
	}

	response := &TagAssetsGetResponse{
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
		slog.ErrorContext(ctx.Request.Context(), "Invalid URI parameters", "error", err.Error())
		dto.BadRequest(ctx, "Invalid SHA256 in URI")
		return
	}

	// Bind JSON payload
	if err := ctx.ShouldBindJSON(&req.TagAssetsGetPayload); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid JSON payload", "error", err.Error())
		dto.BadRequest(ctx, err.Error())
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
		handleGetTagAssetsError(ctx, err, req.Name)
		return
	}

	// Success response
	slog.InfoContext(ctx.Request.Context(), "got tags assets successfully",
		"name", req.Name,
		"total", len(assets),
	)

	response := NewTagAssetsGetResponse(assets, req.Limit, req.Offset)
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
