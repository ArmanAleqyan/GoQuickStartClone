package main

import (
	"fmt"
	"net"

	"ironnode/pkg/config"
	"ironnode/pkg/database"
	"ironnode/pkg/logger"
	"ironnode/pkg/models"
	"ironnode/services/analytics-service/internal/handler"
	"ironnode/services/analytics-service/internal/repository"
	"ironnode/services/analytics-service/internal/service"
	pb "ironnode/services/analytics-service/proto"

	"google.golang.org/grpc"
)

func main() {
	logger.Info("Starting Analytics Service...")

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
	if err := db.AutoMigrate(&models.RequestLog{}); err != nil {
		logger.Fatal("Failed to migrate database:", err)
	}

	// Initialize repository, service, and handler
	analyticsRepo := repository.NewAnalyticsRepository(db)
	analyticsService := service.NewAnalyticsService(analyticsRepo)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterAnalyticsServiceServer(grpcServer, analyticsHandler)

	// Start listening
	address := fmt.Sprintf(":%s", cfg.Services.AnalyticsPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Fatal("Failed to listen:", err)
	}

	logger.Info("Analytics Service is running on", address)
	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal("Failed to serve:", err)
	}
}
