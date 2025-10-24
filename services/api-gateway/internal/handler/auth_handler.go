package handler

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"ironnode/pkg/config"
	"ironnode/pkg/response"
	pb "ironnode/services/auth-service/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthHandler struct {
	authClient pb.AuthServiceClient
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	// Connect to Auth Service via gRPC
	// Use AUTH_SERVICE_HOST from environment if available, otherwise localhost
	authServiceHost := os.Getenv("AUTH_SERVICE_HOST")
	if authServiceHost == "" {
		authServiceHost = "localhost"
	}

	conn, err := grpc.Dial(
		authServiceHost+":"+cfg.Services.AuthServicePort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}

	client := pb.NewAuthServiceClient(conn)

	return &AuthHandler{
		authClient: client,
	}
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.Register(ctx, &pb.RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})

	if err != nil {
		response.InternalServerError(c, "Failed to register user", err)
		return
	}

	response.Success(c, http.StatusCreated, "User registered successfully", gin.H{
		"user_id":    resp.UserId,
		"email":      resp.Email,
		"first_name": resp.FirstName,
		"last_name":  resp.LastName,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.Login(ctx, &pb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		response.Unauthorized(c, "Invalid credentials")
		return
	}

	response.Success(c, http.StatusOK, "Login successful", gin.H{
		"token": resp.Token,
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.GetUser(ctx, &pb.GetUserRequest{
		UserId: userID.(string),
	})

	if err != nil {
		response.InternalServerError(c, "Failed to get user profile", err)
		return
	}

	response.Success(c, http.StatusOK, "Profile retrieved successfully", gin.H{
		"user_id":    resp.UserId,
		"email":      resp.Email,
		"first_name": resp.FirstName,
		"last_name":  resp.LastName,
		"is_active":  resp.IsActive,
	})
}

func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		token := parts[1]

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := h.authClient.ValidateToken(ctx, &pb.ValidateTokenRequest{
			Token: token,
		})

		if err != nil || !resp.Valid {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("user_id", resp.UserId)
		c.Next()
	}
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.ForgotPassword(ctx, &pb.ForgotPasswordRequest{
		Email: req.Email,
	})

	if err != nil {
		response.InternalServerError(c, "Failed to process password reset request", err)
		return
	}

	response.Success(c, http.StatusOK, "Password reset link has been sent to your email", gin.H{
		"message": resp.Message,
		"token":   resp.Token,
	})
}

type VerifyResetTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *AuthHandler) VerifyResetToken(c *gin.Context) {
	var req VerifyResetTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.VerifyResetToken(ctx, &pb.VerifyResetTokenRequest{
		Token: req.Token,
	})

	if err != nil {
		response.BadRequest(c, "Invalid or expired token", err)
		return
	}

	response.Success(c, http.StatusOK, "Token is valid", gin.H{
		"valid":      resp.Valid,
		"expires_at": resp.ExpiresAt,
	})
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.authClient.ResetPassword(ctx, &pb.ResetPasswordRequest{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	})

	if err != nil {
		response.BadRequest(c, "Failed to reset password", err)
		return
	}

	response.Success(c, http.StatusOK, "Password reset successful", gin.H{
		"message": resp.Message,
	})
}
