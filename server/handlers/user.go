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

func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.Service.GetUser(id)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.User

	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	user, err := h.Service.CreateUser(c.Request.Context(), &req)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}
