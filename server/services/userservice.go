package services

import (
	"passwordmanager-server/models"
	"passwordmanager-server/repositories"
)

type UserService struct {
	Repo *repositories.UserRepository
}

func (s *UserService) RegisterUser(dto models.CreateUserDTO) (*models.User, error) {
	newUser := &models.User{
		Name:  dto.Name,
		Email: dto.Email,
	}

	return s.Repo.Save(newUser)
}
