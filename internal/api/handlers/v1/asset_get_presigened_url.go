package v1

import (
	"errors"
	"log/slog"
	"time"

	"github.com/UnivocalX/aether/internal/api/dto"
	"github.com/UnivocalX/aether/internal/api/services/data"
	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/gin-gonic/gin"
)

type AssetGetUrlRequest struct {
	AssetUriParams
}

type AssetGetUrlResponse struct {
	Checksum  string     `json:"checksum"`
	UploadURL string     `json:"upload_url,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

func NewAssetGetUrlResponse(presignedUrl *registry.PresignedUrl) *AssetGetUrlResponse {
	return &AssetGetUrlResponse{
		Checksum: presignedUrl.Checksum,
		UploadURL: presignedUrl.URL.Value(),
		ExpiresAt: &presignedUrl.ExpiresAt,
	}
}

func HandleGetAssetUrl(svc *data.Service, ctx *gin.Context) {
	var req AssetGetUrlRequest

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&req.AssetUriParams); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid URI parameters", "error", err.Error())
		dto.BadRequest(ctx, "Invalid SHA256 in URI")
		return
	}

	presignedUrl, err := svc.GetAssetPresignedUrl(ctx.Request.Context(), req.SHA256)
	if err != nil {
		handleGetAssetUrlError(ctx, err, req.SHA256)
		return
	}

	// Success response
	slog.InfoContext(ctx.Request.Context(), "got asset presigned url successfully",
		"Sha256", req.SHA256,
	)
	
	response := NewAssetGetUrlResponse(presignedUrl)
	dto.OK(ctx, "got asset presigned url successfully", response)
}

func handleGetAssetUrlError(ctx *gin.Context, err error, sha256 string) {
	switch {
	case errors.Is(err, registry.ErrValidation):
		dto.BadRequest(ctx, err.Error())

	case errors.Is(err, data.ErrAssetNotFound):
		dto.NotFound(ctx, err.Error())

	case errors.Is(err, data.ErrAssetIsReady):
		dto.Conflict(ctx, err.Error())

	default:
		slog.ErrorContext(ctx.Request.Context(), "Failed to get asset url",
			"error", err.Error(),
			"sha256", sha256,
		)
		dto.InternalError(ctx, "Failed to get asset url")
	}
}
