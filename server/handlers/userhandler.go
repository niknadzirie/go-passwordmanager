package handlers

import (
	"net/http"
	"passwordmanager-server/models"
	"passwordmanager-server/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	Service *services.UserService
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var input models.CreateUserDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	result, err := h.Service.RegisterUser(input)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}
