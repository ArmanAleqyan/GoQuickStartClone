package repository

import (
	"time"

	"ironnode/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AnalyticsRepository interface {
	LogRequest(log *models.RequestLog) error
	GetRequestsByUser(userID uuid.UUID, limit int) ([]*models.RequestLog, error)
	GetUsageStats(userID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error)
}

type analyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) LogRequest(log *models.RequestLog) error {
	return r.db.Create(log).Error
}

func (r *analyticsRepository) GetRequestsByUser(userID uuid.UUID, limit int) ([]*models.RequestLog, error) {
	var logs []*models.RequestLog
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *analyticsRepository) GetUsageStats(userID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	var totalRequests int64
	var successfulRequests int64
	var totalResponseTime int64

	// Total requests
	r.db.Model(&models.RequestLog{}).
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startDate, endDate).
		Count(&totalRequests)

	// Successful requests
	r.db.Model(&models.RequestLog{}).
		Where("user_id = ? AND status_code = ? AND created_at BETWEEN ? AND ?", userID, 200, startDate, endDate).
		Count(&successfulRequests)

	// Average response time
	r.db.Model(&models.RequestLog{}).
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startDate, endDate).
		Select("COALESCE(SUM(response_time), 0)").
		Scan(&totalResponseTime)

	avgResponseTime := int64(0)
	if totalRequests > 0 {
		avgResponseTime = totalResponseTime / totalRequests
	}

	successRate := float64(0)
	if totalRequests > 0 {
		successRate = (float64(successfulRequests) / float64(totalRequests)) * 100
	}

	stats := map[string]interface{}{
		"total_requests":        totalRequests,
		"successful_requests":   successfulRequests,
		"success_rate":          successRate,
		"average_response_time": avgResponseTime,
	}

	return stats, nil
}
