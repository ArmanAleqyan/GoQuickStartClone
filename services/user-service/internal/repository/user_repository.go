package repository

import (
	"ironnode/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateAPIKey(apiKey *models.APIKey) error
	GetAPIKeysByUser(userID uuid.UUID) ([]*models.APIKey, error)
	GetAPIKeyByKey(key string) (*models.APIKey, error)
	DeleteAPIKey(id uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateAPIKey(apiKey *models.APIKey) error {
	return r.db.Create(apiKey).Error
}

func (r *userRepository) GetAPIKeysByUser(userID uuid.UUID) ([]*models.APIKey, error) {
	var keys []*models.APIKey
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).Find(&keys).Error
	return keys, err
}

func (r *userRepository) GetAPIKeyByKey(key string) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.Where("key = ? AND is_active = ?", key, true).First(&apiKey).Error
	return &apiKey, err
}

func (r *userRepository) DeleteAPIKey(id uuid.UUID) error {
	return r.db.Delete(&models.APIKey{}, id).Error
}
