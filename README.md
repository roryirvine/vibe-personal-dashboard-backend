# Vibe Personal Dashboard Backend

A lightweight, configuration-driven metrics API service built in Go. Reads metrics from SQLite and serves them as JSON via REST API. Perfect for dashboards and monitoring applications that need flexible, query-based metrics without complex infrastructure.

## Quick Start

### Prerequisites
- Go 1.21+
- sqlite3 (for manual database operations)

### Build and Run

```bash
# Build the server
go build -o bin/server ./cmd/server

# Set up test database with sample data
./scripts/setup_test_db.sh

# Run the server (default: listen on :8080, use ./data.db)
./bin/server

# Or with custom configuration
PORT=3000 DB_PATH=/var/data/metrics.db ./bin/server
```

### Quick Test

```bash
# List all metrics
curl http://localhost:8080/metrics

# Get a specific metric
curl http://localhost:8080/metrics/all_users

# Get multiple metrics
curl "http://localhost:8080/metrics?names=server_time,system_info"

# Get a parameterized metric
curl "http://localhost:8080/metrics/user_details?user_id=2"
```

## API Endpoints

### List All Metrics
**Request:**
```
GET /metrics
```

**Response:**
```json
["server_time", "system_info", "all_users", "user_details"]
```

### Get Single Metric
**Request:**
```
GET /metrics/{name}
```

**Example:**
```bash
curl http://localhost:8080/metrics/server_time
```

**Response:**
```json
[
  {
    "name": "server_time",
    "value": "2025-10-30 16:45:33"
  }
]
```

### Get Multiple Metrics
**Request:**
```
GET /metrics?names=metric1,metric2,metric3
```

**Example:**
```bash
curl "http://localhost:8080/metrics?names=server_time,system_info"
```

**Response:**
```json
[
  {
    "name": "server_time",
    "value": "2025-10-30 16:45:33"
  },
  {
    "name": "system_info",
    "value": "running"
  }
]
```

### Parameterized Metrics
Query parameters are passed to all requested metrics. Parameters must match the type defined in configuration.

**Example:**
```bash
curl "http://localhost:8080/metrics/user_details?user_id=2"
```

**Response:**
```json
[
  {
    "name": "user_details",
    "value": [
      {
        "id": 2,
        "name": "Bob Smith",
        "email": "bob@example.com"
      }
    ]
  }
]
```

## Example Metrics

The service includes four example metrics demonstrating different patterns:

### server_time
Single-value metric returning the current server datetime.
```bash
curl http://localhost:8080/metrics/server_time
```

### system_info
Single-value metric returning system status.
```bash
curl http://localhost:8080/metrics/system_info
```

### all_users
Multi-row metric returning all users from the database.
```bash
curl http://localhost:8080/metrics/all_users
```

Returns array of objects:
```json
[
  {
    "name": "all_users",
    "value": [
      {"id": 1, "name": "Alice Johnson", "email": "alice@example.com"},
      {"id": 2, "name": "Bob Smith", "email": "bob@example.com"},
      {"id": 3, "name": "Charlie Brown", "email": "charlie@example.com"},
      {"id": 4, "name": "Diana Prince", "email": "diana@example.com"},
      {"id": 5, "name": "Eve Wilson", "email": "eve@example.com"}
    ]
  }
]
```

### user_details
Multi-row metric with required integer parameter, demonstrating parameterized queries.
```bash
curl "http://localhost:8080/metrics/user_details?user_id=1"
```

Returns specific user:
```json
[
  {
    "name": "user_details",
    "value": [
      {"id": 1, "name": "Alice Johnson", "email": "alice@example.com"}
    ]
  }
]
```

## Configuration

### Environment Variables

**PORT** - HTTP server port (default: 8080)
```bash
PORT=3000 ./bin/server
```

**DB_PATH** - Path to SQLite database file (default: ./data.db)
```bash
DB_PATH=/var/data/metrics.db ./bin/server
```

### Metrics Configuration

Metrics are defined in `config/metrics.toml`. Each metric specifies:
- **name**: Unique identifier for the metric
- **query**: SQL query with positional placeholders (`?`)
- **multi_row**: Boolean (true = return array, false = return scalar)
- **params**: Optional array of parameter definitions

**Important**: All parameters must be marked as `required = true`. Optional parameters are not supported with positional SQL parameters because you cannot conditionally omit a `?` placeholder. If you need variations, create separate metrics:

```toml
# Instead of one metric with optional limit:
[[metrics]]
name = "users_all"
query = "SELECT id, name FROM users"
multi_row = true

[[metrics]]
name = "users_top_10"
query = "SELECT id, name FROM users LIMIT 10"
multi_row = true
```

### Log Level

The service uses structured JSON logging. To change the log level:

1. Edit `cmd/server/main.go` and find the `setupLogging()` function
2. Change the `slog.LevelInfo` to one of:
   - `slog.LevelDebug` - verbose output (debug, info, warn, error)
   - `slog.LevelInfo` - normal output (info, warn, error) - default
   - `slog.LevelWarn` - warnings only (warn, error)
   - `slog.LevelError` - errors only

```go
opts := &slog.HandlerOptions{Level: slog.LevelDebug}
```

## Development

### Build
```bash
go build -o bin/server ./cmd/server
```

### Run Tests
```bash
# All tests
go test ./...

# Verbose output
go test -v ./...

# With coverage
go test -cover ./...
```

### Run Server (Development)
```bash
go run ./cmd/server
```

### Set Up Test Database
```bash
./scripts/setup_test_db.sh
```

This script is idempotent - safe to run multiple times. It creates a fresh database with sample data each time.

## Architecture Overview

The service uses clean, layered architecture:

- **HTTP API** (`internal/api/handlers/`, `internal/api/router.go`): REST endpoints and middleware
- **Service** (`internal/service/`): Business logic, parameter validation, concurrent execution
- **Repository** (`internal/repository/`): Database abstraction layer
- **Models** (`internal/models/`): Data structures and validation
- **Configuration** (`internal/config/`): TOML file parsing

See `DESIGN.md` for complete architectural details.

## Important Limitations and Design Decisions

### Optional Parameters Not Supported
SQL uses positional placeholders (`?`), so all parameters defined in a metric's configuration MUST be provided in every request. Optional parameters would require dynamic query building, which introduces SQL injection risks.

**Workaround**: Create separate metrics for different variations rather than trying to make parameters optional.

### No Configuration Hot-Reload
Configuration changes (adding/modifying metrics) require restarting the service. This keeps the architecture simple and avoids subtle bugs from in-flight requests seeing stale configuration.

### No Caching
Query results are not cached. This simplifies the architecture and is acceptable for the low-volume use case (dashboard metrics queried once per minute). Cache can be added later if needed.

### Error Handling
Errors from any layer (parameter validation, database, configuration) result in a 500 response with a JSON error message. The error message includes full context through wrapped errors:
- Model layer: Base error (e.g., "invalid parameter type")
- Config layer: Adds metric name context
- Service layer: Adds operation context
- Handler layer: Returns as HTTP error

### Concurrent Execution
Multiple metrics requested via `?names=` are executed in parallel using goroutines. If any metric fails, the entire request fails (fail-fast). This means the client either gets all results or an error, never partial results.

## Troubleshooting

**"metric not found" (404)**
- Metric name doesn't exist in `config/metrics.toml`
- Check metric name spelling

**"invalid integer value" (400)**
- Parameter type doesn't match the configured type in metrics.toml
- Ensure `user_id=123` not `user_id=abc` for int parameters

**"sql: no rows in result set" (500)**
- Query executed but returned no results
- For multi-row metrics, zero results are allowed (returns empty array)
- For single-value metrics, at least one row is required

**"failed to connect to database"**
- Check DB_PATH points to a valid SQLite database
- Run `./scripts/setup_test_db.sh` to create sample database

## Files and Structure

```
.
├── README.md                      # This file
├── DESIGN.md                      # Architectural design document
├── cmd/
│   └── server/
│       └── main.go               # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   └── metrics.go        # HTTP handlers
│   │   └── router.go             # Route setup and middleware
│   ├── config/
│   │   └── config.go             # TOML configuration parsing
│   ├── models/
│   │   ├── metric.go             # Metric configuration struct
│   │   ├── param_definition.go   # Parameter definition struct
│   │   ├── param_type.go         # Parameter type enum
│   │   └── metric_result.go      # API response struct
│   ├── repository/
│   │   ├── repository.go         # Repository interface
│   │   └── sqlite.go             # SQLite implementation
│   └── service/
│       ├── metric_service.go     # Service orchestration
│       └── params.go             # Parameter conversion
├── config/
│   └── metrics.toml              # Metric definitions
├── scripts/
│   ├── setup_test_db.sh          # Database setup script
│   └── setup_test_db.sql         # Schema and sample data
├── go.mod                        # Go module definition
└── go.sum                        # Dependency checksums
```

## Next Steps

- **Extend metrics**: Add your own metrics to `config/metrics.toml`
- **Add caching**: Implement per-metric TTL caching in the service layer if performance requires it
- **Authentication**: Add middleware to `internal/api/router.go` if the service needs authentication
- **Testing**: Add integration tests for specific metric queries
- **Deployment**: Package as Docker container or deploy to your preferred platform

## License

This is a personal project. Feel free to use and modify as needed.