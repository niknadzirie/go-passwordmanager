package services

import (
	"context"
	"errors"
	"passwordmanager-server/models"
	"passwordmanager-server/repositories"
)

type UserService struct {
	Repo *repositories.UserRepository
}

func (s *UserService) GetUser(id string) (models.User, error) {
	return s.Repo.FindById(id)
}

func (s *UserService) CreateUser(ctx context.Context, req *models.User) (*models.User, error) {

	user := &models.User{
		Name: req.Name,
	}

	if err := s.Repo.CreateUser(ctx, user); err != nil {
		return nil, errors.New("Internal server error")
	}

	return user, nil
}
