package handlers

import (
	"log/slog"

	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/gin-gonic/gin"

	"github.com/UnivocalX/aether/internal/api/models"
)

type DataHandler struct {
	registry *registry.Engine
}

func NewDataHandler(reg *registry.Engine) *DataHandler {
	return &DataHandler{registry: reg}
}

func (handler *DataHandler) Create(c *gin.Context) {
	var req models.CreateDataRequest

	// Bind URI parameters
	if err := c.ShouldBindUri(&req); err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to bind URI", "error", err)
		BadRequest(c, "Invalid SHA256 in path parameter")
		return
	}

	// Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to bind JSON", "error", err)
		BadRequest(c, "Invalid request payload")
		return
	}

	if err := req.Normalize(); err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to normalize request", "error", err)
		BadRequest(c, err.Error())
		return
	}

	// Business logic
	url, err := handler.registry.PutURL(c.Request.Context(), req.SHA256)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to generate presigned URL",
			"error", err, "sha256", req.SHA256,
		)
		InternalError(c, "Failed to generate upload URL")
		return
	}

	// Success response
	data := models.CreateResponseData{
		SHA256:       req.SHA256,
		PresignedURL: url,
		Expiry:       handler.registry.Config.Storage.TTL.String(),
	}

	slog.InfoContext(c.Request.Context(), "Generated presigned URL", "sha256", req.SHA256, "Expiry", data.Expiry)
	Created(c, "Presigned URL generated successfully", data)
}
