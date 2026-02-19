# Module 05: AI Advisor Logic

## Overview
This document specifies the intelligent failure analysis and recommendation system that detects patterns in task performance and provides actionable schedule optimization suggestions to users.

---

## Architecture

```
┌─────────────────────┐
│   Task History      │
│   Daily Stats       │
│   Success Metrics   │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────────────┐
│   Pattern Detection Engine  │
│   • Chronic failures        │
│   • Time estimation errors  │
│   • Category weaknesses     │
│   • Commute delays          │
│   • Overcommitment signs    │
└──────────┬──────────────────┘
           │
           ▼
┌─────────────────────────────┐
│   Insight Generator (LLM)   │
│   • Contextualize patterns  │
│   • Generate suggestions    │
│   • Prioritize advice       │
└──────────┬──────────────────┘
           │
           ▼
┌─────────────────────────────┐
│   Recommendation Formatter  │
│   • Severity scoring        │
│   • Action items            │
│   • Expected improvements   │
└──────────┬──────────────────┘
           │
           ▼
    ┌──────────────┐
    │  User Nudge  │
    └──────────────┘
```

---

## Core Principles

1. **Proactive, Not Reactive**: Surface insights before users ask
2. **Data-Driven**: All recommendations backed by historical patterns
3. **Actionable**: Provide specific changes, not vague advice
4. **Non-Judgmental**: Frame failures as optimization opportunities
5. **Personalized**: Tailor advice to individual user patterns

---

## 1. Pattern Detection Algorithms

### 1.1 Chronic Task Failure

**Definition**: A task or category consistently underperforming (< 40% success) over 3+ occurrences in a lookback period.

**Special Handling for Recurring Tasks**: Group by recurring_task_id (linking to recurring_tasks table) to analyze all instances of a recurring schedule together.

**Detection Logic**:

Chronic failures are detected by:
1. Filtering tasks from lookback period (default 30 days)
2. Grouping tasks by recurring_task_id (for recurring instances) or normalized name (for one-time tasks)
3. Requiring minimum 3 occurrences to be considered "chronic"
4. Calculating failure rate = (failed_count / total_attempts) * 100
5. Flagging patterns where failure_rate > 60%
6. Computing average success percentage across all attempts
7. Generating specific recommendations based on the pattern
8. Sorting results by failure rate (highest first)

**Example Patterns**:

| Task Name | Attempts | Failed | Failure Rate | Avg Success | Recommendation |
|-----------|----------|--------|--------------|-------------|----------------|
| "Morning workout" | 15 | 12 | 80% | 28.5% | Consider switching to evening workouts or reducing duration by 50% |
| "Deep work session" | 8 | 6 | 75% | 35.2% | Break into smaller 60-min focused blocks instead of 2-hour sessions |
| "Read 5 chapters" | 10 | 7 | 70% | 40.0% | Reduce to 3 chapters per session for more consistent completion |

---

### 1.2 Time Estimation Error Pattern

**Definition**: Consistent mismatch between expected and actual duration, indicating poor planning.

**Detection Logic**:

Time estimation errors are detected by:
1. Filtering time-based and commute tasks from lookback period
2. Requiring both expected_duration and actual_duration fields
3. Grouping by normalized task name
4. Requiring minimum 3 samples per task type
5. Calculating average expected vs average actual duration
6. Computing variance percentage = ((avg_actual - avg_expected) / avg_expected) * 100
7. Flagging patterns where abs(variance) > 30%
8. Classifying as "underestimation" (positive variance) or "overestimation" (negative variance)
9. Generating targeted recommendations based on pattern type
10. Sorting by absolute variance percentage

**Example Patterns**:

| Task Name | Avg Expected | Avg Actual | Variance | Pattern | Recommendation |
|-----------|--------------|------------|----------|---------|----------------|
| "Deep work" | 120 min | 180 min | +50% | Underestimation | Increase estimates to 180 min to prevent schedule overflow |
| "Review emails" | 30 min | 15 min | -50% | Overestimation | Reduce to 15 min estimates; freed time can be reallocated |
| "Commute to work" | 30 min | 42 min | +40% | Underestimation | Add 15-min buffer; consider alternative routes or departure times |

---

### 1.3 Category Weakness Analysis

**Definition**: Entire categories performing poorly compared to user's overall average.

**Detection Logic**:

Category weaknesses are identified by:
1. Filtering completed tasks with success_percentage from lookback period
2. Computing overall average success rate as baseline
3. Grouping tasks by category and calculating category-specific averages
4. Requiring minimum 3 tasks per category for statistical significance
5. Computing performance gap = overall_avg - category_avg
6. Flagging categories where gap > 15 percentage points
7. Generating targeted recommendations for weak categories
8. Sorting by performance gap (largest first)

**Example Patterns**:

| Category | Avg Success | Overall Avg | Gap | Recommendation |
|----------|-------------|-------------|-----|----------------|
| "Exercise" | 58.3% | 84.5% | -26.2% | High struggle area. Consider: 1) Reduce workout intensity, 2) Find accountability partner, 3) Schedule at more energetic time of day |
| "Learning" | 65.1% | 84.5% | -19.4% | Learning tasks often over-scoped. Break into smaller modules (20-30 min) instead of hour-long sessions |

---

### 1.4 Commute Delay Pattern

**Definition**: Recurring delays in commute tasks indicating route/timing issues.

**Detection Logic**:

Commute patterns are detected by:
1. Filtering commute tasks from lookback period
2. Requiring minimum 5 commutes for analysis
3. Counting delays (actual > expected * 1.2)
4. Computing delay rate percentage
5. Calculating average expected vs actual duration
6. Analyzing by time of day (morning vs evening)
7. Identifying worst time of day based on average delay
8. Flagging if delay_rate > 50%
9. Generating specific recommendations (earlier departure, route changes, buffer time)

**Example Pattern**:

```json
{
  "pattern_type": "commute_delay",
  "total_commutes": 22,
  "delayed_count": 14,
  "delay_rate": 63.6,
  "avg_expected": 30,
  "avg_actual": 42,
  "avg_delay": 12,
  "worst_time_of_day": "morning",
  "recommendation": "Your morning commute consistently runs 12 min late (64% of trips). Consider: 1) Departing 15 min earlier, 2) Using alternative route during rush hour, 3) Increasing buffer time in your schedule."
}
```

---

### 1.5 Overcommitment Detection

**Definition**: User scheduling more tasks than realistically completable.

**Detection Logic**:

Overcommitment is detected by:
1. Analyzing daily statistics from lookback period (default 14 days)
2. Computing average completion rate across all days
3. Comparing first week vs second week completion rates to detect decline
4. Counting incomplete high-priority task backlog
5. Calculating average daily task count
6. Flagging if completion_rate < 60% OR declining by 10+ percentage points
7. Generating recommendations to reduce daily task count

**Example Pattern**:

```json
{
  "pattern_type": "overcommitment",
  "avg_completion_rate": 52.3,
  "is_declining": true,
  "avg_daily_tasks": 15.7,
  "incomplete_backlog": 38,
  "recommendation": "You're consistently completing only 52% of planned tasks, with a declining trend. Your backlog has 38 incomplete tasks. Recommendation: Reduce daily task count to 8-10 high-priority items. Focus on completion over quantity."
}
```

---

### 1.6 Recurring Task Adherence Pattern

**Definition**: Analysis of recurring task completion rates to identify habit formation or decay.

**Detection Logic**:

Recurring task adherence is analyzed by:
1. Querying all active recurring task templates from recurring_tasks table
2. For each template, getting all instances from tasks table via recurring_task_id
3. Requiring minimum 5 instances for statistical significance
4. Calculating expected occurrences based on recurrence rules and lookback period
5. Computing actual occurrences and completed occurrences
6. Calculating completion_rate and adherence_rate
7. Comparing first half vs second half completion rates to detect decline
8. Flagging if completion < 50%, adherence < 60%, or declining by 20+ percentage points
9. Generating recommendations (reduce frequency, shorten duration, change timing)
10. Sorting by completion rate (lowest first)

**Example Pattern**:

```json
{
  "pattern_type": "recurring_adherence_decline",
  "task_name": "Morning workout",
  "recurrence_type": "DAILY",
  "expected_occurrences": 30,
  "actual_occurrences": 18,
  "completed_occurrences": 9,
  "completion_rate": 50.0,
  "adherence_rate": 60.0,
  "is_declining": true,
  "recommendation": "Your daily workout adherence has dropped from 70% to 30% over the past 2 weeks. Consider: 1) Reducing frequency to every other day, 2) Shortening workout duration, 3) Changing time of day."
}
```

---

## 2. LLM-Powered Insight Generation

### 2.1 Context Assembly

**Purpose**: Prepare rich context for LLM to generate personalized recommendations.

Context assembly compiles:
- **User profile**: 30-day summary with total tasks, completion rate, success rate, top categories, task type distribution
- **Detected patterns**: All patterns with type, severity, data, and initial recommendations
- **Recent wins**: Top 3 high-performing tasks (success ≥ 90%)
- **Recent struggles**: Top 3 low-performing tasks (success ≤ 40%)

---

### 2.2 LLM System Prompt for Advisor

```text
You are Nudge AI Advisor, a supportive life-management coach. Your role is to analyze user task performance data and provide actionable, encouraging schedule optimization advice.

**INPUT DATA:**
You will receive:
1. User profile (30-day summary)
2. Detected patterns (chronic failures, time estimation errors, etc.)
3. Recent wins (high-performing tasks)
4. Recent struggles (low-performing tasks)

**OUTPUT FORMAT:**
Generate a personalized advisory report with:

1. **Executive Summary** (2-3 sentences)
   - Overall assessment of user's task management
   - Highlight 1-2 major themes

2. **Top 3 Recommendations** (ordered by impact)
   For each:
   - **Issue**: Concise problem statement
   - **Data**: Supporting statistics
   - **Action**: Specific, actionable change
   - **Expected Impact**: Predicted improvement

3. **Celebrate Wins** (1-2 items)
   - Acknowledge areas of strong performance
   - Encourage continuation

4. **Next Steps** (bullet list)
   - 3-5 immediate actions user can take

**TONE GUIDELINES:**
- Supportive, not judgmental
- Data-driven, not prescriptive
- Specific, not vague
- Encouraging, not dismissive of struggles

**CONSTRAINTS:**
- Keep recommendations to 3 max (avoid overwhelming user)
- Prioritize high-impact changes over easy fixes
- Frame failures as learning opportunities
- Use second-person ("You've been..." vs "The user has...")

**EXAMPLE INPUT:**
{
  "user_profile": {
    "total_tasks_30d": 156,
    "avg_completion_rate": 0.68,
    "avg_success_rate": 78.5
  },
  "detected_patterns": [
    {
      "pattern_type": "chronic_failure",
      "data": {
        "task_name": "Morning workout",
        "failure_rate": 80,
        "avg_success": 28.5
      }
    }
  ]
}

**EXAMPLE OUTPUT:**
{
  "executive_summary": "You're maintaining a solid 68% completion rate with strong overall success (78.5%). However, morning workout tasks are a significant struggle point, failing 80% of the time. Let's optimize your schedule to turn this around.",
  
  "recommendations": [
    {
      "issue": "Morning workout chronic failure (80% failure rate)",
      "data": "15 attempts, only 3 completions, avg success 28.5%",
      "action": "Switch workout to evening (6-7 PM) or reduce duration from 60 to 30 minutes",
      "expected_impact": "Based on similar patterns, this could improve success rate to 70%+"
    }
  ],
  
  "celebrate_wins": [
    "Your 'Deep work sessions' have 92% success rate - excellent focus and time estimation!"
  ],
  
  "next_steps": [
    "Reschedule next 5 workouts to evening and track results",
    "Reduce workout time to 30 min for next week as experiment",
    "Review progress in 7 days"
  ]
}
```

---

### 2.3 LLM Request Structure

The AI advisor calls the LLM with:
1. System prompt defining advisor role and output format
2. User message with assembled context (JSON format)
3. Temperature 0.7 for slightly creative personalization
4. Max tokens 1500 for detailed responses
5. JSON response parsing and validation
6. Confidence score calculation based on data quality
7. Returns structured AdvisoryReport with recommendations

---

## 3. Recommendation Severity Scoring

**Purpose**: Prioritize which recommendations to show first.

Severity scoring (1-10 scale):
- **Base severity**: chronic_failure=8, time_estimation_error=5, category_weakness=6, commute_delay=4, overcommitment=9
- **Modifiers**:
  - +1 if failure_rate > 80%
  - +1 if sample_size > 10 (well-established pattern)
  - +2 if pattern is_declining (getting worse)
- **Result**: Clamped to max 10

Higher scores indicate more urgent issues requiring immediate attention.

---

## 4. Notification & Delivery Strategy

### 4.1 Trigger Conditions

| Condition | Trigger | Frequency | Channel |
|-----------|---------|-----------|---------|
| **Critical Issue** | Overcommitment detected | Immediate | Push notification |
| **Weekly Review** | Every Monday 9 AM | Weekly | Email + In-app |
| **Monthly Insights** | 1st of month | Monthly | Email report |
| **Real-Time Nudge** | 3 consecutive task failures | Immediate | In-app banner |
| **Success Milestone** | 90%+ weekly success | Immediate | Celebration notification |

---

### 4.2 Notification Templates

**Critical: Overcommitment**
```
🚨 You've completed only 52% of tasks this week

Your schedule may be overloaded. Tap to see personalized recommendations to reduce stress and improve completion rates.

[View Recommendations]
```

**Weekly Review**
```
📊 Your Week in Review

✅ 45 tasks completed (68% completion rate)
📈 Avg success: 78.5%
🎯 Top category: Work (85% success)

⚠️ 3 insights detected - see how to optimize your schedule

[View Full Report]
```

**Real-Time Nudge**
```
💡 Noticed a pattern

You've struggled with "Deep work" 3 times this week. Consider breaking it into smaller 60-min blocks instead of 2-hour sessions.

[Apply Suggestion] [Dismiss]
```

---

## 5. A/B Testing Framework

### 5.1 Recommendation Variants

Test different advice styles to optimize user adoption:

| Variant | Style | Example |
|---------|-------|---------|
| **A: Direct** | Prescriptive | "Reduce workout to 30 minutes" |
| **B: Suggestive** | Exploratory | "Have you considered shorter workouts?" |
| **C: Data-led** | Analytical | "30-min workouts have 85% success vs 40% for 60-min" |

**Metric to Track**: % of users who implement recommendation within 7 days

---

### 5.2 Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Recommendation Adoption Rate** | > 40% | % users who act on advice within 7 days |
| **Post-Advice Success Improvement** | +15% pts | Compare success rate before/after implementation |
| **User Satisfaction** | > 4.0/5 | Survey: "Was this advice helpful?" |
| **Engagement With Reports** | > 60% | % users who open weekly review |

---

## 6. Edge Cases & Safeguards

### 6.1 Insufficient Data

Data sufficiency checks verify:
- Minimum 10+ completed tasks in lookback period
- At least 5 days with task activity
- At least 2 different task types

If insufficient data: Show generic productivity tips instead of personalized advice.

---

### 6.2 Contradictory Patterns

**Scenario**: User has both "overestimation" and "underestimation" patterns for different tasks.

**Handling**: Segment advice by task category/type:
```
"For 'Deep work' tasks: Increase estimates by 50%"
"For 'Email review' tasks: Reduce estimates to 15 minutes"
```

---

### 6.3 Privacy & Sensitivity

**Guideline**: Avoid diagnosing personal issues (mental health, physical limitations).

**Example - Bad**:
❌ "Your repeated workout failures may indicate depression or lack of motivation."

**Example - Good**:
✅ "Morning workouts have low completion rates. Consider evening scheduling or shorter sessions."

---

## 7. Code Generation Checklist

- [ ] Implement pattern detection algorithms in `advisor/pattern_detection.py`
- [ ] Build LLM integration in `advisor/insight_generator.py`
- [ ] Create severity scoring in `advisor/prioritization.py`
- [ ] Add notification service in `advisor/notifications.py`
- [ ] Implement A/B testing framework in `advisor/experiments.py`
- [ ] Write unit tests for all detection algorithms
- [ ] Add integration tests for end-to-end advice generation
- [ ] Create dashboard view for advisor reports
- [ ] Implement user feedback mechanism (helpful/not helpful)
- [ ] Add analytics tracking for recommendation adoption

---

## 8. API Integration

### 8.1 New Endpoint: GET /advisor/insights

**Request**:
```
GET /api/v1/advisor/insights?lookback_days=30
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "generated_at": "2026-02-18T15:45:00Z",
    "data_sufficiency": {
      "sufficient": true,
      "task_count": 156,
      "date_range": "2026-01-19 to 2026-02-18"
    },
    "executive_summary": "You're maintaining strong performance (78.5% avg success) with room for optimization in exercise routines.",
    "recommendations": [
      {
        "id": "rec_001",
        "severity": 8,
        "pattern_type": "chronic_failure",
        "issue": "Morning workout chronic failure (80% failure rate)",
        "action": "Switch to evening (6-7 PM) or reduce to 30 minutes",
        "expected_impact": "Could improve success rate to 70%+",
        "supporting_data": {
          "attempts": 15,
          "completions": 3,
          "avg_success": 28.5
        }
      }
    ],
    "celebrate_wins": [
      "Your 'Deep work sessions' have 92% success rate!"
    ],
    "next_steps": [
      "Reschedule next 5 workouts to evening",
      "Review progress in 7 days"
    ]
  }
}
```

---

### 8.2 New Endpoint: POST /advisor/feedback

**Purpose**: Track if user found advice helpful

**Request**:
```json
{
  "recommendation_id": "rec_001",
  "helpful": true,
  "implemented": true,
  "notes": "Switched to evening workouts and it's working!"
}
```

---

## 9. Example Advisor Report

### Full Report Example

```markdown
# Your Nudge Insights - Week of Feb 12-18, 2026

## 📊 Performance Summary
- **156 tasks** created (avg 22/day)
- **68% completion rate** (106 completed, 50 incomplete/failed)
- **78.5% average success** on completed tasks

## 🎯 Top 3 Recommendations

### 1. 🏋️ Optimize Morning Workout Schedule (Critical)
**Issue**: Morning workout has 80% failure rate (12 of 15 attempts failed)

**Data**: 
- Expected: 60 min per session
- Actual: Completed only 3 times, avg 25 min
- Success rate: 28.5%

**Action**: 
1. **Switch to evening** (6-7 PM when energy is higher), OR
2. **Reduce to 30 minutes** to match realistic capacity

**Expected Impact**: Based on similar user patterns, this could boost success to 70%+

---

### 2. ⏱️ Adjust Deep Work Time Estimates (Moderate)
**Issue**: Deep work sessions consistently take 50% longer than planned

**Data**:
- Expected: 120 min average
- Actual: 180 min average
- Causes schedule overflow and task delays

**Action**: Increase deep work estimates to 180 min OR break into two 90-min sessions with break

**Expected Impact**: Better schedule accuracy, reduce cascading delays

---

### 3. 🚗 Add Commute Buffer (Low Priority)
**Issue**: Morning commute delayed 64% of the time

**Data**:
- Expected: 30 min
- Actual: 42 min average (+12 min)
- 14 of 22 commutes were late

**Action**: 
1. Depart 15 minutes earlier, OR
2. Update commute estimate to 45 min

**Expected Impact**: Reduce late arrivals and stress

---

## 🎉 Celebrate Your Wins!

✅ **Deep work sessions**: 92% success rate - excellent focus!

✅ **Email management**: Consistently finishing in 15 min (50% faster than estimate)

✅ **Friday productivity**: Best day of the week (95% success rate)

---

## 📝 Next Steps

This week, try:
1. ⚡ Reschedule next 5 workouts to evening (6-7 PM)
2. 📊 Update deep work estimates to 180 min
3. 🕒 Add 15-min buffer to morning commute
4. 📆 Review progress next Monday

---

*Based on 156 tasks from Jan 19 - Feb 18, 2026*
```

---

## 10. Future Enhancements (Phase 2)

| Enhancement | Description | Impact |
|-------------|-------------|--------|
| **Habit Formation Tracking** | Track streak days for recurring tasks | Gamification |
| **Collaborative Insights** | "Users like you improved by..." | Social proof |
| **Predictive Scheduling** | AI suggests optimal time slots | Proactive |
| **Energy Pattern Detection** | Identify peak performance hours | Personalization |
| **Goal Setting Integration** | Align recommendations with user goals | Motivation |

---

**Document Status**: ✅ Complete  
**Dependencies**: `01_data_schema.md`, `04_analytics_engine.md`  
**Project Status**: All specification documents complete ✅
