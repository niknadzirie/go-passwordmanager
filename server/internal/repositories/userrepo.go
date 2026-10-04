// Package repositories database related
package repositories

import (
	"passwordmanager-server/internal/models"

	"gorm.io/gorm"
)

// UserRepository struct
type UserRepository struct {
	Db *gorm.DB
}

// Save to database
func (r *UserRepository) Save(user *models.User) (*models.User, error) {
	result := r.Db.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

// FindByID find by id
func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User

	result := r.Db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil

}

// FindAll find all
func (r *UserRepository) FindAll() ([]models.User, error) {
	var users []models.User
	result := r.Db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil

}

// DeleteByID delete by id
func (r *UserRepository) DeleteByID(id uint) error {
	result := r.Db.Delete(&models.User{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// UpdateUser update user
func (r *UserRepository) UpdateUser(user *models.User) (*models.User, error) {
	result := r.Db.Save(user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}
