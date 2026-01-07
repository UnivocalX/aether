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
		HandleListAssets(svc, ctx)
	})

	// Get a specific asset
	v1.GET("/assets/:checksum", func(ctx *gin.Context) {
		HandleGetAsset(svc, ctx)
	})

	// Get a specific asset tags
	v1.GET("/assets/:checksum/tags", func(ctx *gin.Context) {
		HandleGetAssetTags(svc, ctx)
	})

	// Get an asset put presigned Url
	v1.GET("/assets/:checksum/ingress", func(ctx *gin.Context) {
		HandleGetIngressUrl(svc, ctx)
	})

	// Post an asset
	v1.POST("/assets/:checksum", func(ctx *gin.Context) {
		HandleCreateAsset(svc, ctx)
	})

	// Post assets
	v1.POST("/assets", func(ctx *gin.Context) {
		HandleCreateAssetBatch(svc, ctx)
	})

	// Tag a specific asset
	v1.PUT("/assets/:checksum/tags/:name", func(ctx *gin.Context) {
		HandleTaggingAsset(svc, ctx)
	})

	// UnTag a specific asset
	v1.DELETE("/assets/:checksum/tags/:name", func(ctx *gin.Context) {
		HandleUntagAsset(svc, ctx)
	})

	// Tags
	// List tag assets
	v1.GET("tags/:name/assets", func(ctx *gin.Context) {
		HandleGetTagAssets(svc, ctx)
	})

	// Create tag
	v1.POST("tags/:name", func(ctx *gin.Context) {
		HandleCreateTag(svc, ctx)
	})

	// Datasets
	v1.POST("datasets/:name", func(ctx *gin.Context) {
		HandleCreateDataset(svc, ctx)
	})
}
