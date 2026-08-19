package main

import (
	"context"
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func main() {
	ctx := context.Background()

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Product{})

	prod, err := gorm.G[Product](db).Find(ctx)

	fmt.Println(prod)

}
