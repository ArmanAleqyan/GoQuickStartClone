package service

import (
	"errors"
	"time"

	"ironnode/pkg/models"
	"ironnode/services/billing-service/internal/repository"

	"github.com/google/uuid"
)

type BillingService interface {
	CreateSubscription(userID uuid.UUID, planType models.PlanType) (*models.Subscription, error)
	GetSubscription(userID uuid.UUID) (*models.Subscription, error)
	UpdateSubscription(userID uuid.UUID, planType models.PlanType) error
	CheckQuota(userID uuid.UUID) (bool, error)
	IncrementUsage(userID uuid.UUID) error
}

type billingService struct {
	repo repository.BillingRepository
}

func NewBillingService(repo repository.BillingRepository) BillingService {
	return &billingService{repo: repo}
}

func (s *billingService) CreateSubscription(userID uuid.UUID, planType models.PlanType) (*models.Subscription, error) {
	// Define plan limits
	planLimits := map[models.PlanType]struct {
		requests int
		price    float64
	}{
		models.FreePlan:          {requests: 10000, price: 0},
		models.BasicPlan:         {requests: 100000, price: 29.99},
		models.ProfessionalPlan:  {requests: 1000000, price: 99.99},
		models.EnterprisePlan:    {requests: 10000000, price: 499.99},
	}

	plan, exists := planLimits[planType]
	if !exists {
		return nil, errors.New("invalid plan type")
	}

	subscription := &models.Subscription{
		UserID:           userID,
		PlanType:         planType,
		RequestsPerMonth: plan.requests,
		Price:            plan.price,
		IsActive:         true,
		StartsAt:         time.Now(),
	}

	if err := s.repo.CreateSubscription(subscription); err != nil {
		return nil, err
	}

	return subscription, nil
}

func (s *billingService) GetSubscription(userID uuid.UUID) (*models.Subscription, error) {
	return s.repo.GetSubscriptionByUser(userID)
}

func (s *billingService) UpdateSubscription(userID uuid.UUID, planType models.PlanType) error {
	subscription, err := s.repo.GetSubscriptionByUser(userID)
	if err != nil {
		return err
	}

	subscription.PlanType = planType
	// Update limits based on new plan
	// (similar logic as CreateSubscription)

	return s.repo.UpdateSubscription(subscription)
}

func (s *billingService) CheckQuota(userID uuid.UUID) (bool, error) {
	subscription, err := s.repo.GetSubscriptionByUser(userID)
	if err != nil {
		return false, err
	}

	if subscription.IsExpired() {
		return false, errors.New("subscription expired")
	}

	return subscription.HasRequestsAvailable(), nil
}

func (s *billingService) IncrementUsage(userID uuid.UUID) error {
	return s.repo.IncrementUsage(userID)
}
