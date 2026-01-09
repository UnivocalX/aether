package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/gin-gonic/gin"
)

type CreateDatasetPayload struct {
	Description string `json:"description" binding:"omitempty,max=1000"`
}

type CreateDatasetResponseData struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func CreateDatasetHandler(svc *data.Service, ctx *gin.Context) {
	var uri dto.DatasetUri
	var payload CreateDatasetPayload

	// Bind URI parameters
	if err := ctx.ShouldBindUri(&uri); err != nil {
		slog.ErrorContext(ctx.Request.Context(), "Invalid URI parameters", "error", err.Error())
		dto.HandleErrorResponse(
			ctx,
			"failed to create dataset",
			fmt.Errorf("%w, %w", dto.ErrInvalidUri, err),
		)
		return
	}

	// Bind JSON payload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to create dataset",
			fmt.Errorf("%w, %w", dto.ErrInvalidPayload, err),
		)
		return
	}

	dsv, err := svc.CreateDataset(ctx.Request.Context(), uri.DatasetName, payload.Description)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to create dataset", err)
		return
	}

	data := &CreateDatasetResponseData{
		ID:          dsv.DatasetID,
		Name:        dsv.Dataset.Name,
		Description: dsv.Dataset.Description,
	}
	response := dto.NewResponse(ctx, "dataset created successfully").WithData(data)

	// Success response
	slog.InfoContext(ctx.Request.Context(), response.Message,
		"dataset", dsv,
	)

	response.OK(ctx)
}
