package handlers

import (
	"database/sql"
	"net/http"
	"nudge/internal/database"
	"nudge/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateTaskRequest represents the task creation payload
type CreateTaskRequest struct {
	Name             string  `json:"name" binding:"required"`
	TaskCategory     string  `json:"task_category" binding:"required,oneof=ACTION ANCHOR TRANSIT"`
	Status           *string `json:"status" binding:"omitempty,oneof=CREATED COMPLETED FAILED DEFERRED"`
	ExpectedDuration *int32  `json:"expected_duration" binding:"omitempty,min=1,max=1440"`
	ExpectedUnits    *int32  `json:"expected_units" binding:"omitempty,min=1,max=1000"`
	ActualDuration   *int32  `json:"actual_duration" binding:"omitempty,min=1,max=1440"`
	ActualUnits      *int32  `json:"actual_units" binding:"omitempty,min=0"`
	Category         *string `json:"category"`
	Notes            *string `json:"notes"`
	Deadline         *string `json:"deadline"` // ISO 8601 format
	StartAt          *int32  `json:"start_at" binding:"omitempty,min=0,max=1439"`
}

// UpdateTaskRequest represents the task update payload
type UpdateTaskRequest struct {
	Name             *string `json:"name"`
	TaskCategory     *string `json:"task_category" binding:"omitempty,oneof=ACTION ANCHOR TRANSIT"`
	Status           *string `json:"status" binding:"omitempty,oneof=CREATED COMPLETED FAILED DEFERRED"`
	ExpectedDuration *int32  `json:"expected_duration" binding:"omitempty,min=1,max=1440"`
	ExpectedUnits    *int32  `json:"expected_units" binding:"omitempty,min=1,max=1000"`
	ActualDuration   *int32  `json:"actual_duration" binding:"omitempty,min=1,max=1440"`
	ActualUnits      *int32  `json:"actual_units" binding:"omitempty,min=0"`
	Category         *string `json:"category"`
	Notes            *string `json:"notes"`
	Deadline         *string `json:"deadline"` // ISO 8601 format, empty string to clear
	StartAt          *int32  `json:"start_at" binding:"omitempty,min=0,max=1439"`
}

// TaskResponse represents task data in responses
type TaskResponse struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	Name             string     `json:"name"`
	TaskCategory     string     `json:"task_category"`
	Status           string     `json:"status"`
	ExpectedDuration *int32     `json:"expected_duration,omitempty"`
	ExpectedUnits    *int32     `json:"expected_units,omitempty"`
	ActualDuration   *int32     `json:"actual_duration,omitempty"`
	ActualUnits      *int32     `json:"actual_units,omitempty"`
	Category         *string    `json:"category,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
	Deadline         *time.Time `json:"deadline,omitempty"`
	StartAt          *int32     `json:"start_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

// CreateTask creates a new task for the authenticated user
func CreateTask(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := models.Task{
		UserID:       userID,
		Name:         req.Name,
		TaskCategory: models.TaskCategory(req.TaskCategory),
		Status:       models.TaskStatusCreated,
	}

	// Set optional status
	if req.Status != nil {
		task.Status = models.TaskStatus(*req.Status)
	}

	// Set optional fields
	if req.ExpectedDuration != nil {
		task.ExpectedDuration = sql.NullInt32{Int32: *req.ExpectedDuration, Valid: true}
	}
	if req.ExpectedUnits != nil {
		task.ExpectedUnits = sql.NullInt32{Int32: *req.ExpectedUnits, Valid: true}
	}
	if req.ActualDuration != nil {
		task.ActualDuration = sql.NullInt32{Int32: *req.ActualDuration, Valid: true}
	}
	if req.ActualUnits != nil {
		task.ActualUnits = sql.NullInt32{Int32: *req.ActualUnits, Valid: true}
	}
	if req.StartAt != nil {
		task.StartAt = sql.NullInt32{Int32: *req.StartAt, Valid: true}
	}

	task.Category = req.Category
	task.Notes = req.Notes

	// Parse deadline if provided
	if req.Deadline != nil && *req.Deadline != "" {
		deadline, err := time.Parse(time.RFC3339, *req.Deadline)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deadline format, use ISO 8601"})
			return
		}
		task.Deadline = &deadline
	}

	// Set CompletedAt if status is COMPLETED
	if task.Status == models.TaskStatusCompleted {
		now := time.Now()
		task.CompletedAt = &now
	}

	if err := database.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Task created successfully",
		"task":    taskToResponse(task),
	})
}

// GetTasks retrieves all tasks for the authenticated user
func GetTasks(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	// Query parameters for filtering
	status := c.Query("status")
	taskCategory := c.Query("task_category")
	category := c.Query("category")
	search := c.Query("search")

	query := database.DB.Where("user_id = ?", userID)

	// Apply filters with validation
	// Status filter - validate against allowed values
	if status != "" {
		validStatuses := []string{"CREATED", "COMPLETED", "FAILED", "DEFERRED"}
		if isValidValue(status, validStatuses) {
			query = query.Where("status = ?", status)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
			return
		}
	}

	// Task category filter - validate against allowed values
	if taskCategory != "" {
		validCategories := []string{"ACTION", "ANCHOR", "TRANSIT"}
		if isValidValue(taskCategory, validCategories) {
			query = query.Where("task_category = ?", taskCategory)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task_category value"})
			return
		}
	}

	// Category filter - parameterized query is safe
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// Search filter - parameterized query with LIKE is safe
	// GORM handles the parameter binding safely
	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	var tasks []models.Task
	if err := query.Order("created_at DESC").Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	response := make([]TaskResponse, len(tasks))
	for i, task := range tasks {
		response[i] = taskToResponse(task)
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": response,
		"count": len(response),
	})
}

// GetTask retrieves a single task by ID for the authenticated user
func GetTask(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	taskID := c.Param("id")

	taskUUID, err := uuid.Parse(taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var task models.Task
	if err := database.DB.Where("id = ? AND user_id = ?", taskUUID, userID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task": taskToResponse(task),
	})
}

// UpdateTask updates a task for the authenticated user
func UpdateTask(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	taskID := c.Param("id")

	taskUUID, err := uuid.Parse(taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var task models.Task
	if err := database.DB.Where("id = ? AND user_id = ?", taskUUID, userID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Update fields if provided
	if req.Name != nil {
		task.Name = *req.Name
	}
	if req.TaskCategory != nil {
		task.TaskCategory = models.TaskCategory(*req.TaskCategory)
	}
	if req.Status != nil {
		oldStatus := task.Status
		task.Status = models.TaskStatus(*req.Status)

		// Set CompletedAt when status changes to COMPLETED
		if task.Status == models.TaskStatusCompleted && oldStatus != models.TaskStatusCompleted {
			now := time.Now()
			task.CompletedAt = &now
		}
		// Clear CompletedAt when status changes from COMPLETED
		if task.Status != models.TaskStatusCompleted && oldStatus == models.TaskStatusCompleted {
			task.CompletedAt = nil
		}
	}

	if req.ExpectedDuration != nil {
		task.ExpectedDuration = sql.NullInt32{Int32: *req.ExpectedDuration, Valid: true}
	}
	if req.ExpectedUnits != nil {
		task.ExpectedUnits = sql.NullInt32{Int32: *req.ExpectedUnits, Valid: true}
	}
	if req.ActualDuration != nil {
		task.ActualDuration = sql.NullInt32{Int32: *req.ActualDuration, Valid: true}
	}
	if req.ActualUnits != nil {
		task.ActualUnits = sql.NullInt32{Int32: *req.ActualUnits, Valid: true}
	}
	if req.StartAt != nil {
		task.StartAt = sql.NullInt32{Int32: *req.StartAt, Valid: true}
	}

	if req.Category != nil {
		task.Category = req.Category
	}
	if req.Notes != nil {
		task.Notes = req.Notes
	}

	// Handle deadline update
	if req.Deadline != nil {
		if *req.Deadline == "" {
			task.Deadline = nil
		} else {
			deadline, err := time.Parse(time.RFC3339, *req.Deadline)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deadline format, use ISO 8601"})
				return
			}
			task.Deadline = &deadline
		}
	}

	if err := database.DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task updated successfully",
		"task":    taskToResponse(task),
	})
}

// DeleteTask deletes a task for the authenticated user
func DeleteTask(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	taskID := c.Param("id")

	taskUUID, err := uuid.Parse(taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	result := database.DB.Where("id = ? AND user_id = ?", taskUUID, userID).Delete(&models.Task{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task deleted successfully",
	})
}

// Helper function to convert Task model to TaskResponse
func taskToResponse(task models.Task) TaskResponse {
	response := TaskResponse{
		ID:           task.ID.String(),
		UserID:       task.UserID.String(),
		Name:         task.Name,
		TaskCategory: string(task.TaskCategory),
		Status:       string(task.Status),
		Category:     task.Category,
		Notes:        task.Notes,
		Deadline:     task.Deadline,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
		CompletedAt:  task.CompletedAt,
	}

	if task.ExpectedDuration.Valid {
		response.ExpectedDuration = &task.ExpectedDuration.Int32
	}
	if task.ExpectedUnits.Valid {
		response.ExpectedUnits = &task.ExpectedUnits.Int32
	}
	if task.ActualDuration.Valid {
		response.ActualDuration = &task.ActualDuration.Int32
	}
	if task.ActualUnits.Valid {
		response.ActualUnits = &task.ActualUnits.Int32
	}
	if task.StartAt.Valid {
		response.StartAt = &task.StartAt.Int32
	}

	return response
}

// isValidValue checks if a value exists in the allowed list
func isValidValue(value string, allowedValues []string) bool {
	for _, allowed := range allowedValues {
		if value == allowed {
			return true
		}
	}
	return false
}
