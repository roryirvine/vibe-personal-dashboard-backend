// Implements the repository interface using SQLite.
package repository

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a SQLite repository.
// Path can be a file path or ":memory:" for an in-memory database.
func NewSQLiteRepository(path string) (Repository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Verify connection
	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &SQLiteRepository{db: db}, nil
}

func (r *SQLiteRepository) QuerySingleValue(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	var value interface{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no rows returned")
		}
		return nil, fmt.Errorf("query failed: %w", err)
	}
	return value, nil
}

func (r *SQLiteRepository) QueryMultiRow(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
