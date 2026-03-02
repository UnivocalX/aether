package v1

import (
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/internal/registry"
	"github.com/UnivocalX/aether/pkg/web/api/dto"
	"github.com/UnivocalX/aether/pkg/web/services/data"
	"github.com/gin-gonic/gin"
)

type CreateDatasetRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"omitempty,max=1000"`
}

type CreateDatasetResponse struct {
	dto.Response
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func CreateDatasetHandler(svc *data.Service, ctx *gin.Context) {
	var payload CreateDatasetRequest

	// Bind JSON payload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		dto.HandleErrorResponse(
			ctx,
			"failed to create dataset",
			fmt.Errorf("%w, %w", dto.ErrInvalidPayload, err),
		)
		return
	}

	dsv, err := svc.CreateDataset(ctx.Request.Context(), payload.Name, payload.Description)
	if err != nil {
		dto.HandleErrorResponse(ctx, "failed to create dataset", err)
		return
	}

	response := newCreateDatasetResponse(ctx, dsv)
	response.Created(ctx)
}

func newCreateDatasetResponse(ctx *gin.Context, dsv *registry.DatasetVersion) CreateDatasetResponse {
	response := CreateDatasetResponse{
		Response:    *dto.NewResponse(ctx, "dataset created successfully"),
		ID:          dsv.DatasetID,
		Name:        dsv.Dataset.Name,
		Description: dsv.Dataset.Description,
	}
	slog.InfoContext(ctx.Request.Context(), response.Msg,
		"dataset", dsv,
	)
	return response
}
