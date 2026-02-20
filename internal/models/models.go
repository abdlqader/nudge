package models

// Enums for Task
type TaskType string

const (
	TaskTypeUnitBased TaskType = "UNIT_BASED"
	TaskTypeTimeBased TaskType = "TIME_BASED"
	TaskTypeCommute   TaskType = "COMMUTE"
)

type TaskCategory string

const (
	TaskCategoryAnchor  TaskCategory = "ANCHOR"
	TaskCategoryTransit TaskCategory = "TRANSIT"
	TaskCategoryAction  TaskCategory = "ACTION"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "PENDING"
	TaskStatusInProgress TaskStatus = "IN_PROGRESS"
	TaskStatusCompleted  TaskStatus = "COMPLETED"
	TaskStatusFailed     TaskStatus = "FAILED"
	TaskStatusDeferred   TaskStatus = "DEFERRED"
)

type Priority int

const (
	PriorityLow      Priority = 1
	PriorityMedium   Priority = 2
	PriorityHigh     Priority = 3
	PriorityCritical Priority = 4
)

// Enums for RecurringTask
type RecurrenceType string

const (
	RecurrenceTypeDaily          RecurrenceType = "DAILY"
	RecurrenceTypeWeekly         RecurrenceType = "WEEKLY"
	RecurrenceTypeMonthlyDate    RecurrenceType = "MONTHLY_DATE"
	RecurrenceTypeMonthlyPattern RecurrenceType = "MONTHLY_PATTERN"
)
