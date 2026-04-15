package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RecurringTask is a template that defines a repeating task pattern.
// It does NOT store actual task instances — those live in the tasks table
// linked via recurring_task_id.
type RecurringTask struct {
	ID     uuid.UUID `gorm:"type:char(36);primaryKey"`
	UserID uuid.UUID `gorm:"type:char(36);not null;index"`
	User   User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`

	Name         string       `gorm:"type:varchar(200);not null"`
	TaskCategory TaskCategory `gorm:"type:varchar(50);not null;default:'ACTION'"`

	// Recurrence rule
	RecurrenceType       RecurrenceType `gorm:"type:varchar(50);not null"`
	RecurrenceInterval   sql.NullInt32  `gorm:"type:int"`         // DAILY: every N days (default 1)
	RecurrenceDays       *string        `gorm:"type:text"`        // WEEKLY: JSON e.g. "[1,3]" Mon=1…Sun=7
	RecurrenceDayOfMonth sql.NullInt32  `gorm:"type:int"`         // MONTHLY_DATE: 1–31
	RecurrencePattern    *string        `gorm:"type:varchar(50)"` // MONTHLY_PATTERN: "first_monday" etc.

	// Task defaults (copied into realised instances)
	ExpectedDuration sql.NullInt32 `gorm:"type:int"` // minutes
	ExpectedUnits    sql.NullInt32 `gorm:"type:int"`
	Category         *string       `gorm:"type:varchar(50)"`
	Notes            *string       `gorm:"type:text"`
	StartAt          sql.NullInt32 `gorm:"type:int"` // minutes from midnight (0–1439)

	IsActive  bool      `gorm:"not null;default:true"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime"`
}

func (rt *RecurringTask) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

func (RecurringTask) TableName() string {
	return "recurring_tasks"
}
