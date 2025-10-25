package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/UnivocalX/aether/internal/api/handlers"
	"github.com/UnivocalX/aether/internal/api/middleware"
	"github.com/UnivocalX/aether/pkg/registry"
)

type Options struct {
	Registry   *registry.Options
	Production bool
}

// New returns a router that uses your logging middleware instead of Gin's default logger.
func New(opt *Options) (*gin.Engine, error) {
	slog.Debug("Setting up new API router")
	client, err := registry.New(opt.Registry)

	if err != nil {
		return nil, err
	}

	if opt.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logging())

	// API v1 routes
	v1 := router.Group("v1")
	v1.GET("/health", handlers.HealthCheck())

	// Data endpoints
	dataHandler := handlers.NewDataHandler(client)
	v1.POST("/data/:sha256", dataHandler.Create)

	return router, nil
}
