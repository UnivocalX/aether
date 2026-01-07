package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/gin-gonic/gin"
)

type AssetDeleteTagUriParams struct {
	AssetUriParams
	TagUriParams
}

type AssetDeleteTagRequest struct {
	AssetDeleteTagUriParams
}

func HandleUntagAsset(svc *data.Service, ctx *gin.Context) {
	var req AssetDeleteTagRequest

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&req.AssetDeleteTagUriParams); err != nil {
		dto.HandleErrorResponse(ctx, "failed to untag asset", fmt.Errorf("%w: %w", dto.ErrInvalidUri, err))
		return
	}

	if err := svc.UntagAsset(ctx.Request.Context(), req.Checksum, req.Name); err != nil {
		dto.HandleErrorResponse(ctx, "failed to untag asset", err)
		return
	}

	// Success response
	response := dto.NewResponse(ctx, "untag asset successfully")
	slog.InfoContext(ctx.Request.Context(), response.Message,
		"tagName", req.Name,
		"checksum", req.Checksum,
	)

	response.NoContent(ctx)
}
