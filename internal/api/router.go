package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/UnivocalX/aether/internal/api/handlers"
	"github.com/UnivocalX/aether/internal/api/middleware"
	"github.com/UnivocalX/aether/pkg/registry"
)

// New returns a router that uses your logging middleware instead of Gin's default logger.
func New(engine *registry.Engine, prod bool) *gin.Engine {
	slog.Info("Setting up new API router")

	if prod {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logging())

	// API v1 routes
	v1 := router.Group("v1")
	v1.GET("/health", handlers.HealthCheck())

	// Data endpoints
	registryHandler := handlers.NewRegistryHandler(engine)
	v1.POST("/data/:sha256", registryHandler.CreateAsset)
	v1.POST("/tag/:name", registryHandler.CreateTag)

	return router
}
