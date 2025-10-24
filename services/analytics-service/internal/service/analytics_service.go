package service

import (
	"time"

	"ironnode/pkg/models"
	"ironnode/services/analytics-service/internal/repository"

	"github.com/google/uuid"
)

type AnalyticsService interface {
	LogRequest(userID, apiKeyID uuid.UUID, blockchain, method, endpoint string, statusCode int, responseTime, requestSize, responseSize int64, ipAddress, userAgent, errorMsg string) error
	GetRequestHistory(userID uuid.UUID, limit int) ([]*models.RequestLog, error)
	GetUsageStats(userID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error)
}

type analyticsService struct {
	repo repository.AnalyticsRepository
}

func NewAnalyticsService(repo repository.AnalyticsRepository) AnalyticsService {
	return &analyticsService{repo: repo}
}

func (s *analyticsService) LogRequest(
	userID, apiKeyID uuid.UUID,
	blockchain, method, endpoint string,
	statusCode int,
	responseTime, requestSize, responseSize int64,
	ipAddress, userAgent, errorMsg string,
) error {
	log := &models.RequestLog{
		UserID:       userID,
		APIKeyID:     apiKeyID,
		Blockchain:   blockchain,
		Method:       method,
		Endpoint:     endpoint,
		StatusCode:   statusCode,
		ResponseTime: responseTime,
		RequestSize:  requestSize,
		ResponseSize: responseSize,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Error:        errorMsg,
	}

	return s.repo.LogRequest(log)
}

func (s *analyticsService) GetRequestHistory(userID uuid.UUID, limit int) ([]*models.RequestLog, error) {
	return s.repo.GetRequestsByUser(userID, limit)
}

func (s *analyticsService) GetUsageStats(userID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	return s.repo.GetUsageStats(userID, startDate, endDate)
}
