package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
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
		dto.HandleErrorResponse(
			ctx, 
			"failed to tag asset",
			fmt.Errorf("%w: %w", dto.ErrInvalidUri, err),
		)
		return
	}

	if err := svc.TagAsset(ctx.Request.Context(), req.Checksum, req.Name); err != nil {
		dto.HandleErrorResponse(ctx, "failed to tag asset", err)
		return
	}

	// Success response
	response := dto.NewResponse(ctx, "tag asset successfully")

	slog.InfoContext(ctx.Request.Context(), response.Message,
		"tagName", req.Name,
		"Checksum", req.Checksum,
	)

	response.NoContent(ctx)
}