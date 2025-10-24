package main

import (
	"fmt"
	"net"

	"ironnode/pkg/config"
	"ironnode/pkg/database"
	"ironnode/pkg/logger"
	"ironnode/pkg/models"
	"ironnode/services/blockchain-service/internal/handler"
	"ironnode/services/blockchain-service/internal/repository"
	"ironnode/services/blockchain-service/internal/service"
	pb "ironnode/services/blockchain-service/proto"

	"google.golang.org/grpc"
)

func main() {
	logger.Info("Starting Blockchain Service...")

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
	if err := db.AutoMigrate(&models.BlockchainNode{}); err != nil {
		logger.Fatal("Failed to migrate database:", err)
	}

	// Initialize repository, service, and handler
	nodeRepo := repository.NewNodeRepository(db)
	nodeService := service.NewNodeService(nodeRepo)
	nodeHandler := handler.NewNodeHandler(nodeService)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterBlockchainServiceServer(grpcServer, nodeHandler)

	// Start listening
	address := fmt.Sprintf(":%s", cfg.Services.BlockchainPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Fatal("Failed to listen:", err)
	}

	logger.Info("Blockchain Service is running on", address)
	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal("Failed to serve:", err)
	}
}
