package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlanType string

const (
	FreePlan       PlanType = "free"
	BasicPlan      PlanType = "basic"
	ProfessionalPlan PlanType = "professional"
	EnterprisePlan PlanType = "enterprise"
)

type Subscription struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID            uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	PlanType          PlanType       `gorm:"type:varchar(50);not null" json:"plan_type"`
	RequestsPerMonth  int            `json:"requests_per_month"`
	RequestsUsed      int            `gorm:"default:0" json:"requests_used"`
	Price             float64        `json:"price"`
	IsActive          bool           `gorm:"default:true" json:"is_active"`
	StartsAt          time.Time      `json:"starts_at"`
	EndsAt            *time.Time     `json:"ends_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
	User              User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *Subscription) HasRequestsAvailable() bool {
	return s.RequestsUsed < s.RequestsPerMonth
}

func (s *Subscription) IsExpired() bool {
	if s.EndsAt == nil {
		return false
	}
	return time.Now().After(*s.EndsAt)
}
