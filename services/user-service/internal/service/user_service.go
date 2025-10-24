package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"ironnode/pkg/models"
	"ironnode/services/user-service/internal/repository"

	"github.com/google/uuid"
)

type UserService interface {
	CreateAPIKey(userID uuid.UUID, name, description string) (*models.APIKey, error)
	GetAPIKeys(userID uuid.UUID) ([]*models.APIKey, error)
	ValidateAPIKey(key string) (*models.APIKey, error)
	DeleteAPIKey(id uuid.UUID) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateAPIKey(userID uuid.UUID, name, description string) (*models.APIKey, error) {
	// Generate random API key
	key := generateAPIKey()

	apiKey := &models.APIKey{
		UserID:      userID,
		Key:         key,
		Name:        name,
		Description: description,
		IsActive:    true,
	}

	if err := s.repo.CreateAPIKey(apiKey); err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (s *userService) GetAPIKeys(userID uuid.UUID) ([]*models.APIKey, error) {
	return s.repo.GetAPIKeysByUser(userID)
}

func (s *userService) ValidateAPIKey(key string) (*models.APIKey, error) {
	apiKey, err := s.repo.GetAPIKeyByKey(key)
	if err != nil {
		return nil, err
	}

	if apiKey.IsExpired() {
		return nil, fmt.Errorf("API key has expired")
	}

	return apiKey, nil
}

func (s *userService) DeleteAPIKey(id uuid.UUID) error {
	return s.repo.DeleteAPIKey(id)
}

func generateAPIKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return "qn_" + hex.EncodeToString(bytes)
}
