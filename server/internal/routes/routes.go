// Package routes setup routing
package routes

import (
	"passwordmanager-server/internal/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRouter setup gin router
func SetupRouter(handler *handlers.UserHandler) *gin.Engine {
	r := gin.Default()

	userRoutes := r.Group("/users")
	{
		// GET all user
		userRoutes.GET("", handler.GetAllUsers)
		// GET user by id
		userRoutes.GET("/:id", handler.GetUser)
		// Create user
		userRoutes.POST("", handler.CreateUser)
		// Delete user by id
		userRoutes.DELETE("/:id", handler.DeleteUser)
		// Update user
		userRoutes.PUT("/:id", handler.UpdateUser)
	}

	return r
}
