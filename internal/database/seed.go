package database

import (
	"database/sql"
	"log"
	"nudge/config"
	"nudge/internal/models"
	"time"
)

// Seed populates the database with initial data
func Seed() error {
	// Only seed in development
	if !config.IsDevelopment() {
		log.Println("Skipping seed - not in development mode")
		return nil
	}

	log.Println("Seeding database with sample data...")

	// Sample Tasks
	tasks := []models.Task{
		{
			Name:             "Read book",
			TaskCategory:     models.TaskCategoryAction,
			Status:           models.TaskStatusDeferred,
			ExpectedUnits:    sql.NullInt32{Int32: 3, Valid: true},
			ExpectedDuration: sql.NullInt32{Int32: 90, Valid: true},
			Category:         strPtr("Personal"),
		},
		{
			Name:             "Work on project",
			TaskCategory:     models.TaskCategoryAction,
			Status:           models.TaskStatusCreated,
			ExpectedDuration: sql.NullInt32{Int32: 120, Valid: true},
			Deadline:         timePtr(time.Now().Add(48 * time.Hour)),
			Category:         strPtr("Work"),
		},
		{
			Name:             "Morning jog",
			TaskCategory:     models.TaskCategoryAction,
			Status:           models.TaskStatusCompleted,
			ExpectedDuration: sql.NullInt32{Int32: 30, Valid: true},
			ActualDuration:   sql.NullInt32{Int32: 35, Valid: true},
			CompletedAt:      timePtr(time.Now().Add(-2 * time.Hour)),
			Notes:            strPtr("Felt great, but took a bit longer than expected"),
			Category:         strPtr("Health"),
		},
	}

	for i := range tasks {
		if err := DB.Create(&tasks[i]).Error; err != nil {
			log.Printf("Failed to create task: %v", err)
			return err
		}

		// Calculate and log success percentage for completed tasks
		if tasks[i].Status == models.TaskStatusCompleted {
			success := tasks[i].CalculateSuccess()
			if success != nil {
				log.Printf("Task '%s' - Success: %.2f%%", tasks[i].Name, *success)
			}
		}
	}

	log.Println("Database seeding completed successfully")
	return nil
}

// ClearData removes all data from tables (keeps schema)
func ClearData() error {
	log.Println("WARNING: Clearing all data from database...")

	DB.Exec("DELETE FROM tasks")

	log.Println("All data cleared")
	return nil
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
