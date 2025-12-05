package routes

import (
	"time"

	"ironnode/pkg/middleware"
	"ironnode/services/api-gateway/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func SetupRoutes(
	router *gin.Engine,
	authHandler *handler.AuthHandler,
	blockchainHandler *handler.BlockchainHandler,
	redisClient *redis.Client,
) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API Documentation - serve static HTML
	router.Static("/docs", "./docs")

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/verify-reset-token", authHandler.VerifyResetToken)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// Rate limiter middleware
		rateLimiter := middleware.NewRateLimiter(redisClient, 100, 1*time.Minute)

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(authHandler.AuthMiddleware())
		protected.Use(rateLimiter.Limit())
		{
			// User routes
			protected.GET("/user/profile", authHandler.GetProfile)

			// Blockchain routes
			blockchain := protected.Group("/blockchain")
			{
				blockchain.GET("/nodes", blockchainHandler.ListNodes)
				blockchain.GET("/nodes/:id", blockchainHandler.GetNode)
			}

			// Analytics routes
			analytics := protected.Group("/analytics")
			{
				analytics.GET("/usage", handler.GetUsageStats)
				analytics.GET("/requests", handler.GetRequestHistory)
			}

			// API Keys routes
			apiKeys := protected.Group("/api-keys")
			{
				apiKeys.GET("", handler.ListAPIKeys)
				apiKeys.POST("", handler.CreateAPIKey)
				apiKeys.DELETE("/:id", handler.DeleteAPIKey)
			}

			// Wallet routes
			wallets := protected.Group("/wallets")
			{
				wallets.POST("", handler.CreateWallet)                                  // Создать кошельки (BEP20 + TRC20)
				wallets.GET("", handler.GetWallets)                                     // Получить все кошельки
				wallets.GET("/:id", handler.GetWalletByID)                              // Получить кошелек по ID
				wallets.GET("/client/:client_user_id", handler.GetWalletsByClient)     // Получить кошельки клиента
				wallets.DELETE("/:id", handler.DeactivateWallet)                        // Деактивировать кошелек
			}

			// Balance routes (Tron - TRX и USDT)
			balance := protected.Group("/balance")
			{
				balance.GET("/tron/:address", handler.GetTronBalance)       // Получить TRX и USDT балансы по адресу
				balance.GET("/trx/:address", handler.GetTRXBalanceOnly)     // Получить только TRX баланс
				balance.GET("/usdt/:address", handler.GetUSDTBalanceOnly)   // Получить только USDT баланс
				balance.POST("/check", handler.GetBalancesByAddress)        // Получить балансы (POST версия)
				balance.GET("/wallet/:wallet_id", handler.GetBalanceByWalletID) // Получить балансы по ID кошелька из БД
			}
		}
	}
}
