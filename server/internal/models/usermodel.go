package models

import (
	"time"
)

type User struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	Name      string    `json:"name"`
	Email     string    `json:"email" gorm:"unique;not null"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUserDTO struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type GetUserDTO struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateUserDTO struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}
