package api

import (
	"github.com/gin-gonic/gin"

	"github.com/UnivocalX/aether/internal/api/middleware"
	"github.com/UnivocalX/aether/internal/api/v1/handlers"
	"github.com/UnivocalX/aether/pkg/registry"
)

type Config struct {
	Registry *registry.Config
	Port     string
	Prod     bool
}

func New(cfg *Config) (*gin.Engine, error) {
	if cfg.Prod {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create engine
	engine, err := registry.New(cfg.Registry)
	if err != nil {
		return nil, err
	}

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())

	// Base routes
	router.GET("/health", handlers.HealthCheck())

	// API v1 routes
	v1 := router.Group("v1")
	v1.GET("/health", handlers.HealthCheck())

	// Registry endpoints
	registryHandler := handlers.NewRegistryHandler(engine)
	v1.POST("/assets", registryHandler.BatchCreateAsset)
	v1.POST("/assets/:sha256", registryHandler.CreateAsset)
	v1.GET("/assets/:sha256", registryHandler.GetAsset)
	v1.GET("/assets", registryHandler.ListAssets)

	v1.POST("/tags/:name", registryHandler.CreateTag)
	v1.GET("/tags/:name", registryHandler.GetTag)
	v1.GET("/tags", registryHandler.ListTags)

	return router, nil
}