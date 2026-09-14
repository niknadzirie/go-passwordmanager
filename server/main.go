package main

import (
	"passwordmanager-server/database"
	"passwordmanager-server/handlers"
	"passwordmanager-server/repositories"
	"passwordmanager-server/routes"
	"passwordmanager-server/services"
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
