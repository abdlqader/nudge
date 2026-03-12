package main

import (
	"fmt"
	"log"
	"nudge/config"
	"nudge/internal/database"
	"nudge/internal/routes"
)

func main() {
	log.Println("Nudge - Starting application...")

	// Load configuration
	config.Load()

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create custom indexes
	if err := database.CreateIndexes(); err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	// Seed database (development only)
	if config.IsDevelopment() {
		if err := database.Seed(); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
	}

	log.Println("Database initialized successfully")

	// Setup and start HTTP server
	router := routes.SetupRouter()
	address := fmt.Sprintf(":%s", config.AppConfig.Port)

	log.Printf("Starting server on port %s...", config.AppConfig.Port)
	log.Println("Nudge is ready!")

	if err := router.Run(address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
