# Nudge API Documentation

## Getting Started

### Import Postman Collection
1. Open Postman
2. Click **Import** button
3. Select `Nudge.postman_collection.json`
4. Collection will appear in your workspace

The collection includes:
- `base_url` variable set to `http://localhost:8080`
- `auth_token` variable (empty by default - copy token from login response)

---

## Authentication

All task endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

**Steps:**
1. Register a user or login with existing credentials
2. Copy the `token` from the response
3. Set it in Postman's `auth_token` variable OR add it to the Authorization header

---

## Available Endpoints

### Health Check
**GET** `/health`

Simple endpoint to verify the API is running.

**Response:**
```json
{
  "status": "ok"
}
```

---

## Authentication Endpoints

### Register User
**POST** `/auth/register`

Creates a new user account.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Validation:**
- `email`: Valid email format (required)
- `password`: Minimum 6 characters (required)
- `first_name`: Required
- `last_name`: Required

**Success Response (201):**
```json
{
  "message": "User created successfully",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe"
  }
}
```

**Error Responses:**
- `400`: Invalid request body or validation error
- `409`: Email already registered

---

### Login
**POST** `/auth/login`

Authenticates a user and returns a JWT token.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Success Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe"
  }
}
```

**JWT Token Details:**
- Contains `user_id` in claims
- Expires in 24 hours
- Use in `Authorization` header as: `Bearer <token>`

**Error Responses:**
- `400`: Invalid request body
- `401`: Invalid email or password

---

## Task Endpoints

All task endpoints require authentication.

### Create Task
**POST** `/tasks`

Creates a new task for the authenticated user.

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "Complete project documentation",
  "task_category": "ACTION",
  "status": "CREATED",
  "expected_duration": 120,
  "expected_units": 5,
  "category": "Work",
  "notes": "Focus on API documentation",
  "deadline": "2026-03-15T17:00:00Z",
  "start_at": 540
}
```

**Field Descriptions:**
- `name`: Task name (required)
- `task_category`: One of `ACTION`, `ANCHOR`, `TRANSIT` (required)
- `status`: One of `CREATED`, `COMPLETED`, `FAILED`, `DEFERRED` (optional, defaults to `CREATED`)
- `expected_duration`: Minutes (1-1440) (optional)
- `expected_units`: Quantity (1-1000) (optional)
- `actual_duration`: Minutes (1-1440) (optional)
- `actual_units`: Quantity (0+) (optional)
- `category`: User-defined tag (optional)
- `notes`: Description or comments (optional)
- `deadline`: ISO 8601 datetime (optional)
- `start_at`: Minutes since midnight (0-1439), e.g., 540 = 9:00 AM (optional)

**Success Response (201):**
```json
{
  "message": "Task created successfully",
  "task": {
    "id": "task-uuid",
    "user_id": "user-uuid",
    "name": "Complete project documentation",
    "task_category": "ACTION",
    "status": "CREATED",
    "expected_duration": 120,
    "category": "Work",
    "notes": "Focus on API documentation",
    "start_at": 540,
    "created_at": "2026-03-12T10:30:00Z",
    "updated_at": "2026-03-12T10:30:00Z"
  }
}
```

**Error Responses:**
- `400`: Invalid request body or validation error
- `401`: Missing or invalid token

---

### Get All Tasks
**GET** `/tasks`

Retrieves all tasks for the authenticated user.

**Headers:**
```
Authorization: Bearer <token>
```

**Query Parameters (optional - can be combined):**
- `status`: Filter by task status (e.g., `CREATED`, `COMPLETED`, `FAILED`, `DEFERRED`)
- `task_category`: Filter by task category (e.g., `ACTION`, `ANCHOR`, `TRANSIT`)
- `category`: Filter by user-defined category (e.g., `Work`, `Health`, `Personal`)
- `search`: Search in task name (partial match, case-insensitive)

**Examples:**
```
GET /tasks                                          # Get all tasks
GET /tasks?status=CREATED                          # Filter by status
GET /tasks?task_category=ACTION                    # Filter by task category
GET /tasks?category=Work                           # Filter by user category
GET /tasks?search=documentation                    # Search in task name
GET /tasks?status=CREATED&category=Work            # Combine multiple filters
GET /tasks?task_category=ACTION&search=project     # Combine category and search
```

**Success Response (200):**
```json
{
  "tasks": [
    {
      "id": "task-uuid-1",
      "user_id": "user-uuid",
      "name": "Task 1",
      "task_category": "ACTION",
      "status": "CREATED",
      "created_at": "2026-03-12T10:30:00Z",
      "updated_at": "2026-03-12T10:30:00Z"
    },
    {
      "id": "task-uuid-2",
      "user_id": "user-uuid",
      "name": "Task 2",
      "task_category": "ANCHOR",
      "status": "COMPLETED",
      "completed_at": "2026-03-12T15:00:00Z",
      "created_at": "2026-03-12T09:00:00Z",
      "updated_at": "2026-03-12T15:00:00Z"
    }
  ],
  "count": 2
}
```

---

### Get Task by ID
**GET** `/tasks/:id`

Retrieves a specific task by ID for the authenticated user.

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "task": {
    "id": "task-uuid",
    "user_id": "user-uuid",
    "name": "Complete project documentation",
    "task_category": "ACTION",
    "status": "CREATED",
    "expected_duration": 120,
    "category": "Work",
    "created_at": "2026-03-12T10:30:00Z",
    "updated_at": "2026-03-12T10:30:00Z"
  }
}
```

**Error Responses:**
- `400`: Invalid task ID format
- `401`: Missing or invalid token
- `404`: Task not found or doesn't belong to user

---

### Update Task
**PUT** `/tasks/:id`

Updates a task for the authenticated user. All fields are optional.

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body (all fields optional):**
```json
{
  "name": "Updated task name",
  "status": "COMPLETED",
  "actual_duration": 115,
  "notes": "Completed ahead of schedule"
}
```

**Success Response (200):**
```json
{
  "message": "Task updated successfully",
  "task": {
    "id": "task-uuid",
    "user_id": "user-uuid",
    "name": "Updated task name",
    "task_category": "ACTION",
    "status": "COMPLETED",
    "actual_duration": 115,
    "completed_at": "2026-03-12T12:00:00Z",
    "created_at": "2026-03-12T10:30:00Z",
    "updated_at": "2026-03-12T12:00:00Z"
  }
}
```

**Special Behaviors:**
- Setting `status` to `COMPLETED` automatically sets `completed_at` timestamp
- Changing from `COMPLETED` to another status clears `completed_at`
- Set `deadline` to empty string `""` to clear it

**Error Responses:**
- `400`: Invalid request body or task ID format
- `401`: Missing or invalid token
- `404`: Task not found or doesn't belong to user

---

### Delete Task
**DELETE** `/tasks/:id`

Deletes a task for the authenticated user.

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "message": "Task deleted successfully"
}
```

**Error Responses:**
- `400`: Invalid task ID format
- `401`: Missing or invalid token
- `404`: Task not found or doesn't belong to user

---

## Testing with cURL

### Register:
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'
```

### Login:
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@mk.com",
    "password": "Testing123"
  }'
```

### Create Task:
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "name": "Morning workout",
    "task_category": "ACTION",
    "expected_duration": 45,
    "category": "Health",
    "start_at": 360
  }'
```

### Get All Tasks:
```bash
# Get all tasks
curl -X GET http://localhost:8080/tasks \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Filter by status
curl -X GET "http://localhost:8080/tasks?status=CREATED" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Filter by task category
curl -X GET "http://localhost:8080/tasks?task_category=ACTION" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Search in task name
curl -X GET "http://localhost:8080/tasks?search=documentation" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Combine multiple filters
curl -X GET "http://localhost:8080/tasks?status=CREATED&category=Work&task_category=ACTION" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### Update Task:
```bash
curl -X PUT http://localhost:8080/tasks/TASK_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "status": "COMPLETED",
    "actual_duration": 40
  }'
```

### Delete Task:
```bash
curl -X DELETE http://localhost:8080/tasks/TASK_ID \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

## Data Models

### Task Categories
- `ACTION`: Regular actionable task
- `ANCHOR`: Important fixed-time task
- `TRANSIT`: Transition or travel task

### Task Status
- `CREATED`: Task created but not started
- `COMPLETED`: Task successfully completed
- `FAILED`: Task attempted but failed
- `DEFERRED`: Task postponed

### StartAt Format
Minutes since midnight (0-1439):
- `0` = 12:00 AM (midnight)
- `360` = 6:00 AM
- `540` = 9:00 AM
- `720` = 12:00 PM (noon)
- `1200` = 8:00 PM
- `1439` = 11:59 PM

---

## Seed Data

The application includes a seeded user for testing:

**Email:** `admin@mk.com`  
**Password:** `Testing123`

This user is automatically created in development mode with 3 sample tasks.

---

## Environment Variables

Configure these in your `.env` file:

```env
# Database
DB_URL=file:local.db
DB_TOKEN=

# Server
PORT=8080
ENV=development

# JWT
JWT_SECRET=your-secret-key-change-in-production
```

---

## Running the Server

```bash
go run main.go
```

Server will start on port 8080 (or the port specified in your `.env` file).
