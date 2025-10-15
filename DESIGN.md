# Vibe Dashboard Backend - Design Document

## Overview

This service provides a RESTful API that reads metrics from a database (initially SQLite) and returns them as JSON. It uses a configuration-driven approach to map metric names to SQL queries, allowing flexible metric definitions without code changes.

## Core Requirements

- Read metrics from SQLite database (initially)
- Serve metrics as JSON via REST API
- Configuration-driven metric definitions (TOML)
- Support both single-value and multi-row metrics
- Support parameterized queries with type validation
- Concurrent query execution for multiple metrics
- Simple error handling (500 with JSON error message)
- Structured logging using slog
- No authentication required (internal service)
- No caching (tolerate staleness, low query volumes)

## Architecture

### Overall Structure

The service uses a clean, layered architecture with three main components:

1. **Config Layer**: Reads TOML configuration at `./config/metrics.toml` on startup
2. **Repository Layer**: Interface-based data access abstraction
3. **HTTP Layer**: REST API using chi router

Dependency injection is used throughout - handlers depend on services, services depend on repositories.

### Configuration Format

Metrics are defined in `./config/metrics.toml`:

```toml
[[metrics]]
name = "active_users"
query = "SELECT COUNT(*) FROM users WHERE last_active > datetime('now', '-7 days')"
multi_row = false

[[metrics]]
name = "user_signups_by_day"
query = "SELECT date, count FROM signups WHERE date >= ?"
multi_row = true
params = [
  { name = "start_date", type = "string", required = true }
]

[[metrics]]
name = "revenue_total"
query = "SELECT SUM(amount) FROM transactions"
multi_row = false
```

Configuration fields:
- `name`: Unique metric identifier
- `query`: SQL query with positional placeholders (`?`)
- `multi_row`: Boolean flag (false = single value, true = array of rows)
- `params`: Optional array of parameter definitions
  - `name`: Parameter name (maps to URL query param)
  - `type`: Data type (`string`, `int`, `float`)
  - `required`: Boolean flag

### Repository Interface

The repository layer abstracts database operations:

```go
type Repository interface {
    QuerySingleValue(ctx context.Context, query string, args ...interface{}) (interface{}, error)
    QueryMultiRow(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)
    Close() error
}
```

**SQLite Implementation:**
- Uses `modernc.org/sqlite` (pure Go, no CGO)
- Connection pooling via `database/sql`
- Prepared statements for all queries
- Context-aware for timeouts and cancellation
- `QuerySingleValue`: Returns first column of first row
- `QueryMultiRow`: Returns all rows as `[]map[string]interface{}`
- Handles NULL values appropriately

### Service Layer

The `MetricService` sits between HTTP handlers and the repository:

```go
type MetricService struct {
    repo   Repository
    config []models.Metric
    logger *slog.Logger
}
```

**Key responsibilities:**
- Validate metric names against configuration
- Validate and convert URL parameters based on metric param definitions
- Execute queries concurrently using `golang.org/x/sync/errgroup`
- Return first error encountered (fail-fast)
- Log query execution timing

**Concurrent Execution:**
When multiple metrics are requested, the service uses `errgroup.WithContext` to:
- Execute all queries in parallel
- Cancel remaining queries if one fails
- Return aggregated results or first error

### HTTP API

**Endpoints:**

1. **List all metrics**: `GET /metrics`
   - Returns array of metric names
   - Example: `["active_users", "user_signups_by_day", "revenue_total"]`

2. **Single metric**: `GET /metrics/{name}`
   - Returns single metric with optional query parameters
   - Example: `GET /metrics/user_signups_by_day?start_date=2025-01-01`

3. **Multiple metrics**: `GET /metrics?names=metric1,metric2,metric3`
   - Comma-separated metric names in query parameter
   - Additional query params apply to all metrics
   - Example: `GET /metrics?names=active_users,revenue_total`

**Response Format:**

All metric responses use array-of-objects format:

```json
[
  {
    "name": "active_users",
    "value": 1523
  },
  {
    "name": "user_signups_by_day",
    "value": [
      {"date": "2025-01-01", "count": 45},
      {"date": "2025-01-02", "count": 52}
    ]
  }
]
```

- Single-value metrics: scalar `value`
- Multi-row metrics: array `value`

**Error Responses:**

Returns 500 Internal Server Error with JSON body:

```json
{
  "error": "metric user_signups_by_day failed: sql: no rows in result set"
}
```

### Logging

Uses `slog` with JSON formatting for structured logging:

**Logged events:**
- Startup/shutdown events
- HTTP requests (method, path, status, duration)
- Individual metric query executions with timing
- Errors with full context

### Configuration

**Environment Variables:**
- `PORT`: HTTP server port (default: 8080)
- `DB_PATH`: SQLite database path (default: "./data.db")

**Conventional Paths:**
- Metrics config: `./config/metrics.toml` (fixed location)

**Restart Required:**
- Configuration changes require service restart
- No hot-reload or dynamic configuration

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── models/
│   │   └── metric.go            # Metric, ParamDefinition, MetricResult
│   ├── config/
│   │   └── config.go            # TOML parsing and validation
│   ├── repository/
│   │   ├── repository.go        # Repository interface
│   │   └── sqlite.go            # SQLite implementation
│   ├── service/
│   │   └── metric_service.go    # Business logic and concurrent execution
│   └── api/
│       └── handlers/
│           └── metrics.go       # HTTP handlers
├── config/
│   └── metrics.toml             # Metric definitions
├── go.mod
└── go.sum
```

## Initialization Flow

1. Read `PORT` and `DB_PATH` from environment
2. Parse `./config/metrics.toml` and validate metric definitions
3. Open SQLite connection and verify with ping
4. Wire up dependencies: repository → service → handlers
5. Set up chi router with handlers
6. Start HTTP server
7. Listen for SIGINT/SIGTERM for graceful shutdown

## Design Decisions

**Why TOML?**
- Clean syntax for multi-line SQL queries
- Popular in Go ecosystem
- Good balance of human-readable and structured

**Why no caching?**
- YAGNI principle - query volumes are low
- Can add later if needed
- Simpler architecture, fewer moving parts

**Why interface-based repository?**
- Enables swapping database implementations
- Makes testing easier (mock repositories)
- Decouples business logic from data access

**Why errgroup for concurrency?**
- Built-in context cancellation
- Fail-fast semantics
- Simple API for parallel execution
- Part of official Go extended library

**Why chi router?**
- Lightweight and idiomatic
- Good middleware support
- Not as heavyweight as gin
- Better ergonomics than stdlib ServeMux

**Why no authentication?**
- Internal service assumption
- Keeps architecture simple
- Can add middleware later if needed

## Future Considerations

**Potential enhancements (not implementing now):**
- Caching layer with per-metric TTL configuration
- Hot-reload via HTTP endpoint or file watch
- Authentication/authorization middleware
- Metrics endpoint with Prometheus format
- Query result pagination for large multi-row results
- Database connection pooling tuning
- Request rate limiting
- Support for multiple databases (config per metric)
