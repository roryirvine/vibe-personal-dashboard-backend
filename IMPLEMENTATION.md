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

### Task 3.1: Install TOML Parser Dependency

**What you're doing**: Adding the TOML parsing library.

**Command**:
```bash
go get github.com/BurntSushi/toml
```

**How to verify**:
```bash
cat go.mod | grep toml
```

You should see:
```
github.com/BurntSushi/toml v1.x.x
```

**Commit**:
```bash
git add go.mod go.sum
git commit -m "Add TOML parser dependency"
```

---

### Task 3.2: Create Config Struct and Parser (TDD)

**What you're doing**: Writing code to load and validate the metrics.toml file.

**Step 1: Write the test first** (this is TDD!)

**File to create**: `internal/config/config_test.go`

**Code**:
```go
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/models"
)

func TestLoadConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		// Create a temporary config file
		content := `
[[metrics]]
name = "test_metric"
query = "SELECT COUNT(*) FROM users"
multi_row = false

[[metrics]]
name = "users_list"
query = "SELECT id, name FROM users WHERE created > ?"
multi_row = true

[[metrics.params]]
name = "start_date"
type = "string"
required = true
`
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "metrics.toml")
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test config: %v", err)
		}

		// Load config
		metrics, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		// Verify we got 2 metrics
		if len(metrics) != 2 {
			t.Errorf("got %d metrics, want 2", len(metrics))
		}

		// Verify first metric
		if metrics[0].Name != "test_metric" {
			t.Errorf("first metric name = %s, want test_metric", metrics[0].Name)
		}
		if metrics[0].MultiRow {
			t.Error("first metric should be single-row")
		}

		// Verify second metric
		if metrics[1].Name != "users_list" {
			t.Errorf("second metric name = %s, want users_list", metrics[1].Name)
		}
		if !metrics[1].MultiRow {
			t.Error("second metric should be multi-row")
		}
		if len(metrics[1].Params) != 1 {
			t.Errorf("second metric has %d params, want 1", len(metrics[1].Params))
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := LoadConfig("/nonexistent/path.toml")
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})

	t.Run("invalid toml", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "bad.toml")
		if err := os.WriteFile(configPath, []byte("invalid { toml"), 0644); err != nil {
			t.Fatalf("failed to write test config: %v", err)
		}

		_, err := LoadConfig(configPath)
		if err == nil {
			t.Error("expected error for invalid TOML")
		}
	})

	t.Run("duplicate metric names", func(t *testing.T) {
		content := `
[[metrics]]
name = "duplicate"
query = "SELECT 1"

[[metrics]]
name = "duplicate"
query = "SELECT 2"
`
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "dup.toml")
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test config: %v", err)
		}

		_, err := LoadConfig(configPath)
		if err == nil {
			t.Error("expected error for duplicate metric names")
		}
	})
}
```

**Run the test** (it will fail because we haven't written the code yet):
```bash
go test ./internal/config/
```

You'll see errors like `undefined: LoadConfig`. **This is expected in TDD!**

**Step 2: Write just enough code to make the test pass**

**File to create**: `internal/config/config.go`

**Code**:
```go
// Loads and validates TOML metric configuration files.
package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/models"
)

type Config struct {
	Metrics []models.Metric `toml:"metrics"`
}

func LoadConfig(path string) ([]models.Metric, error) {
	var config Config

	// Parse TOML file
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate all metrics
	if err := validateMetrics(config.Metrics); err != nil {
		return nil, err
	}

	return config.Metrics, nil
}

func validateMetrics(metrics []models.Metric) error {
	if len(metrics) == 0 {
		return fmt.Errorf("no metrics defined in config")
	}

	// Check for duplicate names
	names := make(map[string]bool)
	for _, metric := range metrics {
		if names[metric.Name] {
			return fmt.Errorf("duplicate metric name: %s", metric.Name)
		}
		names[metric.Name] = true

		// Validate each metric
		if err := metric.Validate(); err != nil {
			return fmt.Errorf("invalid metric %s: %w", metric.Name, err)
		}
	}

	return nil
}
```

**Run tests**:
```bash
go test ./internal/config/
```

All tests should pass now!

**Commit**:
```bash
git add internal/config/
git commit -m "Add config loading with TOML parsing and validation"
```

---

## Phase 4: Repository Layer

### Task 4.1: Define Repository Interface

**What you're doing**: Creating an interface so we can swap database implementations later.

**File to create**: `internal/repository/repository.go`

**Code**:
```go
// Defines the database repository interface for metric queries.
package repository

import "context"

type Repository interface {
	// QuerySingleValue executes a query and returns the first column of the first row.
	// Returns an error if the query fails or returns no rows.
	QuerySingleValue(ctx context.Context, query string, args ...interface{}) (interface{}, error)

	// QueryMultiRow executes a query and returns all rows as a slice of maps.
	// Each map represents a row with column names as keys.
	QueryMultiRow(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)

	// Close closes the database connection and releases resources.
	Close() error
}
```

**No test needed yet**: This is just an interface definition.

**Commit**:
```bash
git add internal/repository/repository.go
git commit -m "Define Repository interface"
```

---

### Task 4.2: Install SQLite Driver

**What you're doing**: Adding the pure-Go SQLite driver.

**Command**:
```bash
go get modernc.org/sqlite
```

**Why this driver**: It's pure Go (no CGO), which makes it easier to cross-compile and deploy.

**Commit**:
```bash
git add go.mod go.sum
git commit -m "Add SQLite driver dependency"
```

---

### Task 4.3: Implement SQLite Repository (TDD)

**Step 1: Write tests**

**File to create**: `internal/repository/sqlite_test.go`

**Code**:
```go
package repository

import (
	"context"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) Repository {
	// Use in-memory SQLite database for testing
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create test repository: %v", err)
	}

	// Create a test table
	ctx := context.Background()
	_, err = repo.(*SQLiteRepository).db.ExecContext(ctx, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			age INTEGER,
			balance REAL
		)
	`)
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	// Insert test data
	_, err = repo.(*SQLiteRepository).db.ExecContext(ctx, `
		INSERT INTO users (id, name, age, balance) VALUES
		(1, 'Alice', 30, 100.50),
		(2, 'Bob', 25, 200.75),
		(3, 'Charlie', 35, NULL)
	`)
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	return repo
}

func TestSQLiteRepository_QuerySingleValue(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	ctx := context.Background()

	t.Run("query returns integer", func(t *testing.T) {
		result, err := repo.QuerySingleValue(ctx, "SELECT COUNT(*) FROM users")
		if err != nil {
			t.Fatalf("QuerySingleValue() error = %v", err)
		}

		count, ok := result.(int64)
		if !ok {
			t.Fatalf("expected int64, got %T", result)
		}
		if count != 3 {
			t.Errorf("got count %d, want 3", count)
		}
	})

	t.Run("query returns float", func(t *testing.T) {
		result, err := repo.QuerySingleValue(ctx, "SELECT balance FROM users WHERE id = 1")
		if err != nil {
			t.Fatalf("QuerySingleValue() error = %v", err)
		}

		balance, ok := result.(float64)
		if !ok {
			t.Fatalf("expected float64, got %T", result)
		}
		if balance != 100.50 {
			t.Errorf("got balance %f, want 100.50", balance)
		}
	})

	t.Run("query returns string", func(t *testing.T) {
		result, err := repo.QuerySingleValue(ctx, "SELECT name FROM users WHERE id = 2")
		if err != nil {
			t.Fatalf("QuerySingleValue() error = %v", err)
		}

		name, ok := result.(string)
		if !ok {
			t.Fatalf("expected string, got %T", result)
		}
		if name != "Bob" {
			t.Errorf("got name %s, want Bob", name)
		}
	})

	t.Run("query with WHERE clause", func(t *testing.T) {
		result, err := repo.QuerySingleValue(ctx, "SELECT name FROM users WHERE age > ?", 30)
		if err != nil {
			t.Fatalf("QuerySingleValue() error = %v", err)
		}

		name := result.(string)
		if name != "Charlie" {
			t.Errorf("got name %s, want Charlie", name)
		}
	})

	t.Run("query returns no rows", func(t *testing.T) {
		_, err := repo.QuerySingleValue(ctx, "SELECT name FROM users WHERE id = 999")
		if err == nil {
			t.Error("expected error for no rows, got nil")
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(1 * time.Millisecond) // Ensure context is expired

		_, err := repo.QuerySingleValue(ctx, "SELECT COUNT(*) FROM users")
		if err == nil {
			t.Error("expected context timeout error")
		}
	})
}

func TestSQLiteRepository_QueryMultiRow(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	ctx := context.Background()

	t.Run("query returns multiple rows", func(t *testing.T) {
		rows, err := repo.QueryMultiRow(ctx, "SELECT id, name, age FROM users ORDER BY id")
		if err != nil {
			t.Fatalf("QueryMultiRow() error = %v", err)
		}

		if len(rows) != 3 {
			t.Fatalf("got %d rows, want 3", len(rows))
		}

		// Check first row
		if rows[0]["id"].(int64) != 1 {
			t.Errorf("row 0 id = %v, want 1", rows[0]["id"])
		}
		if rows[0]["name"].(string) != "Alice" {
			t.Errorf("row 0 name = %v, want Alice", rows[0]["name"])
		}

		// Check NULL value handling
		if rows[2]["balance"] != nil {
			t.Errorf("row 2 balance = %v, want nil", rows[2]["balance"])
		}
	})

	t.Run("query with WHERE clause", func(t *testing.T) {
		rows, err := repo.QueryMultiRow(ctx, "SELECT name FROM users WHERE age < ?", 30)
		if err != nil {
			t.Fatalf("QueryMultiRow() error = %v", err)
		}

		if len(rows) != 1 {
			t.Fatalf("got %d rows, want 1", len(rows))
		}
		if rows[0]["name"].(string) != "Bob" {
			t.Errorf("got name %v, want Bob", rows[0]["name"])
		}
	})

	t.Run("query returns empty result", func(t *testing.T) {
		rows, err := repo.QueryMultiRow(ctx, "SELECT * FROM users WHERE id > 1000")
		if err != nil {
			t.Fatalf("QueryMultiRow() error = %v", err)
		}

		if len(rows) != 0 {
			t.Errorf("got %d rows, want 0", len(rows))
		}
	})
}
```

**Run tests** (they'll fail):
```bash
go test ./internal/repository/
```

**Step 2: Implement the SQLite repository**

**File to create**: `internal/repository/sqlite.go`

**Code**:
```go
// Implements the repository interface using SQLite.
package repository

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // Import SQLite driver
)

type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLite repository.
// The path parameter can be a file path or ":memory:" for an in-memory database.
func NewSQLiteRepository(path string) (Repository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &SQLiteRepository{db: db}, nil
}

func (r *SQLiteRepository) QuerySingleValue(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	row := r.db.QueryRowContext(ctx, query, args...)

	var result interface{}
	if err := row.Scan(&result); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("query returned no rows")
		}
		return nil, fmt.Errorf("failed to scan result: %w", err)
	}

	return result, nil
}

func (r *SQLiteRepository) QueryMultiRow(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold each column value
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Build map for this row
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			rowMap[col] = values[i]
		}

		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
```

**Run tests**:
```bash
go test ./internal/repository/ -v
```

All tests should pass!

**Commit**:
```bash
git add internal/repository/
git commit -m "Implement SQLite repository with comprehensive tests"
```

---

## Phase 5: Service Layer

### Task 5.1: Install errgroup Dependency

**What you're doing**: Adding the errgroup package for concurrent execution.

**Command**:
```bash
go get golang.org/x/sync/errgroup
```

**Commit**:
```bash
git add go.mod go.sum
git commit -m "Add errgroup dependency for concurrent execution"
```

---

### Task 5.2: Implement Parameter Conversion Helper (TDD)

**What you're doing**: Converting URL query parameter strings to the correct Go types.

**File to create**: `internal/service/params_test.go`

**Code**:
```go
package service

import (
	"testing"

	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/models"
)

func TestConvertParamValue(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		paramType models.ParamType
		want      interface{}
		wantErr   bool
	}{
		{
			name:      "string value",
			value:     "hello",
			paramType: models.ParamTypeString,
			want:      "hello",
			wantErr:   false,
		},
		{
			name:      "int value",
			value:     "42",
			paramType: models.ParamTypeInt,
			want:      int64(42),
			wantErr:   false,
		},
		{
			name:      "negative int",
			value:     "-10",
			paramType: models.ParamTypeInt,
			want:      int64(-10),
			wantErr:   false,
		},
		{
			name:      "invalid int",
			value:     "not-a-number",
			paramType: models.ParamTypeInt,
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "float value",
			value:     "3.14",
			paramType: models.ParamTypeFloat,
			want:      float64(3.14),
			wantErr:   false,
		},
		{
			name:      "invalid float",
			value:     "abc",
			paramType: models.ParamTypeFloat,
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertParamValue(tt.value, tt.paramType)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertParamValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("convertParamValue() = %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}
```

**Run tests** (will fail):
```bash
go test ./internal/service/
```

**File to create**: `internal/service/params.go`

**Code**:
```go
// Converts URL query parameters to typed values for database queries.
package service

import (
	"fmt"
	"strconv"

	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/models"
)

func convertParamValue(value string, paramType models.ParamType) (interface{}, error) {
	switch paramType {
	case models.ParamTypeString:
		return value, nil

	case models.ParamTypeInt:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid integer value %q: %w", value, err)
		}
		return intVal, nil

	case models.ParamTypeFloat:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float value %q: %w", value, err)
		}
		return floatVal, nil

	default:
		return nil, fmt.Errorf("unsupported parameter type: %s", paramType)
	}
}
```

**Run tests**:
```bash
go test ./internal/service/
```

**Commit**:
```bash
git add internal/service/
git commit -m "Add parameter conversion with type validation"
```

---

### Task 5.3: Implement MetricService (TDD)

This is a larger task, so we'll break it into sub-steps.

**File to create**: `internal/service/metric_service_test.go`

**Code (Part 1 - test setup and simple tests)**:
```go
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/models"
)

type mockRepository struct {
	singleValueFunc func(ctx context.Context, query string, args ...interface{}) (interface{}, error)
	multiRowFunc    func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)
}

func (m *mockRepository) QuerySingleValue(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	if m.singleValueFunc != nil {
		return m.singleValueFunc(ctx, query, args...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) QueryMultiRow(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if m.multiRowFunc != nil {
		return m.multiRowFunc(ctx, query, args...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) Close() error {
	return nil
}

func TestNewMetricService(t *testing.T) {
	repo := &mockRepository{}
	metrics := []models.Metric{
		{Name: "test", Query: "SELECT 1", MultiRow: false},
	}

	service := NewMetricService(repo, metrics, nil)
	if service == nil {
		t.Error("NewMetricService returned nil")
	}
}

func TestMetricService_GetMetricNames(t *testing.T) {
	repo := &mockRepository{}
	metrics := []models.Metric{
		{Name: "metric1", Query: "SELECT 1", MultiRow: false},
		{Name: "metric2", Query: "SELECT 2", MultiRow: false},
		{Name: "metric3", Query: "SELECT 3", MultiRow: false},
	}

	service := NewMetricService(repo, metrics, nil)
	names := service.GetMetricNames()

	if len(names) != 3 {
		t.Errorf("got %d names, want 3", len(names))
	}

	expectedNames := map[string]bool{"metric1": true, "metric2": true, "metric3": true}
	for _, name := range names {
		if !expectedNames[name] {
			t.Errorf("unexpected metric name: %s", name)
		}
	}
}

func TestMetricService_GetMetric_SingleValue(t *testing.T) {
	repo := &mockRepository{
		singleValueFunc: func(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
			if query == "SELECT COUNT(*) FROM users" {
				return int64(42), nil
			}
			return nil, errors.New("unexpected query")
		},
	}

	metrics := []models.Metric{
		{
			Name:     "user_count",
			Query:    "SELECT COUNT(*) FROM users",
			MultiRow: false,
		},
	}

	service := NewMetricService(repo, metrics, nil)
	ctx := context.Background()

	result, err := service.GetMetric(ctx, "user_count", nil)
	if err != nil {
		t.Fatalf("GetMetric() error = %v", err)
	}

	if result.Name != "user_count" {
		t.Errorf("result.Name = %s, want user_count", result.Name)
	}

	value, ok := result.Value.(int64)
	if !ok {
		t.Fatalf("expected int64 value, got %T", result.Value)
	}
	if value != 42 {
		t.Errorf("value = %d, want 42", value)
	}
}

func TestMetricService_GetMetric_MultiRow(t *testing.T) {
	repo := &mockRepository{
		multiRowFunc: func(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{"id": int64(1), "name": "Alice"},
				{"id": int64(2), "name": "Bob"},
			}, nil
		},
	}

	metrics := []models.Metric{
		{
			Name:     "users_list",
			Query:    "SELECT id, name FROM users",
			MultiRow: true,
		},
	}

	service := NewMetricService(repo, metrics, nil)
	ctx := context.Background()

	result, err := service.GetMetric(ctx, "users_list", nil)
	if err != nil {
		t.Fatalf("GetMetric() error = %v", err)
	}

	rows, ok := result.Value.([]map[string]interface{})
	if !ok {
		t.Fatalf("expected []map[string]interface{}, got %T", result.Value)
	}

	if len(rows) != 2 {
		t.Errorf("got %d rows, want 2", len(rows))
	}
}

func TestMetricService_GetMetric_WithParams(t *testing.T) {
	repo := &mockRepository{
		singleValueFunc: func(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
			// Verify parameters were passed correctly
			if len(args) != 1 {
				t.Errorf("expected 1 arg, got %d", len(args))
			}
			if args[0] != int64(5) {
				t.Errorf("expected arg 5, got %v", args[0])
			}
			return int64(100), nil
		},
	}

	metrics := []models.Metric{
		{
			Name:     "users_over_age",
			Query:    "SELECT COUNT(*) FROM users WHERE age > ?",
			MultiRow: false,
			Params: []models.ParamDefinition{
				{Name: "min_age", Type: models.ParamTypeInt, Required: true},
			},
		},
	}

	service := NewMetricService(repo, metrics, nil)
	ctx := context.Background()

	params := map[string]string{"min_age": "5"}
	result, err := service.GetMetric(ctx, "users_over_age", params)
	if err != nil {
		t.Fatalf("GetMetric() error = %v", err)
	}

	if result.Value.(int64) != 100 {
		t.Errorf("value = %v, want 100", result.Value)
	}
}

func TestMetricService_GetMetric_MissingRequiredParam(t *testing.T) {
	repo := &mockRepository{}

	metrics := []models.Metric{
		{
			Name:     "test",
			Query:    "SELECT * FROM users WHERE id = ?",
			MultiRow: false,
			Params: []models.ParamDefinition{
				{Name: "user_id", Type: models.ParamTypeInt, Required: true},
			},
		},
	}

	service := NewMetricService(repo, metrics, nil)
	ctx := context.Background()

	_, err := service.GetMetric(ctx, "test", nil)
	if err == nil {
		t.Error("expected error for missing required parameter")
	}
}

func TestMetricService_GetMetric_NotFound(t *testing.T) {
	repo := &mockRepository{}
	service := NewMetricService(repo, []models.Metric{}, nil)
	ctx := context.Background()

	_, err := service.GetMetric(ctx, "nonexistent", nil)
	if err == nil {
		t.Error("expected error for nonexistent metric")
	}
}
```

**Run tests** (will fail):
```bash
go test ./internal/service/
```

**File to create**: `internal/service/metric_service.go`

**Code**:
```go
// Executes metric queries with parameter validation and concurrent execution.
package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/models"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/repository"
	"golang.org/x/sync/errgroup"
)

type MetricService struct {
	repo    repository.Repository
	metrics map[string]models.Metric
	logger  *slog.Logger
}

func NewMetricService(repo repository.Repository, metrics []models.Metric, logger *slog.Logger) *MetricService {
	// Build a map for fast metric lookup
	metricMap := make(map[string]models.Metric)
	for _, m := range metrics {
		metricMap[m.Name] = m
	}

	return &MetricService{
		repo:    repo,
		metrics: metricMap,
		logger:  logger,
	}
}

func (s *MetricService) GetMetricNames() []string {
	names := make([]string, 0, len(s.metrics))
	for name := range s.metrics {
		names = append(names, name)
	}
	return names
}

func (s *MetricService) GetMetric(ctx context.Context, name string, params map[string]string) (models.MetricResult, error) {
	// Find metric definition
	metric, exists := s.metrics[name]
	if !exists {
		return models.MetricResult{}, fmt.Errorf("metric not found: %s", name)
	}

	// Validate and convert parameters
	args, err := s.prepareParams(metric, params)
	if err != nil {
		return models.MetricResult{}, fmt.Errorf("invalid parameters: %w", err)
	}

	// Execute query based on metric type
	var value interface{}
	if metric.MultiRow {
		value, err = s.repo.QueryMultiRow(ctx, metric.Query, args...)
	} else {
		value, err = s.repo.QuerySingleValue(ctx, metric.Query, args...)
	}

	if err != nil {
		return models.MetricResult{}, fmt.Errorf("query failed: %w", err)
	}

	return models.MetricResult{
		Name:  name,
		Value: value,
	}, nil
}

func (s *MetricService) GetMetrics(ctx context.Context, names []string, params map[string]string) ([]models.MetricResult, error) {
	results := make([]models.MetricResult, len(names))
	g, ctx := errgroup.WithContext(ctx)

	for i, name := range names {
		i, name := i, name // Capture loop variables
		g.Go(func() error {
			result, err := s.GetMetric(ctx, name, params)
			if err != nil {
				return fmt.Errorf("metric %s: %w", name, err)
			}
			results[i] = result
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *MetricService) prepareParams(metric models.Metric, params map[string]string) ([]interface{}, error) {
	if len(metric.Params) == 0 {
		return nil, nil
	}

	args := make([]interface{}, len(metric.Params))

	for i, paramDef := range metric.Params {
		value, exists := params[paramDef.Name]

		// Check if required parameter is missing
		if !exists {
			if paramDef.Required {
				return nil, fmt.Errorf("required parameter missing: %s", paramDef.Name)
			}
			// Optional parameter not provided - use nil
			args[i] = nil
			continue
		}

		// Convert to correct type
		converted, err := convertParamValue(value, paramDef.Type)
		if err != nil {
			return nil, fmt.Errorf("parameter %s: %w", paramDef.Name, err)
		}
		args[i] = converted
	}

	return args, nil
}
```

**Run tests**:
```bash
go test ./internal/service/ -v
```

All tests should pass!

**Add a test for concurrent execution**:

**Add to `metric_service_test.go`**:
```go
func TestMetricService_GetMetrics_Concurrent(t *testing.T) {
	callCount := 0
	repo := &mockRepository{
		singleValueFunc: func(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
			callCount++
			return int64(callCount), nil
		},
	}

	metrics := []models.Metric{
		{Name: "metric1", Query: "SELECT 1", MultiRow: false},
		{Name: "metric2", Query: "SELECT 2", MultiRow: false},
		{Name: "metric3", Query: "SELECT 3", MultiRow: false},
	}

	service := NewMetricService(repo, metrics, nil)
	ctx := context.Background()

	results, err := service.GetMetrics(ctx, []string{"metric1", "metric2", "metric3"}, nil)
	if err != nil {
		t.Fatalf("GetMetrics() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("got %d results, want 3", len(results))
	}

	// Verify all metrics were called
	if callCount != 3 {
		t.Errorf("repository was called %d times, want 3", callCount)
	}
}

func TestMetricService_GetMetrics_ErrorHandling(t *testing.T) {
	repo := &mockRepository{
		singleValueFunc: func(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
			if query == "SELECT 2" {
				return nil, errors.New("database error")
			}
			return int64(1), nil
		},
	}

	metrics := []models.Metric{
		{Name: "metric1", Query: "SELECT 1", MultiRow: false},
		{Name: "metric2", Query: "SELECT 2", MultiRow: false},
	}

	service := NewMetricService(repo, metrics, nil)
	ctx := context.Background()

	_, err := service.GetMetrics(ctx, []string{"metric1", "metric2"}, nil)
	if err == nil {
		t.Error("expected error when one metric fails")
	}
}
```

**Run tests again**:
```bash
go test ./internal/service/ -v
```

**Commit**:
```bash
git add internal/service/
git commit -m "Implement MetricService with concurrent execution and comprehensive tests"
```

---

## Phase 6: HTTP API Layer

### Task 6.1: Install Chi Router

**Command**:
```bash
go get github.com/go-chi/chi/v5
```

**Commit**:
```bash
git add go.mod go.sum
git commit -m "Add chi router dependency"
```

---

### Task 6.2: Implement HTTP Handlers (TDD)

Due to length constraints, I'll provide a condensed version focusing on key patterns.

**File to create**: `internal/api/handlers/metrics_test.go`

**Code (abbreviated - showing pattern)**:
```go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/models"
)

type mockMetricService struct {
	getNamesFunc  func() []string
	getMetricFunc func(ctx context.Context, name string, params map[string]string) (models.MetricResult, error)
	getMetricsFunc func(ctx context.Context, names []string, params map[string]string) ([]models.MetricResult, error)
}

func (m *mockMetricService) GetMetricNames() []string {
	if m.getNamesFunc != nil {
		return m.getNamesFunc()
	}
	return nil
}

func (m *mockMetricService) GetMetric(ctx context.Context, name string, params map[string]string) (models.MetricResult, error) {
	if m.getMetricFunc != nil {
		return m.getMetricFunc(ctx, name, params)
	}
	return models.MetricResult{}, nil
}

func (m *mockMetricService) GetMetrics(ctx context.Context, names []string, params map[string]string) ([]models.MetricResult, error) {
	if m.getMetricsFunc != nil {
		return m.getMetricsFunc(ctx, names, params)
	}
	return nil, nil
}

func TestListMetrics(t *testing.T) {
	service := &mockMetricService{
		getNamesFunc: func() []string {
			return []string{"metric1", "metric2"}
		},
	}

	handler := NewMetricsHandler(service, nil)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ListMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var names []string
	if err := json.NewDecoder(w.Body).Decode(&names); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(names) != 2 {
		t.Errorf("got %d names, want 2", len(names))
	}
}

func TestGetSingleMetric(t *testing.T) {
	service := &mockMetricService{
		getMetricFunc: func(ctx context.Context, name string, params map[string]string) (models.MetricResult, error) {
			return models.MetricResult{
				Name:  name,
				Value: int64(42),
			}, nil
		},
	}

	handler := NewMetricsHandler(service, nil)

	req := httptest.NewRequest(http.MethodGet, "/metrics/test_metric", nil)
	w := httptest.NewRecorder()

	// Set up chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("name", "test_metric")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetSingleMetric(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var results []models.MetricResult
	if err := json.NewDecoder(w.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
	if results[0].Name != "test_metric" {
		t.Errorf("name = %s, want test_metric", results[0].Name)
	}
}

```

**File to create**: `internal/api/handlers/metrics.go`

**Code**:
```go
// HTTP handlers for metrics API endpoints.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/models"
)

type MetricService interface {
	GetMetricNames() []string
	GetMetric(ctx context.Context, name string, params map[string]string) (models.MetricResult, error)
	GetMetrics(ctx context.Context, names []string, params map[string]string) ([]models.MetricResult, error)
}

type MetricsHandler struct {
	service MetricService
	logger  *slog.Logger
}

func NewMetricsHandler(service MetricService, logger *slog.Logger) *MetricsHandler {
	return &MetricsHandler{
		service: service,
		logger:  logger,
	}
}

func (h *MetricsHandler) ListMetrics(w http.ResponseWriter, r *http.Request) {
	names := h.service.GetMetricNames()
	h.respondJSON(w, http.StatusOK, names)
}

func (h *MetricsHandler) GetSingleMetric(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	params := extractQueryParams(r)

	result, err := h.service.GetMetric(r.Context(), name, params)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, fmt.Sprintf("metric %s failed: %v", name, err))
		return
	}

	h.respondJSON(w, http.StatusOK, []models.MetricResult{result})
}

func (h *MetricsHandler) GetMultipleMetrics(w http.ResponseWriter, r *http.Request) {
	// Parse comma-separated metric names from query param
	namesParam := r.URL.Query().Get("names")
	if namesParam == "" {
		h.respondError(w, http.StatusBadRequest, "missing 'names' query parameter")
		return
	}

	names := strings.Split(namesParam, ",")
	for i, name := range names {
		names[i] = strings.TrimSpace(name)
	}

	params := extractQueryParams(r)
	delete(params, "names") // Remove 'names' from params

	results, err := h.service.GetMetrics(r.Context(), names, params)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, results)
}

func extractQueryParams(r *http.Request) map[string]string {
	params := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0] // Take first value if multiple provided
		}
	}
	return params
}

func (h *MetricsHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		if h.logger != nil {
			h.logger.Error("failed to encode JSON response", "error", err)
		}
	}
}

func (h *MetricsHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
```

**Run tests**:
```bash
go test ./internal/api/handlers/
```

**Commit**:
```bash
git add internal/api/handlers/
git commit -m "Implement HTTP handlers with tests"
```

---

### Task 6.3: Create Router Setup

**File to create**: `internal/api/router.go`

**Code**:
```go
// Configures HTTP routes and middleware for the metrics API.
package api

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/api/handlers"
)

func NewRouter(metricsHandler *handlers.MetricsHandler, logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(requestLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Routes
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// Check if 'names' param exists to determine which handler to use
		if r.URL.Query().Get("names") != "" {
			metricsHandler.GetMultipleMetrics(w, r)
		} else {
			metricsHandler.ListMetrics(w, r)
		}
	})
	r.Get("/metrics/{name}", metricsHandler.GetSingleMetric)

	return r
}

func requestLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
```

**Add missing import**:
```go
import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/api/handlers"
)
```

**Commit**:
```bash
git add internal/api/
git commit -m "Add router configuration with logging middleware"
```

---

## Phase 7: Main Application

### Task 7.1: Implement main.go

**File to create**: `cmd/server/main.go`

**Code**:
```go
// Main entry point for the metrics API server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/api"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/api/handlers"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/config"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/repository"
	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/service"
)

func main() {
	// Set up structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("starting vibe dashboard backend")

	// Load configuration
	port := getEnv("PORT", "8080")
	dbPath := getEnv("DB_PATH", "./data.db")
	configPath := "./config/metrics.toml"

	logger.Info("configuration",
		"port", port,
		"db_path", dbPath,
		"config_path", configPath,
	)

	// Load metrics configuration
	metrics, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	logger.Info("loaded metrics", "count", len(metrics))

	// Initialize repository
	repo, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		logger.Error("failed to initialize repository", "error", err)
		os.Exit(1)
	}
	defer repo.Close()
	logger.Info("database connection established")

	// Initialize service
	metricService := service.NewMetricService(repo, metrics, logger)

	// Initialize handlers
	metricsHandler := handlers.NewMetricsHandler(metricService, logger)

	// Set up router
	router := api.NewRouter(metricsHandler, logger)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("server starting", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("server shutting down")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}

	logger.Info("server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
```

**Commit**:
```bash
git add cmd/server/main.go
git commit -m "Implement main application with graceful shutdown"
```

---

## Phase 8: Example Configuration and Data

### Task 8.1: Create Example Metrics Config

**File to create**: `config/metrics.toml`

**Code**:
```toml
# Example metrics configuration

[[metrics]]
name = "total_users"
query = "SELECT COUNT(*) FROM users"
multi_row = false

[[metrics]]
name = "active_users"
query = "SELECT COUNT(*) FROM users WHERE last_active > datetime('now', '-7 days')"
multi_row = false

[[metrics]]
name = "users_by_status"
query = "SELECT status, COUNT(*) as count FROM users GROUP BY status ORDER BY status"
multi_row = true

[[metrics]]
name = "users_since_date"
query = "SELECT id, name, email, created FROM users WHERE created >= ? ORDER BY created DESC"
multi_row = true

[[metrics.params]]
name = "start_date"
type = "string"
required = true

[[metrics]]
name = "user_balance_sum"
query = "SELECT SUM(balance) FROM users WHERE balance > ?"
multi_row = false

[[metrics.params]]
name = "min_balance"
type = "float"
required = false
```

**Commit**:
```bash
git add config/metrics.toml
git commit -m "Add example metrics configuration"
```

---

### Task 8.2: Create Test Database Setup Script

**File to create**: `scripts/setup_test_db.sql`

**Code**:
```sql
-- Test database schema and data

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'active',
    balance REAL,
    created TEXT NOT NULL DEFAULT (datetime('now')),
    last_active TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Insert test data
INSERT INTO users (name, email, status, balance, created, last_active) VALUES
    ('Alice Smith', 'alice@example.com', 'active', 1500.50, '2025-01-15 10:00:00', '2025-10-14 15:30:00'),
    ('Bob Jones', 'bob@example.com', 'active', 2300.75, '2025-02-20 14:30:00', '2025-10-15 09:00:00'),
    ('Charlie Brown', 'charlie@example.com', 'inactive', 0.00, '2024-12-01 09:15:00', '2025-01-05 12:00:00'),
    ('Diana Prince', 'diana@example.com', 'active', 5000.00, '2025-03-10 11:45:00', '2025-10-15 08:00:00'),
    ('Eve Wilson', 'eve@example.com', 'suspended', NULL, '2025-01-25 16:20:00', '2025-09-30 10:15:00');
```

**File to create**: `scripts/setup_test_db.sh`

**Code**:
```bash
#!/bin/bash
# Create test database

DB_PATH="${1:-./data.db}"

echo "Creating test database at $DB_PATH"

# Remove existing database
rm -f "$DB_PATH"

# Create database and load schema
sqlite3 "$DB_PATH" < scripts/setup_test_db.sql

echo "Test database created successfully"
echo "Run the server with: DB_PATH=$DB_PATH go run cmd/server/main.go"
```

**Make script executable**:
```bash
chmod +x scripts/setup_test_db.sh
```

**Commit**:
```bash
git add scripts/
git commit -m "Add test database setup script"
```

---

## Phase 9: Documentation and Testing

### Task 9.1: Create README for Running the Project

**File to create**: `README.md` (update existing)

**Code**:
```markdown
# Vibe Personal Dashboard Backend

A RESTful API service for querying metrics from a database using configuration-driven metric definitions.

## Quick Start

### 1. Set up test database

```bash
./scripts/setup_test_db.sh
```

### 2. Run the server

```bash
go run cmd/server/main.go
```

The server will start on port 8080 by default.

### 3. Try the API

```bash
# List all available metrics
curl http://localhost:8080/metrics

# Get a single metric
curl http://localhost:8080/metrics/total_users

# Get multiple metrics
curl "http://localhost:8080/metrics?names=total_users,active_users"

# Get a parameterized metric
curl "http://localhost:8080/metrics/users_since_date?start_date=2025-01-01"
```

## Configuration

### Environment Variables

- `PORT` - HTTP server port (default: 8080)
- `DB_PATH` - SQLite database path (default: ./data.db)

### Metrics Configuration

Edit `config/metrics.toml` to define your metrics.

## Development

### Run tests

```bash
go test ./...
```

### Run tests with coverage

```bash
go test -cover ./...
```

### Build

```bash
go build -o bin/server ./cmd/server
```

## Architecture

See [DESIGN.md](DESIGN.md) for detailed architecture documentation.

See [IMPLEMENTATION.md](IMPLEMENTATION.md) for step-by-step implementation guide.
```

**Commit**:
```bash
git add README.md
git commit -m "Update README with usage instructions"
```

---

### Task 9.2: End-to-End Manual Test

**Commands to run**:

```bash
# 1. Set up test database
./scripts/setup_test_db.sh

# 2. Start server (in one terminal)
go run cmd/server/main.go

# 3. In another terminal, test endpoints
curl http://localhost:8080/metrics
curl http://localhost:8080/metrics/total_users
curl "http://localhost:8080/metrics?names=total_users,active_users"
curl "http://localhost:8080/metrics/users_since_date?start_date=2025-01-01"
curl http://localhost:8080/metrics/users_by_status
```

**Expected results**:
- First request returns list of metric names
- Second returns single value (count)
- Third returns two metrics
- Fourth returns array of users
- Fifth returns grouped counts

**If all work, commit**:
```bash
git commit --allow-empty -m "Manual end-to-end testing complete"
```

---

## Summary

You've now built a complete metrics API service following TDD, YAGNI, and KISS principles!

**Key achievements**:
- ✅ Test-driven development throughout
- ✅ Clean architecture with separated concerns
- ✅ Type-safe parameter handling
- ✅ Concurrent query execution
- ✅ Comprehensive error handling
- ✅ Structured logging
- ✅ Graceful shutdown
- ✅ Well-tested codebase

**Next steps (if needed)**:
- Add integration tests
- Add caching layer
- Add metrics/observability
- Deploy to production

Remember to commit frequently and keep your tests passing!
