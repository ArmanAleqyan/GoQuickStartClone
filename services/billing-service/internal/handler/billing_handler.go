package handler

import (
	"context"

	"ironnode/pkg/models"
	"ironnode/services/billing-service/internal/service"
	pb "ironnode/services/billing-service/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BillingHandler struct {
	pb.UnimplementedBillingServiceServer
	billingService service.BillingService
}

func NewBillingHandler(billingService service.BillingService) *BillingHandler {
	return &BillingHandler{billingService: billingService}
}

func (h *BillingHandler) CreateSubscription(ctx context.Context, req *pb.CreateSubscriptionRequest) (*pb.SubscriptionResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	planType := models.PlanType(req.PlanType)

	subscription, err := h.billingService.CreateSubscription(userID, planType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create subscription: %v", err)
	}

	return &pb.SubscriptionResponse{
		Id:               subscription.ID.String(),
		UserId:           subscription.UserID.String(),
		PlanType:         string(subscription.PlanType),
		RequestsPerMonth: int32(subscription.RequestsPerMonth),
		RequestsUsed:     int32(subscription.RequestsUsed),
		Price:            subscription.Price,
		IsActive:         subscription.IsActive,
	}, nil
}

func (h *BillingHandler) GetSubscription(ctx context.Context, req *pb.GetSubscriptionRequest) (*pb.SubscriptionResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	subscription, err := h.billingService.GetSubscription(userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "subscription not found: %v", err)
	}

	return &pb.SubscriptionResponse{
		Id:               subscription.ID.String(),
		UserId:           subscription.UserID.String(),
		PlanType:         string(subscription.PlanType),
		RequestsPerMonth: int32(subscription.RequestsPerMonth),
		RequestsUsed:     int32(subscription.RequestsUsed),
		Price:            subscription.Price,
		IsActive:         subscription.IsActive,
	}, nil
}

func (h *BillingHandler) CheckQuota(ctx context.Context, req *pb.CheckQuotaRequest) (*pb.CheckQuotaResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	hasQuota, err := h.billingService.CheckQuota(userID)
	if err != nil {
		return &pb.CheckQuotaResponse{
			HasQuota: false,
			Message:  err.Error(),
		}, nil
	}

	return &pb.CheckQuotaResponse{
		HasQuota: hasQuota,
		Message:  "OK",
	}, nil
}

func (h *BillingHandler) IncrementUsage(ctx context.Context, req *pb.IncrementUsageRequest) (*pb.IncrementUsageResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	if err := h.billingService.IncrementUsage(userID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to increment usage: %v", err)
	}

	return &pb.IncrementUsageResponse{
		Success: true,
	}, nil
}
