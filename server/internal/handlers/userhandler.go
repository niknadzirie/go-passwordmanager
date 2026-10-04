package handlers

import (
	"errors"
	"net/http"
	"passwordmanager-server/internal/models"
	"passwordmanager-server/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserHandler struct
type UserHandler struct {
	Service *services.UserService
}

// CreateUser create user
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

// GetUser get user
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	result, err := h.Service.GetUserByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			handleError(c, http.StatusNotFound, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, result)

}

// GetAllUsers get user
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	result, err := h.Service.GetAllUsers()
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteUser delete user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = h.Service.RemoveUser(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			handleError(c, http.StatusNotFound, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// UpdateUser update user
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var input models.UpdateUserDTO

	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	result, err := h.Service.UpdateUser(uint(id), input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			handleError(c, http.StatusBadRequest, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
