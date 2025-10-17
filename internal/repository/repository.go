// Defines the database repository interface for metric queries.
package repository

import "context"

// Repository abstracts database operations from business logic.
type Repository interface {
	QuerySingleValue(ctx context.Context, query string, args ...interface{}) (interface{}, error)
	QueryMultiRow(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)
	Close() error
}
