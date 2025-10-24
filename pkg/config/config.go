package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Database    DatabaseConfig
	Redis       RedisConfig
	RabbitMQ    RabbitMQConfig
	JWT         JWTConfig
	Services    ServicesConfig
	Email       EmailConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type ServicesConfig struct {
	APIGatewayPort      string
	AuthServicePort     string
	UserServicePort     string
	BlockchainPort      string
	AnalyticsPort       string
	BillingServicePort  string
}

type EmailConfig struct {
	From string
}

func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5433"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "ironnode"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		RabbitMQ: RabbitMQConfig{
			Host:     getEnv("RABBITMQ_HOST", "localhost"),
			Port:     getEnv("RABBITMQ_PORT", "5672"),
			User:     getEnv("RABBITMQ_USER", "guest"),
			Password: getEnv("RABBITMQ_PASSWORD", "guest"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key"),
			Expiry: 24 * time.Hour,
		},
		Services: ServicesConfig{
			APIGatewayPort:     getEnv("API_GATEWAY_PORT", "8080"),
			AuthServicePort:    getEnv("AUTH_SERVICE_PORT", "50051"),
			UserServicePort:    getEnv("USER_SERVICE_PORT", "50052"),
			BlockchainPort:     getEnv("BLOCKCHAIN_SERVICE_PORT", "50053"),
			AnalyticsPort:      getEnv("ANALYTICS_SERVICE_PORT", "50055"),
			BillingServicePort: getEnv("BILLING_SERVICE_PORT", "50056"),
		},
		Email: EmailConfig{
			From: getEnv("EMAIL_FROM", "noreply@ironnode.com"),
		},
	}

	return config, nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func (c *RabbitMQConfig) URL() string {
	return fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		c.User, c.Password, c.Host, c.Port,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
