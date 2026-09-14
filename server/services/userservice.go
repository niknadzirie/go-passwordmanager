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

func (s *UserService) GetUserById(id uint) (*models.GetUserDTO, error) {
	savedUser, err := s.Repo.FindById(id)
	if err != nil {
		return nil, err
	}

	response := &models.GetUserDTO{
		ID:        savedUser.ID,
		Name:      savedUser.Name,
		Email:     savedUser.Email,
		CreatedAt: savedUser.CreatedAt,
	}

	return response, nil

}

func (s *UserService) GetAllUsers() ([]models.GetUserDTO, error) {
	users, err := s.Repo.FindAll()
	if err != nil {
		return nil, err
	}

	response := make([]models.GetUserDTO, 0)

	for _, user := range users {
		dto := models.GetUserDTO{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		}
		response = append(response, dto)
	}

	return response, nil
}
