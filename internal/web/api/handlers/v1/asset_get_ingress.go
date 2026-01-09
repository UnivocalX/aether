package v1

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/gin-gonic/gin"
)

type AssetIngressResponseData struct {
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
	data := NewAssetIngressResponseData(ingressUrl)
	response := dto.NewResponse(ctx, "got asset ingress url successfully").WithData(data)

	slog.InfoContext(ctx.Request.Context(), response.Message,
		"checksum", uri.AssetChecksum,
	)

	response.OK(ctx)
}

func NewAssetIngressResponseData(presignedUrl *registry.PresignedUrl) *AssetIngressResponseData {
	return &AssetIngressResponseData{
		Checksum:  presignedUrl.Checksum,
		UploadURL: presignedUrl.URL.Value(),
		ExpiresAt: &presignedUrl.ExpiresAt,
	}
}
