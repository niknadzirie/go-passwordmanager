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
	database.ConnectDatabase()

	// GIN Routing
	r := gin.Default()

	// Register
	repo := &repositories.UserRepository{}
	service := &services.UserService{Repo: repo}
	handler := &handlers.UserHandler{Service: service}

	r.POST("/users", handler.CreateUser)

	r.Run(":8080")

}
