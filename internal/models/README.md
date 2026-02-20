# Nudge Models Documentation

## Overview

This document describes the data models implemented for the Nudge task management system.

## File Structure

```
internal/models/
├── models.go           # Enum type definitions
├── task.go             # Task model
└── recurring_task.go   # RecurringTask model (templates)
```

## Enums

### TaskType
- `UNIT_BASED`: Tasks measured in discrete units (e.g., "read 3 chapters")
- `TIME_BASED`: Tasks measured in duration (e.g., "study for 2 hours")
- `COMMUTE`: Special time-based task for travel

### TaskCategory
- `ANCHOR`: Non-negotiable time blocks (sleep, family time, meals)
- `TRANSIT`: Movement between locations (commute, drive, travel)
- `ACTION`: Productive tasks (work, deliverables, reports)

### TaskStatus
- `PENDING`: Not started
- `IN_PROGRESS`: Currently active
- `COMPLETED`: Finished (may be partial or full success)
- `FAILED`: Explicitly marked as failed or abandoned
- `DEFERRED`: Postponed to future date

### Priority
- `LOW`: 1
- `MEDIUM`: 2
- `HIGH`: 3
- `CRITICAL`: 4

### RecurrenceType
- `DAILY`: Repeats every day or every N days
- `WEEKLY`: Repeats on specific days of the week
- `MONTHLY_DATE`: Repeats on specific day of month (e.g., 15th)
- `MONTHLY_PATTERN`: Repeats on pattern (e.g., "first_monday")

## Models

### RecurringTask

Template for recurring tasks. These are NOT actual tasks, but patterns used to generate task instances.

**Fields:**
- `ID` (UUID): Primary key, auto-generated
- `Name` (string): Template name/label
- `RecurrenceType` (enum): Type of recurrence pattern
- `RecurrenceInterval` (*int): For DAILY - repeat every N days (default: 1)
- `RecurrenceDays` (JSON): For WEEKLY - array of weekdays [0-6] where 0=Sunday
- `RecurrenceDayOfMonth` (*int): For MONTHLY_DATE - day of month (1-31)
- `RecurrencePattern` (*string): For MONTHLY_PATTERN - pattern like "first_monday"
- `RecurrenceEndDate` (*time.Time): Optional end date for recurrence
- `IsActive` (bool): Enable/disable template without deleting
- Timestamps: `CreatedAt`, `UpdatedAt`

**Relationships:**
- One-to-many with Tasks (CASCADE delete)

**Examples:**
```go
// Daily standup every day
RecurringTask{
    Name: "Daily Standup",
    RecurrenceType: RecurrenceTypeDaily,
    RecurrenceInterval: intPtr(1),
}

// Team meeting on Monday and Wednesday
RecurringTask{
    Name: "Team Meeting",
    RecurrenceType: RecurrenceTypeWeekly,
    RecurrenceDays: datatypes.JSON([]byte(`[1,3]`)),
}

// Pay rent on 1st of month
RecurringTask{
    Name: "Pay Rent",
    RecurrenceType: RecurrenceTypeMonthlyDate,
    RecurrenceDayOfMonth: intPtr(1),
}

// Board meeting first Monday of month
RecurringTask{
    Name: "Board Meeting",
    RecurrenceType: RecurrenceTypeMonthlyPattern,
    RecurrencePattern: strPtr("first_monday"),
}
```

### Task

Individual actionable item with complete task data and success tracking.

**Fields:**
- `ID` (UUID): Primary key, auto-generated
- `RecurringTaskID` (*UUID): Link to template (NULL for standalone tasks)
- `Name` (string): Task description
- `TaskType` (enum): Classification (UNIT_BASED, TIME_BASED, COMMUTE)
- `TaskCategory` (enum): Category (ANCHOR, TRANSIT, ACTION)
- `IsCommute` (bool): Auto-set for COMMUTE tasks
- `Status` (enum): Current state
- `Priority` (enum): Importance level
- `ExpectedDuration` (sql.NullInt32): Expected minutes (1-1440)
- `ExpectedUnits` (sql.NullInt32): Expected quantity (1-1000)
- `ActualDuration` (sql.NullInt32): Actual minutes taken
- `ActualUnits` (sql.NullInt32): Actual quantity completed
- `Category` (*string): User-defined tag (max 50 chars)
- `Notes` (*string): Additional context (max 1000 chars)
- `Deadline` (*time.Time): Due date/time
- Timestamps: `CreatedAt`, `UpdatedAt`, `CompletedAt`
- `SuccessPercentage` (*float64): Computed field (not stored)

**Business Rules:**

1. **UNIT_BASED tasks:**
   - MUST have `ExpectedUnits`
   - Success = (ActualUnits / ExpectedUnits) × 100, capped at 100%

2. **TIME_BASED tasks (non-commute):**
   - MUST have `ExpectedDuration`
   - Success = (ExpectedDuration / ActualDuration) × 100, capped at 150%
   - Faster completion = higher success

3. **COMMUTE tasks:**
   - MUST have `ExpectedDuration`
   - Auto-sets: `IsCommute=true`, `TaskCategory=TRANSIT`
   - On-time or early = 100%
   - Late = (ExpectedDuration / ActualDuration) × 100, capped at 100%

**Methods:**
- `BeforeCreate()`: Generates UUID, enforces commute properties
- `BeforeSave()`: Validates business rules
- `CalculateSuccess()`: Computes success percentage

**Examples:**
```go
// Unit-based task
Task{
    Name: "Read 3 chapters",
    TaskType: TaskTypeUnitBased,
    TaskCategory: TaskCategoryAction,
    ExpectedUnits: sql.NullInt32{Int32: 3, Valid: true},
    ExpectedDuration: sql.NullInt32{Int32: 150, Valid: true},
}

// Time-based task
Task{
    Name: "Deep work session",
    TaskType: TaskTypeTimeBased,
    TaskCategory: TaskCategoryAction,
    ExpectedDuration: sql.NullInt32{Int32: 120, Valid: true},
}

// Commute task
Task{
    Name: "Morning commute",
    TaskType: TaskTypeCommute,
    ExpectedDuration: sql.NullInt32{Int32: 30, Valid: true},
    // Auto-set: IsCommute=true, TaskCategory=TRANSIT
}

// Anchor task
Task{
    Name: "Sleep",
    TaskType: TaskTypeTimeBased,
    TaskCategory: TaskCategoryAnchor,
    Priority: PriorityCritical,
    ExpectedDuration: sql.NullInt32{Int32: 480, Valid: true}, // 8 hours
}
```

## Database Indexes

The following indexes are created for performance:

**Tasks:**
- `idx_tasks_created_at`: For date-based queries
- `idx_tasks_status`: For filtering by status
- `idx_tasks_completed_at`: For completion tracking
- `idx_tasks_recurring_id`: For recurring task lookups
- `idx_tasks_is_commute`: For commute analysis

**RecurringTasks:**
- `idx_recurring_tasks_active`: For active template queries

## Success Calculation Examples

```go
// Unit-based: Read 3 chapters, completed 2
// Success = (2 / 3) × 100 = 66.67%

// Time-based: Expected 120 min, took 90 min
// Success = (120 / 90) × 100 = 133.33%

// Commute: Expected 30 min, took 35 min
// Success = (30 / 35) × 100 = 85.71%

// Sleep (time-based): Expected 480 min, got 450 min
// Success = (480 / 450) × 100 = 106.67%
```

## Testing

Run the application to see seed data:

```bash
go run main.go
```

This creates:
- 4 recurring task templates (daily, weekly, monthly patterns)
- 6 sample tasks (unit-based, time-based, commute, anchor tasks)
- Success calculations for completed tasks

## Next Steps

1. **API Handlers**: Create REST endpoints for CRUD operations
2. **Task Generation**: Implement cron job to generate tasks from templates
3. **Statistics**: Add queries for success rate aggregation
4. **Validation**: Add comprehensive validation middleware
