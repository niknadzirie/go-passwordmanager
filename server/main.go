package main

import (
	"passwordmanager-server/database"
	"passwordmanager-server/handlers"
	"passwordmanager-server/repositories"
	"passwordmanager-server/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Connect database
	db := database.InitDB()

	// Register
	repo := &repositories.UserRepository{Db: db}
	service := &services.UserService{Repo: repo}
	handler := &handlers.UserHandler{Service: service}

	// GIN Router
	r := gin.Default()
	// Get all users
	r.GET("/users")
	// Get by id
	r.GET("/users/:id")
	// Create and save user to DB
	r.POST("/users", handler.CreateUser)

	// Run the application
	r.Run(":8080")

}
