// Package transaction provides transaction management for database operations.
// It implements the Unit of Work pattern to ensure atomic operations across
// multiple aggregates and repositories.
package transaction

import "context"

// Manager defines the interface for managing database transactions.
// It provides a way to execute multiple operations atomically.
type Manager interface {
	// Execute executes a function within a transaction.
	// If the function returns an error, the transaction is rolled back.
	// If the function returns nil, the transaction is committed.
	Execute(ctx context.Context, fn func(ctx context.Context) error) error

	// ExecuteWithResult executes a function within a transaction and returns a result.
	// If the function returns an error, the transaction is rolled back.
	// If the function returns nil, the transaction is committed.
	ExecuteWithResult(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)
}
