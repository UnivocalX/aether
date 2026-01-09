package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/gin-gonic/gin"
)

func TagAssetHandler(svc *data.Service, ctx *gin.Context) {
	var uri dto.AssetTagUri

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&uri); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to tag asset",
			fmt.Errorf("%w, %w", dto.ErrInvalidUri, err),
		)
		return
	}

	if err := svc.TagAsset(ctx.Request.Context(), uri.AssetChecksum, uri.TagName); err != nil {
		dto.HandleErrorResponse(ctx, "failed to tag asset", err)
		return
	}

	// Success response
	response := dto.NewResponse(ctx, "tag asset successfully")

	slog.InfoContext(ctx.Request.Context(), response.Message,
		"tagName", uri.TagName,
		"Checksum", uri.AssetChecksum,
	)

	response.NoContent(ctx)
}
