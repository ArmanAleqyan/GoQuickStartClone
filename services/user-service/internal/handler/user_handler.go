package handler

import (
	"context"

	"ironnode/services/user-service/internal/service"
	pb "ironnode/services/user-service/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) CreateAPIKey(ctx context.Context, req *pb.CreateAPIKeyRequest) (*pb.APIKeyResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	apiKey, err := h.userService.CreateAPIKey(userID, req.Name, req.Description)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create API key: %v", err)
	}

	return &pb.APIKeyResponse{
		Id:          apiKey.ID.String(),
		UserId:      apiKey.UserID.String(),
		Key:         apiKey.Key,
		Name:        apiKey.Name,
		Description: apiKey.Description,
		IsActive:    apiKey.IsActive,
	}, nil
}

func (h *UserHandler) GetAPIKeys(ctx context.Context, req *pb.GetAPIKeysRequest) (*pb.GetAPIKeysResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	keys, err := h.userService.GetAPIKeys(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get API keys: %v", err)
	}

	var pbKeys []*pb.APIKeyResponse
	for _, key := range keys {
		pbKeys = append(pbKeys, &pb.APIKeyResponse{
			Id:          key.ID.String(),
			UserId:      key.UserID.String(),
			Key:         key.Key,
			Name:        key.Name,
			Description: key.Description,
			IsActive:    key.IsActive,
		})
	}

	return &pb.GetAPIKeysResponse{
		ApiKeys: pbKeys,
	}, nil
}

func (h *UserHandler) ValidateAPIKey(ctx context.Context, req *pb.ValidateAPIKeyRequest) (*pb.ValidateAPIKeyResponse, error) {
	apiKey, err := h.userService.ValidateAPIKey(req.Key)
	if err != nil {
		return &pb.ValidateAPIKeyResponse{
			Valid: false,
		}, nil
	}

	return &pb.ValidateAPIKeyResponse{
		Valid:  true,
		UserId: apiKey.UserID.String(),
	}, nil
}

func (h *UserHandler) DeleteAPIKey(ctx context.Context, req *pb.DeleteAPIKeyRequest) (*pb.DeleteAPIKeyResponse, error) {
	keyID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid API key ID: %v", err)
	}

	if err := h.userService.DeleteAPIKey(keyID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete API key: %v", err)
	}

	return &pb.DeleteAPIKeyResponse{
		Success: true,
	}, nil
}
