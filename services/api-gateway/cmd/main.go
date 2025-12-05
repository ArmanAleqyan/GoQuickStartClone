package main

import (
	"ironnode/pkg/config"
	"ironnode/pkg/logger"
	"ironnode/pkg/middleware"
	"ironnode/services/api-gateway/internal/handler"
	"ironnode/services/api-gateway/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger.Info("Starting API Gateway...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration:", err)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Initialize Gin router
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORS())

	// Initialize handlers
	authHandler := handler.NewAuthHandler(cfg)
	blockchainHandler := handler.NewBlockchainHandler(cfg)

	// Initialize wallet service
	if err := handler.InitWalletService(cfg.Database.DSN()); err != nil {
		logger.Fatal("Failed to initialize wallet service:", err)
	}
	logger.Info("Wallet service initialized successfully")

	// Initialize Tron client
	tronNodeURL := "http://78.46.94.60:8090"
	handler.InitTronClient(tronNodeURL)
	logger.Info("Tron client initialized successfully. Node:", tronNodeURL)

	// Setup routes
	routes.SetupRoutes(router, authHandler, blockchainHandler, redisClient)

	// Start server
	address := ":" + cfg.Services.APIGatewayPort
	logger.Info("API Gateway is running on", address)
	if err := router.Run(address); err != nil {
		logger.Fatal("Failed to start server:", err)
	}
}
