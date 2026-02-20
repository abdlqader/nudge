package database

import (
	"log"
	"nudge/config"
)

// Seed populates the database with initial data
func Seed() error {
	// Only seed in development
	if !config.IsDevelopment() {
		log.Println("Skipping seed - not in development mode")
		return nil
	}

	log.Println("Seeding database with sample data...")

	// TODO: Add seed data when models are created
	// Example seed data structure:
	
	// Sample Recurring Tasks
	// recurringTasks := []models.RecurringTask{
	//     {
	//         Name:            "Daily Standup",
	//         RecurrenceType:  "DAILY",
	//         RecurrenceInterval: 1,
	//         IsActive:        true,
	//     },
	//     {
	//         Name:            "Team Meeting",
	//         RecurrenceType:  "WEEKLY",
	//         RecurrenceDays:  datatypes.JSON([]byte(`[1,3]`)), // Monday, Wednesday
	//         IsActive:        true,
	//     },
	// }
	//
	// for _, rt := range recurringTasks {
	//     if err := DB.Create(&rt).Error; err != nil {
	//         return err
	//     }
	// }

	// Sample Tasks
	// tasks := []models.Task{
	//     {
	//         Name:             "Read 3 chapters",
	//         TaskType:         "UNIT_BASED",
	//         TaskCategory:     "ACTION",
	//         Status:           "PENDING",
	//         Priority:         2,
	//         ExpectedUnits:    sql.NullInt32{Int32: 3, Valid: true},
	//         ExpectedDuration: sql.NullInt32{Int32: 150, Valid: true},
	//     },
	//     {
	//         Name:             "Morning commute",
	//         TaskType:         "COMMUTE",
	//         TaskCategory:     "TRANSIT",
	//         IsCommute:        true,
	//         Status:           "PENDING",
	//         Priority:         2,
	//         ExpectedDuration: sql.NullInt32{Int32: 30, Valid: true},
	//     },
	// }
	//
	// for _, task := range tasks {
	//     if err := DB.Create(&task).Error; err != nil {
	//         return err
	//     }
	// }

	log.Println("Database seeding completed successfully")
	return nil
}

// ClearData removes all data from tables (keeps schema)
func ClearData() error {
	log.Println("WARNING: Clearing all data from database...")

	// TODO: Add table clearing when models are created
	// DB.Exec("DELETE FROM tasks")
	// DB.Exec("DELETE FROM recurring_tasks")

	log.Println("All data cleared")
	return nil
}
