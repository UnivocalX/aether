package v1

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/api/dto"
	"github.com/UnivocalX/aether/internal/api/services/data"
	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/gin-gonic/gin"
)

type TagPostUriParams struct {
	Name string `uri:"name" binding:"required,min=1,max=100"`
}

type TagPostRequest struct {
	TagPostUriParams
}

type TagPostResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func HandleCreateTag(svc *data.Service, ctx *gin.Context) {
	var req TagPostRequest

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&req.TagPostUriParams); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid URI parameters", "error", err.Error())
		dto.BadRequest(ctx, "Invalid Name in URI")
		return
	}

	// Execute business logic
	result := svc.CreateTag(ctx.Request.Context(), data.CreateTagParams{Name: req.Name})
	if result.Err != nil {
		handleCreateTagError(ctx, result.Err, req.Name)
		return
	}

	// Success response
	slog.InfoContext(ctx.Request.Context(), "tag created successfully",
		"tagName", result.Tag.Name,
		"tagId", result.Tag.ID,
	)

	response := &TagPostResponse{ID: result.Tag.ID, Name: result.Tag.Name}
	dto.Created(ctx, "Successfully created tag", response)
}

// handleCreateAssetError maps business errors to HTTP responses
func handleCreateTagError(ctx *gin.Context, err error, name string) {
	switch {
	case errors.Is(err, registry.ErrValidation):
		dto.BadRequest(ctx, err.Error())

	case errors.Is(err, data.ErrTagAlreadyExists):
		dto.Conflict(ctx, fmt.Sprintf("tag %s already exists.", name))

	default:
		slog.ErrorContext(ctx.Request.Context(), "Failed to create tag",
			"error", err.Error(),
			"tag", name,
		)
		dto.InternalError(ctx, "Failed to create tag")
	}
}
