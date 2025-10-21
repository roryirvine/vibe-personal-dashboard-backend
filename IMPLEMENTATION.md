# Detailed Implementation Guide

## Overview

This guide walks you through implementing a metrics API service in Go from scratch. Follow each task in order. **Make a git commit after completing each task** - this allows you to track progress and revert if needed.

## Development Principles

**TDD (Test-Driven Development)**: Write a failing test first, then write just enough code to make it pass. This ensures your code is testable and does what you expect.

**YAGNI (You Aren't Gonna Need It)**: Only implement what's required now. Don't add features "just in case" - we can always add them later.

**KISS (Keep It Simple, Stupid)**: Choose the simplest solution that works. Avoid clever abstractions or complex patterns unless absolutely necessary.

**Commit Frequently**: After each task or logical unit of work, commit your changes. Small commits are easier to review and debug.

## Prerequisites

- Go 1.21+ installed
- Basic understanding of Go syntax
- Familiarity with HTTP/REST concepts
- Access to this repository

---

## Phase 1: Foundation

### Task 1.1: Initialize Go Module

**What you're doing**: Creating a Go module so we can manage dependencies.

**Files to create/modify**:
- `go.mod` (will be created)

**Commands to run**:
```bash
cd /workspaces/vibe-personal-dashboard-backend
go mod init github.com/roryirvine/vibe-personal-dashboard-backend
```

**How to verify**:
```bash
cat go.mod
```

You should see a file that starts with:
```
module github.com/roryirvine/vibe-personal-dashboard-backend

go 1.2
```

**Commit**:
```bash
git add go.mod
git commit -m "Initialize Go module"
```

---

### Task 1.2: Create Directory Structure

**What you're doing**: Setting up the project layout following Go conventions.

**Directories to create**:
```bash
mkdir -p cmd/server
mkdir -p internal/models
mkdir -p internal/config
mkdir -p internal/repository
mkdir -p internal/service
mkdir -p internal/api/handlers
mkdir -p config
```

**Why this structure**:
- `cmd/server/`: Contains `main.go` - the application entry point
- `internal/`: Private application code that can't be imported by other projects
- `internal/models/`: Data structures (structs) used throughout the app
- `internal/config/`: Configuration file parsing logic
- `internal/repository/`: Database access layer
- `internal/service/`: Business logic
- `internal/api/handlers/`: HTTP request handlers
- `config/`: Configuration files (not code)

**How to verify**:
```bash
tree -L 3 -d
```

**Commit**:
```bash
git add -A
git commit -m "Create project directory structure"
```

---

### Task 1.3: Update .gitignore

**What you're doing**: Telling git to ignore generated files and databases.

**File to modify**: `.gitignore`

**Add these lines**:
```
# Binaries
/bin/
/server
*.exe

# Databases
*.db
*.sqlite
*.sqlite3

# Test coverage
coverage.out
coverage.html

# IDE
.vscode/
.idea/

# OS
.DS_Store
```

**Why**: We don't want to commit binaries, databases with test data, or IDE-specific files.

**Commit**:
```bash
git add .gitignore
git commit -m "Update .gitignore for Go project artifacts"
```

---

## Phase 2: Models Package (Data Structures)

### Task 2.1: Define ParamType Enum

**What you're doing**: Creating a type-safe enum for parameter types (string, int, float).

**File to create**: `internal/models/param_type.go`

**Code**:
```go
// Defines parameter types supported in metric queries.
package models

type ParamType string

const (
	ParamTypeString ParamType = "string"
	ParamTypeInt    ParamType = "int"
	ParamTypeFloat  ParamType = "float"
)

func (pt ParamType) IsValid() bool {
	switch pt {
	case ParamTypeString, ParamTypeInt, ParamTypeFloat:
		return true
	}
	return false
}
```

**Why this design**:
- Using a custom type (`ParamType`) instead of plain strings provides type safety
- Constants prevent typos
- `IsValid()` method allows validation

**How to test**: We'll write a test file.

**File to create**: `internal/models/param_type_test.go`

**Code**:
```go
package models

import "testing"

func TestParamType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		paramType ParamType
		want     bool
	}{
		{"string is valid", ParamTypeString, true},
		{"int is valid", ParamTypeInt, true},
		{"float is valid", ParamTypeFloat, true},
		{"invalid type", ParamType("boolean"), false},
		{"empty type", ParamType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.paramType.IsValid(); got != tt.want {
				t.Errorf("ParamType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

**How to run the test**:
```bash
go test ./internal/models/
```

You should see:
```
ok      github.com/roryirvine/vibe-personal-dashboard-backend/internal/models
```

**Commit**:
```bash
git add internal/models/
git commit -m "Add ParamType enum with validation"
```

---

### Task 2.2: Define ParamDefinition Struct

**What you're doing**: Creating a struct to represent a parameter definition from the config file.

**File to create**: `internal/models/param_definition.go`

**Code**:
```go
package models

type ParamDefinition struct {
	Name     string    `toml:"name"`
	Type     ParamType `toml:"type"`
	Required bool      `toml:"required"`
}

func (pd ParamDefinition) Validate() error {
	if pd.Name == "" {
		return ErrParamNameEmpty
	}
	if !pd.Type.IsValid() {
		return ErrInvalidParamType
	}
	return nil
}
```

**File to modify**: `internal/models/param_definition.go` (add errors at the top)

**Add these error definitions**:
```go
// Defines parameter definitions for metric queries with validation.
package models

import "errors"

var (
	ErrParamNameEmpty    = errors.New("parameter name cannot be empty")
	ErrInvalidParamType  = errors.New("parameter type must be string, int, or float")
)

```

**File to create**: `internal/models/param_definition_test.go`

**Code**:
```go
package models

import (
	"testing"
)

func TestParamDefinition_Validate(t *testing.T) {
	tests := []struct {
		name    string
		param   ParamDefinition
		wantErr error
	}{
		{
			name: "valid required string param",
			param: ParamDefinition{
				Name:     "user_id",
				Type:     ParamTypeString,
				Required: true,
			},
			wantErr: nil,
		},
		{
			name: "valid optional int param",
			param: ParamDefinition{
				Name:     "limit",
				Type:     ParamTypeInt,
				Required: false,
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			param: ParamDefinition{
				Name:     "",
				Type:     ParamTypeString,
				Required: true,
			},
			wantErr: ErrParamNameEmpty,
		},
		{
			name: "invalid type",
			param: ParamDefinition{
				Name:     "test",
				Type:     ParamType("boolean"),
				Required: true,
			},
			wantErr: ErrInvalidParamType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.param.Validate()
			if err != tt.wantErr {
				t.Errorf("ParamDefinition.Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
```

**Run tests**:
```bash
go test ./internal/models/
```

**Commit**:
```bash
git add internal/models/
git commit -m "Add ParamDefinition with validation"
```

---

### Task 2.3: Define Metric Struct

**What you're doing**: Creating the main struct that represents a metric configuration.

**File to create**: `internal/models/metric.go`

**Code**:
```go
// Defines metric configuration structure with query and parameter definitions.
package models

type Metric struct {
	Name     string            `toml:"name"`
	Query    string            `toml:"query"`
	MultiRow bool              `toml:"multi_row"`
	Params   []ParamDefinition `toml:"params"`
}

func (m Metric) Validate() error {
	if m.Name == "" {
		return ErrMetricNameEmpty
	}
	if m.Query == "" {
		return ErrMetricQueryEmpty
	}

	// Validate all parameter definitions
	for _, param := range m.Params {
		if err := param.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (m Metric) GetParamByName(name string) (ParamDefinition, bool) {
	for _, param := range m.Params {
		if param.Name == name {
			return param, true
		}
	}
	return ParamDefinition{}, false
}
```

**Add these errors** to the top of the file:
```go
var (
	ErrMetricNameEmpty  = errors.New("metric name cannot be empty")
	ErrMetricQueryEmpty = errors.New("metric query cannot be empty")
)
```

**File to create**: `internal/models/metric_test.go`

**Code**:
```go
package models

import "testing"

func TestMetric_Validate(t *testing.T) {
	tests := []struct {
		name    string
		metric  Metric
		wantErr error
	}{
		{
			name: "valid metric without params",
			metric: Metric{
				Name:     "active_users",
				Query:    "SELECT COUNT(*) FROM users",
				MultiRow: false,
				Params:   nil,
			},
			wantErr: nil,
		},
		{
			name: "valid metric with params",
			metric: Metric{
				Name:     "users_by_date",
				Query:    "SELECT * FROM users WHERE created > ?",
				MultiRow: true,
				Params: []ParamDefinition{
					{Name: "start_date", Type: ParamTypeString, Required: true},
				},
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			metric: Metric{
				Name:  "",
				Query: "SELECT 1",
			},
			wantErr: ErrMetricNameEmpty,
		},
		{
			name: "empty query",
			metric: Metric{
				Name:  "test",
				Query: "",
			},
			wantErr: ErrMetricQueryEmpty,
		},
		{
			name: "invalid param",
			metric: Metric{
				Name:  "test",
				Query: "SELECT 1",
				Params: []ParamDefinition{
					{Name: "", Type: ParamTypeString, Required: true},
				},
			},
			wantErr: ErrParamNameEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metric.Validate()
			if err != tt.wantErr {
				t.Errorf("Metric.Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetric_GetParamByName(t *testing.T) {
	metric := Metric{
		Name:  "test",
		Query: "SELECT * FROM users WHERE id = ? AND status = ?",
		Params: []ParamDefinition{
			{Name: "user_id", Type: ParamTypeInt, Required: true},
			{Name: "status", Type: ParamTypeString, Required: false},
		},
	}

	t.Run("existing param", func(t *testing.T) {
		param, found := metric.GetParamByName("user_id")
		if !found {
			t.Error("expected to find user_id param")
		}
		if param.Name != "user_id" || param.Type != ParamTypeInt {
			t.Errorf("got param %+v, want user_id int param", param)
		}
	})

	t.Run("non-existing param", func(t *testing.T) {
		_, found := metric.GetParamByName("nonexistent")
		if found {
			t.Error("expected not to find nonexistent param")
		}
	})
}
```

**Run tests**:
```bash
go test ./internal/models/
```

**Commit**:
```bash
git add internal/models/
git commit -m "Add Metric struct with validation and param lookup"
```

---

### Task 2.4: Define MetricResult Struct

**What you're doing**: Creating the struct for API responses.

**File to create**: `internal/models/metric_result.go`

**Code**:
```go
// Defines the API response structure for metric results.
package models

type MetricResult struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}
```

**Why `interface{}`**: The value can be a single scalar (int, float, string) for single-row metrics, or a slice of maps for multi-row metrics. Using `interface{}` gives us that flexibility.

**No test needed**: This is a simple data structure with no logic.

**Commit**:
```bash
git add internal/models/metric_result.go
git commit -m "Add MetricResult struct for API responses"
```

---

## Phase 3: Config Package

The config package is responsible for reading the metrics.toml file and transforming it into validated Go data structures that the rest of the application can use. This separation keeps file parsing concerns away from business logic.

### Task 3.1: Install TOML Parser Dependency

**Why we need this**: The metrics.toml file uses TOML format because it's more human-readable than JSON for configuration files, especially when defining arrays of objects with parameters. The BurntSushi/toml library is the most mature and widely-used TOML parser in Go.

**What you're achieving**: Adding a third-party dependency that can deserialize TOML text into Go structs.

**Command**:
```bash
go get github.com/BurntSushi/toml
```

This updates `go.mod` and `go.sum` to track the dependency. The go.mod file is Go's way of managing dependencies - similar to package.json in Node or requirements.txt in Python.

**Commit after verifying the dependency appears in go.mod**

---

### Task 3.2: Create Config Struct and Parser (TDD)

**The problem we're solving**: We need to:
1. Read an external TOML file from disk
2. Parse it into our models.Metric structs (with all their nested ParamDefinitions)
3. Validate that the configuration makes sense (no duplicate names, all required fields present)
4. Return a usable slice of Metric objects

**Why validation at load time**: We want to fail fast during application startup if the configuration is invalid, rather than discovering problems when a user makes an API request. This is better for operations - the server won't start with a bad config.

**Architecture decisions**:

1. **Single responsibility**: The config package's only job is to load and validate the TOML file. It doesn't execute queries or handle HTTP requests.

2. **Validation layering**: We already have `Validate()` methods on our models (Metric, ParamDefinition). The config loader calls these, but also adds config-level validation like checking for duplicate metric names. This is different from model validation because it's checking relationships between metrics, not individual metric validity.

3. **Error handling philosophy**: Use Go's error wrapping (`fmt.Errorf("context: %w", err)`) to preserve the underlying error while adding context. This makes debugging easier - you'll see both "failed to parse config file" and the specific TOML syntax error.

**Design approach (following TDD)**:

**Step 1: Write tests that describe the behavior you need**

Think about what cases matter:
- **Happy path**: A valid TOML file with multiple metrics, some with parameters
- **File not found**: What happens if the path is wrong?
- **Malformed TOML**: What if the file has syntax errors?
- **Duplicate names**: What if two metrics have the same name? (This would break our API routing)
- **Empty config**: Is a config file with no metrics valid? (Probably not - why run the server?)

Write test cases that verify each scenario. For the happy path, use `t.TempDir()` and `os.WriteFile()` to create an actual TOML file on disk during the test. This tests the real file I/O, not just mocked behavior.

**Step 2: Implement the minimum code to pass tests**

You'll need:

1. **A Config struct**: This is a wrapper around `[]models.Metric` with TOML struct tags. The TOML library needs these tags to know how to map the TOML structure to Go fields.

   ```
   [[metrics]]      <-- This TOML array syntax
   name = "foo"
   ```

   needs to map to a Go field with tag `toml:"metrics"`

2. **A LoadConfig function**:
   - Takes a file path string
   - Uses `toml.DecodeFile()` to parse the file into the Config struct
   - Calls a validation function
   - Returns `[]models.Metric` and an error

3. **A validateMetrics helper function**:
   - Checks that at least one metric exists
   - Uses a `map[string]bool` to detect duplicate names as it iterates
   - Calls `Validate()` on each metric to ensure individual metrics are valid
   - Returns descriptive errors

**Key Go patterns to use**:
- Return errors, don't panic - let the caller (main.go) decide whether to exit
- Use `fmt.Errorf()` with `%w` verb to wrap errors while preserving the original
- Validate early and fail fast - do all validation in LoadConfig before returning

**Testing strategy**:
Run tests first (they'll fail - that's expected in TDD). Then implement. Then run tests again until they pass. Don't write more code than needed to pass the tests.

**What success looks like**: When tests pass, you have a LoadConfig function that can read a TOML file, deserialize it into your domain models, validate the configuration, and return clean data or descriptive errors.

**Commit**: Once tests pass and you've verified the implementation handles all test cases correctly.

---

## Phase 4: Repository Layer

The repository layer abstracts database access, isolating SQL and database-specific code from business logic. This layer provides a clean interface for the service layer to query metrics without knowing implementation details.

### Task 4.1: Define Repository Interface

**The problem we're solving**: The service layer needs to execute two types of queries:
1. **Single-value queries**: Return one scalar (e.g., `SELECT COUNT(*)`)
2. **Multi-row queries**: Return multiple rows as structured data (e.g., `SELECT id, name FROM users`)

But the service layer shouldn't need to know about database drivers, connection pooling, SQL types, or NULL handling. That's infrastructure concern, not business logic.

**Why use an interface**:
- **Testability**: The service layer can use a mock repository in tests without touching a real database
- **Flexibility**: We can swap SQLite for PostgreSQL later without changing service code
- **Dependency inversion**: High-level code (services) doesn't depend on low-level code (SQL), both depend on the interface

**Design decisions**:

1. **Context for cancellation**: All query methods accept `context.Context` so requests can be cancelled if the client disconnects or times out

2. **Variadic args for parameters**: `args ...interface{}` allows passing any number of query parameters with proper SQL injection prevention

3. **Return types**:
   - `QuerySingleValue` returns `interface{}` because the value could be int64, float64, string, or nil
   - `QueryMultiRow` returns `[]map[string]interface{}` where each map is one row with column names as keys
   - These generic types let us handle any query without knowing the schema ahead of time

4. **Error handling**: Return errors rather than panicking - let callers decide how to handle failures

**What to implement**:

Create `internal/repository/repository.go` with:
- A `Repository` interface defining three methods: `QuerySingleValue`, `QueryMultiRow`, and `Close`
- Document what each method does and when it returns errors
- Use `context.Context` as the first parameter (Go convention)

**Commit** after creating the interface.

---

### Task 4.2: Install SQLite Driver

**Why SQLite**:
- **Zero configuration**: No separate database server to install or manage
- **Perfect for local development**: The entire database is a single file
- **Good enough for production**: Handles thousands of requests per second for read-heavy workloads like dashboards
- **ACID compliant**: Full transaction support despite being "lite"

**Why modernc.org/sqlite specifically**:
- **Pure Go implementation**: No CGO dependency, which means:
  - Easier cross-compilation (build for any OS from any OS)
  - Simpler deployment (no native library dependencies)
  - Better debugging (no C code)
- **Drop-in replacement** for mattn/go-sqlite3 (the CGO version)
- **Active maintenance** and good performance

**Command**:
```bash
go get modernc.org/sqlite
```

This is a direct dependency (unlike TOML which we imported first), so `go mod tidy` shouldn't be needed.

**Commit** after verifying it appears in go.mod.

---

### Task 4.3: Implement SQLite Repository (TDD)

**The problem we're solving**: We need concrete implementations of `QuerySingleValue` and `QueryMultiRow` that:
1. Handle Go's `database/sql` package correctly (it's verbose and easy to get wrong)
2. Support different SQL types (INTEGER, TEXT, REAL, NULL)
3. Convert `sql.Rows` into our generic `[]map[string]interface{}` format
4. Respect context cancellation
5. Handle edge cases (no rows, NULL values, connection errors)

**Why this is complex**: Go's `database/sql` package is low-level. For `QueryMultiRow`, we need to:
- Get column names from the result set
- Create interface{} pointers for scanning (you can't scan into interface{} directly)
- Build maps dynamically for each row
- Handle NULL properly (it becomes nil in Go)

**Architecture decisions**:

1. **In-memory testing**: Use `:memory:` as the database path for tests. This gives us:
   - Real SQL execution (not mocked)
   - Fast tests (no disk I/O)
   - Isolated tests (each test gets a fresh database)
   - No test fixtures to maintain

2. **Connection pooling**: Configure `SetMaxOpenConns` and `SetMaxIdleConns` for production use

3. **Error wrapping**: Wrap all errors with context about what operation failed

4. **Blank import for driver**: Use `_ "modernc.org/sqlite"` to register the driver without directly using the package

**Testing strategy** (following TDD):

**What scenarios to test**:

For **QuerySingleValue**:
- Returns integer (COUNT queries)
- Returns float (SUM of REAL columns)
- Returns string (SELECT name queries)
- Works with WHERE clauses and parameters
- Returns error when no rows found (this is an error condition for single-value queries)
- Respects context timeout

For **QueryMultiRow**:
- Returns multiple rows with correct structure
- Each row is a `map[string]interface{}` with column names as keys
- Handles NULL values (they become nil)
- Works with WHERE clauses and parameters
- Returns empty slice (not error) for zero results
- Column types are preserved (int64, float64, string)

**Setup pattern**:
Create a helper function `setupTestDB(t *testing.T)` that:
1. Creates an in-memory repository
2. Creates a test table with various column types
3. Inserts sample data including NULL values
4. Returns the repository for use in tests

**Implementation approach**:

Create `internal/repository/sqlite.go` with:

1. **SQLiteRepository struct**:
   - Holds a `*sql.DB` connection
   - Implements the Repository interface

2. **NewSQLiteRepository constructor**:
   - Takes a path (file or ":memory:")
   - Opens database with `sql.Open("sqlite", path)`
   - Configures connection pool
   - Pings to verify connection
   - Returns error if connection fails

3. **QuerySingleValue implementation**:
   - Use `db.QueryRowContext()` (returns single row)
   - Scan into `interface{}`
   - Check for `sql.ErrNoRows` specifically (this should be an error)
   - Return the scanned value

4. **QueryMultiRow implementation**:
   - Use `db.QueryContext()` (returns multiple rows)
   - Call `rows.Columns()` to get column names
   - For each row:
     - Create slice of `interface{}` values
     - Create slice of pointers to those values (for Scan)
     - Call `rows.Scan(valuePtrs...)`
     - Build `map[string]interface{}` pairing columns with values
   - Check `rows.Err()` after iteration
   - Return slice of maps

5. **Close implementation**:
   - Simply call `db.Close()`

**Key Go patterns**:
- Use `defer rows.Close()` immediately after getting rows
- The double-indirection pattern for scanning: `valuePtrs[i] = &values[i]`
- Check `rows.Err()` after the loop (errors can occur during iteration)
- Use context-aware methods (`QueryRowContext`, `QueryContext`)

**What success looks like**: Tests pass, and you've verified:
- Different SQL types work correctly
- NULL handling works
- Parameter passing works
- Error cases are handled
- Context cancellation works

**Commit** after all tests pass.

---

## Phase 5: Service Layer

The service layer contains business logic that orchestrates between repositories and HTTP handlers. It validates parameters, executes queries, and handles concurrent execution of multiple metrics.

### Task 5.1: Install errgroup Dependency

**Why we need this**: The `errgroup` package provides coordinated error handling for goroutines. When fetching multiple metrics concurrently, we need to wait for all goroutines to complete and collect any errors that occur.

**What you're achieving**: Adding golang.org/x/sync/errgroup for concurrent metric execution.

**Command**:
```bash
go get golang.org/x/sync/errgroup
```

**Commit** after verifying dependency appears in go.mod.

---

### Task 5.2: Implement Parameter Conversion Helper (TDD)

**The problem we're solving**: HTTP query parameters arrive as strings (e.g., "?min_age=25"), but our database queries need typed values (int64, float64, string). We need type-safe conversion that fails fast with clear errors when the conversion isn't valid.

**Why this matters**: Without proper validation, "?min_age=abc" would silently fail or cause runtime panics. We want to return a 400 Bad Request with a clear error message instead.

**Architecture decisions**:

1. **Type safety**: The `convertParamValue` function takes a `ParamType` enum (not a string) to ensure only valid types are converted

2. **Error wrapping**: Use `fmt.Errorf` with `%w` to preserve the underlying strconv error while adding context about which parameter failed

3. **Return interface{}**: Since the caller doesn't know the type at compile time, we return `interface{}` but document the possible return types (int64, float64, string)

**Testing strategy** (following TDD):

Write tests covering:
- Valid conversions for each type (string, int, float)
- Edge cases (negative numbers, decimals)
- Invalid conversions (letters when expecting numbers)
- Boundary values (very large numbers, special characters)

Create `internal/service/params_test.go` with test cases for each scenario. Then implement `internal/service/params.go` with a `convertParamValue` function that uses Go's `strconv` package.

**What success looks like**: Tests pass, and the implementation correctly converts strings to typed values or returns descriptive errors.

**Commit** after tests pass.

---

### Task 5.3: Implement MetricService (TDD)

**The problem we're solving**: The HTTP layer needs a service that can:
1. Return a list of available metric names
2. Execute a single metric query with parameter validation
3. Execute multiple metrics concurrently and collect results
4. Handle errors from any layer (parameter validation, database queries)

**Why concurrent execution**: Dashboard UIs often request multiple metrics at once. Sequential execution would mean total latency = sum of all query times. Concurrent execution with `errgroup` means total latency ≈ slowest query time.

**Architecture decisions**:

1. **Map-based metric lookup**: Store metrics in a `map[string]models.Metric` for O(1) lookup by name instead of linear search through a slice

2. **Interface-based repository**: The service depends on a `repository.Repository` interface, not a concrete SQLite implementation. This makes testing trivial - we can use a mock repository without touching a database.

3. **Context propagation**: All methods accept `context.Context` as the first parameter (Go convention). This allows:
   - Request cancellation if the HTTP client disconnects
   - Timeout enforcement
   - Trace ID propagation (for distributed tracing)

4. **Parameter preparation**: Extract parameter validation and conversion into a separate `prepareParams` method. This method:
   - Checks that required parameters are present
   - Converts string values to the correct types
   - Returns a `[]interface{}` that can be passed directly to the repository's query methods

5. **Concurrent execution with errgroup**: For `GetMetrics`, use `errgroup.WithContext` which:
   - Runs each query in its own goroutine
   - Cancels all remaining queries if any query fails (fail-fast)
   - Collects the first error encountered
   - Waits for all goroutines to complete before returning

**Testing strategy** (following TDD):

**Step 1: Write tests**

Create `internal/service/metric_service_test.go` with:

- **Mock repository**: A test double that implements `repository.Repository` with configurable behavior
- **Test constructor**: Verify `NewMetricService` builds the metric map correctly
- **Test GetMetricNames**: Verify it returns all metric names
- **Test GetMetric (single value)**: Verify it calls the repository correctly for non-multi-row metrics
- **Test GetMetric (multi row)**: Verify it handles multi-row results
- **Test GetMetric (with parameters)**: Verify parameters are converted and passed correctly
- **Test GetMetric (missing required param)**: Verify validation errors are returned
- **Test GetMetric (not found)**: Verify error when metric name doesn't exist
- **Test GetMetrics (concurrent)**: Verify multiple metrics are fetched
- **Test GetMetrics (error handling)**: Verify that if one metric fails, the whole batch fails

**Step 2: Implement**

Create `internal/service/metric_service.go` with:

1. **MetricService struct**: Holds repository, metrics map, and logger
2. **NewMetricService**: Builds the map from the slice
3. **GetMetricNames**: Iterates the map and returns keys
4. **GetMetric**: Looks up metric, validates/converts params, calls appropriate repository method
5. **GetMetrics**: Uses errgroup to execute queries concurrently
6. **prepareParams helper**: Validates presence of required params and converts types

**Key Go patterns**:
- Use `i, name := i, name` to capture loop variables in goroutines (Go 1.21 and earlier)
- Check the `exists` boolean from map lookups
- Use `make([]interface{}, len(metric.Params))` to pre-allocate the args slice

**What success looks like**: All tests pass, including the concurrent execution test that verifies multiple metrics are fetched in parallel.

**Commit** after all tests pass.

---

## Phase 6: HTTP API Layer

The HTTP layer translates between HTTP requests/responses and the service layer. It handles routing, parameter extraction, JSON serialization, and error responses.

### Task 6.1: Install Chi Router

**Why Chi**: Chi is a lightweight, idiomatic HTTP router for Go that:
- Uses only standard library types (`http.Handler`, `http.ResponseWriter`)
- Supports middleware composition
- Provides URL parameter extraction (`/metrics/{name}`)
- Has no external dependencies beyond the standard library

**Command**:
```bash
go get github.com/go-chi/chi/v5
```

**Commit** after verifying dependency appears in go.mod.

---

### Task 6.2: Implement HTTP Handlers (TDD)

**The problem we're solving**: We need HTTP handlers that:
1. Extract URL parameters and query strings
2. Call the service layer with proper context
3. Serialize results to JSON
4. Return appropriate HTTP status codes and error responses

**Architecture decisions**:

1. **Handler struct with dependencies**: Create a `MetricsHandler` struct that holds a `MetricService` interface (not the concrete type) and a logger. This allows testing with a mock service.

2. **Interface for the service**: Define a `MetricService` interface in the handlers package that declares only the methods the handlers need. This is the Interface Segregation Principle - handlers shouldn't know about internal service details.

3. **Consistent error responses**: All errors return JSON in the format `{"error": "message"}`. This makes client-side error handling predictable.

4. **Route design**:
   - `GET /metrics` - Returns list of metric names (or multiple metrics if `?names=` param exists)
   - `GET /metrics/{name}` - Returns a single metric result
   - Both endpoints extract query parameters and pass them to the service for validation

**Testing strategy** (following TDD):

Create `internal/api/handlers/metrics_test.go` with:

- **Mock service**: Implements the `MetricService` interface with configurable responses
- **httptest for HTTP testing**: Use `httptest.NewRequest` and `httptest.NewRecorder` to test handlers without starting a real server
- **Chi context for URL params**: Use `chi.NewRouteContext()` to simulate URL parameter extraction
- **JSON decoding assertions**: Decode the response body and verify the structure

Test cases:
- ListMetrics returns JSON array of names
- GetSingleMetric extracts name from URL and returns result
- GetMultipleMetrics parses comma-separated names
- Error cases return proper status codes and error JSON

**Implementation**:

Create `internal/api/handlers/metrics.go` with:

1. **MetricService interface**: Declares `GetMetricNames`, `GetMetric`, `GetMetrics`
2. **MetricsHandler struct**: Holds service and logger
3. **List handler**: Calls service.GetMetricNames() and returns JSON
4. **Single metric handler**: Extracts `{name}` from URL using `chi.URLParam`, extracts query params, calls service
5. **Multiple metrics handler**: Extracts `?names=` param, splits by comma, calls service
6. **Helper methods**: `respondJSON` and `respondError` for consistent response format

**What success looks like**: All handler tests pass, and the implementation correctly translates between HTTP and the service layer.

**Commit** after tests pass.

---

### Task 6.3: Create Router Setup

**The problem we're solving**: We need to wire up:
- HTTP routes to handlers
- Middleware for logging, recovery, timeouts
- Request lifecycle management

**Architecture decisions**:

1. **Middleware chain**: Use Chi's middleware composition to add:
   - Request ID generation (for tracing)
   - Real IP extraction (for logging behind proxies)
   - Request logging (log method, path, status, duration)
   - Panic recovery (convert panics to 500 errors)
   - Timeouts (prevent long-running requests)

2. **Route organization**: Use a factory function `NewRouter` that takes dependencies (handler, logger) and returns a configured `*chi.Mux`

3. **Conditional routing**: The `/metrics` endpoint checks for the presence of `?names=` to decide whether to list all metrics or fetch specific ones

**Implementation**:

Create `internal/api/router.go` with:

1. **NewRouter function**: Takes handlers and logger, returns configured router
2. **Middleware setup**: Add Chi middleware in order
3. **Custom request logger**: Wraps Chi's response writer to capture status code and timing
4. **Route definitions**: Map HTTP methods and paths to handlers

**What success looks like**: The router correctly routes requests to handlers and middleware executes in order.

**Commit** after creating the router.

---

## Phase 7: Main Application

The main application wires everything together: configuration loading, dependency initialization, HTTP server setup, and graceful shutdown.

### Task 7.1: Implement main.go

**The problem we're solving**: We need an entry point that:
1. Loads configuration from disk and environment
2. Initializes all layers (repository, service, handlers)
3. Starts an HTTP server
4. Handles graceful shutdown on SIGINT/SIGTERM

**Architecture decisions**:

1. **Dependency injection**: Pass dependencies explicitly (repository to service, service to handlers). No global variables or singletons.

2. **Fail fast on startup**: If configuration is invalid or database connection fails, exit immediately with os.Exit(1). Don't start the server in a broken state.

3. **Graceful shutdown**: Listen for OS signals and give in-flight requests 30 seconds to complete before forcing shutdown.

4. **Structured logging**: Use Go's `slog` package for JSON-formatted logs with structured fields.

**Implementation**:

Create `cmd/server/main.go` with:

1. **Setup logging**: Create a JSON logger and set it as the default
2. **Load config**: Get PORT and DB_PATH from environment (with defaults), load metrics.toml
3. **Initialize layers**: Create repository → create service → create handlers → create router
4. **Configure HTTP server**: Set up timeouts properly:
   - `ReadTimeout: 10s` - Clients should send requests quickly
   - `WriteTimeout: 30s` - Allow time to write response after middleware timeout
   - `middleware.Timeout: 25s` - Request processing timeout (fires before WriteTimeout)
   - Important: Middleware timeout must be shorter than WriteTimeout to allow clean cancellation
5. **Start server**: Run `server.ListenAndServe()` in a goroutine
6. **Wait for signal**: Block on a signal channel
7. **Shutdown**: Call `server.Shutdown()` with a 30-second timeout context

**Key Go patterns**:
- Use `defer repo.Close()` to ensure the database connection is closed
- Use `signal.Notify()` to listen for SIGINT/SIGTERM
- Use `context.WithTimeout()` for the shutdown deadline

**What success looks like**: The server starts, responds to requests, and shuts down cleanly when receiving SIGINT (Ctrl+C).

**Commit** after verifying the server starts.

---

## Phase 8: Example Configuration and Data

Provide example configuration and test data so the implementer can immediately run and test the application.

### Task 8.1: Create Example Metrics Config

**What you're providing**: A complete `config/metrics.toml` file with realistic examples of:
- Simple metrics without parameters
- Multi-row metrics
- Parameterized metrics with different types

**Important**: All parameters should be marked as `required = true`. Optional parameters are not supported with SQL positional parameters (`?`) - if a parameter might not be provided, create separate metrics instead.

Create `config/metrics.toml` with 4-5 example metrics that demonstrate different features.

**Commit** after creating the file.

---

### Task 8.2: Create Test Database Setup Script

**What you're providing**: A shell script and SQL file that create a SQLite database with sample data matching the example metrics.

Create:
1. `scripts/setup_test_db.sql` - SQL schema and INSERT statements
2. `scripts/setup_test_db.sh` - Bash script that runs sqlite3 with the SQL file

Make the script executable with `chmod +x`.

**What success looks like**: Running `./scripts/setup_test_db.sh` creates `data.db` with sample data.

**Commit** after creating the scripts.

---

## Phase 9: Documentation and Testing

### Task 9.1: Create README

**What you're providing**: A README.md with:
- Quick start instructions (setup DB, run server, test endpoints)
- Configuration options (environment variables)
- Development commands (tests, build)

**Configuration Documentation Requirements**:

1. **Environment Variables**: Document how to set environment variables before running the server:
   - `PORT` - Server port (default: 8080)
   - `DB_PATH` - Path to SQLite database (default: ./data.db)
   - Example: `PORT=3000 DB_PATH=/var/data/metrics.db go run ./cmd/server`

2. **Log Level Configuration**: Document how to change the log level in `cmd/server/main.go`:
   - Default: `slog.LevelInfo` (shows INFO, WARN, ERROR)
   - For debugging: Change to `slog.LevelDebug` (shows DEBUG, INFO, WARN, ERROR)
   - Location: Line ~88 in `cmd/server/main.go` in the `setupLogging()` function

3. **Metrics Configuration**: Document the architectural constraint about parameters:
   - All parameters must be marked `required = true` in `config/metrics.toml`
   - Optional parameters are not supported with SQL positional parameters (`?`)
   - If a parameter might not be provided, create separate metrics instead
   - Example: Instead of one metric with optional `limit`, create `users_all` and `users_top_10` as separate metrics

Update `README.md` with clear, actionable instructions.

**Commit** after updating README.

---

### Task 9.2: End-to-End Manual Test

**What you're doing**: Manually verify the entire system works by:
1. Setting up the test database
2. Starting the server
3. Making HTTP requests with curl
4. Verifying responses match expectations

Run the commands in the README and verify each endpoint returns the correct data.

**What success looks like**: All curl commands return the expected JSON responses.

**Commit** (empty commit is fine) after successful manual testing.

---

## Summary

You've now built a complete metrics API service following TDD, YAGNI, and KISS principles!

**Key achievements**:
- Test-driven development throughout
- Clean architecture with separated concerns
- Type-safe parameter handling
- Concurrent query execution
- Comprehensive error handling
- Structured logging
- Graceful shutdown
- Well-tested codebase

**Next steps (if needed)**:
- Add integration tests
- Add caching layer
- Add metrics/observability
- Deploy to production

Remember to commit frequently and keep your tests passing!
