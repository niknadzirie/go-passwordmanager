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

func (r *UserRepository) DeleteById(id uint) error {
	result := r.Db.Delete(&models.User{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *UserRepository) UpdateUser(user *models.User) (*models.User, error) {
	result := r.Db.Save(user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}
