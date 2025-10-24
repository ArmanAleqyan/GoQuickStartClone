package handler

import (
	"context"

	"ironnode/services/auth-service/internal/service"
	pb "ironnode/services/auth-service/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := h.authService.Register(req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	return &pb.RegisterResponse{
		UserId:    user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to login: %v", err)
	}

	return &pb.LoginResponse{
		Token: token,
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	userID, err := h.authService.ValidateToken(req.Token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	return &pb.ValidateTokenResponse{
		UserId: userID.String(),
		Valid:  true,
	}, nil
}

func (h *AuthHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &pb.GetUserResponse{
		UserId:    user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsActive:  user.IsActive,
	}, nil
}

func (h *AuthHandler) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*pb.ForgotPasswordResponse, error) {
	token, err := h.authService.ForgotPassword(req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to process password reset: %v", err)
	}

	return &pb.ForgotPasswordResponse{
		Message: "If the email exists, a password reset link has been sent",
		Token:   token,
	}, nil
}

func (h *AuthHandler) VerifyResetToken(ctx context.Context, req *pb.VerifyResetTokenRequest) (*pb.VerifyResetTokenResponse, error) {
	reset, err := h.authService.VerifyResetToken(req.Token)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired token: %v", err)
	}

	return &pb.VerifyResetTokenResponse{
		Valid:     true,
		ExpiresAt: reset.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (h *AuthHandler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	err := h.authService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to reset password: %v", err)
	}

	return &pb.ResetPasswordResponse{
		Message: "Password has been reset successfully",
	}, nil
}
