package handlers

import (
	"net/http"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/UnivocalX/aether/internal/api/models"
)

func HealthCheck(c *gin.Context) {
	response := models.Response{
		Message: "Service is healthy",
		Data:    gin.H{"status": "ok", "timestamp": time.Now()},
	}
	c.JSON(http.StatusOK, response)
}