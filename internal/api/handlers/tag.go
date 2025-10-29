package handlers

import (
	"context"
	"errors"
	"log/slog"

	"github.com/UnivocalX/aether/pkg/registry/models"
	"github.com/gin-gonic/gin"

	"github.com/UnivocalX/aether/internal/api/schemas"
)

// Custom errors
var (
	ErrTagAlreadyExists = errors.New("tag already exists")
)

// CreateTag handles the HTTP request/response cycle
func (handler *RegistryHandler) CreateTag(ctx *gin.Context) {
	var req schemas.CreateTagRequest

	// Bind and validate request
	if err := ctx.ShouldBindUri(req); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid URI parameter", "error", err)
		BadRequest(ctx, "Invalid name in path parameter")
		return
	}

	// Execute business logic
	response, err := handler.createTag(ctx.Request.Context(), req)
	if err != nil {
		handler.handleCreateTagError(ctx, err, req.Name)
		return
	}

	// Success response
	slog.InfoContext(ctx.Request.Context(), "Tag created successfully",
		"tag_name", req.Name,
		"tag_id", response.ID,
	)
	Created(ctx, "Tag created successfully", response)
}

// createTag contains the core business logic (no HTTP concerns)
func (handler *RegistryHandler) createTag(ctx context.Context, req schemas.CreateTagRequest) (*schemas.CreateTagResponse, error) {
	// Check if tag already exists
	existingTag, err := handler.registry.GetTagRecord(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	if existingTag != nil {
		slog.InfoContext(ctx, "Tag already exists",
			"tag_name", existingTag.Name,
			"tag_id", existingTag.ID,
		)
		return nil, ErrTagAlreadyExists
	}

	// Create new tag
	tag, err := handler.registry.CreateTagRecord(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	return &schemas.CreateTagResponse{
		Name: tag.Name,
		ID:   tag.ID,
	}, nil
}

// handleCreateTagError maps business errors to HTTP responses
func (handler *RegistryHandler) handleCreateTagError(ctx *gin.Context, err error, tagName string) {
	switch {
	case errors.Is(err, ErrTagAlreadyExists):
		Conflict(ctx, "Tag already exists")

	case errors.Is(err, models.ErrValidation):
		BadRequest(ctx, err.Error())

	default:
		slog.ErrorContext(ctx.Request.Context(), "Failed to create tag",
			"error", err,
			"tag_name", tagName,
		)
		InternalError(ctx, "Failed to create tag")
	}
}
