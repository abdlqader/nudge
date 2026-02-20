package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RecurringTask defines recurrence patterns for task schedules
// This table contains ONLY recurrence configuration
type RecurringTask struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey"`
	Name      string    `gorm:"type:varchar(200);not null"`
	
	// Recurrence Configuration
	RecurrenceType        RecurrenceType     `gorm:"type:varchar(50);not null"`
	RecurrenceInterval    *int               `gorm:"default:1"` // For DAILY (every N days)
	RecurrenceDays        datatypes.JSON     `gorm:"type:json"` // For WEEKLY [0-6] where 0=Sunday
	RecurrenceDayOfMonth  *int               `gorm:"type:int"`  // For MONTHLY_DATE (1-31)
	RecurrencePattern     *string            `gorm:"type:varchar(50)"` // For MONTHLY_PATTERN
	RecurrenceEndDate     *time.Time         `gorm:"type:datetime"`
	
	// Status
	IsActive   bool      `gorm:"not null;default:true"`
	CreatedAt  time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt  time.Time `gorm:"not null;autoUpdateTime"`
	
	// Relationships
	Tasks []Task `gorm:"foreignKey:RecurringTaskID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook to generate UUID
func (rt *RecurringTask) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (RecurringTask) TableName() string {
	return "recurring_tasks"
}
