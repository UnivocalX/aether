package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/gin-gonic/gin"
)

type TagUriParams struct {
	Name string `uri:"name" binding:"required,min=1,max=100"`
}

type TagPostRequest struct {
	TagUriParams
}

type TagPostResponseData struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func HandleCreateTag(svc *data.Service, ctx *gin.Context) {
	var req TagPostRequest

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&req.TagUriParams); err != nil {
		dto.HandleErrorResponse(
			ctx, 
			"failed to create new tag",
			fmt.Errorf("%w: %w", dto.ErrInvalidUri, err),
		)
		return
	}

	// Execute business logic
	result := svc.CreateTag(ctx.Request.Context(), data.CreateTagParams{Name: req.Name})
	if result.Err != nil {
		dto.HandleErrorResponse(ctx, "failed to create new tag", result.Err)
		return
	}

	// Success response
	response := dto.NewResponse(ctx, "tag created successfully")
	response.Data = &TagPostResponseData{ID: result.Tag.ID, Name: result.Tag.Name}

	slog.InfoContext(ctx.Request.Context(), response.Message,
		"tagName", result.Tag.Name,
		"tagId", result.Tag.ID,
	)

	response.Created(ctx)
}