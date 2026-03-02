package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/registry"
	"github.com/UnivocalX/aether/pkg/web/api/dto"
	"github.com/UnivocalX/aether/pkg/web/services/data"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type ListAssetsRequest struct {
	Cursor       uint     `json:"cursor" binding:"omitempty,gte=0"`
	Limit        uint     `json:"limit" binding:"omitempty,gte=1,lte=1000"`
	MimeType     string   `json:"mime_type" binding:"omitempty"`
	State        string   `json:"state" binding:"omitempty,oneof=pending active archived"`
	IncludedTags []string `json:"included_tags" binding:"omitempty,dive,min=1,max=100"`
	ExcludedTags []string `json:"excluded_tags" binding:"omitempty,dive,min=1,max=100"`
}

type ListAssetsResponse struct {
	dto.Response
	Total      int             `json:"total"`
	NextCursor *uint           `json:"next_cursor,omitempty"`
	Assets     []*AssetDetails `json:"assets"`
}

type AssetDetails struct {
	ID        uint           `json:"id"`
	Checksum  string         `json:"checksum"`
	Display   string         `json:"display"`
	Extra     datatypes.JSON `json:"extra"`
	MimeType  string         `json:"mime_type"`
	SizeBytes int64          `json:"size_bytes"`
	State     string         `json:"state"`
	Tags      []string       `json:"tags"`
}

func ListAssetsHandler(svc *data.Service, ctx *gin.Context) {
	var request ListAssetsRequest

	// Bind JSON payload
	if err := ctx.ShouldBindJSON(&request); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to list assets",
			fmt.Errorf("%w, %w", dto.ErrInvalidPayload, err),
		)
		return
	}

	assets, err := svc.ListAssets(ctx.Request.Context(), ToSearchOptions(&request)...)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to list assets", err)
		return
	}

	// Success response
	response := newListAssetsResponse(ctx, assets, request.Limit)
	response.OK(ctx)
}

func ToSearchOptions(req *ListAssetsRequest) []registry.SearchAssetsOption {
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

func newListAssetsResponse(ctx *gin.Context, assets []*registry.Asset, limit uint) ListAssetsResponse {
	items := make([]*AssetDetails, 0, len(assets))

	for _, asset := range assets {
		tags := make([]string, 0, len(asset.Tags))
		for _, tag := range asset.Tags {
			tags = append(tags, tag.Name)
		}

		items = append(items, &AssetDetails{
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

	response := ListAssetsResponse{
		Response:   *dto.NewResponse(ctx, "listed assets successfully"),
		Total:      len(assets),
		Assets:     items,
		NextCursor: nextCursor,
	}
	slog.InfoContext(ctx.Request.Context(), response.Msg,
		"total", len(assets),
	)
	return response
}
