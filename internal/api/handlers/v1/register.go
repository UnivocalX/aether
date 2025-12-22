package v1

import (
	"log/slog"

	"github.com/UnivocalX/aether/internal/api/services/data"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r gin.IRouter, svc *data.Service) {
	slog.Info("Registering V1 routes")

	v1 := r.Group("/v1")
	v1.GET("/assets/:sha256", func(ctx *gin.Context) {
		HandleGetAsset(svc, ctx)
	})

	v1.GET("/assets/:sha256/tags", func(ctx *gin.Context) {
		HandleGetAssetTags(svc, ctx)
	})

	v1.POST("/assets/:sha256", func(ctx *gin.Context) {
		HandleCreateAsset(svc, ctx)
	})
	
	v1.POST("tags/:name", func(ctx *gin.Context) {
		HandleCreateTag(svc, ctx)
	})

	v1.PUT("/assets/:sha256/tags/:name", func(ctx *gin.Context) {
		HandleTaggingAsset(svc, ctx)
	})
}
