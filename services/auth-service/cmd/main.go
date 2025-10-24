package main

import (
	"fmt"
	"net"

	"ironnode/pkg/config"
	"ironnode/pkg/database"
	"ironnode/pkg/email"
	"ironnode/pkg/logger"
	"ironnode/pkg/models"
	"ironnode/services/auth-service/internal/handler"
	"ironnode/services/auth-service/internal/repository"
	"ironnode/services/auth-service/internal/service"
	pb "ironnode/services/auth-service/proto"

	"google.golang.org/grpc"
)

func main() {
	logger.Info("Starting Auth Service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration:", err)
	}

	// Connect to database
	db, err := database.NewPostgresConnection(cfg.Database.DSN())
	if err != nil {
		logger.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&models.User{}, &models.PasswordReset{}); err != nil {
		logger.Fatal("Failed to migrate database:", err)
	}

	// Initialize email service
	emailService := email.NewEmailService(cfg.Email.From)

	// Initialize repository, service, and handler
	authRepo := repository.NewAuthRepository(db)
	authService := service.NewAuthService(authRepo, emailService, cfg.JWT.Secret, cfg.JWT.Expiry)
	authHandler := handler.NewAuthHandler(authService)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authHandler)

	// Start listening
	address := fmt.Sprintf(":%s", cfg.Services.AuthServicePort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Fatal("Failed to listen:", err)
	}

	logger.Info("Auth Service is running on", address)
	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal("Failed to serve:", err)
	}
}
