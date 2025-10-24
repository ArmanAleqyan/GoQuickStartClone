package main

import (
	"fmt"
	"net"

	"ironnode/pkg/config"
	"ironnode/pkg/database"
	"ironnode/pkg/logger"
	"ironnode/pkg/models"
	"ironnode/services/user-service/internal/handler"
	"ironnode/services/user-service/internal/repository"
	"ironnode/services/user-service/internal/service"
	pb "ironnode/services/user-service/proto"

	"google.golang.org/grpc"
)

func main() {
	logger.Info("Starting User Service...")

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
	if err := db.AutoMigrate(&models.APIKey{}); err != nil {
		logger.Fatal("Failed to migrate database:", err)
	}

	// Initialize repository, service, and handler
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, userHandler)

	// Start listening
	address := fmt.Sprintf(":%s", cfg.Services.UserServicePort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Fatal("Failed to listen:", err)
	}

	logger.Info("User Service is running on", address)
	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal("Failed to serve:", err)
	}
}
