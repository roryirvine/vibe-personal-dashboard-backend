package repository

import (
	"context"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) Repository {
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create test repository: %v", err)
	}

	// Create test table with various column types
	_, err = repo.(*SQLiteRepository).db.Exec(`
		CREATE TABLE test_data (
			id INTEGER PRIMARY KEY,
			name TEXT,
			count INTEGER,
			amount REAL,
			optional TEXT
		)
	`)
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	// Insert test data including NULL values
	_, err = repo.(*SQLiteRepository).db.Exec(`
		INSERT INTO test_data (id, name, count, amount, optional) VALUES
		(1, 'Alice', 100, 50.5, 'value1'),
		(2, 'Bob', 200, 100.25, NULL),
		(3, 'Charlie', 300, 150.75, 'value3')
	`)
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	return repo
}

func TestNewSQLiteRepository_Memory(t *testing.T) {
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory repository: %v", err)
	}
	defer repo.Close()

	if repo == nil {
		t.Error("expected non-nil repository")
	}
}

func TestNewSQLiteRepository_BadPath(t *testing.T) {
	// Try to open a database in a nonexistent directory
	_, err := NewSQLiteRepository("/nonexistent/path/db.sqlite")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestQuerySingleValue_Integer(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	val, err := repo.QuerySingleValue(context.Background(), "SELECT COUNT(*) FROM test_data")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if val != int64(3) {
		t.Errorf("expected 3, got %v (type %T)", val, val)
	}
}

func TestQuerySingleValue_String(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	val, err := repo.QuerySingleValue(context.Background(), "SELECT name FROM test_data WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if val != "Alice" {
		t.Errorf("expected 'Alice', got %v", val)
	}
}

func TestQuerySingleValue_Float(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	val, err := repo.QuerySingleValue(context.Background(), "SELECT SUM(amount) FROM test_data")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	expected := 301.5
	if val != expected {
		t.Errorf("expected %v, got %v", expected, val)
	}
}

func TestQuerySingleValue_NoRows(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	_, err := repo.QuerySingleValue(context.Background(), "SELECT name FROM test_data WHERE id = ?", 999)
	if err == nil {
		t.Error("expected error for no rows")
	}
}

func TestQuerySingleValue_WithTimeout(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	val, err := repo.QuerySingleValue(ctx, "SELECT COUNT(*) FROM test_data")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if val != int64(3) {
		t.Errorf("expected 3, got %v", val)
	}
}

func TestQueryMultiRow_AllRows(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	rows, err := repo.QueryMultiRow(context.Background(), "SELECT id, name, count FROM test_data ORDER BY id")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if len(rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(rows))
	}

	if rows[0]["name"] != "Alice" {
		t.Errorf("expected 'Alice', got %v", rows[0]["name"])
	}
	if rows[1]["name"] != "Bob" {
		t.Errorf("expected 'Bob', got %v", rows[1]["name"])
	}
}

func TestQueryMultiRow_WithFilter(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	rows, err := repo.QueryMultiRow(context.Background(), "SELECT name, count FROM test_data WHERE count > ?", 150)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}
}

func TestQueryMultiRow_NoRows(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	rows, err := repo.QueryMultiRow(context.Background(), "SELECT name FROM test_data WHERE id = ?", 999)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("expected empty slice, got %d rows", len(rows))
	}
}

func TestQueryMultiRow_NullHandling(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	rows, err := repo.QueryMultiRow(context.Background(), "SELECT id, optional FROM test_data ORDER BY id")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	// First row has a value
	if rows[0]["optional"] != "value1" {
		t.Errorf("expected 'value1', got %v", rows[0]["optional"])
	}

	// Second row has NULL (should be nil)
	if rows[1]["optional"] != nil {
		t.Errorf("expected nil for NULL value, got %v", rows[1]["optional"])
	}

	// Third row has a value
	if rows[2]["optional"] != "value3" {
		t.Errorf("expected 'value3', got %v", rows[2]["optional"])
	}
}

func TestQueryMultiRow_ColumnTypes(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	rows, err := repo.QueryMultiRow(context.Background(), "SELECT id, count, amount FROM test_data WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	row := rows[0]

	// Verify types
	idVal := row["id"]
	if _, ok := idVal.(int64); !ok {
		t.Errorf("expected int64 for id, got %T", idVal)
	}

	countVal := row["count"]
	if _, ok := countVal.(int64); !ok {
		t.Errorf("expected int64 for count, got %T", countVal)
	}

	amountVal := row["amount"]
	if _, ok := amountVal.(float64); !ok {
		t.Errorf("expected float64 for amount, got %T", amountVal)
	}
}

func TestQueryMultiRow_ColumnNames(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	rows, err := repo.QueryMultiRow(context.Background(), "SELECT id, name, count FROM test_data WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	row := rows[0]

	// Verify all expected columns are present
	if _, ok := row["id"]; !ok {
		t.Error("expected 'id' column")
	}
	if _, ok := row["name"]; !ok {
		t.Error("expected 'name' column")
	}
	if _, ok := row["count"]; !ok {
		t.Error("expected 'count' column")
	}

	// Verify no unexpected columns
	if len(row) != 3 {
		t.Errorf("expected 3 columns, got %d", len(row))
	}
}

func TestClose(t *testing.T) {
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	err = repo.Close()
	if err != nil {
		t.Fatalf("failed to close repository: %v", err)
	}

	// Verify subsequent queries fail after close
	_, err = repo.QuerySingleValue(context.Background(), "SELECT 1")
	if err == nil {
		t.Error("expected error when querying closed database")
	}
}
