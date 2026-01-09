package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/gin-gonic/gin"
)

func UntagAssetHandler(svc *data.Service, ctx *gin.Context) {
	var uri dto.AssetTagUri

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&uri); err != nil {
		dto.HandleErrorResponse(ctx, "failed to untag asset", fmt.Errorf("%w, %w", dto.ErrInvalidUri, err))
		return
	}

	if err := svc.UntagAsset(ctx.Request.Context(), uri.AssetChecksum, uri.TagName); err != nil {
		dto.HandleErrorResponse(ctx, "failed to untag asset", err)
		return
	}

	// Success response
	response := dto.NewResponse(ctx, "untag asset successfully")
	slog.InfoContext(ctx.Request.Context(), response.Message,
		"tagName", uri.TagName,
		"assetChecksum", uri.AssetChecksum,
	)

	response.NoContent(ctx)
}
