# Nudge

A task management and productivity tracking application.

## Project Structure

```
nudge/
├── config/              # Application configuration
│   └── config.go        # Environment variables and config loading
├── internal/
│   ├── database/        # Database layer
│   │   ├── database.go  # Connection logic (SQLite/Turso)
│   │   ├── migration.go # Schema migrations
│   │   └── seed.go      # Seed data for development
│   └── models/          # Database models (GORM)
│       └── models.go    # Model definitions
├── docs/                # Documentation
├── main.go              # Application entry point
├── go.mod               # Go module definition
├── .env                 # Environment variables (local)
├── .env.example         # Environment variables template
└── .gitignore           # Git ignore rules
```

## Setup

### Prerequisites

- Go 1.21 or higher
- SQLite (for local development)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd nudge
```

2. Copy the example environment file:
```bash
cp .env.example .env
```

3. Install dependencies:
```bash
go mod download
```

4. Run the application:
```bash
go run main.go
```

## Database Configuration

The application supports two database environments:

### Development (Local SQLite)

Uses a local SQLite database file. Configuration in `.env`:

```env
ENV=development
DB_URL=file:local.db
DB_TOKEN=
```

### Production (Turso)

Uses Turso (libSQL) cloud database. Configuration in `.env`:

```env
ENV=production
DB_URL=libsql://your-database.turso.io
DB_TOKEN=your-turso-auth-token
```

## Database Operations

The application automatically:
- Connects to the database based on environment variables
- Runs migrations to create/update schema
- Creates performance indexes
- Seeds sample data (development only)

### Manual Database Operations

When models are added, you can use the database package functions:

```go
// Run migrations
database.Migrate()

// Migrate specific models
database.MigrateModels(&models.Task{}, &models.RecurringTask{})

// Seed database
database.Seed()

// Clear all data
database.ClearData()
```

## Next Steps

1. **Add Models**: Create model files in `internal/models/` based on the data schema
2. **Update Migrations**: Add models to `Migrate()` function in `database/migration.go`
3. **Add Seed Data**: Implement seed data in `database/seed.go`
4. **Create API Handlers**: Build REST API endpoints for CRUD operations

## Development

```bash
# Run the application
go run main.go

# Build binary
go build -o nudge

# Run tests (when added)
go test ./...
```

## License

[Your License Here]
