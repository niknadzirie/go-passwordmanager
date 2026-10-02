// Package handlers handle error
package handlers

import (
	"passwordmanager-server/internal/models"

	"github.com/gin-gonic/gin"
)

func handleError(c *gin.Context, status int, err error) {
	c.JSON(status, models.ErrorResponse{Code: status, Message: err.Error()})
}
