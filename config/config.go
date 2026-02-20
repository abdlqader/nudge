package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl   string
	DBToken string
	Env     string
}

var AppConfig *Config

// Load initializes the application configuration
func Load() {
	// Load .env file if it exists (ignore error in production)
	_ = godotenv.Load()

	AppConfig = &Config{
		DBUrl:   getEnv("DB_URL", "file:local.db"),
		DBToken: getEnv("DB_TOKEN", ""),
		Env:     getEnv("ENV", "development"),
	}

	log.Printf("Configuration loaded - Environment: %s, DB: %s", AppConfig.Env, AppConfig.DBUrl)
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsDevelopment checks if running in development mode
func IsDevelopment() bool {
	return AppConfig.Env == "development"
}

// IsProduction checks if running in production mode
func IsProduction() bool {
	return AppConfig.Env == "production"
}
