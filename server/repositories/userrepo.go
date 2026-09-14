package repositories

import (
	"passwordmanager-server/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	Db *gorm.DB
}

func (r *UserRepository) Save(user *models.User) (*models.User, error) {
	result := r.Db.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

func (r *UserRepository) FindById(id uint) (*models.User, error) {
	var user models.User

	result := r.Db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil

}

func (r *UserRepository) FindAll() ([]models.User, error) {
	var users []models.User
	result := r.Db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil

}
