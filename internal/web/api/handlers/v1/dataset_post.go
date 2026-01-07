package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/gin-gonic/gin"
)

type DatasetUriParams struct {
	Name string `uri:"name" binding:"required,len=100"`
}

type DatasetPostPayload struct {
	description string `uri:"description" binding:"omitempty,max=1000"`
}

type DatasetPostRequest struct {
	DatasetUriParams
	DatasetPostPayload
}

type DatasetPostResponseData struct {
	ID          uint
	Name        string
	Description string
}

func HandleCreateDataset(svc *data.Service, ctx *gin.Context) {
	var req DatasetPostRequest

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&req.DatasetUriParams); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid URI parameters", "error", err.Error())
		dto.HandleErrorResponse(
			ctx,
			"failed to create dataset",
			fmt.Errorf("%w: %w", dto.ErrInvalidUri, err),
		)
		return
	}

	dsv, err := svc.CreateDataset(ctx.Request.Context(), req.Name, req.description)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to create dataset", err)
		return
	}

	response := dto.NewResponse(ctx, "dataset created successfully")
	response.Data = &DatasetPostResponseData{
		ID: dsv.DatasetID, 
		Name: dsv.Dataset.Name, 
		Description: 
		dsv.Dataset.Description,
	}

	// Success response
	slog.InfoContext(ctx.Request.Context(), response.Message,
		"dataset", dsv,
	)

	response.OK(ctx)
}
