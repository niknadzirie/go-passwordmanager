// Package services bussines logic (service layer)
package services

import (
	"passwordmanager-server/internal/models"
	"passwordmanager-server/internal/repositories"
)

// UserService struct
type UserService struct {
	Repo *repositories.UserRepository
}

// RegisterUser register user
func (s *UserService) RegisterUser(dto models.CreateUserDTO) (*models.User, error) {
	newUser := &models.User{
		Name:  dto.Name,
		Email: dto.Email,
	}

	return s.Repo.Save(newUser)
}

// GetUserByID get user by id
func (s *UserService) GetUserByID(id uint) (*models.GetUserDTO, error) {
	savedUser, err := s.Repo.FindByID(id)
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

// GetAllUsers get all user
func (s *UserService) GetAllUsers() ([]models.GetUserDTO, error) {
	users, err := s.Repo.FindAll()
	if err != nil {
		return nil, err
	}

	response := make([]models.GetUserDTO, 0)

	for _, user := range users {
		dto := models.GetUserDTO(user)
		response = append(response, dto)
	}

	return response, nil
}

// RemoveUser remove user
func (s *UserService) RemoveUser(id uint) error {
	return s.Repo.DeleteByID(id)
}

// UpdateUser update user
func (s *UserService) UpdateUser(id uint, dto models.UpdateUserDTO) (*models.GetUserDTO, error) {
	existingUser, err := s.Repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	existingUser.Name = dto.Name
	existingUser.Email = dto.Email

	updatedUser, err := s.Repo.UpdateUser(existingUser)
	if err != nil {
		return nil, err
	}

	response := &models.GetUserDTO{
		ID:        updatedUser.ID,
		Name:      updatedUser.Name,
		Email:     updatedUser.Email,
		CreatedAt: updatedUser.CreatedAt,
	}

	return response, nil
}

func toUserDTO(u)
