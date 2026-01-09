package v1

import (
	"fmt"

	"github.com/UnivocalX/aether/internal/web/api/dto"
	"github.com/UnivocalX/aether/internal/web/services/data"
	"github.com/gin-gonic/gin"
)

func GetAssetIngressURLsBatchHandler(svc *data.Service, ctx *gin.Context) {
	dto.HandleErrorResponse(
		ctx,
		"Route not implemented",
		fmt.Errorf("this endpoint is a placeholder: implementation in progress"),
	)
}
