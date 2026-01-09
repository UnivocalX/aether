package web

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	v1 "github.com/UnivocalX/aether/internal/web/api/handlers/v1"
	"github.com/UnivocalX/aether/internal/web/middleware"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/UnivocalX/aether/pkg/registry"
)

var (
	MaxMultipartMemory int64 = 8 << 20 // 8 MiB
	MaxRequestSize     int64 = 4 << 20 // 4 MiB for JSON batch requests
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

func NewServer(prod bool, engine *registry.Engine) *Server {
	// set gin mode
	if prod {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.MaxRequestSizeLimit(MaxRequestSize))
	router.MaxMultipartMemory = MaxMultipartMemory

	server := &Server{
		Registry: engine,
		Router:   router,
		DataSvc:  data.NewService(engine),
		Prod:     prod,
	}

	// Register Routes
	server.RegisterRoutes()
	return server
}

func (s *Server) RegisterRoutes() {
	slog.Info("Registering API routes")
	api := s.Router.Group("/api")
	v1.RegisterRoutes(api, s.DataSvc)
}
