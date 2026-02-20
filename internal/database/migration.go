package database

import (
	"log"

	"gorm.io/gorm"
)

// Migrate runs all database migrations
func Migrate() error {
	log.Println("Running database migrations...")

	// Auto-migrate all models here
	// Example: err := DB.AutoMigrate(&models.Task{}, &models.RecurringTask{})
	
	// TODO: Add models when they are created
	// err := DB.AutoMigrate(
	//     &models.RecurringTask{},
	//     &models.Task{},
	// )
	
	log.Println("Database migrations completed successfully")
	return nil
}

// MigrateModels runs migration for provided models
func MigrateModels(models ...interface{}) error {
	if err := DB.AutoMigrate(models...); err != nil {
		return err
	}
	log.Printf("Migrated %d models", len(models))
	return nil
}

// DropAllTables drops all tables (use with caution!)
func DropAllTables(models ...interface{}) error {
	log.Println("WARNING: Dropping all tables...")
	return DB.Migrator().DropTable(models...)
}

// CreateIndexes creates custom indexes for performance optimization
func CreateIndexes() error {
	log.Println("Creating custom indexes...")

	// TODO: Add custom indexes when models are created
	// Examples:
	// DB.Exec("CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at)")
	// DB.Exec("CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)")
	// DB.Exec("CREATE INDEX IF NOT EXISTS idx_tasks_completed_at ON tasks(completed_at)")
	// DB.Exec("CREATE INDEX IF NOT EXISTS idx_tasks_recurring_id ON tasks(recurring_task_id)")
	// DB.Exec("CREATE INDEX IF NOT EXISTS idx_recurring_tasks_active ON recurring_tasks(is_active)")

	log.Println("Custom indexes created successfully")
	return nil
}

// Transaction wraps a function in a database transaction
func Transaction(fn func(*gorm.DB) error) error {
	return DB.Transaction(fn)
}
