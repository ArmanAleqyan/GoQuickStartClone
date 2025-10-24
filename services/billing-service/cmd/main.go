package main

import (
	"fmt"
	"net"

	"ironnode/pkg/config"
	"ironnode/pkg/database"
	"ironnode/pkg/logger"
	"ironnode/pkg/models"
	"ironnode/services/billing-service/internal/handler"
	"ironnode/services/billing-service/internal/repository"
	"ironnode/services/billing-service/internal/service"
	pb "ironnode/services/billing-service/proto"

	"google.golang.org/grpc"
)

func main() {
	logger.Info("Starting Billing Service...")

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
	if err := db.AutoMigrate(&models.Subscription{}); err != nil {
		logger.Fatal("Failed to migrate database:", err)
	}

	// Initialize repository, service, and handler
	billingRepo := repository.NewBillingRepository(db)
	billingService := service.NewBillingService(billingRepo)
	billingHandler := handler.NewBillingHandler(billingService)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterBillingServiceServer(grpcServer, billingHandler)

	// Start listening
	address := fmt.Sprintf(":%s", cfg.Services.BillingServicePort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Fatal("Failed to listen:", err)
	}

	logger.Info("Billing Service is running on", address)
	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal("Failed to serve:", err)
	}
}
