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
