package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	v1 "github.com/UnivocalX/aether/internal/api/handlers/v1"
	"github.com/UnivocalX/aether/internal/api/middleware"
	"github.com/UnivocalX/aether/internal/api/services/data"
	"github.com/UnivocalX/aether/pkg/registry"
)

type Server struct {
	Registry *registry.Engine
	Router   *gin.Engine
	DataSvc  *data.Service
	Prod     bool
}

func (s *Server) Run(port string) error {
	slog.Info("Starting server...", "port", port, "production", s.Prod)

	httpServer := &http.Server{
		Addr:           ":" + port,
		Handler:        s.Router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return httpServer.ListenAndServe()
}

func New(prod bool, engine *registry.Engine) *Server {
	// set gin mode
	if prod {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())

	server := &Server{
		Registry: engine,
		Router:   router,
		DataSvc:  data.NewService(engine),
		Prod:     prod,
	}

	// Register Routes
	v1.RegisterRoutes(server.Router, server.DataSvc)
	return server
}
