package models

import (
	"time"
)

type User struct {
	ID          uint         `json:"id" gorm:"primarykey"`
	Name        string       `json:"name"`
	Email       string       `json:"email" gorm:"unique;not null"`
	Password    string       `json:"-"`
	CreatedAt   time.Time    `json:"created_at"`
	Credentials []Credential `json:"credentials" gorm:"foreignkey:UserID"`
}

type CreateUserDTO struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type GetUserDTO struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateUserDTO struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
