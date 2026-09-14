package repositories

import (
	"passwordmanager-server/models"

	"gorm.io/gorm"
)

type CredentialRepository struct {
	Db *gorm.DB
}

func (r *CredentialRepository) Save(credential *models.Credential) (*models.Credential, error) {
	result := r.Db.Create(credential)
	if result.Error != nil {
		return nil, result.Error
	}

	return credential, nil
}
