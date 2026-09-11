package repositories

import (
	"context"
	"errors"
	"passwordmanager-server/database"
	"passwordmanager-server/models"
)

type UserRepository struct{}

func (r *UserRepository) FindById(id string) (models.User, error) {

	var user models.User

	database.DB.First(&user, id)

	return user, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {

	result := database.DB.Create(user)

	if result.Error != nil {
		return errors.New("could not save user: " + result.Error.Error())
	}

	return nil

}
