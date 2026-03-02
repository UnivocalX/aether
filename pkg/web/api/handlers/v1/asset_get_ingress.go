package v1

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/UnivocalX/aether/internal/registry"
	"github.com/UnivocalX/aether/pkg/web/api/dto"
	"github.com/UnivocalX/aether/pkg/web/services/data"
	"github.com/gin-gonic/gin"
)

type AssetIngressResponse struct {
	dto.Response
	Checksum  string     `json:"checksum"`
	UploadURL string     `json:"upload_url,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

func GetAssetIngressHandler(svc *data.Service, ctx *gin.Context) {
	var uri dto.AssetUri

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&uri); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to get asset ingress url",
			fmt.Errorf("%w, %w", dto.ErrInvalidUri, err),
		)
		return
	}

	// business logic
	ingressUrl, err := svc.GetAssetIngressUrl(ctx.Request.Context(), uri.AssetChecksum)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to get asset ingress url", err)
		return
	}

	// Success response
	response := newAssetIngressResponse(ctx, ingressUrl)
	response.OK(ctx)
}

func newAssetIngressResponse(ctx *gin.Context, presignedUrl *registry.PresignedUrl) AssetIngressResponse {
	response := AssetIngressResponse{
		Response: *dto.NewResponse(ctx, "got asset ingress url successfully"),
		Checksum:  presignedUrl.Checksum,
		UploadURL: presignedUrl.URL.Value(),
		ExpiresAt: &presignedUrl.ExpiresAt,
	}
	slog.InfoContext(ctx.Request.Context(), response.Msg,
		"checksum", presignedUrl.Checksum,
	)
	return response
}
