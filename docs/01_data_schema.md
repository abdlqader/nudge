# Module 01: Data Schema Design

## Overview
This document defines the complete data architecture for the Nudge application, including Pydantic models for API validation, PostgreSQL schema for persistence, and business logic constraints.

---

## Core Principles

1. **Immutability**: Tasks maintain history through status changes, not deletions
2. **Audit Trail**: All modifications are timestamped
3. **Type Safety**: Strong typing through Pydantic
4. **Flexibility**: Support both unit-based and time-based tasks
5. **Commute Integration**: Commutes are tasks with special flags

---

## Entity Relationship Diagram

```
┌─────────────────────┐
│       User          │
│  (Future Phase)     │
└──────────┬──────────┘
           │ 1:N
           ├────────────────────────┐
           │                        │
           ▼                        ▼
┌─────────────────────┐    ┌─────────────────────┐
│  Recurring_Tasks    │    │       Task          │
│  - id (PK)          │    │  - id (PK)          │
│  - name             │    │  - name             │
│  - recurrence_type  │    │  - task_type        │
│  - recurrence_*     │    │  - is_commute       │
│  - task_type        │    │  - status           │
│  - expected_*       │    │  - expected_*       │
└──────────┬──────────┘    │  - actual_*         │
           │               │  - recurring_task_id│
           │ 1:N           │  - created_at       │
           └──────────────►│  - completed_at     │
                           └─────────────────────┘

Note: Daily statistics are computed on-demand via queries on the tasks table.
```

---

## 1. Recurring_Tasks Entity

### Purpose
Defines recurring task templates with schedule patterns. Instances are automatically generated in the `tasks` table.

### Pydantic Model Specification

RecurrenceType (Enum):
- DAILY: Repeats every day or every N days
- WEEKLY: Repeats on specific days of the week
- MONTHLY_DATE: Repeats on specific day of month (e.g., 15th of every month)
- MONTHLY_PATTERN: Repeats on pattern (e.g., "first Monday", "last Friday")

### Fields Table

| Field Name | Type | Nullable | Default | Description | Validation Rules |
|------------|------|----------|---------|-------------|------------------|
| `id` | UUID | No | auto | Unique identifier | Primary Key |
| `name` | String | No | - | Task description | 1-200 characters |
| `task_type` | TaskType | No | - | Classification | Enum: UNIT_BASED, TIME_BASED, COMMUTE |
| `task_category` | TaskCategory | No | ACTION | Category type | Enum: ANCHOR, TRANSIT, ACTION |
| `priority` | Priority | No | MEDIUM | Importance level | Enum value (1-4) |
| `expected_duration` | Integer | Yes | null | Expected minutes | 1-1440 (max 24 hours) |
| `expected_units` | Integer | Yes | null | Expected quantity | 1-1000 |
| `category` | String | Yes | null | User-defined tag | Max 50 chars |
| `notes` | Text | Yes | null | Additional context | Max 1000 chars |
| `recurrence_type` | RecurrenceType | No | - | Recurrence pattern | Enum value |
| `recurrence_interval` | Integer | Yes | 1 | Repeat every N periods | For DAILY (e.g., every 2 days) |
| `recurrence_days` | JSON | Yes | null | Days of week | Array: [0-6] where 0=Sunday |
| `recurrence_day_of_month` | Integer | Yes | null | Day of month | 1-31 for MONTHLY_DATE |
| `recurrence_pattern` | String | Yes | null | Monthly pattern | "first_monday", "last_friday", etc. |
| `recurrence_end_date` | DateTime | Yes | null | Stop recurring after | Optional end date |
| `is_active` | Boolean | No | True | Enable/disable schedule | Can pause without deleting |
| `created_at` | DateTime | No | now() | Creation timestamp | ISO 8601 |
| `updated_at` | DateTime | No | now() | Last modified | ISO 8601, auto-update |

### Business Logic Constraints

**Task Type Rules**:
```
IF task_type == UNIT_BASED:
    REQUIRE: expected_units IS NOT NULL
    OPTIONAL: expected_duration
    
IF task_type == TIME_BASED:
    REQUIRE: expected_duration IS NOT NULL
    
IF task_type == COMMUTE:
    REQUIRE: expected_duration IS NOT NULL
    SET: task_category = TRANSIT
```

**Recurrence Rules**:
```
IF recurrence_type == DAILY:
    OPTIONAL: recurrence_interval (default 1 = every day, 2 = every 2 days)
    EXAMPLE: "Daily standup", "Take vitamins every day"
    
IF recurrence_type == WEEKLY:
    REQUIRE: recurrence_days IS NOT NULL (array of 0-6)
    EXAMPLE: "Team meeting every Monday and Wednesday" → [1, 3]
    
IF recurrence_type == MONTHLY_DATE:
    REQUIRE: recurrence_day_of_month (1-31)
    EXAMPLE: "Pay rent on the 1st" → day_of_month = 1
    
IF recurrence_type == MONTHLY_PATTERN:
    REQUIRE: recurrence_pattern
    PATTERNS: "first_monday", "first_tuesday", ... "first_sunday",
              "second_monday", ... "second_sunday",
              "third_monday", ... "third_sunday",
              "fourth_monday", ... "fourth_sunday",
              "last_monday", ... "last_sunday"
    EXAMPLE: "Board meeting first Monday of month" → "first_monday"
```

---

## 2. Task Entity

### Purpose
Represents a single actionable item with success tracking capabilities. Can be standalone or an instance of a recurring schedule.

### Pydantic Model Specification

TaskType (Enum):
- UNIT_BASED: Tasks measured in discrete units (e.g., "read 3 chapters")
- TIME_BASED: Tasks measured in duration (e.g., "study for 2 hours")
- COMMUTE: Special time-based task for travel

TaskCategory (Enum):
- ANCHOR: Non-negotiable time blocks (sleep, family time, meals)
- TRANSIT: Movement between locations (commute, drive, travel)
- ACTION: Productive tasks (work, deliverables, reports)

TaskStatus (Enum):
- PENDING: Not started
- IN_PROGRESS: Currently active
- COMPLETED: Finished (may be partial or full success)
- FAILED: Explicitly marked as failed or abandoned
- DEFERRED: Postponed to future date

Priority (Enum):
- LOW: 1
- MEDIUM: 2
- HIGH: 3
- CRITICAL: 4

### Fields Table

| Field Name | Type | Nullable | Default | Description | Validation Rules |
|------------|------|----------|---------|-------------|------------------|
| `id` | UUID | No | auto | Unique identifier | Primary Key |
| `name` | String | No | - | Task description | 1-200 characters |
| `task_type` | TaskType | No | - | Classification | Enum value |
| `task_category` | TaskCategory | No | ACTION | Category type | Enum: ANCHOR, TRANSIT, ACTION |
| `is_commute` | Boolean | No | False | Commute flag | Auto-set if task_type=COMMUTE |
| `status` | TaskStatus | No | PENDING | Current state | Enum value |
| `priority` | Priority | No | MEDIUM | Importance level | Enum value |
| `expected_duration` | Integer | Yes | null | Expected minutes | 1-1440 (max 24 hours) |
| `actual_duration` | Integer | Yes | null | Actual minutes | 1-1440, set on completion |
| `expected_units` | Integer | Yes | null | Expected quantity | 1-1000 |
| `actual_units` | Integer | Yes | null | Completed quantity | 0-expected_units |
| `category` | String | Yes | null | User-defined tag | Max 50 chars |
| `notes` | Text | Yes | null | Additional context | Max 1000 chars |
| `deadline` | DateTime | Yes | null | Due date/time | Must be future |
| `created_at` | DateTime | No | now() | Creation timestamp | ISO 8601 |
| `updated_at` | DateTime | No | now() | Last modified | ISO 8601, auto-update |
| `completed_at` | DateTime | Yes | null | Completion timestamp | Set when status=COMPLETED |
| `success_percentage` | Float | Yes | null | Computed success | 0.0-100.0, calculated field |
| `recurring_task_id` | UUID | Yes | null | Link to recurring template | References recurring_tasks(id) |

### Business Logic Constraints

**Task Type Rules**:
```
IF task_type == UNIT_BASED:
    REQUIRE: expected_units IS NOT NULL
    OPTIONAL: expected_duration
    
IF task_type == TIME_BASED:
    REQUIRE: expected_duration IS NOT NULL
    SET: expected_units = NULL
    
IF task_type == COMMUTE:
    REQUIRE: expected_duration IS NOT NULL
    SET: is_commute = TRUE
    SET: expected_units = NULL
    SET: task_category = TRANSIT
```

**Task Category Rules**:
```
IF task_category == ANCHOR:
    DEFAULT: priority = CRITICAL (anchors are foundational)
    EXAMPLES: "Sleep 8 hours", "Family dinner", "Morning routine"
    
IF task_category == TRANSIT:
    REQUIRE: expected_duration IS NOT NULL
    AUTO-SET: If task_type == COMMUTE, set task_category = TRANSIT
    EXAMPLES: "Drive to office", "Flight to NYC", "Commute home"
    
IF task_category == ACTION:
    DEFAULT: Most common category
    EXAMPLES: "Write report", "Read chapters", "Review PRs"
```

**Recurring Task Relationship**:
```
IF task is instance of recurring schedule:
    REQUIRE: recurring_task_id IS NOT NULL
    INHERIT: name, task_type, task_category, expected_* from recurring_tasks
    SET: status = PENDING (can be updated independently)
    
IF task is standalone:
    SET: recurring_task_id = NULL
    EXAMPLES: "Buy groceries today", "Call dentist"
```

**Success Calculation** (Computed Field):
```
WHEN status == COMPLETED:
    IF task_type == UNIT_BASED:
        success_percentage = (actual_units / expected_units) × 100
        CLAMP to [0, 100]
        
    IF task_type == TIME_BASED AND NOT is_commute:
        # Faster is better
        success_percentage = (expected_duration / actual_duration) × 100
        CLAMP to [0, 150]  # Allow overachievement
        
    IF task_type == COMMUTE:
        # On-time is ideal
        IF actual_duration <= expected_duration:
            success_percentage = 100
        ELSE:
            success_percentage = (expected_duration / actual_duration) × 100
            CLAMP to [0, 100]
ELSE:
    success_percentage = NULL
```

---

## 3. Category_Stats Entity (Optional - Phase 2)

### Purpose
Track success rates by user-defined categories (Work, Personal, Health, etc.)

| Field Name | Type | Nullable | Description |
|------------|------|----------|-------------|
| `id` | UUID | No | Primary Key |
| `category` | String | No | Category name |
| `date` | Date | No | Reference date |
| `task_count` | Integer | No | Tasks in category |
| `avg_success` | Float | Yes | Average success % |

---

## PostgreSQL Schema Definition

### Extensions

Requires UUID generation extensions (uuid-ossp or pgcrypto).

### Table: recurring_tasks

The recurring_tasks table stores recurring task templates with:
- Basic task properties (name, type, category, priority, expected_duration, expected_units)
- Recurrence configuration (type, interval, days, day_of_month, pattern, end_date)
- Status tracking (is_active, created_at, updated_at)
- Validation constraints ensuring data consistency
- Performance indexes on commonly queried fields

### Table: tasks

The tasks table stores individual task instances with:
- Task properties (name, type, category, status, priority)
- Expectation and actual values (duration, units)
- Success tracking (success_percentage, completed_at)
- Optional deadline and notes
- Link to recurring template (recurring_task_id)
- Validation constraints ensuring proper task completion
- Performance indexes on commonly queried fields

### Table: category_stats (Phase 2)

The category_stats table aggregates performance by category with:
- Category name and date
- Task count and average success for that category/date
- Unique constraint on category and date combination
- Indexes for efficient date and category queries

### Triggers

Automatic triggers update the updated_at timestamp on all tables when records are modified.

---

## Migration Strategy

### Version 1.0 → 1.1 (Adding User Support)

Future migration will add:
- users table with email and timestamps
- user_id foreign key to tasks and recurring_tasks tables
- Indexes on user_id columns for performance

---

## Data Validation Rules

### Pre-Insert Validation (Pydantic)

| Rule | Description | Error Message |
|------|-------------|---------------|
| **Name Required** | Task name cannot be empty | "Task name is required" |
| **Type Consistency** | Task type must match duration/unit fields | "TIME_BASED tasks require expected_duration" |
| **Future Deadline** | Deadline must be in future | "Deadline must be after current time" |
| **Commute Flag** | is_commute must match task_type | "Commute flag inconsistent with task type" |

### Post-Completion Validation

| Rule | Description | Action |
|------|-------------|--------|
| **Actuals Required** | Must provide actual values when marking complete | Reject completion request |
| **Units Bounds** | actual_units cannot exceed expected_units | Clamp to expected_units |
| **Success Range** | success_percentage must be 0-150 | Automatically clamp |

---

## Example Data Scenarios

### Scenario 1: Unit-Based Task (ACTION)
```json
{
    "name": "Read chapters for exam",
    "task_type": "UNIT_BASED",
    "task_category": "ACTION",
    "expected_units": 5,
    "expected_duration": 150,
    "actual_units": 4,
    "actual_duration": 140,
    "success_percentage": 80.0
}
```

### Scenario 2: Time-Based Task (ACTION)
```json
{
    "name": "Deep work session",
    "task_type": "TIME_BASED",
    "task_category": "ACTION",
    "expected_duration": 120,
    "actual_duration": 90,
    "success_percentage": 133.33  // Finished faster
}
```

### Scenario 3: Commute Task (TRANSIT)
```json
{
    "name": "Morning commute to office",
    "task_type": "COMMUTE",
    "task_category": "TRANSIT",
    "is_commute": true,
    "expected_duration": 30,
    "actual_duration": 45,
    "success_percentage": 66.67  // Late arrival
}
```

### Scenario 4: Mixed Task (Units + Time) - ACTION
```json
{
    "name": "Review 3 PRs",
    "task_type": "UNIT_BASED",
    "task_category": "ACTION",
    "expected_units": 3,
    "expected_duration": 60,
    "actual_units": 3,
    "actual_duration": 75,
    "success_percentage": 100.0  // Based on units, not time
}
```

### Scenario 5: Anchor Task (ANCHOR)
```json
{
    "name": "Sleep",
    "task_type": "TIME_BASED",
    "task_category": "ANCHOR",
    "expected_duration": 480,
    "actual_duration": 450,
    "priority": 4,
    "success_percentage": 93.75  // Got most of needed sleep
}
```

### Scenario 6: Family Time (ANCHOR)
```json
{
    "name": "Family dinner",
    "task_type": "TIME_BASED",
    "task_category": "ANCHOR",
    "expected_duration": 60,
    "actual_duration": 60,
    "priority": 4,
    "success_percentage": 100.0
}
```

### Scenario 7: Daily Recurring Task Template
```json
// recurring_tasks table
{
    "name": "Daily standup",
    "task_type": "TIME_BASED",
    "task_category": "ACTION",
    "expected_duration": 15,
    "priority": 3,
    "recurrence_type": "DAILY",
    "recurrence_interval": 1,
    "is_active": true
}

// Generated task instance (in tasks table)
{
    "name": "Daily standup",
    "task_type": "TIME_BASED",
    "task_category": "ACTION",
    "expected_duration": 15,
    "priority": 3,
    "status": "PENDING",
    "recurring_task_id": "<uuid-of-recurring-task>",
    "created_at": "2026-02-18T00:00:00Z"
}
```

### Scenario 8: Weekly Recurring Task Template
```json
// recurring_tasks table
{
    "name": "Team meeting",
    "task_type": "TIME_BASED",
    "task_category": "ACTION",
    "expected_duration": 60,
    "priority": 3,
    "recurrence_type": "WEEKLY",
    "recurrence_days": [1, 3],
    "notes": "Every Monday and Wednesday",
    "is_active": true
}
```

### Scenario 9: Monthly Date Recurring Task Template
```json
// recurring_tasks table
{
    "name": "Pay rent",
    "task_type": "UNIT_BASED",
    "task_category": "ACTION",
    "expected_units": 1,
    "expected_duration": 10,
    "priority": 4,
    "recurrence_type": "MONTHLY_DATE",
    "recurrence_day_of_month": 1,
    "notes": "First day of every month",
    "is_active": true
}
```

### Scenario 10: Monthly Pattern Recurring Task Template
```json
// recurring_tasks table
{
    "name": "Board meeting",
    "task_type": "TIME_BASED",
    "task_category": "ACTION",
    "expected_duration": 120,
    "priority": 4,
    "recurrence_type": "MONTHLY_PATTERN",
    "recurrence_pattern": "first_monday",
    "notes": "First Monday of every month",
    "is_active": true
}
```

---

## Code Generation Checklist (For Coding Agent)

- [ ] Create Pydantic models in `models/task.py`
- [ ] Create Pydantic models in `models/recurring_task.py`
- [ ] Create SQLAlchemy ORM models in `db/models.py`
- [ ] Write migration script `migrations/001_initial_schema.sql`
- [ ] Write migration script `migrations/002_add_recurring_tasks.sql`
- [ ] Implement success calculation in `utils/success_calculator.py`
- [ ] **Implement recurring task instance generator in `tasks/recurring_generator.py`**
- [ ] **Add cron job for daily task instance generation (runs at midnight)**
- [ ] **Add function to calculate next occurrence date for each recurrence type**
- [ ] **Implement direct query functions for statistics aggregation in `analytics/queries.py`**
- [ ] Add validation decorators for business logic
- [ ] Write unit tests for all validation rules
- [ ] **Write unit tests for recurrence date calculations**
- [ ] Create database initialization script `db/init_db.py`
- [ ] Add example seed data script `db/seed_data.py`

---

## Recurring Task Instance Generation

### Overview

Recurring tasks work on a **template + instances** model:
- **Template**: The recurring task definition (stored in `recurring_tasks` table)
- **Instances**: Individual occurrences created automatically in `tasks` table (linked via `recurring_task_id`)

### Generation Logic

**Cron Job**: Runs daily at 12:00 AM (midnight)

The daily cron job:
1. Queries all active recurring task templates
2. For each template, determines if an instance should be generated for today based on recurrence rules
3. Creates task instances in the tasks table linked to their templates via recurring_task_id
4. Checks for duplicates to avoid creating multiple instances for the same day
5. Sets default deadline to end of day

Recurrence logic:
- **DAILY**: Checks if today falls on the correct interval (e.g., every 2 days)
- **WEEKLY**: Checks if today's weekday is in the recurrence_days array
- **MONTHLY_DATE**: Checks if today matches the specified day of month
- **MONTHLY_PATTERN**: Matches patterns like "first_monday" or "last_friday" of the month

### Edge Cases

| Scenario | Handling | Example |
|----------|----------|---------|
| **Invalid date** | Skip that month | Feb 31st monthly task skips February |
| **Already exists** | Don't create duplicate | Instance already created for that date |
| **Skipped occurrence** | Mark in skip table (future) | User explicitly skipped via API |
| **End date reached** | Stop generation | recurrence_end_date has passed |
| **Deactivated template** | Stop generation | is_active = False in recurring_tasks |
| **Deleted template** | CASCADE delete instances | ON DELETE CASCADE removes all instances |

---

## Performance Considerations

| Scenario | Query Pattern | Optimization |
|----------|---------------|--------------|
| **Dashboard Load** | Fetch today's stats | Direct query on tasks with date indexes |
| **Task Filtering** | Filter by status/category | Indexed columns (status, category, created_at) |
| **Weekly Reports** | Aggregate 7 days | Query tasks with date range, use indexes |
| **Monthly Stats** | Aggregate 30 days | Query tasks with date range, consider caching |
| **Commute Analysis** | Filter is_commute=true | Dedicated index on is_commute |

**Query Optimization**:
- All date-based queries use indexed `created_at` and `completed_at` columns
- Consider implementing Redis cache for frequently accessed date ranges
- Use PostgreSQL partial indexes for common query patterns
- JSONB columns (recurrence_days) indexed with GIN for efficient queries
- Use EXPLAIN ANALYZE for query optimization

**Expected Load**:
- ~50 tasks per user per day
- ~10 status updates per task
- 1 dashboard query per minute (worst case)
- Date range queries typically scan < 1000 rows

**Database Size Estimation**:
- Tasks: ~18,000/year/user (~5 MB)
- Recurring_Tasks: ~50 templates/user (~10 KB)
- Total: < 6 MB per user per year

---

**Document Status**: ✅ Complete  
**Dependencies**: None  
**Next Document**: `02_llm_parsing_logic.md`
