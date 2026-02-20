package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Task represents a single actionable item with success tracking
type Task struct {
	ID               uuid.UUID     `gorm:"type:char(36);primaryKey"`
	RecurringTaskID  *uuid.UUID    `gorm:"type:char(36);index"` // NULL for standalone tasks
	
	// Basic Information
	Name         string       `gorm:"type:varchar(200);not null"`
	TaskType     TaskType     `gorm:"type:varchar(50);not null"`
	TaskCategory TaskCategory `gorm:"type:varchar(50);not null;default:'ACTION'"`
	IsCommute    bool         `gorm:"not null;default:false"`
	
	// Status and Priority
	Status   TaskStatus `gorm:"type:varchar(50);not null;default:'PENDING';index"`
	Priority Priority   `gorm:"type:int;not null;default:2"`
	
	// Expected Values
	ExpectedDuration sql.NullInt32 `gorm:"type:int"` // minutes (1-1440)
	ExpectedUnits    sql.NullInt32 `gorm:"type:int"` // quantity (1-1000)
	
	// Actual Values
	ActualDuration sql.NullInt32 `gorm:"type:int"` // minutes (1-1440)
	ActualUnits    sql.NullInt32 `gorm:"type:int"` // quantity (0-expected_units)
	
	// Metadata
	Category *string `gorm:"type:varchar(50)"` // User-defined tag
	Notes    *string `gorm:"type:text"`
	Deadline *time.Time `gorm:"type:datetime"`
	
	// Timestamps
	CreatedAt   time.Time  `gorm:"not null;autoCreateTime;index"`
	UpdatedAt   time.Time  `gorm:"not null;autoUpdateTime"`
	CompletedAt *time.Time `gorm:"type:datetime;index"`
	
	// Computed Field (not stored, calculated on-demand)
	SuccessPercentage *float64 `gorm:"-"`
	
	// Relationships
	RecurringTask *RecurringTask `gorm:"foreignKey:RecurringTaskID;references:ID"`
}

// BeforeCreate hook to generate UUID and enforce business rules
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	
	// Auto-set commute properties
	if t.TaskType == TaskTypeCommute {
		t.IsCommute = true
		t.TaskCategory = TaskCategoryTransit
	}
	
	return nil
}

// BeforeSave hook to validate business rules
func (t *Task) BeforeSave(tx *gorm.DB) error {
	// Validate task type requirements
	if t.TaskType == TaskTypeUnitBased && !t.ExpectedUnits.Valid {
		return gorm.ErrInvalidValue
	}
	
	if (t.TaskType == TaskTypeTimeBased || t.TaskType == TaskTypeCommute) && !t.ExpectedDuration.Valid {
		return gorm.ErrInvalidValue
	}
	
	return nil
}

// CalculateSuccess computes the success percentage based on task type
func (t *Task) CalculateSuccess() *float64 {
	if t.Status != TaskStatusCompleted {
		return nil
	}
	
	var success float64
	
	switch t.TaskType {
	case TaskTypeUnitBased:
		// Formula: (actual_units / expected_units) × 100, capped at 100%
		if !t.ExpectedUnits.Valid || !t.ActualUnits.Valid {
			return nil
		}
		success = (float64(t.ActualUnits.Int32) / float64(t.ExpectedUnits.Int32)) * 100
		if success > 100 {
			success = 100
		}
		
	case TaskTypeTimeBased:
		if t.IsCommute {
			// Commute: On-time or early = 100%, late is penalized, capped at 100%
			if !t.ExpectedDuration.Valid || !t.ActualDuration.Valid {
				return nil
			}
			if t.ActualDuration.Int32 <= t.ExpectedDuration.Int32 {
				success = 100
			} else {
				success = (float64(t.ExpectedDuration.Int32) / float64(t.ActualDuration.Int32)) * 100
			}
		} else {
			// Regular time-based: Faster = higher success, capped at 150%
			if !t.ExpectedDuration.Valid || !t.ActualDuration.Valid {
				return nil
			}
			success = (float64(t.ExpectedDuration.Int32) / float64(t.ActualDuration.Int32)) * 100
			if success > 150 {
				success = 150
			}
		}
		
	case TaskTypeCommute:
		// Same as commute logic above
		if !t.ExpectedDuration.Valid || !t.ActualDuration.Valid {
			return nil
		}
		if t.ActualDuration.Int32 <= t.ExpectedDuration.Int32 {
			success = 100
		} else {
			success = (float64(t.ExpectedDuration.Int32) / float64(t.ActualDuration.Int32)) * 100
		}
	}
	
	return &success
}

// TableName specifies the table name
func (Task) TableName() string {
	return "tasks"
}
