package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"ironnode/pkg/config"
	"ironnode/pkg/database"
	"ironnode/pkg/email"
	"ironnode/pkg/logger"
	"ironnode/pkg/middleware"
	"ironnode/pkg/models"
	"ironnode/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	db           *gorm.DB
	jwtSecret    string
	emailService *email.EmailService
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

func main() {
	logger.Info("Starting Standalone API for testing...")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config:", err)
	}

	jwtSecret = cfg.JWT.Secret

	// Connect to database
	db, err = database.NewPostgresConnection(cfg.Database.DSN())
	if err != nil {
		logger.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate password reset table
	if err := db.AutoMigrate(&models.PasswordReset{}); err != nil {
		logger.Fatal("Failed to migrate password reset table:", err)
	}

	// Initialize email service
	emailService = email.NewEmailService("noreply@ironnode.com")

	// Setup router
	router := gin.Default()
	router.Use(middleware.CORS())
	router.Use(middleware.RequestLogger())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "API is running"})
	})

	// API Documentation
	router.GET("/docs", func(c *gin.Context) {
		c.File("docs/api-documentation.html")
	})

	// Root redirect to docs
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/docs")
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handleRegister)
			auth.POST("/login", handleLogin)
			auth.POST("/forgot-password", handleForgotPassword)
			auth.POST("/reset-password", handleResetPassword)
			auth.POST("/verify-reset-token", handleVerifyResetToken)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(authMiddleware())
		{
			protected.GET("/user/profile", handleGetProfile)
			protected.GET("/blockchain/nodes", handleListNodes)
			protected.GET("/blockchain/nodes/:id", handleGetNode)
			protected.GET("/analytics/usage", handleGetUsage)
			protected.GET("/analytics/requests", handleGetRequests)
			protected.GET("/api-keys", handleListAPIKeys)
			protected.POST("/api-keys", handleCreateAPIKey)
			protected.DELETE("/api-keys/:id", handleDeleteAPIKey)
		}
	}

	// Start server
	port := cfg.Services.APIGatewayPort
	logger.Info("🚀 IronNode API starting...")
	logger.Info("📡 Server: http://localhost:" + port)
	logger.Info("📚 Docs: http://localhost:" + port + "/docs")
	logger.Info("👤 Demo: demo@example.com / password123")

	srv := &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server:", err)
		}
	}()

	logger.Info("✅ Server started successfully")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutting down server...")

	// Shutdown with 30s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown email service
	emailService.Shutdown()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", err)
	}

	logger.Info("✅ Server exited gracefully")
}

func handleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err)
		return
	}

	// Check if user exists
	var existingUser models.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		response.BadRequest(c, "User already exists", nil)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.InternalServerError(c, "Failed to hash password", err)
		return
	}

	// Create user
	user := &models.User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  true,
	}

	if err := db.Create(user).Error; err != nil {
		response.InternalServerError(c, "Failed to create user", err)
		return
	}

	response.Success(c, http.StatusCreated, "User registered successfully", gin.H{
		"user_id":    user.ID.String(),
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	})
}

func handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err)
		return
	}

	// Find user
	var user models.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		response.Unauthorized(c, "Invalid credentials")
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		response.Unauthorized(c, "Invalid credentials")
		return
	}

	// Generate token
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		response.InternalServerError(c, "Failed to generate token", err)
		return
	}

	response.Success(c, http.StatusOK, "Login successful", gin.H{
		"token": tokenString,
	})
}

func handleGetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user models.User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		response.NotFound(c, "User not found")
		return
	}

	response.Success(c, http.StatusOK, "Profile retrieved successfully", gin.H{
		"user_id":    user.ID.String(),
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"is_active":  user.IsActive,
	})
}

func handleListNodes(c *gin.Context) {
	var nodes []models.BlockchainNode
	if err := db.Where("is_active = ?", true).Find(&nodes).Error; err != nil {
		response.InternalServerError(c, "Failed to fetch nodes", err)
		return
	}

	var result []gin.H
	for _, node := range nodes {
		result = append(result, gin.H{
			"id":       node.ID.String(),
			"name":     node.Name,
			"type":     string(node.Type),
			"network":  node.Network,
			"is_active": node.IsActive,
			"priority": node.Priority,
		})
	}

	response.Success(c, http.StatusOK, "Nodes retrieved successfully", result)
}

func handleGetNode(c *gin.Context) {
	nodeID := c.Param("id")

	var node models.BlockchainNode
	if err := db.Where("id = ?", nodeID).First(&node).Error; err != nil {
		response.NotFound(c, "Node not found")
		return
	}

	response.Success(c, http.StatusOK, "Node retrieved successfully", gin.H{
		"id":       node.ID.String(),
		"name":     node.Name,
		"type":     string(node.Type),
		"network":  node.Network,
		"url":      node.URL,
		"is_active": node.IsActive,
		"priority": node.Priority,
	})
}

func handleGetUsage(c *gin.Context) {
	response.Success(c, http.StatusOK, "Usage stats retrieved", gin.H{
		"total_requests":        150,
		"requests_today":        25,
		"requests_this_month":   450,
		"success_rate":          98.5,
		"average_response_time": 125,
	})
}

func handleGetRequests(c *gin.Context) {
	response.Success(c, http.StatusOK, "Request history retrieved", []gin.H{
		{
			"id":            "1",
			"blockchain":    "ethereum",
			"method":        "eth_blockNumber",
			"status_code":   200,
			"response_time": 125,
			"timestamp":     time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		},
		{
			"id":            "2",
			"blockchain":    "polygon",
			"method":        "eth_getBalance",
			"status_code":   200,
			"response_time": 98,
			"timestamp":     time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		},
	})
}

func handleListAPIKeys(c *gin.Context) {
	response.Success(c, http.StatusOK, "API keys retrieved", []gin.H{
		{
			"id":          "1",
			"name":        "Production Key",
			"key":         "qn_***************",
			"is_active":   true,
			"created_at":  time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339),
		},
	})
}

func handleCreateAPIKey(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err)
		return
	}

	response.Success(c, http.StatusCreated, "API key created", gin.H{
		"id":          uuid.New().String(),
		"name":        req.Name,
		"description": req.Description,
		"key":         "qn_" + uuid.New().String(),
		"is_active":   true,
		"created_at":  time.Now().Format(time.RFC3339),
	})
}

func handleDeleteAPIKey(c *gin.Context) {
	keyID := c.Param("id")
	response.Success(c, http.StatusOK, "API key deleted", gin.H{
		"id": keyID,
	})
}

// Password Reset handlers

func handleForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err)
		return
	}

	// Find user
	var user models.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Не сообщаем существует ли email (security best practice)
		response.Success(c, http.StatusOK, "If email exists, reset link has been sent", nil)
		return
	}

	// Generate reset token
	token, err := generateResetToken()
	if err != nil {
		response.InternalServerError(c, "Failed to generate reset token", err)
		return
	}

	// Save reset token to database
	passwordReset := &models.PasswordReset{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour), // Token valid for 1 hour
	}

	if err := db.Create(passwordReset).Error; err != nil {
		response.InternalServerError(c, "Failed to create reset token", err)
		return
	}

	// Send email with reset link
	resetURL := "http://localhost/reset-password" // In production use actual domain
	if err := emailService.SendPasswordResetEmail(user.Email, token, resetURL); err != nil {
		logger.Error("Failed to send email:", err)
	}

	response.Success(c, http.StatusOK, "Password reset link has been sent to your email", gin.H{
		"message": "Check console for reset link (email service not configured)",
		"token":   token, // Remove in production!
	})
}

func handleVerifyResetToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err)
		return
	}

	var passwordReset models.PasswordReset
	if err := db.Where("token = ?", req.Token).First(&passwordReset).Error; err != nil {
		response.BadRequest(c, "Invalid reset token", nil)
		return
	}

	if !passwordReset.IsValid() {
		response.BadRequest(c, "Reset token expired or already used", nil)
		return
	}

	response.Success(c, http.StatusOK, "Token is valid", gin.H{
		"valid":      true,
		"expires_at": passwordReset.ExpiresAt.Format(time.RFC3339),
	})
}

func handleResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err)
		return
	}

	// Find reset token
	var passwordReset models.PasswordReset
	if err := db.Preload("User").Where("token = ?", req.Token).First(&passwordReset).Error; err != nil {
		response.BadRequest(c, "Invalid reset token", nil)
		return
	}

	// Validate token
	if !passwordReset.IsValid() {
		response.BadRequest(c, "Reset token expired or already used", nil)
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.InternalServerError(c, "Failed to hash password", err)
		return
	}

	// Update user password
	if err := db.Model(&passwordReset.User).Update("password", string(hashedPassword)).Error; err != nil {
		response.InternalServerError(c, "Failed to update password", err)
		return
	}

	// Mark token as used
	now := time.Now()
	passwordReset.UsedAt = &now
	if err := db.Save(&passwordReset).Error; err != nil {
		logger.Error("Failed to mark token as used:", err)
	}

	// Send notification email
	emailService.SendPasswordChangedEmail(passwordReset.User.Email)

	response.Success(c, http.StatusOK, "Password reset successful", gin.H{
		"message": "You can now login with your new password",
	})
}

func generateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
