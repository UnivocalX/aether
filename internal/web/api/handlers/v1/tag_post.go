package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/gin-gonic/gin"
)

type CreateTagPayload struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

type AddTagResponseData struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func CreateTagHandler(svc *data.Service, ctx *gin.Context) {
	var payload CreateTagPayload

	// Bind JSON payload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to add tag",
			fmt.Errorf("%w, %w", dto.ErrInvalidUri, err),
		)
		return
	}

	// Execute business logic
	tag, err := svc.CreateTag(ctx.Request.Context(), payload.Name)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to add tag", err)
		return
	}

	// Success response
	data := &AddTagResponseData{ID: tag.ID, Name: tag.Name}
	response := dto.NewResponse(ctx, "tag added successfully").WithData(data)

	slog.InfoContext(ctx.Request.Context(), response.Message,
		"name", tag.Name,
		"Id", tag.ID,
	)

	response.Created(ctx)
}
