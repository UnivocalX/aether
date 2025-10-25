package api

import (
	"github.com/gin-gonic/gin"

	"github.com/UnivocalX/aether/internal/api/handlers"
	"github.com/UnivocalX/aether/internal/api/middleware"
)

// New returns a router that uses your logging middleware instead of Gin's default logger.
func New(production bool) *gin.Engine {
	if production {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Attach recovery (no logging) and your own logging middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logging())

	// V1 API routes
	v1 := router.Group("v1")
	v1.GET("/health", handlers.HealthCheck)

	return router
}
