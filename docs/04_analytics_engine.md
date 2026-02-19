# Module 04: Analytics Engine

## Overview
This document specifies the analytics computation layer responsible for calculating success percentages, aggregating task data, and generating visualization-ready datasets for the Nudge dashboard.

---

## Architecture

```
┌──────────────────────┐
│   Raw Task Data      │
│   (tasks table)      │
└─────────┬────────────┘
          │
          ▼
┌──────────────────────────────┐
│   Success Calculation Engine │
│   • Unit-based formula       │
│   • Time-based formula       │
│   • Commute-specific logic   │
└─────────┬────────────────────┘
          │
          ▼
┌──────────────────────────────┐
│   Aggregation Layer          │
│   • Daily rollups            │
│   • Weekly summaries         │
│   • Monthly trends           │
└─────────┬────────────────────┘
          │
          ▼
┌──────────────────────────────┐
│   Visualization Formatter    │
│   • Pie chart data           │
│   • Bar graph datasets       │
│   • Time series data         │
└─────────┬────────────────────┘
          │
          ▼
    ┌────────────┐
    │  Dashboard │
    └────────────┘
```

---

## Core Principles

1. **Real-Time Accuracy**: Calculate success immediately upon task completion
2. **Caching**: Pre-compute daily aggregates for performance
3. **Consistency**: Use same formulas across all views
4. **Transparency**: Expose calculation methodology in API responses
5. **Flexibility**: Support custom time ranges and filters

---

## 1. Success Percentage Calculation

### Formula Specifications

#### 1.1 Unit-Based Tasks

**Use Case**: Tasks measured by discrete items (chapters, emails, meetings)

**Formula**:
$$
\text{Success\%} = \frac{\text{actual\_units}}{\text{expected\_units}} \times 100
$$

**Constraints**:
- Result clamped to [0, 100]
- Partial credit awarded (e.g., 2/3 chapters = 66.67%)

**Examples**:

| Expected Units | Actual Units | Success % | Interpretation |
|----------------|--------------|-----------|----------------|
| 5 | 5 | 100.0 | Full completion |
| 5 | 4 | 80.0 | Partial completion |
| 5 | 0 | 0.0 | Not started |
| 3 | 3 | 100.0 | Full completion |

**Edge Cases**:

Edge cases are handled appropriately:
- actual_units > expected_units: Capped at 100%
- expected_units = 0: Prevented by schema validation
- Null actual_units: Only calculated for COMPLETED tasks

---

#### 1.2 Time-Based Tasks (Non-Commute)

**Use Case**: Tasks focused on duration (study, work sessions)

**Philosophy**: Finishing faster = higher success (efficiency metric)

**Formula**:
$$
\text{Success\%} = \frac{\text{expected\_duration}}{\text{actual\_duration}} \times 100
$$

**Constraints**:
- Result clamped to [0, 150] (allow 50% overachievement)
- Inverse ratio: faster is better

**Examples**:

| Expected Duration (min) | Actual Duration (min) | Success % | Interpretation |
|------------------------|-----------------------|-----------|----------------|
| 120 | 120 | 100.0 | Exactly on time |
| 120 | 90 | 133.3 | 33% faster (efficient) |
| 120 | 180 | 66.7 | Took 50% longer |
| 60 | 30 | 150.0 | 50% faster (max cap) |
| 60 | 20 | 150.0 | Capped at 150% |

**Rationale**:
- Rewards time efficiency
- Penalizes scope creep
- 150% cap prevents unrealistic scores from very short tasks

**Edge Cases**:

Edge cases are handled:
- actual_duration = 0: Raises validation error
- Very fast completion: Capped at 150%
- Extremely long completion: Results in very low score (e.g., 20%)

---

#### 1.3 Commute Tasks

**Use Case**: Travel time tracking (predictability metric)

**Philosophy**: On-time = 100%, delays penalized, early arrival not rewarded

**Formula**:
$$
\text{Success\%} = 
\begin{cases} 
100 & \text{if } \text{actual} \leq \text{expected} \\
\frac{\text{expected}}{\text{actual}} \times 100 & \text{if } \text{actual} > \text{expected}
\end{cases}
$$

**Constraints**:
- Result clamped to [0, 100]
- No bonus for early arrival
- Penalizes delays proportionally

**Examples**:

| Expected Duration (min) | Actual Duration (min) | Success % | Interpretation |
|------------------------|-----------------------|-----------|----------------|
| 30 | 30 | 100.0 | Exactly on time |
| 30 | 25 | 100.0 | Early (no bonus) |
| 30 | 45 | 66.7 | 15 min delay |
| 30 | 60 | 50.0 | 30 min delay (50% late) |
| 45 | 40 | 100.0 | Early arrival |

**Rationale**:
- Commutes should be predictable, not faster
- Early arrival doesn't indicate better performance
- Focus on reliability over speed

**Implementation**:

The commute success calculation:
- Returns 100.0 if actual <= expected (on-time or early)
- Returns (expected / actual) * 100 if delayed, clamped to [0, 100]

---

### 1.4 Unified Calculation Function

The success percentage calculation handles all three task types:
- Returns None if task is not completed
- For UNIT_BASED: (actual_units / expected_units) * 100, capped at 100
- For TIME_BASED (non-commute): (expected_duration / actual_duration) * 100, capped at 150
- For COMMUTE: 100 if on-time, otherwise (expected / actual) * 100
- Validates required fields and raises errors for invalid data
- Rounds result to 2 decimal places

---

## 2. Aggregation Layer

### 2.1 Daily Statistics Query

**Purpose**: Compute daily statistics on-demand from tasks table

**Note on Recurring Tasks**: Only count task instances (actual occurrences in `tasks` table), not the recurring templates in `recurring_tasks` table. A recurring daily task creates 1 instance per day in the `tasks` table.

**Computed Fields**:

| Field | Formula | Description |
|-------|---------|-------------|
| `total_tasks` | `COUNT(*)` | Tasks created on date (includes recurring instances) |
| `completed_tasks` | `COUNT(status=COMPLETED)` | Finished tasks |
| `failed_tasks` | `COUNT(status=FAILED)` | Abandoned tasks |
| `in_progress_tasks` | `COUNT(status=IN_PROGRESS)` | Active tasks |
| `avg_success_rate` | `AVG(success_percentage)` | Mean success (completed only) |
| `total_commute_time` | `SUM(actual_duration WHERE is_commute=true)` | Total commute minutes |
| `total_productive_time` | `SUM(actual_duration WHERE is_commute=false)` | Total work minutes |
| `high_priority_completed` | `COUNT(priority>=3 AND status=COMPLETED)` | Critical/High done |

**SQL Implementation**:

Daily statistics are computed with a single SQL query that:
- Counts total tasks created on the target date
- Counts tasks by status (completed, failed, in_progress)
- Calculates average success percentage for completed tasks
- Sums commute time and productive time separately
- Counts high-priority completed tasks (priority >= 3)
- Groups results by date

**Implementation**:

Daily stats are retrieved via direct database query:
- For past dates: Results cached for 24 hours
- For current date: Shorter TTL or no cache to reflect real-time updates
- Cache invalidated when tasks are updated/completed

---

### 2.2 Weekly Aggregation

**Purpose**: Generate 7-day rolling summary

**Computed On-the-Fly** (direct query on tasks table)

Weekly stats aggregate task data for specified date range:
- Single SQL query groups by date within range
- Calculates totals and completion rates across all days
- Computes average success rate from daily averages
- Creates daily trend array with date, completed count, and success rate
- Identifies best and worst days based on success rate

---

### 2.3 Monthly Aggregation

**Purpose**: Calendar month summary with weekly breakdown

Monthly stats aggregate data for entire calendar month:
- Queries all days in the specified month from tasks table
- Groups daily stats by ISO week number
- Creates weekly breakdown with completion counts and success rates
- Calculates overall totals and completion rates
- Returns structured monthly summary

---

## 3. Visualization Data Formatters

### 3.1 Pie Chart: Task Status Distribution

**Use Case**: Show proportion of Completed/In Progress/Failed tasks

**Data Structure**:

```json
{
  "chart_type": "pie",
  "data": [
    {"label": "Completed", "value": 65, "percentage": 77.38, "color": "#10B981"},
    {"label": "In Progress", "value": 12, "percentage": 14.29, "color": "#3B82F6"},
    {"label": "Failed", "value": 7, "percentage": 8.33, "color": "#EF4444"}
  ],
  "total": 84
}
```

**SQL Query**:

Task status distribution is computed by:
- Counting tasks by status within date range
- Calculating percentage of each status relative to total
- Rounding percentages to 2 decimal places

**Formatter**:

Pie chart data is formatted with:
- Status labels (converted from ENUM to readable text)
- Task counts for each status
- Percentage calculations
- Predefined colors for each status (green for completed, blue for in_progress, red for failed, orange for pending)

---

### 3.2 Bar Chart: Category Performance

**Use Case**: Compare success rates across categories (Work, Personal, Health)

**Data Structure**:

```json
{
  "chart_type": "bar",
  "x_axis": "Category",
  "y_axis": "Average Success %",
  "data": [
    {"label": "Work", "value": 83.2, "count": 189, "color": "#3B82F6"},
    {"label": "Personal", "value": 88.9, "count": 87, "color": "#10B981"},
    {"label": "Health", "value": 81.0, "count": 42, "color": "#F59E0B"}
  ]
}
```

**SQL Query**:

Category performance is computed by:
- Selecting category, task count, and average success percentage
- Filtering for completed tasks with success_percentage values
- Filtering by date range (completed_at >= cutoff date)
- Grouping by category
- Ordering by average success descending

---

### 3.3 Line Chart: Daily Success Trend

**Use Case**: Show success rate over time (7-day, 30-day trend)

**Data Structure**:

```json
{
  "chart_type": "line",
  "x_axis": "Date",
  "y_axis": "Success %",
  "data": {
    "labels": ["2026-02-12", "2026-02-13", "2026-02-14", "2026-02-15"],
    "datasets": [
      {
        "label": "Average Success Rate",
        "data": [81.5, 88.0, 68.1, 95.3],
        "borderColor": "#3B82F6",
        "backgroundColor": "rgba(59, 130, 246, 0.1)"
      },
      {
        "label": "Completion Rate",
        "data": [75.0, 83.3, 50.0, 100.0],
        "borderColor": "#10B981",
        "backgroundColor": "rgba(16, 185, 129, 0.1)"
      }
    ]
  }
}
```

**Formatter**:

Line chart data is formatted with:
- Date labels (ISO format strings)
- Two datasets: average success rate and completion rate
- Color coding and background fills for visual distinction
- Calculated completion rate = (completed / total) * 100 for each day

---

### 3.4 Heatmap: Priority vs. Completion

**Use Case**: Visualize which priority levels have best completion rates

**Data Structure**:

```json
{
  "chart_type": "heatmap",
  "rows": ["CRITICAL", "HIGH", "MEDIUM", "LOW"],
  "columns": ["Completed", "Failed"],
  "data": [
    [{"priority": "CRITICAL", "status": "Completed", "count": 18, "color_intensity": 0.9}],
    [{"priority": "CRITICAL", "status": "Failed", "count": 2, "color_intensity": 0.1}],
    [{"priority": "HIGH", "status": "Completed", "count": 45, "color_intensity": 0.8}]
  ]
}
```

---

## 4. Performance Optimization

### 4.1 Caching Strategy

**Cache Layer**: Redis for frequently accessed stats

Caching configuration:
- **daily_stats**: 24-hour TTL for past dates
- **weekly_stats**: 12-hour TTL (refresh twice daily)
- **monthly_stats**: 24-hour TTL (refresh daily)
- **trend_analysis**: 24-hour TTL (refresh daily)

Caching decorator logic:
- For past dates: Cache for 24 hours since data won't change
- For current date: Use shorter TTL or skip cache to reflect real-time updates

**Cache Invalidation**:
- On task completion → Invalidate today's daily_stats cache
- On task update → Invalidate relevant date's daily_stats cache
- On task deletion → Invalidate relevant date's daily_stats cache
- Past dates rarely need invalidation since data is immutable

---

### 4.2 Database Indexing

**Critical Indexes**:

Performance indexes include:
- `idx_tasks_created_at`: For date range queries
- `idx_tasks_completed_at`: For completed tasks queries (with WHERE clause)
- `idx_tasks_success`: For success percentage queries (with WHERE clause)
- `idx_tasks_status_created`: Composite index for status filtering with dates
- `idx_tasks_category_success`: For category aggregations (with WHERE clause)
- `idx_tasks_commute_date`: For commute-specific analysis (with WHERE clause)

---

### 4.3 Query Optimization

**Inefficient Query** (N+1 problem):
Making individual queries for each date in a range (e.g., 30 separate queries for 30 days)

**Optimized Query** (Single batch query):
Single SQL query with date range filter and GROUP BY date clause

**Performance Comparison**:
- N+1 queries: ~100ms * 30 days = 3 seconds
- Batch query: ~150ms for 30 days
- **20x improvement** for monthly views

---

## 5. Analytics API Response Examples

### Example 1: Daily Stats with Visualizations

```json
{
  "status": "success",
  "data": {
    "date": "2026-02-18",
    "summary": {
      "total_tasks": 12,
      "completed_tasks": 8,
      "avg_success_rate": 87.5
    },
    "visualizations": {
      "status_distribution": {
        "chart_type": "pie",
        "data": [
          {"label": "Completed", "value": 8, "percentage": 66.67},
          {"label": "In Progress", "value": 3, "percentage": 25.0},
          {"label": "Failed", "value": 1, "percentage": 8.33}
        ]
      },
      "category_performance": {
        "chart_type": "bar",
        "data": [
          {"label": "Work", "value": 85.2, "count": 5},
          {"label": "Personal", "value": 92.1, "count": 3}
        ]
      }
    }
  }
}
```

---

### 6.1 Handling Missing Data

| Scenario | Handling | Rationale |
|----------|----------|-----------|
| No tasks on date | Return zeros, not error | Valid state |
| All tasks incomplete | `avg_success_rate = null` | Cannot compute average |
| Single outlier (0% or 150%) | Include in average | Accurate representation |
| Past dates with no data | Return empty structure | Consistent API |
| Recurring schedule with no generated instances | Template exists in recurring_tasks, 0 instances in tasks | Valid state - cron will generate |

---

### 6.2 Success Rate Outliers

**Detection**:

Outliers are identified using IQR (Interquartile Range) method:
- Calculate Q1 (25th percentile) and Q3 (75th percentile)
- Calculate IQR = Q3 - Q1
- Lower bound = Q1 - 1.5 * IQR
- Upper bound = Q3 + 1.5 * IQR
- Values outside these bounds are flagged as outliers

**Action**: Flag in analytics response, don't exclude from calculations

---

## Code Generation Checklist

- [ ] Implement `SuccessCalculator` class in `analytics/success_calculator.py`
- [ ] Build `AggregationEngine` with direct query functions in `analytics/aggregation.py`
- [ ] Create visualization formatters in `analytics/formatters.py`
- [ ] Add caching layer (Redis) in `analytics/cache.py`
- [ ] Write optimized SQL queries for date ranges in `analytics/queries.py`
- [ ] **Implement recurring task instance generator cron job (runs at midnight)**
- [ ] **Add logic to skip invalid dates (e.g., Feb 31st) for monthly recurrences**
- [ ] Add unit tests for all success formulas
- [ ] Create integration tests for aggregations and date range queries
- [ ] **Add tests for recurring task instance creation**
- [ ] Add performance benchmarks comparing cached vs uncached queries
- [ ] Implement cache invalidation on task updates
- [ ] Document query optimization strategies

---

## 8. Testing Strategy

### Unit Tests

Test cases cover:
- **Unit-based success**: Expected 5, Actual 5 = 100%; Expected 5, Actual 4 = 80%; Expected 10, Actual 7 = 70%
- **Time-based success**: Expected 120, Actual 120 = 100%; Expected 120, Actual 90 = 133.33%; Expected 60, Actual 20 = 150% (capped)
- **Commute success**: Expected 30, Actual 30 = 100%; Expected 30, Actual 25 = 100% (early); Expected 30, Actual 45 = 66.67% (delayed)

### Integration Tests

- End-to-end: Create task → Complete → Verify aggregation
- Range queries: Weekly/Monthly stats across date boundaries
- Caching: Verify cache hits and invalidation

---

**Document Status**: ✅ Complete  
**Dependencies**: `01_data_schema.md`, `03_backend_api_map.md`  
**Next Document**: `05_ai_advisor_logic.md`
