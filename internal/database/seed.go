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

	// Create a sample user (check if already exists)
	var existingUser models.User
	err := DB.Where("email = ?", "admin@mk.com").First(&existingUser).Error
	
	var user models.User
	if err != nil {
		// User doesn't exist, create it
		user = models.User{
			Email:     "admin@mk.com",
			FirstName: "Mo",
			LastName:  "Ka",
		}
		user.HashPassword("Testing123")

		// Create the user in database
		if err := DB.Create(&user).Error; err != nil {
			log.Printf("Failed to create user: %v", err)
			return err
		}
		log.Printf("Created user: %s %s (%s)", user.FirstName, user.LastName, user.Email)
	} else {
		user = existingUser
		log.Printf("User already exists: %s %s (%s)", user.FirstName, user.LastName, user.Email)
	}

	// Check if tasks already exist for this user
	var taskCount int64
	DB.Model(&models.Task{}).Where("user_id = ?", user.ID).Count(&taskCount)
	
	if taskCount > 0 {
		log.Printf("Tasks already exist for user (%d tasks found), skipping task seeding", taskCount)
	} else {
		// Sample Tasks (linked to the user)
		tasks := []models.Task{
			{
				UserID:           user.ID,
				Name:             "Read book",
				TaskCategory:     models.TaskCategoryAction,
				Status:           models.TaskStatusDeferred,
				ExpectedUnits:    sql.NullInt32{Int32: 3, Valid: true},
				ExpectedDuration: sql.NullInt32{Int32: 90, Valid: true},
				Category:         strPtr("Personal"),
				StartAt:          sql.NullInt32{Int32: 1200, Valid: true}, // 8:00 PM (20:00)
			},
			{
				UserID:           user.ID,
				Name:             "Work on project",
				TaskCategory:     models.TaskCategoryAction,
				Status:           models.TaskStatusCreated,
				ExpectedDuration: sql.NullInt32{Int32: 120, Valid: true},
				Deadline:         timePtr(time.Now().Add(48 * time.Hour)),
				Category:         strPtr("Work"),
				StartAt:          sql.NullInt32{Int32: 570, Valid: true}, // 9:30 AM
			},
			{
				UserID:           user.ID,
				Name:             "Morning jog",
				TaskCategory:     models.TaskCategoryAction,
				Status:           models.TaskStatusCompleted,
				ExpectedDuration: sql.NullInt32{Int32: 30, Valid: true},
				ActualDuration:   sql.NullInt32{Int32: 35, Valid: true},
				CompletedAt:      timePtr(time.Now().Add(-2 * time.Hour)),
				Notes:            strPtr("Felt great, but took a bit longer than expected"),
				Category:         strPtr("Health"),
				StartAt:          sql.NullInt32{Int32: 360, Valid: true}, // 6:00 AM
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
		log.Printf("Created %d sample tasks", len(tasks))
	}

	log.Println("Database seeding completed successfully")
	return nil
}

// ClearData removes all data from tables (keeps schema)
func ClearData() error {
	log.Println("WARNING: Clearing all data from database...")

	DB.Exec("DELETE FROM tasks")
	DB.Exec("DELETE FROM users")

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
