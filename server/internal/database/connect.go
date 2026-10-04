// Package database provides tools for connecting to and interacting with the database.
package database

import (
	"passwordmanager-server/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// InitDB initializes the database connection pool and runs migrations.
func InitDB() *gorm.DB {
	database, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// Migrate the schema
	err = database.AutoMigrate(&models.User{})
	if err != nil {
		panic("failed to migrate database")
	}

	return database
}
