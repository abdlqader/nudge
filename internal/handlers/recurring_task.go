package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"nudge/internal/database"
	"nudge/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ── Request / Response types ─────────────────────────────────────────────

type CreateRecurringTaskRequest struct {
	Name                 string  `json:"name" binding:"required"`
	TaskCategory         string  `json:"task_category" binding:"required,oneof=ACTION ANCHOR TRANSIT"`
	RecurrenceType       string  `json:"recurrence_type" binding:"required,oneof=DAILY WEEKLY MONTHLY_DATE MONTHLY_PATTERN"`
	RecurrenceInterval   *int32  `json:"recurrence_interval" binding:"omitempty,min=1,max=365"`
	RecurrenceDays       []int   `json:"recurrence_days"` // [1..7] Mon=1
	RecurrenceDayOfMonth *int32  `json:"recurrence_day_of_month" binding:"omitempty,min=1,max=31"`
	RecurrencePattern    *string `json:"recurrence_pattern"`
	ExpectedDuration     *int32  `json:"expected_duration" binding:"omitempty,min=1,max=1440"`
	ExpectedUnits        *int32  `json:"expected_units" binding:"omitempty,min=1,max=1000"`
	Category             *string `json:"category"`
	Notes                *string `json:"notes"`
	StartAt              *int32  `json:"start_at" binding:"omitempty,min=0,max=1439"`
}

type UpdateRecurringTaskRequest struct {
	Name                 *string `json:"name"`
	TaskCategory         *string `json:"task_category" binding:"omitempty,oneof=ACTION ANCHOR TRANSIT"`
	RecurrenceType       *string `json:"recurrence_type" binding:"omitempty,oneof=DAILY WEEKLY MONTHLY_DATE MONTHLY_PATTERN"`
	RecurrenceInterval   *int32  `json:"recurrence_interval" binding:"omitempty,min=1,max=365"`
	RecurrenceDays       []int   `json:"recurrence_days"`
	RecurrenceDayOfMonth *int32  `json:"recurrence_day_of_month" binding:"omitempty,min=1,max=31"`
	RecurrencePattern    *string `json:"recurrence_pattern"`
	ExpectedDuration     *int32  `json:"expected_duration" binding:"omitempty,min=1,max=1440"`
	ExpectedUnits        *int32  `json:"expected_units" binding:"omitempty,min=1,max=1000"`
	Category             *string `json:"category"`
	Notes                *string `json:"notes"`
	StartAt              *int32  `json:"start_at" binding:"omitempty,min=0,max=1439"`
	IsActive             *bool   `json:"is_active"`
}

type RecurringTaskResponse struct {
	ID                   string  `json:"id"`
	UserID               string  `json:"user_id"`
	Name                 string  `json:"name"`
	TaskCategory         string  `json:"task_category"`
	RecurrenceType       string  `json:"recurrence_type"`
	RecurrenceInterval   *int32  `json:"recurrence_interval,omitempty"`
	RecurrenceDays       []int   `json:"recurrence_days,omitempty"`
	RecurrenceDayOfMonth *int32  `json:"recurrence_day_of_month,omitempty"`
	RecurrencePattern    *string `json:"recurrence_pattern,omitempty"`
	ExpectedDuration     *int32  `json:"expected_duration,omitempty"`
	ExpectedUnits        *int32  `json:"expected_units,omitempty"`
	Category             *string `json:"category,omitempty"`
	Notes                *string `json:"notes,omitempty"`
	StartAt              *int32  `json:"start_at,omitempty"`
	IsActive             bool    `json:"is_active"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
}

// ── Helpers ──────────────────────────────────────────────────────────────

func toNullInt32(v *int32) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *v, Valid: true}
}

func recurringTaskToResponse(rt models.RecurringTask) RecurringTaskResponse {
	resp := RecurringTaskResponse{
		ID:             rt.ID.String(),
		UserID:         rt.UserID.String(),
		Name:           rt.Name,
		TaskCategory:   string(rt.TaskCategory),
		RecurrenceType: string(rt.RecurrenceType),
		Category:       rt.Category,
		Notes:          rt.Notes,
		IsActive:       rt.IsActive,
		CreatedAt:      rt.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      rt.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if rt.RecurrenceInterval.Valid {
		resp.RecurrenceInterval = &rt.RecurrenceInterval.Int32
	}
	if rt.RecurrenceDays != nil && *rt.RecurrenceDays != "" {
		var days []int
		if err := json.Unmarshal([]byte(*rt.RecurrenceDays), &days); err == nil {
			resp.RecurrenceDays = days
		}
	}
	if rt.RecurrenceDayOfMonth.Valid {
		resp.RecurrenceDayOfMonth = &rt.RecurrenceDayOfMonth.Int32
	}
	resp.RecurrencePattern = rt.RecurrencePattern
	if rt.ExpectedDuration.Valid {
		resp.ExpectedDuration = &rt.ExpectedDuration.Int32
	}
	if rt.ExpectedUnits.Valid {
		resp.ExpectedUnits = &rt.ExpectedUnits.Int32
	}
	if rt.StartAt.Valid {
		resp.StartAt = &rt.StartAt.Int32
	}
	return resp
}

// ── Handlers ─────────────────────────────────────────────────────────────

// CreateRecurringTask creates a new recurring task template.
func CreateRecurringTask(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req CreateRecurringTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rt := models.RecurringTask{
		UserID:         userID,
		Name:           req.Name,
		TaskCategory:   models.TaskCategory(req.TaskCategory),
		RecurrenceType: models.RecurrenceType(req.RecurrenceType),
		IsActive:       true,
	}

	if req.RecurrenceInterval != nil {
		rt.RecurrenceInterval = toNullInt32(req.RecurrenceInterval)
	}
	if len(req.RecurrenceDays) > 0 {
		b, _ := json.Marshal(req.RecurrenceDays)
		s := string(b)
		rt.RecurrenceDays = &s
	}
	if req.RecurrenceDayOfMonth != nil {
		rt.RecurrenceDayOfMonth = toNullInt32(req.RecurrenceDayOfMonth)
	}
	rt.RecurrencePattern = req.RecurrencePattern
	if req.ExpectedDuration != nil {
		rt.ExpectedDuration = toNullInt32(req.ExpectedDuration)
	}
	if req.ExpectedUnits != nil {
		rt.ExpectedUnits = toNullInt32(req.ExpectedUnits)
	}
	if req.StartAt != nil {
		rt.StartAt = toNullInt32(req.StartAt)
	}
	rt.Category = req.Category
	rt.Notes = req.Notes

	if err := database.DB.Create(&rt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create recurring task"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Recurring task created successfully",
		"recurring_task": recurringTaskToResponse(rt),
	})
}

// GetRecurringTasks lists all recurring task templates for the user.
func GetRecurringTasks(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var rts []models.RecurringTask
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&rts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recurring tasks"})
		return
	}

	resp := make([]RecurringTaskResponse, len(rts))
	for i, rt := range rts {
		resp[i] = recurringTaskToResponse(rt)
	}

	c.JSON(http.StatusOK, gin.H{
		"recurring_tasks": resp,
		"count":           len(resp),
	})
}

// GetRecurringTask retrieves a single recurring task template.
func GetRecurringTask(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	rtID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recurring task ID"})
		return
	}

	var rt models.RecurringTask
	if err := database.DB.Where("id = ? AND user_id = ?", rtID, userID).First(&rt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recurring task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"recurring_task": recurringTaskToResponse(rt)})
}

// UpdateRecurringTask updates a recurring task template.
// Only affects future (unrealized) occurrences — past Task records are untouched.
func UpdateRecurringTask(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	rtID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recurring task ID"})
		return
	}

	var req UpdateRecurringTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var rt models.RecurringTask
	if err := database.DB.Where("id = ? AND user_id = ?", rtID, userID).First(&rt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recurring task not found"})
		return
	}

	if req.Name != nil {
		rt.Name = *req.Name
	}
	if req.TaskCategory != nil {
		rt.TaskCategory = models.TaskCategory(*req.TaskCategory)
	}
	if req.RecurrenceType != nil {
		rt.RecurrenceType = models.RecurrenceType(*req.RecurrenceType)
	}
	if req.RecurrenceInterval != nil {
		rt.RecurrenceInterval = toNullInt32(req.RecurrenceInterval)
	}
	if req.RecurrenceDays != nil {
		if len(req.RecurrenceDays) > 0 {
			b, _ := json.Marshal(req.RecurrenceDays)
			s := string(b)
			rt.RecurrenceDays = &s
		} else {
			rt.RecurrenceDays = nil
		}
	}
	if req.RecurrenceDayOfMonth != nil {
		rt.RecurrenceDayOfMonth = toNullInt32(req.RecurrenceDayOfMonth)
	}
	if req.RecurrencePattern != nil {
		rt.RecurrencePattern = req.RecurrencePattern
	}
	if req.ExpectedDuration != nil {
		rt.ExpectedDuration = toNullInt32(req.ExpectedDuration)
	}
	if req.ExpectedUnits != nil {
		rt.ExpectedUnits = toNullInt32(req.ExpectedUnits)
	}
	if req.StartAt != nil {
		rt.StartAt = toNullInt32(req.StartAt)
	}
	if req.Category != nil {
		rt.Category = req.Category
	}
	if req.Notes != nil {
		rt.Notes = req.Notes
	}
	if req.IsActive != nil {
		rt.IsActive = *req.IsActive
	}

	if err := database.DB.Save(&rt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recurring task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Recurring task updated successfully",
		"recurring_task": recurringTaskToResponse(rt),
	})
}

// DeleteRecurringTask removes a recurring task template.
// Realized Task records (past occurrences) are kept, their recurring_task_id remains pointing to the deleted template.
func DeleteRecurringTask(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	rtID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recurring task ID"})
		return
	}

	result := database.DB.Where("id = ? AND user_id = ?", rtID, userID).Delete(&models.RecurringTask{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recurring task"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recurring task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recurring task deleted successfully"})
}
