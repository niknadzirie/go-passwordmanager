package main

import (
	"passwordmanager-server/internal/database"
	"passwordmanager-server/internal/handlers"
	"passwordmanager-server/internal/repositories"
	"passwordmanager-server/internal/routes"
	"passwordmanager-server/internal/services"
)

func main() {
	// Connect database
	db := database.InitDB()

	// Register
	repo := &repositories.UserRepository{Db: db}
	service := &services.UserService{Repo: repo}
	handler := &handlers.UserHandler{Service: service}

	r := routes.SetupRouter(handler)

	// Run the application
	r.Run(":8080")

}
