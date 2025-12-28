package v1

import (
	"log/slog"

	"github.com/UnivocalX/aether/internal/api/services/data"
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
	v1.GET("/assets/:sha256", func(ctx *gin.Context) {
		HandleGetAsset(svc, ctx)
	})

	// Get a specific asset tags
	v1.GET("/assets/:sha256/tags", func(ctx *gin.Context) {
		HandleGetAssetTags(svc, ctx)
	})

	// Get a specific asset presigned Url
	v1.GET("/assets/:sha256/presignedUrl", func(ctx *gin.Context) {
		HandleGetAssetUrl(svc, ctx)
	})

	// Post an asset
	v1.POST("/assets/:sha256", func(ctx *gin.Context) {
		HandleCreateAsset(svc, ctx)
	})

	// Tag a specific asset 
	v1.PUT("/assets/:sha256/tags/:name", func(ctx *gin.Context) {
		HandleTaggingAsset(svc, ctx)
	})

	// UnTag a specific asset 
	v1.DELETE("/assets/:sha256/tags/:name", func(ctx *gin.Context) {
		HandleUntaggingAsset(svc, ctx)
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

}
