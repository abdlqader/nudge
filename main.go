package main

import (
	"log"
	"nudge/config"
	"nudge/internal/database"
)

func main() {
	log.Println("Nudge - Starting application...")

	// Load configuration
	config.Load()

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

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
	log.Println("Nudge is ready!")
}
