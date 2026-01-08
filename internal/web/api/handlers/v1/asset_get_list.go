package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
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

type AssetListGetResponseData struct {
	Total      int              `json:"total"`
	NextCursor *uint            `json:"next_cursor,omitempty"`
	Assets     []*AssetListItem `json:"assets"`
}

func HandleListAssets(svc *data.Service, ctx *gin.Context) {
	var req AssetListGetRequest

	// Bind JSON payload
	if err := ctx.ShouldBindJSON(&req.AssetListGetPayload); err != nil {
		dto.HandleErrorResponse(
			ctx, 
			"failed to list assets",
			fmt.Errorf("%w: %w", dto.ErrInvalidPayload, err),
		)
		return
	}

	assets, err := svc.ListAssets(ctx.Request.Context(), ToSearchOptions(&req)...)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to list assets", err)
		return
	}

	// Success response
	data := NewAssetListGetResponseData(assets, req.Limit)
	response := dto.NewResponse(ctx, "listed assets successfully").WithData(data)

	slog.InfoContext(ctx.Request.Context(), response.Message,
		"total", len(assets),
	)

	response.OK(ctx)
}

func NewAssetListGetResponseData(assets []*registry.Asset, limit uint) *AssetListGetResponseData {
	items := make([]*AssetListItem, 0, len(assets))

	for _, asset := range assets {
		tags := make([]string, 0, len(asset.Tags))
		for _, tag := range asset.Tags {
			tags = append(tags, tag.Name)
		}

		items = append(items, &AssetListItem{
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
	// Only include next_cursor if we got a full page (might be more)
	if len(assets) == int(limit) && len(assets) > 0 {
		nextCursor = &assets[len(assets)-1].ID
	}

	return &AssetListGetResponseData{
		Total:      len(assets),
		Assets:     items,
		NextCursor: nextCursor,
	}
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
