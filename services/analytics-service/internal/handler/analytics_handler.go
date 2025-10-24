package handler

import (
	"context"
	"time"

	"ironnode/services/analytics-service/internal/service"
	pb "ironnode/services/analytics-service/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AnalyticsHandler struct {
	pb.UnimplementedAnalyticsServiceServer
	analyticsService service.AnalyticsService
}

func NewAnalyticsHandler(analyticsService service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

func (h *AnalyticsHandler) LogRequest(ctx context.Context, req *pb.LogRequestRequest) (*pb.LogRequestResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	apiKeyID, err := uuid.Parse(req.ApiKeyId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid API key ID: %v", err)
	}

	err = h.analyticsService.LogRequest(
		userID,
		apiKeyID,
		req.Blockchain,
		req.Method,
		req.Endpoint,
		int(req.StatusCode),
		req.ResponseTime,
		req.RequestSize,
		req.ResponseSize,
		req.IpAddress,
		req.UserAgent,
		req.Error,
	)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to log request: %v", err)
	}

	return &pb.LogRequestResponse{
		Success: true,
	}, nil
}

func (h *AnalyticsHandler) GetRequestHistory(ctx context.Context, req *pb.GetRequestHistoryRequest) (*pb.GetRequestHistoryResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	logs, err := h.analyticsService.GetRequestHistory(userID, int(req.Limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get request history: %v", err)
	}

	var pbLogs []*pb.RequestLog
	for _, log := range logs {
		pbLogs = append(pbLogs, &pb.RequestLog{
			Id:           log.ID.String(),
			UserId:       log.UserID.String(),
			ApiKeyId:     log.APIKeyID.String(),
			Blockchain:   log.Blockchain,
			Method:       log.Method,
			Endpoint:     log.Endpoint,
			StatusCode:   int32(log.StatusCode),
			ResponseTime: log.ResponseTime,
			RequestSize:  log.RequestSize,
			ResponseSize: log.ResponseSize,
			IpAddress:    log.IPAddress,
			UserAgent:    log.UserAgent,
			Error:        log.Error,
			CreatedAt:    log.CreatedAt.Format(time.RFC3339),
		})
	}

	return &pb.GetRequestHistoryResponse{
		Logs: pbLogs,
	}, nil
}

func (h *AnalyticsHandler) GetUsageStats(ctx context.Context, req *pb.GetUsageStatsRequest) (*pb.GetUsageStatsResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid start date: %v", err)
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid end date: %v", err)
	}

	stats, err := h.analyticsService.GetUsageStats(userID, startDate, endDate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get usage stats: %v", err)
	}

	return &pb.GetUsageStatsResponse{
		TotalRequests:       stats["total_requests"].(int64),
		SuccessfulRequests:  stats["successful_requests"].(int64),
		SuccessRate:         stats["success_rate"].(float64),
		AverageResponseTime: stats["average_response_time"].(int64),
	}, nil
}
