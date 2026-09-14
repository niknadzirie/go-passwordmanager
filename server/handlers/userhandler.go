package handlers

import (
	"errors"
	"net/http"
	"passwordmanager-server/models"
	"passwordmanager-server/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	result, err := h.Service.GetUserById(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			handleError(c, http.StatusNotFound, err)
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, result)

}

func (h *UserHandler) GetAllUser(c *gin.Context) {
	result, err := h.Service.GetAllUsers()
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
