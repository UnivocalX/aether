package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/registry"
	"github.com/UnivocalX/aether/pkg/web/api/dto"
	"github.com/UnivocalX/aether/pkg/web/services/data"
	"github.com/gin-gonic/gin"
)

type CreateTagRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

type AddTagResponse struct {
	dto.Response
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func CreateTagHandler(svc *data.Service, ctx *gin.Context) {
	var request CreateTagRequest

	// Bind JSON payload
	if err := ctx.ShouldBindJSON(&request); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to add tag",
			fmt.Errorf("%w, %w", dto.ErrInvalidUri, err),
		)
		return
	}

	// Execute business logic
	tag, err := svc.CreateTag(ctx.Request.Context(), request.Name)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to add tag", err)
		return
	}

	response := newAddTagResponse(ctx, tag)
	response.Created(ctx)
}

func newAddTagResponse(ctx *gin.Context, tag *registry.Tag) AddTagResponse {
	response := AddTagResponse{
		Response: *dto.NewResponse(ctx, "tag added successfully"),
		ID:       tag.ID,
		Name:     tag.Name,
	}
	slog.InfoContext(ctx.Request.Context(), response.Msg,
		"Id", tag.ID,
		"name", tag.Name,
	)
	return response
}
