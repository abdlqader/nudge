package database

import (
	"database/sql"
	"fmt"
	"log"
	"nudge/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var DB *gorm.DB

// Connect establishes database connection based on environment
func Connect() error {
	var err error
	var dialector gorm.Dialector

	cfg := config.AppConfig

	// Determine connection type based on environment
	if cfg.DBToken != "" {
		// Production: Turso with authentication
		log.Println("Connecting to Turso database...")
		connector, err := NewTursoConnector(cfg.DBUrl, cfg.DBToken)
		if err != nil {
			return fmt.Errorf("failed to create Turso connector: %w", err)
		}
		dialector = sqlite.Dialector{Conn: connector}
	} else {
		// Development: Local SQLite
		log.Println("Connecting to local SQLite database...")
		dialector = sqlite.Open(cfg.DBUrl)
	}

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(getLogLevel()),
	}

	DB, err = gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool for SQLite
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// SQLite specific settings
	sqlDB.SetMaxOpenConns(1) // SQLite works best with single connection
	sqlDB.SetMaxIdleConns(1)

	log.Println("Database connection established successfully")
	return nil
}

// NewTursoConnector creates a connector for Turso/libSQL with authentication
func NewTursoConnector(url, token string) (gorm.ConnPool, error) {
	// Construct connection string with authentication token
	connectionString := fmt.Sprintf("%s?authToken=%s", url, token)
	
	db, err := sql.Open("libsql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open libsql connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// getLogLevel returns appropriate log level based on environment
func getLogLevel() logger.LogLevel {
	if config.IsDevelopment() {
		return logger.Info
	}
	return logger.Warn
}
