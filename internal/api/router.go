package api

import (
	"log/slog"
	"fmt"
	
	"github.com/gin-gonic/gin"

	"github.com/UnivocalX/aether/internal/api/handlers"
	"github.com/UnivocalX/aether/internal/api/middleware"
	"github.com/UnivocalX/aether/internal/logging"
	"github.com/UnivocalX/aether/pkg/registry"
)

type Config struct {
	Registry *registry.Config
	Logging  *logging.Config
	Port     string
}

func (cfg Config) String() string {
    return fmt.Sprintf(
        "Config{Registry:%v, Logging:%v, Port:%q}",
        cfg.Registry, cfg.Logging, cfg.Port,
    )
}

// New returns a router that uses your logging middleware instead of Gin's default logger.
func New(cfg *Config) (*gin.Engine, error) {
	slog.Debug("Setting up new API router")

	if cfg.Logging.Prod {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create engine
	engine, err := registry.New(cfg.Registry)
	if err != nil {
		return nil, err
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(cfg.Logging))

	// API v1 routes
	v1 := router.Group("v1")
	v1.GET("/health", handlers.HealthCheck())

	// Data endpoints
	registryHandler := handlers.NewRegistryHandler(engine)
	v1.POST("/data/:sha256", registryHandler.CreateAsset)
	v1.POST("/tag/:name", registryHandler.CreateTag)

	return router, nil
}
