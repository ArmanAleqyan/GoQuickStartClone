package main

import (
	"log"
	"time"

	"ironnode/pkg/config"
	"ironnode/pkg/database"
	"ironnode/pkg/models"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	log.Println("Starting database seeding...")

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

	// Seed demo user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	demoUser := &models.User{
		Email:     "demo@example.com",
		Password:  string(hashedPassword),
		FirstName: "Demo",
		LastName:  "User",
		IsActive:  true,
	}

	if err := db.FirstOrCreate(demoUser, models.User{Email: "demo@example.com"}).Error; err != nil {
		log.Fatal("Failed to create demo user:", err)
	}
	log.Println("Created demo user: demo@example.com / password123")

	// Seed blockchain nodes
	nodes := []models.BlockchainNode{
		{
			Name:     "Ethereum Mainnet",
			Type:     models.Ethereum,
			Network:  "mainnet",
			URL:      "https://mainnet.infura.io/v3/YOUR-PROJECT-ID",
			IsActive: true,
			Priority: 100,
		},
		{
			Name:     "Polygon Mainnet",
			Type:     models.Polygon,
			Network:  "mainnet",
			URL:      "https://polygon-rpc.com",
			IsActive: true,
			Priority: 90,
		},
		{
			Name:     "BSC Mainnet",
			Type:     models.BSC,
			Network:  "mainnet",
			URL:      "https://bsc-dataseed.binance.org",
			IsActive: true,
			Priority: 80,
		},
	}

	for _, node := range nodes {
		if err := db.FirstOrCreate(&node, models.BlockchainNode{Name: node.Name}).Error; err != nil {
			log.Printf("Failed to create node %s: %v\n", node.Name, err)
		} else {
			log.Printf("Created blockchain node: %s\n", node.Name)
		}
	}

	// Seed subscription for demo user
	subscription := &models.Subscription{
		UserID:           demoUser.ID,
		PlanType:         models.FreePlan,
		RequestsPerMonth: 10000,
		RequestsUsed:     0,
		Price:            0,
		IsActive:         true,
		StartsAt:         time.Now(),
	}

	if err := db.FirstOrCreate(subscription, models.Subscription{UserID: demoUser.ID}).Error; err != nil {
		log.Fatal("Failed to create subscription:", err)
	}
	log.Println("Created free subscription for demo user")

	log.Println("Database seeding completed successfully!")
}
