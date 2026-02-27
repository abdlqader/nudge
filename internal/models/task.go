package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Task represents a single actionable item with success tracking
type Task struct {
	ID uuid.UUID `gorm:"type:char(36);primaryKey"`

	// User relationship
	UserID uuid.UUID `gorm:"type:char(36);not null;index"`
	User   User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`

	// Basic Information
	Name         string       `gorm:"type:varchar(200);not null"`
	TaskCategory TaskCategory `gorm:"type:varchar(50);not null;default:'ACTION'"`

	// Status and Priority
	Status TaskStatus `gorm:"type:varchar(50);not null;default:'CREATED';index"`

	// Expected Values
	ExpectedDuration sql.NullInt32 `gorm:"type:int"` // minutes (1-1440)
	ExpectedUnits    sql.NullInt32 `gorm:"type:int"` // quantity (1-1000)

	// Actual Values
	ActualDuration sql.NullInt32 `gorm:"type:int"` // minutes (1-1440)
	ActualUnits    sql.NullInt32 `gorm:"type:int"` // quantity (0-expected_units)

	// Metadata
	Category *string    `gorm:"type:varchar(50)"` // User-defined tag
	Notes    *string    `gorm:"type:text"`        // Description or comments about the task
	Deadline *time.Time `gorm:"type:datetime"`

	// Timestamps
	CreatedAt   time.Time  `gorm:"not null;autoCreateTime;index"`
	UpdatedAt   time.Time  `gorm:"not null;autoUpdateTime"`
	CompletedAt *time.Time `gorm:"type:datetime;index"`

	// Computed Field (not stored, calculated on-demand)
	SuccessPercentage *float64 `gorm:"-"`
}

// BeforeCreate hook to generate UUID and enforce business rules
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}

	return nil
}

// BeforeSave hook to validate business rules
func (t *Task) BeforeSave(tx *gorm.DB) error {
	// Validate task type requirements
	return nil
}

// CalculateSuccess computes the success percentage based on task type
func (t *Task) CalculateSuccess() *float64 {

	var successTime float64
	var successUnit float64

	if t.ExpectedDuration.Valid || t.ActualDuration.Valid {
		successTime = (float64(t.ExpectedDuration.Int32) / float64(t.ActualDuration.Int32)) * 100
		if successTime > 100 {
			successTime = 100
		}
	}

	if t.ExpectedUnits.Valid || t.ActualUnits.Valid {
		successUnit = (float64(t.ActualUnits.Int32) / float64(t.ExpectedUnits.Int32)) * 100
		if successUnit > 100 {
			successUnit = 100
		}

		result := successUnit + successTime/2
		return &result
	}

	return &successTime
}

// TableName specifies the table name
func (Task) TableName() string {
	return "tasks"
}
