# Nudge - Project Master Plan & Development Roadmap

## Executive Summary
**Nudge** is a frictionless, AI-first life-management engine that allows users to interact via natural language to manage their tasks, optimize "dead time" (commutes), and track success analytics through intelligent percentage-based metrics.

---

## System Architecture Overview

```
┌─────────────────┐
│   User Input    │
│ (Natural Lang)  │
└────────┬────────┘
         │
         ▼
┌─────────────────────────┐
│  LLM Layer (Qwen 8B)    │
│  • Parse Intent         │
│  • Extract JSON         │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  FastAPI Backend        │
│  • Task Management      │
│  • CRUD Operations      │
│  • Analytics Engine     │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Database Layer         │
│  PostgreSQL             │
│  • Tasks                │
│  • Recurring_Tasks      │
│  • Indexed for queries  │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Analytics Dashboard    │
│  • Success Rates        │
│  • Pie Charts           │
│  • Trend Analysis       │
│  • Cached queries       │
└─────────────────────────┘
```

---

## Core Concepts

### 1. **Commutes as Tasks**
Commutes are treated as standard tasks with duration properties, allowing the AI to optimize "dead time" and suggest productive activities during transit.

### 2. **Task Categories**
Every task is classified into one of three categories:
- **ANCHOR**: Non-negotiable foundational time blocks (sleep, family time, meals)
- **TRANSIT**: Movement between locations (commutes, drives, travel)
- **ACTION**: Productive work tasks (deliverables, reports, learning)

This categorization enables better schedule optimization and pattern analysis.

### 3. **Recurring Tasks**
Tasks can repeat on flexible schedules:
- **Daily**: Every day or every N days
- **Weekly**: Specific days of the week (e.g., Monday and Wednesday)
- **Monthly (Date)**: Specific day of month (e.g., 1st, 15th)
- **Monthly (Pattern)**: Pattern-based (e.g., "first Monday", "last Friday")

Recurring tasks use a **template + instances** architecture:
- Templates stored in `recurring_tasks` table (define the schedule)
- Instances auto-generated daily in `tasks` table (individual occurrences)
- Each instance links to its template via `recurring_task_id`

This enables independent tracking of each occurrence while maintaining the schedule definition.

### 4. **Sliding Scale Success Metric**
Success is calculated as a percentage based on:
- **Unit-based tasks**: `(actual_units / expected_units) × 100`
- **Time-based tasks**: `(expected_duration / actual_duration) × 100` (inverted - finishing faster = more successful)
- **Commutes**: `actual_duration / expected_duration` (on-time = 100%)

### 3. **AI-Driven Insights**
The system analyzes failure trends (Success < 40%) and provides actionable schedule modifications.

---

## Technology Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Backend** | Go (Golang) | HTTP server, business logic |
| **Database (Dev)** | SQLite (libSQL) | Local development database |
| **Database (Prod)** | Turso (libSQL) | Production cloud database |
| **Database Driver** | go-libsql | Unified driver for SQLite and Turso |
| **LLM** | Qwen 8B / Gemini 1.5 | Natural language understanding |
| **Frontend** | React/Vue (future) | Dashboard and visualizations |
| **Hosting** | Docker + Cloud Run | Containerized deployment |

### Environment Configuration

The application requires environment variables for database configuration:

**Development (Local SQLite)**
- `DB_URL`: `file:local.db`
- `DB_TOKEN`: Empty (not required)

**Production (Turso)**
- `DB_URL`: `libsql://your-database.turso.io`
- `DB_TOKEN`: Authentication token from Turso
- Easy local development without cloud dependencies
- Seamless deployment to Turso production

### **Phase 1: Foundation Layer** (Week 1-2)

#### Module 01: Data Schema Design
**File**: `01_data_schema.md`

**Deliverables**:
- Data models for all entities
- libSQL database schema (SQLite/Turso compatible)
- Migration strategy
- Validation rules
- Database connection logic with environment variable support

**Key Entities**:
- Tasks (task instances with task_category, linked to recurring_tasks via recurring_task_id)
- Recurring_Tasks (schedule templates in separate table)
- Success values (computed on-demand from tasks table)

---

#### Module 02: LLM Parsing Logic
**File**: `02_llm_parsing_logic.md`

**Deliverables**:
- System prompt engineering
- Few-shot examples for task extraction (including recurring patterns)
- JSON schema for LLM output
- Error handling for ambiguous inputs

**Capabilities**:
- Extract task name, duration, units, priority, task_category
- Detect recurrence patterns (daily, weekly, monthly)
- Identify commutes and anchor tasks from context
- Handle multi-task inputs

---

### **Phase 2: API Development** (Week 3-4)

#### Module 03: Backend API Map
**File**: `03_backend_api_map.md`

**Deliverables**:
- HTTP server specifications
- REST API endpoint definitions
- Request/Response schemas
- Authentication flow (optional Phase 2)
- Rate limiting strategy

**Core Endpoints**:
- `POST /nudge` - Natural language input processing
- `GET /stats` - Retrieve analytics (daily/weekly/monthly)
- `PATCH /complete` - Update task completion
- `GET /tasks` - List active tasks
- `DELETE /tasks/{id}` - Remove tasks

---

### **Phase 3: Intelligence Layer** (Week 5-6)

#### Module 04: Analytics Engine
**File**: `04_analytics_engine.md`

**Deliverables**:
- Success percentage calculation logic
- Time-series aggregation functions
- Data grouping for visualizations
- Export formats (JSON, CSV)

**Visualizations**:
- Success rate pie charts (Completed/In Progress/Failed)
- Weekly trend lines
- Category breakdown (Work/Personal/Commute)

---

#### Module 05: AI Advisor Logic
**File**: `05_ai_advisor_logic.md`

**Deliverables**:
- Failure pattern recognition
- Schedule optimization algorithms
- Proactive suggestions engine
- Notification triggers

**Analysis Criteria**:
- Chronic task failure (< 40% success for 3+ days)
- Time estimation accuracy
- Commute delays pattern
- Overcommitment detection

---

## Development Workflow

### Stage 1: Documentation (Current Phase)
- [ ] Create all specification documents
- [ ] Review and validate with stakeholders
- [ ] Finalize data models

### Stage 2: Backend Development
1. Setup Go project structure with modules
2. Implement GORM models and auto-migration
3. Create database connection with environment variable support
4. Build HTTP server with Gin framework
5. Implement REST API handlers
6. Integrate LLM parsing layer
7. Develop analytics engine
8. Add middleware (CORS, rate limiting, auth)

**Initial Setup:**
1. Initialize Go module
2. Install required dependencies (database driver, HTTP framework, environment config)
3. Create .env file with DB_URL and DB_TOKEN
4. Run database migrations
5. Start HTTP server

### Stage 3: Testing & Validation
1. Unit tests for all API endpoints
2. LLM prompt testing with edge cases
3. Load testing for analytics queries
4. End-to-end workflow validation

### Stage 4: Dashboard Development
1. Design UI/UX mockups
2. Implement visualization components
3. Connect to backend APIs
4. User testing and feedback

### Stage 5: Deployment
1. Dockerize application
2. Setup CI/CD pipeline
3. Deploy to cloud infrastructure
4. Monitoring and logging setup

---

## Success Metrics for the Project

| Metric | Target | Measurement |
|--------|--------|-------------|
| **LLM Parse Accuracy** | > 95% | Correct JSON extraction rate |
| **API Response Time** | < 200ms | P95 latency |
| **Success Calc Accuracy** | 100% | Unit test coverage |
| **User Satisfaction** | > 4.5/5 | Beta user feedback |

---

## Risk Mitigation

| Risk | Impact | Mitigation Strategy |
|------|--------|---------------------|
| LLM hallucination | High | Strict JSON schema validation, fallback to clarification prompts |
| Database scaling | Medium | PostgreSQL from day one, optimized for production |
| Ambiguous user input | High | Implement multi-turn conversation for clarification |
| Poor success metric | Medium | A/B test different calculation formulas with users |

---

## File Dependencies

```
Summary_Roadmap.md (YOU ARE HERE)
    │
    ├── 01_data_schema.md
    │   └── Required by: 02, 03, 04, 05
    │
    ├── 02_llm_parsing_logic.md
    │   └── Required by: 03
    │
    ├── 03_backend_api_map.md
    │   └── Depends on: 01, 02
    │   └── Required by: 04, 05
    │
    ├── 04_analytics_engine.md
    │   └── Depends on: 01, 03
    │
    └── 05_ai_advisor_logic.md
        └── Depends on: 01, 03, 04
```

---

## Next Steps

1. **Review this roadmap** and validate the scope
2. **Proceed to Module 01** - Data Schema Design
3. **Establish naming conventions** and coding standards
4. **Setup development environment** (optional, if moving to code immediately)

---

## Notes for Coding Agent

When implementing this plan:
- Follow the dependency graph strictly
- Write tests alongside each module
- Document all API changes in OpenAPI format
- Use type hints throughout Python code
- Optimize database queries early (use EXPLAIN)
- Log all LLM interactions for debugging

---

**Document Status**: ✅ Complete  
**Last Updated**: February 18, 2026  
**Next Document**: `01_data_schema.md`
