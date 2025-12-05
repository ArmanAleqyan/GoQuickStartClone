package main

import (
	"fmt"
	"log"
	"os"

	"ironnode/pkg/config"
	"ironnode/pkg/database"
	"ironnode/pkg/models"

	"gorm.io/gorm"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/migrate/main.go [up|down]")
	}

	command := os.Args[1]

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Connect to database
	db, err := database.NewPostgresConnection(cfg.Database.DSN())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}
	defer sqlDB.Close()

	switch command {
	case "up":
		fmt.Println("Running migrations...")
		if err := runMigrations(db); err != nil {
			log.Fatal("Migration failed:", err)
		}
		fmt.Println("Migrations completed successfully!")

	case "down":
		fmt.Println("Rolling back migrations...")
		if err := rollbackMigrations(db); err != nil {
			log.Fatal("Rollback failed:", err)
		}
		fmt.Println("Rollback completed successfully!")

	default:
		log.Fatal("Invalid command. Use 'up' or 'down'")
	}
}

func runMigrations(db *gorm.DB) error {
	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return err
	}

	// Auto-migrate all models
	return db.AutoMigrate(
		&models.User{},
		&models.APIKey{},
		&models.BlockchainNode{},
		&models.RequestLog{},
		&models.Subscription{},
		&models.Wallet{},
		&models.PasswordReset{},
	)
}

func rollbackMigrations(db *gorm.DB) error {
	// Drop all tables (use with caution!)
	return db.Migrator().DropTable(
		&models.PasswordReset{},
		&models.Wallet{},
		&models.Subscription{},
		&models.RequestLog{},
		&models.BlockchainNode{},
		&models.APIKey{},
		&models.User{},
	)
}
