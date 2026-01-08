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

type IngressUrlGetRequest struct {
	AssetUriParams
}

type IngressUrlGetResponseData struct {
	Checksum  string     `json:"checksum"`
	UploadURL string     `json:"upload_url,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

func NewIngressUrlGetResponseData(presignedUrl *registry.PresignedUrl) *IngressUrlGetResponseData {
	return &IngressUrlGetResponseData{
		Checksum:  presignedUrl.Checksum,
		UploadURL: presignedUrl.URL.Value(),
		ExpiresAt: &presignedUrl.ExpiresAt,
	}
}

func HandleGetIngressUrl(svc *data.Service, ctx *gin.Context) {
	var req IngressUrlGetRequest

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&req.AssetUriParams); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to get asset ingress url",
			fmt.Errorf("%w: %w", dto.ErrInvalidUri, err),
		)
		return
	}

	// business logic
	presignedUrl, err := svc.GetAssetIngressUrl(ctx.Request.Context(), req.Checksum)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to get asset ingress url", err)
		return
	}

	// Success response
	data := NewIngressUrlGetResponseData(presignedUrl)
	response := dto.NewResponse(ctx, "got asset ingress url successfully").WithData(data)
	
	slog.InfoContext(ctx.Request.Context(), response.Message,
		"checksum", req.Checksum,
	)

	response.OK(ctx)
}
