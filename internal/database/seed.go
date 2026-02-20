package database

import (
	"database/sql"
	"log"
	"nudge/config"
	"nudge/internal/models"
	"time"

	"gorm.io/datatypes"
)

// Seed populates the database with initial data
func Seed() error {
	// Only seed in development
	if !config.IsDevelopment() {
		log.Println("Skipping seed - not in development mode")
		return nil
	}

	log.Println("Seeding database with sample data...")

	// Sample Recurring Tasks
	recurringTasks := []models.RecurringTask{
		{
			Name:               "Daily Standup",
			RecurrenceType:     models.RecurrenceTypeDaily,
			RecurrenceInterval: intPtr(1),
			IsActive:           true,
		},
		{
			Name:           "Team Meeting",
			RecurrenceType: models.RecurrenceTypeWeekly,
			RecurrenceDays: datatypes.JSON([]byte(`[1,3]`)), // Monday, Wednesday
			IsActive:       true,
		},
		{
			Name:                 "Pay Rent",
			RecurrenceType:       models.RecurrenceTypeMonthlyDate,
			RecurrenceDayOfMonth: intPtr(1), // 1st of every month
			IsActive:             true,
		},
		{
			Name:              "Board Meeting",
			RecurrenceType:    models.RecurrenceTypeMonthlyPattern,
			RecurrencePattern: strPtr("first_monday"),
			IsActive:          true,
		},
	}

	for _, rt := range recurringTasks {
		if err := DB.Create(&rt).Error; err != nil {
			log.Printf("Failed to create recurring task: %v", err)
			return err
		}
	}

	// Sample Tasks
	tasks := []models.Task{
		{
			Name:             "Read 3 chapters for exam",
			TaskType:         models.TaskTypeUnitBased,
			TaskCategory:     models.TaskCategoryAction,
			Status:           models.TaskStatusPending,
			Priority:         models.PriorityMedium,
			ExpectedUnits:    sql.NullInt32{Int32: 3, Valid: true},
			ExpectedDuration: sql.NullInt32{Int32: 150, Valid: true},
		},
		{
			Name:             "Deep work session",
			TaskType:         models.TaskTypeTimeBased,
			TaskCategory:     models.TaskCategoryAction,
			Status:           models.TaskStatusPending,
			Priority:         models.PriorityHigh,
			ExpectedDuration: sql.NullInt32{Int32: 120, Valid: true},
		},
		{
			Name:             "Morning commute to office",
			TaskType:         models.TaskTypeCommute,
			TaskCategory:     models.TaskCategoryTransit,
			IsCommute:        true,
			Status:           models.TaskStatusCompleted,
			Priority:         models.PriorityMedium,
			ExpectedDuration: sql.NullInt32{Int32: 30, Valid: true},
			ActualDuration:   sql.NullInt32{Int32: 35, Valid: true},
			CompletedAt:      timePtr(time.Now().Add(-1 * time.Hour)),
		},
		{
			Name:             "Sleep",
			TaskType:         models.TaskTypeTimeBased,
			TaskCategory:     models.TaskCategoryAnchor,
			Status:           models.TaskStatusCompleted,
			Priority:         models.PriorityCritical,
			ExpectedDuration: sql.NullInt32{Int32: 480, Valid: true}, // 8 hours
			ActualDuration:   sql.NullInt32{Int32: 450, Valid: true}, // 7.5 hours
			CompletedAt:      timePtr(time.Now().Add(-8 * time.Hour)),
		},
		{
			Name:             "Family dinner",
			TaskType:         models.TaskTypeTimeBased,
			TaskCategory:     models.TaskCategoryAnchor,
			Status:           models.TaskStatusCompleted,
			Priority:         models.PriorityCritical,
			ExpectedDuration: sql.NullInt32{Int32: 60, Valid: true},
			ActualDuration:   sql.NullInt32{Int32: 60, Valid: true},
			CompletedAt:      timePtr(time.Now().Add(-2 * time.Hour)),
		},
		{
			Name:             "Review 3 PRs",
			TaskType:         models.TaskTypeUnitBased,
			TaskCategory:     models.TaskCategoryAction,
			Status:           models.TaskStatusCompleted,
			Priority:         models.PriorityHigh,
			ExpectedUnits:    sql.NullInt32{Int32: 3, Valid: true},
			ActualUnits:      sql.NullInt32{Int32: 3, Valid: true},
			ExpectedDuration: sql.NullInt32{Int32: 60, Valid: true},
			ActualDuration:   sql.NullInt32{Int32: 75, Valid: true},
			CompletedAt:      timePtr(time.Now().Add(-3 * time.Hour)),
			Category:         strPtr("Work"),
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
	DB.Exec("DELETE FROM recurring_tasks")

	log.Println("All data cleared")
	return nil
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
