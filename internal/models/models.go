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
