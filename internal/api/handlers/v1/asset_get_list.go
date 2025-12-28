package v1

import (
	"errors"
	"log/slog"

	"github.com/UnivocalX/aether/internal/api/dto"
	"github.com/UnivocalX/aether/internal/api/services/data"
	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type AssetListGetPayload struct {
	Cursor       uint     `json:"cursor" binding:"omitempty,gte=0"`
	Limit        uint     `json:"limit" binding:"omitempty,gte=1,lte=1000"`
	MimeType     string   `json:"mime_type" binding:"omitempty"`
	State        string   `json:"state" binding:"omitempty,oneof=pending active archived"`
	IncludedTags []string `json:"included_tags" binding:"omitempty,dive,min=1,max=100"`
	ExcludedTags []string `json:"excluded_tags" binding:"omitempty,dive,min=1,max=100"`
}

type AssetListGetRequest struct {
	AssetListGetPayload
}

type AssetListItem struct {
	ID        uint           `json:"id"`
	Checksum  string         `json:"checksum"`
	Display   string         `json:"display"`
	Extra     datatypes.JSON `json:"extra"`
	MimeType  string         `json:"mime_type"`
	SizeBytes int64          `json:"size_bytes"`
	State     string         `json:"state"`
	Tags      []string       `json:"tags"`
}

type AssetListGetResponse struct {
	Total      int             `json:"total"`
	NextCursor *uint           `json:"next_cursor,omitempty"`
	Assets     []AssetListItem `json:"assets"`
}

func NewAssetListGetResponse(assets []*registry.Asset) *AssetListGetResponse {
	items := make([]AssetListItem, 0, len(assets))

	for _, asset := range assets {
		tags := make([]string, 0, len(asset.Tags))
		for _, tag := range asset.Tags {
			tags = append(tags, tag.Name)
		}

		items = append(items, AssetListItem{
			ID:        asset.ID,
			Checksum:  asset.Checksum,
			Display:   asset.Display,
			Extra:     asset.Extra,
			MimeType:  asset.MimeType,
			SizeBytes: asset.SizeBytes,
			State:     string(asset.State),
			Tags:      tags,
		})
	}

	var nextCursor *uint
	if len(assets) > 0 {
		lastID := assets[len(assets)-1].ID
		nextCursor = &lastID
	}

	return &AssetListGetResponse{
		Total:      len(assets),
		Assets:     items,
		NextCursor: nextCursor,
	}
}

func HandleListAssets(svc *data.Service, ctx *gin.Context) {
	var req AssetListGetRequest

	// Bind JSON payload
	if err := ctx.ShouldBindJSON(&req.AssetListGetPayload); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid JSON payload", "error", err.Error())
		dto.BadRequest(ctx, err.Error())
		return
	}

	assets, err := svc.ListAssets(ctx.Request.Context(), ToSearchOptions(&req)...)
	if err != nil {
		handleListAssetsError(ctx, err)
		return
	}

	// Success response
	slog.InfoContext(ctx.Request.Context(), "listed assets successfully",
		"total", len(assets),
	)

	response := NewAssetListGetResponse(assets)
	dto.OK(ctx, "Successfully listed assets", response)
}

func ToSearchOptions(req *AssetListGetRequest) []registry.SearchAssetsOption {
	var opts []registry.SearchAssetsOption

	// Helper to conditionally add options
	addIfSet := func(condition bool, opt registry.SearchAssetsOption) {
		if condition {
			opts = append(opts, opt)
		}
	}

	addIfSet(req.Cursor > 0, registry.WithCursor(req.Cursor))
	addIfSet(req.Limit > 0, registry.WithLimit(req.Limit))
	addIfSet(req.MimeType != "", registry.WithMimeType(req.MimeType))
	addIfSet(req.State != "", registry.WithState(registry.Status(req.State)))
	addIfSet(len(req.IncludedTags) > 0, registry.WithIncludedTags(req.IncludedTags...))
	addIfSet(len(req.ExcludedTags) > 0, registry.WithExcludedTags(req.ExcludedTags...))

	return opts
}

func handleListAssetsError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, registry.ErrValidation):
		dto.BadRequest(ctx, err.Error())

	default:
		slog.ErrorContext(ctx.Request.Context(), "Failed to list assets",
			"error", err.Error(),
		)
		dto.InternalError(ctx, "Failed to list assets")
	}
}
