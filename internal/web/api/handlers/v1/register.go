package v1

import (
	"log/slog"

	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r gin.IRouter, svc *data.Service) {
	slog.Info("Registering V1 routes")

	v1 := r.Group("/v1")

	// Assets
	// List assets
	v1.GET("/assets", func(ctx *gin.Context) {
		ListAssetsHandler(svc, ctx)
	})

	// Get a specific asset
	v1.GET("/assets/:asset_checksum", func(ctx *gin.Context) {
		GetAssetHandler(svc, ctx)
	})

	// Get a specific asset tags
	v1.GET("/assets/:checksum/tags", func(ctx *gin.Context) {
		ListAssetTagsHandler(svc, ctx)
	})

	// Get an asset ingress Url
	v1.GET("/assets/:checksum/ingress", func(ctx *gin.Context) {
		GetAssetIngressHandler(svc, ctx)
	})

	// Tag a specific asset
	v1.PUT("/assets/:checksum/tags/:tag_name", func(ctx *gin.Context) {
		TagAssetHandler(svc, ctx)
	})

	// Untag a specific asset
	v1.DELETE("/assets/:checksum/tags/:tag_name", func(ctx *gin.Context) {
		UntagAssetHandler(svc, ctx)
	})

	// Tags
	// List tag assets
	v1.GET("/tags/:tag_name/assets", func(ctx *gin.Context) {
		ListTagAssetsHandler(svc, ctx)
	})

	// Add tag
	v1.PUT("/tags", func(ctx *gin.Context) {
		AddTagHandler(svc, ctx)
	})

	// Datasets
	// Create dataset
	v1.POST("/datasets/:dataset_name", func(ctx *gin.Context) {
		CreateDatasetHandler(svc, ctx)
	})

	// Batch
	// Post assets
	v1.POST("/batch/assets", func(ctx *gin.Context) {
		CreateAssetsBatchHandler(svc, ctx)
	})

	// Get assets ingress Urls
	v1.GET("/batch/assets/ingress", func(ctx *gin.Context) {
		GetAssetIngressURLsBatchHandler(svc, ctx)
	})
}
