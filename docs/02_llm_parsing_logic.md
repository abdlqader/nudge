# Module 02: LLM Parsing Logic

## Overview
This document specifies the natural language processing pipeline using Qwen 8B (primary) or Gemini 1.5 (fallback) to convert unstructured user input into structured JSON task definitions.

---

## Architecture

```
User Input (Text)
      ↓
┌─────────────────────────┐
│   Preprocessing Layer   │
│   • Sanitize input      │
│   • Detect intent       │
└────────────┬────────────┘
             ↓
┌─────────────────────────┐
│    LLM Layer (Qwen 8B)  │
│   • System Prompt       │
│   • Few-Shot Examples   │
│   • JSON Schema         │
└────────────┬────────────┘
             ↓
┌─────────────────────────┐
│   Validation Layer      │
│   • Schema validation   │
│   • Business rules      │
│   • Confidence scoring  │
└────────────┬────────────┘
             ↓
       Structured JSON
```

---

## Core Objectives

1. **Extract Task Metadata**: Name, type, duration, units, priority
2. **Handle Ambiguity**: Request clarification when confidence < 70%
3. **Support Multi-Task**: Parse multiple tasks from single input
4. **Identify Commutes**: Detect travel-related tasks automatically
5. **Preserve Context**: Maintain conversation history for complex interactions

---

## System Prompt Design

### Base System Prompt

```text
You are an AI assistant for "Nudge," a life-management application. Your role is to parse natural language task descriptions into structured JSON.

**EXTRACTION RULES:**

1. Task Types:
   • UNIT_BASED: Countable items (chapters, meetings, emails)
   • TIME_BASED: Duration-focused (study for 2 hours, workout 30 min)
   • COMMUTE: Travel between locations

2. Task Categories:
   • ANCHOR: Non-negotiable time blocks (sleep, family time, meals, personal care)
   • TRANSIT: Movement between locations (commute, drive, travel, flights)
   • ACTION: Productive work tasks (default for most tasks)
   
   Category Detection Keywords:
   - ANCHOR: "sleep", "family time", "family dinner", "breakfast", "lunch", "dinner", "morning routine"
   - TRANSIT: "commute", "drive", "travel", "flight", "go to", "heading to"
   - ACTION: Default if no anchor/transit keywords detected

3. Duration Conversion:
   • "2 hours" = 120 minutes
   • "half an hour" = 30 minutes
   • "quick" = 15 minutes (estimate)
   • "all day" = 480 minutes (8 hours)

3. Recurrence Detection:
   • DAILY: "every day", "daily", "each day"
   • WEEKLY: "every Monday", "every week", "Mondays and Wednesdays"
   • MONTHLY_DATE: "on the 15th", "every 1st of the month"
   • MONTHLY_PATTERN: "first Monday", "last Friday", "second Tuesday"
   
   Recurrence Interval:
   • "every 2 days" = interval 2
   • "every other day" = interval 2
   
   Days of Week Mapping:
   • Sunday=0, Monday=1, Tuesday=2, Wednesday=3, Thursday=4, Friday=5, Saturday=6
   
   **IMPORTANT: Recurring tasks generate entries in TWO tables:**
   - A recurrence pattern in `recurring_tasks` (ONLY recurrence config: type, interval, days, etc.)
   - Task instances in `tasks` (FULL task data: name, type, duration, units, priority, etc.)
   
   **Field Organization**: 
   - `recurring_tasks` table: Contains ONLY the recurrence configuration fields
   - `tasks` table: Contains ALL task fields for each instance (each task is complete and independent)
   - When LLM parses recurring task input, it extracts ALL task fields + recurrence pattern
   - Backend creates recurrence pattern record + generates task instances with full task data
   
   Mark `is_recurring: true` when you detect recurring patterns.

4. Priority Detection:
   • CRITICAL (4): "urgent", "emergency", "ASAP", "critical"
   • HIGH (3): "important", "must do", "priority"
   • MEDIUM (2): Default (no keywords)
   • LOW (1): "when I have time", "eventually", "low priority"

5. Ambiguity Handling:
   • If duration/units unclear, set to null
   • Add "needs_clarification" flag
   • Provide "clarification_question" field

**OUTPUT FORMAT:**
Return ONLY valid JSON. No explanations, no markdown.

{
  "tasks": [
    {
      "name": "string",
      "task_type": "UNIT_BASED | TIME_BASED | COMMUTE",
      "task_category": "ANCHOR | TRANSIT | ACTION",
      "expected_duration": integer or null,
      "expected_units": integer or null,
      "priority": 1-4,
      "category": "string or null",
      "deadline": "ISO 8601 or null",
      "notes": "string or null",
      "is_recurring": boolean,
      "recurrence_type": "NONE | DAILY | WEEKLY | MONTHLY_DATE | MONTHLY_PATTERN",
      "recurrence_interval": integer or null,
      "recurrence_days": array or null,
      "recurrence_day_of_month": integer or null,
      "recurrence_pattern": "string or null",
      "recurrence_end_date": "ISO 8601 or null",
      "needs_clarification": boolean,
      "clarification_question": "string or null"
    }
  ]
}

**CONSTRAINTS:**
• UNIT_BASED must have expected_units
• TIME_BASED/COMMUTE must have expected_duration
• name: 1-200 characters
• category: max 50 characters
• notes: max 1000 characters
• Deadlines must be future timestamps

**RECURRING TASK HANDLING:**
When `is_recurring: true`, the backend will:
1. Create an entry in `recurring_tasks` table with the recurrence fields
2. Generate the first task instance in `tasks` table automatically
3. Link the instance to the template via `recurring_task_id`

**EXAMPLES:**
[See few-shot examples below]
```

---

## Few-Shot Examples

### Example 1: Simple Unit-Based Task

**User Input:**
```
Read 3 chapters of "Atomic Habits" today
```

**Expected Output:**
```json
{
  "tasks": [
    {
      "name": "Read 3 chapters of Atomic Habits",
      "task_type": "UNIT_BASED",
      "task_category": "ACTION",
      "expected_duration": 90,
      "expected_units": 3,
      "priority": 2,
      "category": "Reading",
      "deadline": "2026-02-18T23:59:00Z",
      "notes": null,
      "needs_clarification": false,
      "clarification_question": null
    }
  ]
}
```

**Reasoning:**
- "Read 3 chapters" → UNIT_BASED (expected_units = 3)
- Estimated 30 min per chapter → 90 min total
- "today" → deadline set to end of day
- No priority keywords → MEDIUM (2)

---

### Example 2: Time-Based Task with Priority

**User Input:**
```
URGENT: Prepare presentation slides for 2 hours before tomorrow's meeting
```

**Expected Output:**
```json
{
  "tasks": [
    {
      "name": "Prepare presentation slides",
      "task_type": "TIME_BASED",
      "task_category": "ACTION",
      "expected_duration": 120,
      "expected_units": null,
      "priority": 4,
      "category": "Work",
      "deadline": "2026-02-19T09:00:00Z",
      "notes": "Before tomorrow's meeting",
      "needs_clarification": false,
      "clarification_question": null
    }
  ]
}
```

**Reasoning:**
- "for 2 hours" → TIME_BASED (120 minutes)
- "URGENT" → CRITICAL priority (4)
- "tomorrow's meeting" → Set reasonable deadline (9AM next day)
- "presentation" → Categorized as Work

---

### Example 3: Commute Detection

**User Input:**
```
Drive to office, usually takes 35 minutes
```

**Expected Output:**
```json
{
  "tasks": [
    {
      "name": "Drive to office",
      "task_type": "COMMUTE",
      "task_category": "TRANSIT",
      "expected_duration": 35,
      "expected_units": null,
      "priority": 2,
      "category": "Commute",
      "deadline": null,
      "notes": null,
      "needs_clarification": false,
      "clarification_question": null
    }
  ]
}
```

**Reasoning:**
- "Drive to" → COMMUTE keyword detected
- task_category → TRANSIT (auto-set for commutes)
- "35 minutes" → expected_duration
- Auto-categorized as "Commute"

---

### Example 4: Multi-Task Input

**User Input:**
```
Today: workout 45 min, review 5 emails, and commute home (1 hour)
```

**Expected Output:**
```json
{
  "tasks": [
    {
      "name": "Workout",
      "task_type": "TIME_BASED",
      "task_category": "ACTION",
      "expected_duration": 45,
      "expected_units": null,
      "priority": 2,
      "category": "Health",
      "deadline": null,
      "notes": null,
      "needs_clarification": false,
      "clarification_question": null
    },
    {
      "name": "Review emails",
      "task_type": "UNIT_BASED",
      "task_category": "ACTION",
      "expected_duration": 25,
      "expected_units": 5,
      "priority": 2,
      "category": "Work",
      "deadline": null,
      "notes": null,
      "needs_clarification": false,
      "clarification_question": null
    },
    {
      "name": "Commute home",
      "task_type": "COMMUTE",
      "task_category": "TRANSIT",
      "expected_duration": 60,
      "expected_units": null,
      "priority": 2,
      "category": "Commute",
      "deadline": null,
      "notes": null,
      "needs_clarification": false,
      "clarification_question": null
    }
  ]
}
```

**Reasoning:**
- Three distinct tasks separated by commas
- "workout 45 min" → TIME_BASED
- "review 5 emails" → UNIT_BASED (estimated 5 min/email)
- "commute home" → COMMUTE detected

---

### Example 5: Ambiguous Input (Needs Clarification)

**User Input:**
```
Finish the project soon
```

**Expected Output:**
```json
{
  "tasks": [
    {
      "name": "Finish the project",
      "task_type": "TIME_BASED",
      "task_category": "ACTION",
      "expected_duration": null,
      "expected_units": null,
      "priority": 2,
      "category": null,
      "deadline": null,
      "notes": null,
      "needs_clarification": true,
      "clarification_question": "How much time do you expect this to take? (e.g., 2 hours, half a day)"
    }
  ]
}
```

**Reasoning:**
- "soon" is vague → No clear duration
- Missing measurable metric
- System requests clarification

---

### Example 6: Deadline with Relative Time

**User Input:**
```
Call mom in 3 hours
```

**Expected Output:**
```json
{
  "tasks": [
    {
      "name": "Call mom",
      "task_type": "TIME_BASED",
      "task_category": "ACTION",
      "expected_duration": 15,
      "expected_units": null,
      "priority": 2,
      "category": "Personal",
      "deadline": "2026-02-18T16:30:00Z",
      "notes": null,
      "needs_clarification": false,
      "clarification_question": null
    }
  ]
}
```

**Reasoning:**
- "in 3 hours" → Deadline = current_time + 3 hours
- Phone calls typically ~15 min (estimated duration)
- Categorized as Personal

---

### Example 7: Complex Deadline with Time

**User Input:**
```
Submit report by Friday 5pm - needs 3 hours of work
```

**Expected Output:**
```json
{
  "tasks": [
    {
      "name": "Submit report",
      "task_type": "TIME_BASED",
      "task_category": "ACTION",
      "expected_duration": 180,
      "expected_units": null,
      "priority": 3,
      "category": "Work",
      "deadline": "2026-02-21T17:00:00Z",
      "notes": "Deadline: Friday 5pm",
      "needs_clarification": false,
      "clarification_question": null
    }
  ]
}
```

**Reasoning:**
- "by Friday 5pm" → Specific deadline (assume next Friday)
- "needs 3 hours" → 180 minutes
- "submit" implies HIGH priority (3)

---

### Example 8: Anchor Task - Sleep

**User Input:**
```
Sleep 8 hours tonight
```

**Expected Output:**
```json
{
  "tasks": [
    {
      "name": "Sleep 8 hours",
      "task_type": "TIME_BASED",
      "task_category": "ANCHOR",
      "expected_duration": 480,
      "expected_units": null,
      "priority": 4,
      "category": "Health",
      "deadline": null,
      "notes": null,
      "needs_clarification": false,
      "clarification_question": null
    }
  ]
}
```

**Reasoning:**
- "Sleep" → ANCHOR keyword detected
- "8 hours" → 480 minutes
- ANCHOR tasks default to CRITICAL priority (4)
- Essential foundational time block

---

### Example 9: Anchor Task - Family Time

**User Input:**
```
Family dinner at 7pm for 1 hour
```

**Expected Output:**
```json
{
  "tasks": [
    {
      "name": "Family dinner",
      "task_type": "TIME_BASED",
      "task_category": "ANCHOR",
      "expected_duration": 60,
      "expected_units": null,
      "priority": 4,
      "category": "Family",
      "deadline": "2026-02-18T19:00:00Z",
      "notes": null,
      "needs_clarification": false,
      "clarification_question": null
    }
  ]
}
```

**Reasoning:**
- "Family dinner" → ANCHOR category (non-negotiable time)
- "at 7pm" → deadline set
- Automatically assigned CRITICAL priority

---

### Example 10: Daily Recurring Task

**User Input:**
```
Daily standup meeting at 9am, takes 15 minutes
```

**Expected Output:**
```json
{
  "tasks": [
    {
      "name": "Daily standup meeting",
      "task_type": "TIME_BASED",
      "task_category": "ACTION",
      "expected_duration": 15,
      "expected_units": null,
      "priority": 3,
      "category": "Work",
      "deadline": "2026-02-18T09:00:00Z",
      "notes": null,
      "is_recurring": true,
      "recurrence_type": "DAILY",
      "recurrence_interval": 1,
      "recurrence_days": null,
      "recurrence_day_of_month": null,
      "recurrence_pattern": null,
      "recurrence_end_date": null,
      "needs_clarification": false,
      "clarification_question": null
    }
  ]
}
```

**Reasoning:**
- "Daily" → DAILY recurrence detected
- "at 9am" → deadline for first occurrence
- "15 minutes" → expected_duration
- recurrence_interval defaults to 1 (every day)

---

### Example 11: Weekly Recurring Task

**User Input:**
```
Team sync every Monday and Wednesday at 2pm for 1 hour
```

**Expected Output:**
```json
{
  "tasks": [
    {
      "name": "Team sync",
      "task_type": "TIME_BASED",
      "task_category": "ACTION",
      "expected_duration": 60,
      "expected_units": null,
      "priority": 3,
      "category": "Work",
      "deadline": null,
      "notes": "Every Monday and Wednesday at 2pm",
      "is_recurring": true,
      "recurrence_type": "WEEKLY",
      "recurrence_interval": 1,
      "recurrence_days": [1, 3],
      "recurrence_day_of_month": null,
      "recurrence_pattern": null,
      "recurrence_end_date": null,
      "needs_clarification": false,
      "clarification_question": null
    }
  ]
}
```

**Reasoning:**
- "every Monday and Wednesday" → WEEKLY recurrence
- Monday=1, Wednesday=3 → recurrence_days: [1, 3]
- "1 hour" → 60 minutes

---

### Example 12: Monthly Pattern Recurring Task

**User Input:**
```
Board meeting on the first Monday of every month, 2 hours
```

**Expected Output:**
```json
{
  "tasks": [
    {
      "name": "Board meeting",
      "task_type": "TIME_BASED",
      "task_category": "ACTION",
      "expected_duration": 120,
      "expected_units": null,
      "priority": 3,
      "category": "Work",
      "deadline": null,
      "notes": null,
      "is_recurring": true,
      "recurrence_type": "MONTHLY_PATTERN",
      "recurrence_interval": null,
      "recurrence_days": null,
      "recurrence_day_of_month": null,
      "recurrence_pattern": "first_monday",
      "recurrence_end_date": null,
      "needs_clarification": false,
      "clarification_question": null
    }
  ]
}
```

**Reasoning:**
- "first Monday of every month" → MONTHLY_PATTERN
- Pattern: "first_monday"
- "2 hours" → 120 minutes

---

### Example 13: Monthly Date Recurring Task

**User Input:**
```
Pay rent on the 1st of every month
```

**Expected Output:**
```json
{
  "tasks": [
    {
      "name": "Pay rent",
      "task_type": "TIME_BASED",
      "task_category": "ACTION",
      "expected_duration": 10,
      "expected_units": null,
      "priority": 4,
      "category": "Finance",
      "deadline": null,
      "notes": null,
      "is_recurring": true,
      "recurrence_type": "MONTHLY_DATE",
      "recurrence_interval": null,
      "recurrence_days": null,
      "recurrence_day_of_month": 1,
      "recurrence_pattern": null,
      "recurrence_end_date": null,
      "needs_clarification": false,
      "clarification_question": null
    }
  ]
}
```

**Reasoning:**
- "on the 1st of every month" → MONTHLY_DATE
- recurrence_day_of_month: 1
- High priority for financial obligations

---

## Edge Cases & Handling

### Edge Case Table

| Input Pattern | Challenge | Handling Strategy | Example |
|---------------|-----------|-------------------|---------|
| **Vague quantities** | "Read some articles" | Set needs_clarification=true | Ask: "How many articles?" |
| **Implied duration** | "Quick workout" | Use defaults (15 min) | Estimate conservatively |
| **Past dates** | "Finish yesterday's task" | Reject or set to today | Error message |
| **Contradictions** | "5 chapters in 10 min" | Flag unrealistic | Add warning note |
| **Multiple deadlines** | "By Friday or Saturday" | Use earliest | Take Friday |
| **Emoji/slang** | "Gym 💪 45m" | Sanitize input | Strip emoji, parse "45m" |
| **Ambiguous category** | "Morning routine" | Default to ANCHOR | Classify as foundational time |
| **Transit vs Action** | "Drive to gym" | Check context | If "to [location]" = TRANSIT |
| **Ambiguous recurrence** | "Meeting next few Mondays" | Ask for clarification | "How many weeks?" or set end_date |
| **Invalid day of month** | "Every 31st" (in Feb) | Accept, skip invalid months | Skip February automatically |
| **Multiple recurrence patterns** | "Every Mon and 1st of month" | Choose primary pattern | Prioritize weekly over monthly |

---

## Preprocessing Pipeline

### Step 1: Input Sanitization

Input text is sanitized by:
- Removing excessive whitespace
- Optionally stripping emojis (though they may carry meaning)
- Expanding common abbreviations (min → minutes, hr → hour, etc.)
- Normalizing case while preserving context for LLM

### Step 2: Intent Detection (Pre-LLM)

The system detects user intent before LLM processing:
- **create_task**: Keywords like "add", "create", "new task", "remind me"
- **query_stats**: Keywords like "how did I do", "show me", "statistics"
- **update_task**: Keywords like "mark done", "complete", "finished"
- **delete_task**: Keywords like "remove", "delete", "cancel"

Requests are routed to appropriate handlers based on detected intent.

---

## LLM Configuration

### Model Parameters (Qwen 8B)

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| `temperature` | 0.2 | Low variance for consistent JSON |
| `max_tokens` | 1024 | Sufficient for multi-task responses |
| `top_p` | 0.9 | Balanced creativity and determinism |
| `frequency_penalty` | 0.3 | Reduce repetition in task names |
| `presence_penalty` | 0.0 | No penalty for recurring concepts |

### Fallback Strategy (Qwen 8B → Gemini 1.5)

```
IF qwen_response.is_invalid_json OR qwen_confidence < 0.6:
    RETRY with Gemini 1.5 Flash
    USE same system prompt
    
IF gemini_response.is_invalid_json:
    FALLBACK to clarification mode
    ASK user: "Could you rephrase that? E.g., 'Read 3 chapters for 90 minutes'"
```

---

## Post-Processing & Validation

### JSON Schema Validation

The system validates LLM output against a comprehensive JSON schema ensuring:
- Required fields are present (name, task_type, priority)
- Field types match expectations (strings, integers, enums)
- Value ranges are within bounds (duration 1-1440, priority 1-4)
- Enum values are valid
- Date-time strings follow ISO 8601 format
- Recurrence fields match recurrence_type requirements

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["tasks"],
  "properties": {
    "tasks": {
      "type": "array",
      "minItems": 1,
      "maxItems": 10,
      "items": {
        "type": "object",
        "required": ["name", "task_type", "priority"],
        "properties": {
          "name": {
            "type": "string",
            "minLength": 1,
            "maxLength": 200
          },
          "task_type": {
            "type": "string",
            "enum": ["UNIT_BASED", "TIME_BASED", "COMMUTE"]
          },
          "expected_duration": {
            "type": ["integer", "null"],
            "minimum": 1,
            "maximum": 1440
          },
          "expected_units": {
            "type": ["integer", "null"],
            "minimum": 1,
            "maximum": 1000
          },
          "priority": {
            "type": "integer",
            "minimum": 1,
            "maximum": 4
          },
          "category": {
            "type": ["string", "null"],
            "maxLength": 50
          },
          "deadline": {
            "type": ["string", "null"],
            "format": "date-time"
          },
          "notes": {
            "type": ["string", "null"],
            "maxLength": 1000
          },
          "is_recurring": {
            "type": "boolean"
          },
          "recurrence_type": {
            "type": "string",
            "enum": ["NONE", "DAILY", "WEEKLY", "MONTHLY_DATE", "MONTHLY_PATTERN"]
          },
          "recurrence_interval": {
            "type": ["integer", "null"],
            "minimum": 1
          },
          "recurrence_days": {
            "type": ["array", "null"],
            "items": {
              "type": "integer",
              "minimum": 0,
              "maximum": 6
            }
          },
          "recurrence_day_of_month": {
            "type": ["integer", "null"],
            "minimum": 1,
            "maximum": 31
          },
          "recurrence_pattern": {
            "type": ["string", "null"],
            "enum": [null, "first_monday", "first_tuesday", "first_wednesday", "first_thursday", "first_friday", "first_saturday", "first_sunday",
                     "second_monday", "second_tuesday", "second_wednesday", "second_thursday", "second_friday", "second_saturday", "second_sunday",
                     "third_monday", "third_tuesday", "third_wednesday", "third_thursday", "third_friday", "third_saturday", "third_sunday",
                     "fourth_monday", "fourth_tuesday", "fourth_wednesday", "fourth_thursday", "fourth_friday", "fourth_saturday", "fourth_sunday",
                     "last_monday", "last_tuesday", "last_wednesday", "last_thursday", "last_friday", "last_saturday", "last_sunday"]
          },
          "recurrence_end_date": {
            "type": ["string", "null"],
            "format": "date-time"
          },
          "needs_clarification": {
            "type": "boolean"
          },
          "clarification_question": {
            "type": ["string", "null"]
          }
        }
      }
    }
  }
}
```

### Business Rule Validation

Post-LLM validation checks enforce business rules:
- **Task type consistency**: UNIT_BASED requires expected_units, TIME_BASED/COMMUTE require expected_duration
- **Deadline validation**: Deadlines must be in the future
- **Realistic estimates**: Warns if time per unit is less than 1 minute
- **Recurrence field validation**:
  - WEEKLY requires recurrence_days array
  - MONTHLY_DATE requires recurrence_day_of_month
  - MONTHLY_PATTERN requires recurrence_pattern
  - Recurrence end dates must be in the future

---

## Confidence Scoring

### Scoring Algorithm

Confidence is calculated based on:
- **Base score**: 1.0
- **Deductions**:
  - -0.3 if needs clarification
  - -0.2 if missing both duration and units
  - -0.2 for vague task names ("thing", "stuff", etc.)
- **Bonuses**:
  - +0.1 if input contains explicit numbers
- **Result**: Clamped to [0.0, 1.0]

### Confidence Thresholds

| Confidence | Action | UI Behavior |
|------------|--------|-------------|
| **> 0.8** | Auto-create task | Silent success, show confirmation |
| **0.6 - 0.8** | Create with warning | Show parsed result, ask "Is this correct?" |
| **0.4 - 0.6** | Request clarification | Show clarification_question |
| **< 0.4** | Reject parsing | "I didn't understand. Could you rephrase?" |

---

## Conversation Context Management

### Multi-Turn Clarification Flow

```
TURN 1:
User: "Finish the project"
Bot: "How much time do you expect this to take? (e.g., 2 hours, half a day)"

TURN 2:
User: "3 hours"
Bot: [Combines context from Turn 1 + Turn 2]
     Creates task: "Finish the project" (180 minutes)

Context Payload:
{
  "conversation_id": "uuid",
  "history": [
    {"role": "user", "content": "Finish the project"},
    {"role": "assistant", "content": "How much time..."},
    {"role": "user", "content": "3 hours"}
  ],
  "pending_task": {
    "name": "Finish the project",
    "needs_clarification": true
  }
}
```

---

## Error Handling Matrix

| Error Type | Detection | Response | Retry Strategy |
|------------|-----------|----------|----------------|
| **Invalid JSON** | JSON parse error | Return 400 error | Retry with Gemini |
| **Missing fields** | Schema validation | Add defaults | Auto-correct |
| **LLM timeout** | 10s timeout | Return 503 | Queue for retry |
| **Hallucination** | Confidence < 0.4 | Request clarification | N/A |
| **Rate limit** | API 429 error | Queue request | Exponential backoff |

---

## Testing Strategy

### Unit Tests (Per Example)

```python
test_cases = [
    {
        "input": "Read 3 chapters of Atomic Habits today",
        "expected": {
            "name": "Read 3 chapters of Atomic Habits",
            "task_type": "UNIT_BASED",
            "expected_units": 3,
            "is_recurring": false
        }
    },
    {
        "input": "Daily standup meeting at 9am, takes 15 minutes",
        "expected": {
            "name": "Daily standup meeting",
            "task_type": "TIME_BASED",
            "is_recurring": true,
            "recurrence_type": "DAILY",
            "recurrence_interval": 1
        }
    },
    {
        "input": "Team sync every Monday and Wednesday at 2pm for 1 hour",
        "expected": {
            "name": "Team sync",
            "task_type": "TIME_BASED",
            "is_recurring": true,
            "recurrence_type": "WEEKLY",
            "recurrence_days": [1, 3]
        }
    },
    {
        "input": "Board meeting on the first Monday of every month, 2 hours",
        "expected": {
            "name": "Board meeting",
            "task_type": "TIME_BASED",
            "is_recurring": true,
            "recurrence_type": "MONTHLY_PATTERN",
            "recurrence_pattern": "first_monday"
        }
    },
    # ... all 13 examples above
]

for case in test_cases:
    response = parse_with_llm(case["input"])
    assert response["tasks"][0]["name"] == case["expected"]["name"]
    assert response["tasks"][0]["task_type"] == case["expected"]["task_type"]
```

### Integration Tests

1. **End-to-End Flow**: User input → LLM → Validation → DB insertion
2. **Fallback Testing**: Force Qwen failure, verify Gemini takeover
3. **Load Testing**: 100 concurrent parse requests
4. **Adversarial Inputs**: SQL injection attempts, XXL strings

---

## Code Generation Checklist

- [ ] Implement `llm/qwen_client.py` with system prompt
- [ ] Implement `llm/gemini_fallback.py`
- [ ] Create `llm/preprocessing.py` for sanitization
- [ ] Add JSON schema validator in `llm/validation.py`
- [ ] Build confidence scoring in `llm/confidence.py`
- [ ] Create conversation context manager `llm/context.py`
- [ ] **Add recurrence pattern parser in `llm/recurrence_parser.py`**
- [ ] **Implement natural language to cron pattern converter**
- [ ] Write unit tests for all 13 examples (including recurring)
- [ ] Add integration test suite
- [ ] Create LLM cost tracking (tokens used)
- [ ] Implement rate limiting middleware

---

## Performance & Cost Optimization

### Cost Estimation (Monthly)

| Model | Cost/1M Tokens | Avg Tokens/Request | Requests/Day | Monthly Cost |
|-------|----------------|---------------------|--------------|--------------|
| **Qwen 8B** | $0.15 | 500 | 100 | $2.25 |
| **Gemini 1.5 Flash** | $0.35 | 500 | 10 (fallback) | $0.53 |
| **Total** | - | - | - | **$2.78/user** |

### Caching Strategy

```python
# Cache common task patterns
cache_key = hash(sanitized_input)
if cache_key in redis_cache:
    return cached_response  # Skip LLM call
else:
    response = call_llm(input)
    redis_cache.set(cache_key, response, ttl=3600)
```

---

**Document Status**: ✅ Complete  
**Dependencies**: `01_data_schema.md` (for task structure)  
**Next Document**: `03_backend_api_map.md`
