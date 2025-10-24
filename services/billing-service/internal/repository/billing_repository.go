package repository

import (
	"ironnode/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BillingRepository interface {
	CreateSubscription(subscription *models.Subscription) error
	GetSubscriptionByUser(userID uuid.UUID) (*models.Subscription, error)
	UpdateSubscription(subscription *models.Subscription) error
	IncrementUsage(userID uuid.UUID) error
}

type billingRepository struct {
	db *gorm.DB
}

func NewBillingRepository(db *gorm.DB) BillingRepository {
	return &billingRepository{db: db}
}

func (r *billingRepository) CreateSubscription(subscription *models.Subscription) error {
	return r.db.Create(subscription).Error
}

func (r *billingRepository) GetSubscriptionByUser(userID uuid.UUID) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).First(&subscription).Error
	return &subscription, err
}

func (r *billingRepository) UpdateSubscription(subscription *models.Subscription) error {
	return r.db.Save(subscription).Error
}

func (r *billingRepository) IncrementUsage(userID uuid.UUID) error {
	return r.db.Model(&models.Subscription{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Update("requests_used", gorm.Expr("requests_used + ?", 1)).Error
}
