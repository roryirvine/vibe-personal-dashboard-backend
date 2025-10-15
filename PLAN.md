# Implementation Plan

## Phase 1: Foundation

### Task 1: Initialize Go module and create project directory structure
- Run `go mod init` with appropriate module path
- Create directory structure:
  - `cmd/server/` - main.go
  - `internal/models/` - domain models
  - `internal/config/` - configuration parsing
  - `internal/repository/` - data access
  - `internal/service/` - business logic
  - `internal/api/handlers/` - HTTP handlers
  - `config/` - metrics.toml location
- Update `.gitignore` for binaries and data files

### Task 2: Implement models package
- Define `Metric` struct with fields: Name, Query, MultiRow, Params
- Define `ParamDefinition` struct with: Name, Type (enum: string/int/float), Required
- Define `MetricResult` struct for API responses
- Keep it simple - just the data structures, no logic

### Task 3: Implement config package
- Create config parser using `github.com/BurntSushi/toml`
- Validate metric definitions on load (no duplicate names, valid param types)
- Return structured errors for invalid config
- Expose `LoadConfig(path string) ([]models.Metric, error)` function

## Phase 2: Data Layer

### Task 4: Implement repository interface
- Define the Repository interface with QuerySingleValue, QueryMultiRow, Close
- Keep it in `internal/repository/repository.go`
- Add clear godoc comments explaining the interface contract

### Task 5: Implement SQLite repository
- Use `modernc.org/sqlite` driver
- Implement connection management with reasonable pool settings
- Implement QuerySingleValue - execute query, scan first column of first row
- Implement QueryMultiRow - use `sql.Rows`, get column names, build []map[string]interface{}
- Handle NULL values appropriately (return nil in Go)
- Add context timeout handling

## Phase 3: Business Logic

### Task 6: Implement service layer
- Create `MetricService` struct holding repository, config, logger
- Implement `GetMetrics(ctx, []string, map[string]string) ([]MetricResult, error)`
- Use `errgroup.WithContext` for concurrent query execution
- Implement parameter validation and type conversion
- Map params to SQL positional arguments in correct order
- Add timing logs for each query execution

## Phase 4: HTTP Layer

### Task 7: Implement HTTP handlers
- Use `github.com/go-chi/chi/v5` router
- Implement three handlers:
  - ListMetrics - return metric names as JSON array
  - GetSingleMetric - extract name from URL path, query params from request
  - GetMultipleMetrics - parse `names` query param (comma-separated)
- Add middleware for request logging (method, path, status, duration)
- Return proper error responses with JSON body

### Task 8: Implement main.go
- Read PORT (default 8080) and DB_PATH (default "./data.db") from env
- Set up slog with JSON handler
- Load metrics config from `./config/metrics.toml`
- Initialize SQLite repository and ping database
- Wire up service and handlers
- Start HTTP server with graceful shutdown (context cancel on SIGINT/SIGTERM)
- Log startup/shutdown events

## Phase 5: Testing & Documentation

### Task 9: Create example metrics.toml
- Include 3-4 example metrics showing different patterns:
  - Simple single-value query
  - Multi-row query
  - Parameterized query
  - Query with multiple params

### Task 10: Create example SQLite database
- Write a simple script or SQL file to create test tables
- Populate with sample data that the example metrics can query
- Document how to set it up

### Task 11: Write unit tests
- Test config parsing with valid and invalid TOML
- Test parameter validation in service layer
- Test repository query methods (may need in-memory SQLite)
- Focus on critical logic, not integration tests for now

### Task 12: End-to-end test
- Start the server with example config and database
- Hit all three endpoints and verify responses
- Test parameterized queries
- Test error cases (invalid metric name, missing required param)
- Verify concurrent execution improves response time

## Dependencies

Required Go packages:
- `github.com/BurntSushi/toml` - TOML parsing
- `github.com/go-chi/chi/v5` - HTTP router
- `modernc.org/sqlite` - SQLite driver (pure Go)
- `golang.org/x/sync/errgroup` - Concurrent execution

Standard library:
- `log/slog` - Structured logging
- `database/sql` - Database interface
- `net/http` - HTTP server
- `context` - Context management
- `os/signal` - Graceful shutdown

## Execution Order

1. Tasks 1-3 can proceed linearly (foundation and models)
2. Tasks 4-5 depend on Task 2 (models)
3. Task 6 depends on Tasks 4-5 (repository) and Task 2 (models)
4. Task 7 depends on Task 6 (service)
5. Task 8 depends on Tasks 3, 6, 7 (config, service, handlers)
6. Tasks 9-10 can be done anytime after Task 1
7. Task 11 can be done incrementally after each implementation task
8. Task 12 requires all previous tasks complete

## Success Criteria

- Service starts successfully and reads configuration
- Can list all available metrics
- Can fetch single metrics (with and without parameters)
- Can fetch multiple metrics concurrently
- Errors are handled gracefully with JSON responses
- Logs are structured and informative
- Graceful shutdown works correctly
- Tests pass for critical components
