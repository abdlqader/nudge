package models

// Enums for Task
type TaskCategory string

const (
	TaskCategoryAnchor  TaskCategory = "ANCHOR"
	TaskCategoryTransit TaskCategory = "TRANSIT"
	TaskCategoryAction  TaskCategory = "ACTION"
)

type TaskStatus string

const (
	TaskStatusCreated   TaskStatus = "CREATED"
	TaskStatusCompleted TaskStatus = "COMPLETED"
	TaskStatusFailed    TaskStatus = "FAILED"
	TaskStatusDeferred  TaskStatus = "DEFERRED"
)

type RecurrenceType string

const (
	RecurrenceTypeDaily          RecurrenceType = "DAILY"
	RecurrenceTypeWeekly         RecurrenceType = "WEEKLY"
	RecurrenceTypeMonthlyDate    RecurrenceType = "MONTHLY_DATE"
	RecurrenceTypeMonthlyPattern RecurrenceType = "MONTHLY_PATTERN"
)
