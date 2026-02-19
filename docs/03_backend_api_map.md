# Module 03: Backend API Map

## Overview
This document defines the complete REST API architecture for Nudge, including all FastAPI endpoints, request/response schemas, authentication, rate limiting, and error handling.

---

## API Architecture

```
┌─────────────────────┐
│   API Gateway       │
│   (Rate Limiting)   │
└──────────┬──────────┘
           │
    ┌──────┴──────┬──────────┬──────────┬──────────┐
    ▼             ▼          ▼          ▼          ▼
┌────────┐  ┌─────────┐  ┌──────┐  ┌────────┐  ┌────────┐
│ Tasks  │  │  Stats  │  │ LLM  │  │ Health │  │ Admin  │
│ Routes │  │ Routes  │  │Input │  │ Check  │  │ Routes │
└────────┘  └─────────┘  └──────┘  └────────┘  └────────┘
     │           │           │          │          │
     └───────────┴───────────┴──────────┴──────────┘
                      │
                      ▼
            ┌──────────────────┐
            │  Database Layer  │
            │   (PostgreSQL)   │
            └──────────────────┘
```

---

## Base Configuration

### API Base URL
- **Development**: `http://localhost:8000/api/v1`
- **Production**: `https://api.nudge.app/api/v1`

### Global Headers

| Header | Required | Value | Purpose |
|--------|----------|-------|---------|
| `Content-Type` | Yes | `application/json` | JSON request body |
| `Authorization` | No (Phase 2) | `Bearer <token>` | User authentication |
| `X-Request-ID` | No | UUID | Request tracing |
| `User-Agent` | No | String | Client identification |

### Global Response Schema

All responses follow this envelope:

```json
{
  "status": "success | error",
  "data": { /* endpoint-specific data */ },
  "message": "Human-readable message",
  "timestamp": "2026-02-18T13:45:00Z",
  "request_id": "uuid"
}
```

---

## Endpoint Specifications

### 1. POST /nudge - Natural Language Input

**Purpose**: Process natural language text and create tasks using LLM parsing.

#### Request Schema

```json
{
  "text": "string (required, 1-2000 chars)",
  "context": {
    "conversation_id": "uuid (optional)",
    "previous_clarification": "string (optional)"
  },
  "preferences": {
    "auto_confirm": "boolean (default: true)",
    "default_category": "string (optional)"
  }
}
```

#### Response Schema (Success)

```json
{
  "status": "success",
  "data": {
    "tasks": [
      {
        "id": "uuid",
        "name": "Read 3 chapters",
        "task_type": "UNIT_BASED",
        "task_category": "ACTION",
        "expected_duration": 90,
        "expected_units": 3,
        "priority": 2,
        "status": "PENDING",
        "recurring_task_id": null,
        "created_at": "2026-02-18T13:45:00Z"
      }
    ],
    "recurring_task_created": false,
    "confidence": 0.85,
    "needs_clarification": false,
    "clarification_question": null,
    "conversation_id": "uuid"
  },
  "message": "Successfully created 1 task",
  "timestamp": "2026-02-18T13:45:00Z"
}
```

#### Response Schema (Recurring Task Created)

```json
{
  "status": "success",
  "data": {
    "tasks": [
      {
        "id": "uuid",
        "name": "Daily standup",
        "task_type": "TIME_BASED",
        "task_category": "ACTION",
        "expected_duration": 15,
        "priority": 3,
        "status": "PENDING",
        "recurring_task_id": "uuid-of-recurring-task-template",
        "created_at": "2026-02-18T13:45:00Z"
      }
    ],
    "recurring_task_created": true,
    "recurring_task": {
      "id": "uuid-of-recurring-task-template",
      "name": "Daily standup",
      "recurrence_type": "DAILY",
      "recurrence_interval": 1,
      "is_active": true
    },
    "confidence": 0.92,
    "needs_clarification": false,
    "clarification_question": null,
    "conversation_id": "uuid"
  },
  "message": "Successfully created recurring task and first instance",
  "timestamp": "2026-02-18T13:45:00Z"
}
```

**Note**: When `is_recurring: true` in LLM output:
1. Backend creates entry in `recurring_tasks` table with recurrence fields
2. Backend creates first task instance in `tasks` table linked via `recurring_task_id`
3. Future instances generated automatically by cron job

#### Response Schema (Needs Clarification)

```json
{
  "status": "success",
  "data": {
    "tasks": [],
    "confidence": 0.45,
    "needs_clarification": true,
    "clarification_question": "How much time do you expect this to take?",
    "conversation_id": "uuid",
    "partial_task": {
      "name": "Finish the project",
      "task_type": "TIME_BASED"
    }
  },
  "message": "Clarification needed",
  "timestamp": "2026-02-18T13:45:00Z"
}
```

#### Error Responses

| Status Code | Error Code | Description | Example |
|-------------|------------|-------------|---------|
| 400 | `INVALID_INPUT` | Malformed request | Empty text field |
| 422 | `PARSING_FAILED` | LLM couldn't parse | Gibberish input |
| 429 | `RATE_LIMIT_EXCEEDED` | Too many requests | 100 req/min hit |
| 500 | `LLM_SERVICE_ERROR` | LLM service down | Qwen API timeout |
| 503 | `SERVICE_UNAVAILABLE` | Database down | Connection pool exhausted |

#### Implementation

The endpoint:
1. Sanitizes input text
2. Calls LLM parsing layer
3. Validates parsed JSON
4. If confidence > 0.6: Creates tasks in DB
5. If confidence < 0.6: Returns clarification request
6. Returns response

---

### 2. GET /tasks - List Task Instances

**Purpose**: Retrieve task instances with filtering and pagination. This endpoint returns tasks from the `tasks` table only (instances), not recurring task templates.

#### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `status` | string | all | Filter: `pending`, `in_progress`, `completed`, `failed` |
| `category` | string | all | Filter by user-defined category |
| `task_category` | string | all | Filter: `anchor`, `transit`, `action` |
| `from_recurring` | boolean | all | Filter instances from recurring tasks |
| `recurring_task_id` | UUID | all | Show instances of specific recurring task |
| `is_commute` | boolean | false | Show only commutes |
| `start_date` | ISO date | today | Start of date range |
| `end_date` | ISO date | today | End of date range |
| `sort_by` | string | created_at | Sort field: `priority`, `deadline`, `created_at` |
| `order` | string | desc | Sort order: `asc`, `desc` |
| `page` | integer | 1 | Page number |
| `per_page` | integer | 20 | Items per page (max 100) |

#### Example Requests

```
GET /api/v1/tasks?status=pending&sort_by=priority&order=desc&page=1&per_page=10
```

```
GET /api/v1/tasks?task_category=anchor&start_date=2026-02-18
```

```
GET /api/v1/tasks?from_recurring=true
# Returns all task instances that are part of recurring schedules
```

```
GET /api/v1/tasks?recurring_task_id=<uuid>
# Returns all instances of a specific recurring task template
```

#### Response Schema

```json
{
  "status": "success",
  "data": {
    "tasks": [
      {
        "id": "uuid",
        "name": "Submit quarterly report",
        "task_type": "TIME_BASED",
        "task_category": "ACTION",
        "status": "IN_PROGRESS",
        "priority": 4,
        "expected_duration": 180,
        "actual_duration": null,
        "category": "Work",
        "deadline": "2026-02-21T17:00:00Z",
        "created_at": "2026-02-18T09:00:00Z",
        "updated_at": "2026-02-18T14:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "per_page": 10,
      "total_items": 47,
      "total_pages": 5
    }
  },
  "message": "Retrieved 10 tasks",
  "timestamp": "2026-02-18T14:35:00Z"
}
```

---

### 3. GET /tasks/{task_id} - Get Single Task

**Purpose**: Retrieve detailed information for a specific task.

#### Path Parameters

- `task_id` (UUID, required)

#### Response Schema

```json
{
  "status": "success",
  "data": {
    "task": {
      "id": "uuid",
      "name": "Read 3 chapters",
      "task_type": "UNIT_BASED",
      "task_category": "ACTION",
      "is_commute": false,
      "status": "COMPLETED",
      "priority": 2,
      "expected_duration": 90,
      "actual_duration": 105,
      "expected_units": 3,
      "actual_units": 3,
      "success_percentage": 100.0,
      "category": "Reading",
      "notes": "Finished all chapters",
      "deadline": null,
      "created_at": "2026-02-18T09:00:00Z",
      "updated_at": "2026-02-18T11:45:00Z",
      "completed_at": "2026-02-18T11:45:00Z"
    }
  },
  "message": "Task retrieved successfully",
  "timestamp": "2026-02-18T14:40:00Z"
}
```

#### Error Response (404)

```json
{
  "status": "error",
  "data": null,
  "message": "Task not found",
  "error_code": "TASK_NOT_FOUND",
  "timestamp": "2026-02-18T14:40:00Z"
}
```

---

### 4. PATCH /tasks/{task_id}/complete - Mark Task Complete

**Purpose**: Update task status to completed and calculate success percentage.

#### Path Parameters

- `task_id` (UUID, required)

#### Request Schema

```json
{
  "actual_duration": 105,  // Required for TIME_BASED/COMMUTE
  "actual_units": 3,       // Required for UNIT_BASED
  "notes": "Took longer than expected",  // Optional
  "completed_at": "2026-02-18T11:45:00Z"  // Optional (defaults to now)
}
```

#### Response Schema

```json
{
  "status": "success",
  "data": {
    "task": {
      "id": "uuid",
      "name": "Read 3 chapters",
      "status": "COMPLETED",
      "actual_duration": 105,
      "actual_units": 3,
      "success_percentage": 100.0,
      "completed_at": "2026-02-18T11:45:00Z"
    },
    "success_breakdown": {
      "expected": "3 units in 90 minutes",
      "actual": "3 units in 105 minutes",
      "calculation": "Units completed (100%) - Time overrun noted"
    }
  },
  "message": "Task marked as completed with 100% success",
  "timestamp": "2026-02-18T14:45:00Z"
}
```

#### Validation Rules

1. Cannot complete a task with status `COMPLETED` or `FAILED`
2. Must provide `actual_duration` for TIME_BASED/COMMUTE tasks
3. Must provide `actual_units` for UNIT_BASED tasks
4. `actual_units` cannot exceed `expected_units`
5. `completed_at` cannot be before `created_at`

---

### 5. PATCH /tasks/{task_id} - Update Task

**Purpose**: Modify task properties (name, deadline, priority, etc.).

#### Request Schema (Partial Updates Allowed)

```json
{
  "name": "Read 4 chapters instead",
  "expected_units": 4,
  "priority": 3,
  "deadline": "2026-02-19T18:00:00Z",
  "notes": "Extended scope"
}
```

#### Response Schema

```json
{
  "status": "success",
  "data": {
    "task": { /* full updated task object */ }
  },
  "message": "Task updated successfully",
  "timestamp": "2026-02-18T14:50:00Z"
}
```

#### Constraints

- Cannot change `task_type` after creation
- Cannot change `task_category` after creation (linked to task_type)
- Cannot modify `actual_duration` or `actual_units` (use `POST /tasks/{id}/complete` endpoint)
- Cannot change `id`, `created_at`, `recurring_task_id`
- Task instances (with `recurring_task_id` set) inherit properties from template
- To modify recurrence schedule, use `PUT /recurring-tasks/{id}` on the template

---

### 6. GET /recurring-tasks - List Recurring Task Templates

**Purpose**: Retrieve recurring task templates from the `recurring_tasks` table (not instances).

#### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `recurrence_type` | string | all | Filter: `daily`, `weekly`, `monthly_date`, `monthly_pattern` |
| `is_active` | boolean | all | Filter by active status |
| `task_category` | string | all | Filter: `anchor`, `transit`, `action` |

#### Response Schema

```json
{
  "status": "success",
  "data": {
    "recurring_tasks": [
      {
        "id": "uuid",
        "name": "Daily standup",
        "task_type": "TIME_BASED",
        "task_category": "ACTION",
        "expected_duration": 15,
        "priority": 3,
        "recurrence_type": "DAILY",
        "recurrence_interval": 1,
        "is_active": true,
        "next_occurrence": "2026-02-19T00:00:00Z",
        "total_instances": 45,
        "completed_instances": 38,
        "created_at": "2026-01-01T00:00:00Z"
      },
      {
        "id": "uuid",
        "name": "Team meeting",
        "task_type": "TIME_BASED",
        "task_category": "ACTION",
        "expected_duration": 60,
        "priority": 3,
        "recurrence_type": "WEEKLY",
        "recurrence_days": [1, 3],
        "is_active": true,
        "next_occurrence": "2026-02-19T00:00:00Z",
        "total_instances": 16,
        "completed_instances": 14
      }
    ]
  },
  "message": "Retrieved 2 recurring tasks",
  "timestamp": "2026-02-18T15:30:00Z"
}
```

**Note**: `total_instances` and `completed_instances` are calculated by querying the `tasks` table for entries with matching `recurring_task_id`.

---

### 7. POST /recurring-tasks - Create Recurring Task Template

**Purpose**: Manually create a recurring task template without using natural language input.

#### Request Schema

```json
{
  "name": "Weekly team meeting",
  "task_type": "TIME_BASED",
  "task_category": "ACTION",
  "expected_duration": 60,
  "priority": 3,
  "recurrence_type": "WEEKLY",
  "recurrence_days": [1, 3],
  "notes": "Every Monday and Wednesday",
  "create_initial_instance": true
}
```

#### Response Schema

```json
{
  "status": "success",
  "data": {
    "recurring_task": {
      "id": "uuid",
      "name": "Weekly team meeting",
      "recurrence_type": "WEEKLY",
      "recurrence_days": [1, 3],
      "is_active": true,
      "created_at": "2026-02-18T15:40:00Z"
    },
    "initial_instance": {
      "id": "uuid",
      "name": "Weekly team meeting",
      "status": "PENDING",
      "recurring_task_id": "uuid-of-template",
      "created_at": "2026-02-18T15:40:00Z"
    }
  },
  "message": "Recurring task created successfully",
  "timestamp": "2026-02-18T15:40:00Z"
}
```

---

### 8. PUT /recurring-tasks/{id} - Update Recurring Task Template

**Purpose**: Update a recurring task template. Changes apply to future instances only.

#### Path Parameters

- `id` (UUID, required) - Recurring task template ID

#### Request Schema

```json
{
  "name": "Updated task name",
  "expected_duration": 90,
  "priority": 4,
  "recurrence_interval": 2,
  "is_active": false
}
```

#### Response Schema

```json
{
  "status": "success",
  "data": {
    "recurring_task": {
      "id": "uuid",
      "name": "Updated task name",
      "expected_duration": 90,
      "priority": 4,
      "recurrence_interval": 2,
      "is_active": false,
      "updated_at": "2026-02-18T15:45:00Z"
    }
  },
  "message": "Recurring task updated successfully",
  "timestamp": "2026-02-18T15:45:00Z"
}
```

**Note**: Cannot modify `recurrence_type` after creation. Delete and recreate instead.

---

### 9. POST /recurring-tasks/{id}/skip - Skip Next Occurrence

**Purpose**: Skip the next scheduled occurrence of a recurring task.

#### Path Parameters

- `id` (UUID, required) - Recurring task template ID

#### Request Schema

```json
{
  "skip_date": "2026-02-19",
  "reason": "Holiday - office closed"
}
```

#### Response Schema

```json
{
  "status": "success",
  "data": {
    "skipped_date": "2026-02-19",
    "next_occurrence": "2026-02-20T00:00:00Z",
    "reason": "Holiday - office closed"
  },
  "message": "Occurrence skipped successfully",
  "timestamp": "2026-02-18T15:35:00Z"
}
```

**Implementation Note**: Create a `skipped_occurrences` table to track skipped dates, checked by cron job before generating instances.

---

### 10. DELETE /recurring-tasks/{id} - Delete Recurring Task Template

**Purpose**: Delete a recurring task template and optionally handle existing instances.

#### Path Parameters

- `id` (UUID, required) - Recurring task template ID

#### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `delete_instances` | boolean | false | If true, cascade delete all task instances |
| `future_only` | boolean | true | If true, only affect future instances |

#### Response Schema

```json
{
  "status": "success",
  "data": {
    "deleted_template_id": "uuid",
    "instances_affected": 12,
    "instances_deleted": 0
  },
  "message": "Recurring task template deleted successfully",
  "timestamp": "2026-02-18T15:50:00Z"
}
```

**Note**: By default, past instances are preserved for analytics. Future instances are marked as CANCELLED or deleted based on `delete_instances` parameter.

---

### 11. DELETE /tasks/{task_id} - Delete Task Instance

**Purpose**: Soft delete a task instance (sets status to DELETED, doesn't remove from DB).

#### Path Parameters

- `task_id` (UUID, required) - Task instance ID

#### Response Schema

```json
{
  "status": "success",
  "data": {
    "deleted_task_id": "uuid"
  },
  "message": "Task deleted successfully",
  "timestamp": "2026-02-18T14:55:00Z"
}
```

**Note**: 
- Soft deletion preserves data for analytics
- Hard deletion available via admin endpoint
- Deleting a task instance does NOT affect the recurring task template (if it's an instance)
- To stop recurring tasks, use `DELETE /recurring-tasks/{id}` or set `is_active: false`

---

### 12. GET /stats/daily - Daily Statistics

**Purpose**: Retrieve aggregated stats for a specific day.

#### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `date` | ISO date | today | Target date (YYYY-MM-DD) |

#### Example Request

```
GET /api/v1/stats/daily?date=2026-02-18
```

#### Response Schema

```json
{
  "status": "success",
  "data": {
    "date": "2026-02-18",
    "summary": {
      "total_tasks": 12,
      "completed_tasks": 8,
      "failed_tasks": 1,
      "in_progress_tasks": 3,
      "completion_rate": 66.67,
      "avg_success_rate": 87.5
    },
    "time_breakdown": {
      "total_productive_time": 420,  // minutes
      "total_commute_time": 90,
      "avg_task_duration": 52.5
    },
    "priority_breakdown": [
      {"priority": "CRITICAL", "count": 2, "completed": 2},
      {"priority": "HIGH", "count": 5, "completed": 4},
      {"priority": "MEDIUM", "count": 4, "completed": 2},
      {"priority": "LOW", "count": 1, "completed": 0}
    ],
    "category_breakdown": [
      {"category": "Work", "count": 7, "avg_success": 85.2},
      {"category": "Personal", "count": 3, "avg_success": 92.1},
      {"category": "Commute", "count": 2, "avg_success": 100.0}
    ],
    "task_category_breakdown": [
      {"task_category": "ACTION", "count": 8, "avg_success": 86.5},
      {"task_category": "TRANSIT", "count": 2, "avg_success": 100.0},
      {"task_category": "ANCHOR", "count": 2, "avg_success": 95.0}
    ]
  },
  "message": "Daily statistics retrieved",
  "timestamp": "2026-02-18T15:00:00Z"
}
```

---

### 10. GET /stats/weekly - Weekly Statistics

**Purpose**: Retrieve aggregated stats for a 7-day period.

#### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `start_date` | ISO date | 7 days ago | Start of week |
| `end_date` | ISO date | today | End of week |

#### Response Schema

```json
{
  "status": "success",
  "data": {
    "period": {
      "start_date": "2026-02-12",
      "end_date": "2026-02-18"
    },
    "summary": {
      "total_tasks": 84,
      "completed_tasks": 65,
      "completion_rate": 77.38,
      "avg_success_rate": 83.2
    },
    "daily_trend": [
      {"date": "2026-02-12", "completed": 9, "avg_success": 81.5},
      {"date": "2026-02-13", "completed": 10, "avg_success": 88.0},
      // ... remaining days
    ],
    "best_day": {
      "date": "2026-02-15",
      "completed_tasks": 12,
      "avg_success": 95.3
    },
    "worst_day": {
      "date": "2026-02-14",
      "completed_tasks": 6,
      "avg_success": 68.1
    }
  },
  "message": "Weekly statistics retrieved",
  "timestamp": "2026-02-18T15:05:00Z"
}
```

---

### 11. GET /stats/monthly - Monthly Statistics

**Purpose**: Retrieve aggregated stats for a calendar month.

#### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `month` | integer | current | Month (1-12) |
| `year` | integer | current | Year (YYYY) |

#### Example Request

```
GET /api/v1/stats/monthly?month=2&year=2026
```

#### Response Schema

```json
{
  "status": "success",
  "data": {
    "period": {
      "month": 2,
      "year": 2026,
      "total_days": 28
    },
    "summary": {
      "total_tasks": 336,
      "completed_tasks": 268,
      "completion_rate": 79.76,
      "avg_success_rate": 84.5
    },
    "weekly_breakdown": [
      {"week": 1, "completed": 65, "avg_success": 82.1},
      {"week": 2, "completed": 72, "avg_success": 86.3},
      {"week": 3, "completed": 68, "avg_success": 83.9},
      {"week": 4, "completed": 63, "avg_success": 85.7}
    ],
    "top_categories": [
      {"category": "Work", "count": 189, "avg_success": 83.2},
      {"category": "Personal", "count": 87, "avg_success": 88.9},
      {"category": "Health", "count": 42, "avg_success": 81.0}
    ]
  },
  "message": "Monthly statistics retrieved",
  "timestamp": "2026-02-18T15:10:00Z"
}
```

---

### 12. GET /stats/trends - Failure Trend Analysis

**Purpose**: Identify patterns in failed/low-success tasks for AI advisor.

#### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `lookback_days` | integer | 30 | Analysis period |
| `success_threshold` | float | 40.0 | Below this = "failure" |

#### Response Schema

```json
{
  "status": "success",
  "data": {
    "analysis_period": {
      "start_date": "2026-01-19",
      "end_date": "2026-02-18",
      "total_days": 30
    },
    "failure_patterns": [
      {
        "pattern_type": "chronic_category_failure",
        "category": "Exercise",
        "failure_count": 12,
        "avg_success": 32.5,
        "recommendation": "Consider reducing exercise duration or frequency"
      },
      {
        "pattern_type": "time_estimation_error",
        "task_name_pattern": "Deep work",
        "avg_expected": 120,
        "avg_actual": 180,
        "recommendation": "Increase time estimates by 50% for deep work tasks"
      },
      {
        "pattern_type": "commute_delays",
        "avg_expected": 30,
        "avg_actual": 42,
        "delayed_days": 18,
        "recommendation": "Add 15-minute buffer to commute times"
      }
    ],
    "success_insights": [
      {
        "category": "Reading",
        "avg_success": 94.2,
        "insight": "Consistently high success rate - keep current schedule"
      }
    ]
  },
  "message": "Trend analysis complete",
  "timestamp": "2026-02-18T15:15:00Z"
}
```

---

### 13. GET /health - Health Check

**Purpose**: Verify service availability and dependencies.

#### Response Schema

```json
{
  "status": "success",
  "data": {
    "service": "nudge-api",
    "version": "1.0.0",
    "uptime": 3600,  // seconds
    "dependencies": {
      "database": {
        "status": "healthy",
        "latency_ms": 12
      },
      "llm_service": {
        "status": "healthy",
        "latency_ms": 450
      }
    }
  },
  "message": "All systems operational",
  "timestamp": "2026-02-18T15:20:00Z"
}
```

---

## Rate Limiting

### Limits by Endpoint

| Endpoint | Limit | Window | Overage Action |
|----------|-------|--------|----------------|
| `POST /nudge` | 60 requests | 1 minute | 429 error |
| `GET /tasks` | 120 requests | 1 minute | 429 error |
| `GET /stats/*` | 30 requests | 1 minute | 429 error |
| `PATCH /tasks/*` | 100 requests | 1 minute | 429 error |
| `DELETE /tasks/*` | 20 requests | 1 minute | 429 error |

### Rate Limit Headers

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 42
X-RateLimit-Reset: 1645196400  // Unix timestamp
```

### Rate Limit Error Response

```json
{
  "status": "error",
  "data": null,
  "message": "Rate limit exceeded. Try again in 23 seconds.",
  "error_code": "RATE_LIMIT_EXCEEDED",
  "retry_after": 23,
  "timestamp": "2026-02-18T15:25:00Z"
}
```

---

## Authentication (Phase 2)

### JWT Token Structure

```json
{
  "sub": "user_uuid",
  "email": "user@example.com",
  "iat": 1645196400,
  "exp": 1645200000,
  "scopes": ["tasks:read", "tasks:write", "stats:read"]
}
```

### Protected Endpoints

All endpoints except `/health` require authentication in Phase 2.

---

## Error Handling

### Standard Error Response

```json
{
  "status": "error",
  "data": null,
  "message": "Human-readable error description",
  "error_code": "ERROR_CODE_ENUM",
  "details": {
    "field": "expected_units",
    "issue": "Cannot be null for UNIT_BASED tasks"
  },
  "timestamp": "2026-02-18T15:30:00Z"
}
```

### Error Code Enum

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_INPUT` | 400 | Malformed request |
| `VALIDATION_ERROR` | 422 | Business rule violation |
| `TASK_NOT_FOUND` | 404 | Task ID doesn't exist |
| `UNAUTHORIZED` | 401 | Missing/invalid auth |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `LLM_SERVICE_ERROR` | 500 | LLM service failure |
| `DATABASE_ERROR` | 500 | Database connection issue |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily down |

---

## Database Transaction Patterns

### Create Task Flow

```python
async def create_task_endpoint(request: NudgeRequest):
    async with db.transaction():
        # 1. Parse LLM output
        parsed = await llm.parse(request.text)
        
        # 2. If recurring, create template and instance
        if parsed.is_recurring:
            template = await db.insert_recurring_task(parsed)
            task = await db.insert_task_instance(parsed, template.id)
        else:
            # 3. Insert one-time task
            task = await db.insert_task(parsed)
        
        # 4. Invalidate cache for today's stats
        await cache.invalidate(f"daily_stats:{datetime.now().date()}")
        
        # 5. Return response
        return TaskResponse(task)
```

### Complete Task Flow

```python
async def complete_task_endpoint(task_id: UUID, request: CompleteRequest):
    async with db.transaction():
        # 1. Fetch task
        task = await db.get_task(task_id)
        
        # 2. Calculate success percentage
        success = calculate_success(task, request.actual_duration, request.actual_units)
        
        # 3. Update task
        task = await db.update_task(task_id, {
            "status": "COMPLETED",
            "actual_duration": request.actual_duration,
            "actual_units": request.actual_units,
            "success_percentage": success,
            "completed_at": now()
        })
        
        # 4. Invalidate cache for completion date
        await cache.invalidate(f"daily_stats:{task.completed_at.date()}")
        
        # 5. Return response
        return CompleteResponse(task, success_breakdown)
```

---

## CORS Configuration

### Allowed Origins (Development)

```python
CORS_ORIGINS = [
    "http://localhost:3000",  # React dev server
    "http://localhost:5173",  # Vite dev server
]
```

### Allowed Methods

```python
CORS_METHODS = ["GET", "POST", "PATCH", "DELETE", "OPTIONS"]
```

### Allowed Headers

```python
CORS_HEADERS = [
    "Content-Type",
    "Authorization",
    "X-Request-ID"
]
```

---

## API Versioning Strategy

### URL-Based Versioning

- Current: `/api/v1/*`
- Future: `/api/v2/*`

### Deprecation Policy

1. Announce deprecation 3 months in advance
2. Support old version for 6 months post-deprecation
3. Return `Sunset` header with EOL date

```
Sunset: Sat, 31 Aug 2026 23:59:59 GMT
```

---

## Code Generation Checklist

- [ ] Setup FastAPI app in `main.py`
- [ ] Create routers in `routers/tasks.py`, `routers/stats.py`
- [ ] **Create router for recurring tasks in `routers/recurring.py`**
- [ ] Implement Pydantic request/response models in `schemas/`
- [ ] **Add recurrence request/response schemas in `schemas/recurrence.py`**
- [ ] Add middleware for rate limiting in `middleware/rate_limiter.py`
- [ ] Add middleware for error handling in `middleware/error_handler.py`
- [ ] Implement CORS configuration
- [ ] Add request ID tracking middleware
- [ ] Create database transaction wrapper `db/transaction.py`
- [ ] **Implement recurring task instance creation logic in `services/recurring_service.py`**
- [ ] **Add background job for generating daily task instances**
- [ ] Write OpenAPI documentation (auto-generated by FastAPI)
- [ ] Add endpoint integration tests
- [ ] **Add tests for recurring task skip functionality**
- [ ] Implement logging for all endpoints
- [ ] Setup monitoring (Prometheus metrics)

---

## Testing Strategy

### Unit Tests

- Test each endpoint with valid/invalid inputs
- Mock database and LLM service
- Verify response schemas

### Integration Tests

- End-to-end flows (create → update → complete → delete)
- Rate limiting behavior
- Error handling scenarios

### Load Tests

```
Target: 100 concurrent users
Endpoints to test:
- POST /nudge: 50 req/s
- GET /tasks: 200 req/s
- GET /stats/daily: 30 req/s
```

---

**Document Status**: ✅ Complete  
**Dependencies**: `01_data_schema.md`, `02_llm_parsing_logic.md`  
**Next Document**: `04_analytics_engine.md`
