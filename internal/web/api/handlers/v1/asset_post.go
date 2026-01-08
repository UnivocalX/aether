package v1

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"

	"github.com/gin-gonic/gin"
)

type AssetUriParams struct {
	Checksum string `uri:"checksum" binding:"required,len=64,hexadecimal"`
}

type AssetPostPayload struct {
	Display string         `json:"display" binding:"omitempty,max=120"`
	Tags    []string       `json:"tags" binding:"omitempty,dive,min=1,max=100"`
	Extra   map[string]any `json:"extra" binding:"omitempty"`
}

type AssetPostRequest struct {
	AssetUriParams
	AssetPostPayload
}

type AssetPostResponseData struct {
	ID        uint       `json:"id"`
	Checksum  string     `json:"checksum"`
	State     string     `json:"state"`
	UploadURL string     `json:"upload_url,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// HandleCreateAsset handles the HTTP Asset post request/response cycle
func HandleCreateAsset(svc *data.Service, ctx *gin.Context) {
	var req AssetPostRequest

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&req.AssetUriParams); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to create new asset",
			fmt.Errorf("%w: %w", dto.ErrInvalidUri, err),
		)
		return
	}

	// Bind JSON payload
	if err := ctx.ShouldBindJSON(&req.AssetPostPayload); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to create new asset",
			fmt.Errorf("%w: %w", dto.ErrInvalidPayload, err),
		)
		return
	}

	// Execute business logic
	result := svc.CreateAsset(ctx.Request.Context(), data.CreateAssetParams{
		Checksum: req.Checksum,
		Display:  req.Display,
		Tags:     req.Tags,
		Extra:    req.Extra,
	})

	// Handle errors
	if result.Err != nil {
		dto.HandleErrorResponse(ctx, "failed to create new asset", result.Err)
		return
	}

	// Success response
	data := NewAssetPostResponseData(result)
	response := dto.NewResponse(ctx, "asset created successfully").WithData(data)

	slog.InfoContext(ctx.Request.Context(), response.Message,
		"checksum", result.Asset.Checksum,
	)

	response.Created(ctx)
}

func NewAssetPostResponseData(result *data.CreateAssetResult) *AssetPostResponseData {
	response := &AssetPostResponseData{
		ID:       result.Asset.ID,
		Checksum: result.Asset.Checksum,
		State:    string(result.Asset.State),
	}

	if result.UploadURL != nil {
		response.UploadURL = result.UploadURL.URL.Value()
		response.ExpiresAt = &result.UploadURL.ExpiresAt
	}

	return response
}
