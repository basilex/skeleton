package transaction

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// contextKey is the type for context keys in this package.
type contextKey struct{}

// txKey is the context key for storing the current transaction.
var txKey contextKey = contextKey{}

// PgxTransactionManager implements the Manager interface using pgxpool.
// It provides transaction management for PostgreSQL databases.
type PgxTransactionManager struct {
	pool *pgxpool.Pool
}

// NewPgxTransactionManager creates a new transaction manager using pgxpool.
func NewPgxTransactionManager(pool *pgxpool.Pool) *PgxTransactionManager {
	return &PgxTransactionManager{pool: pool}
}

// Execute executes a function within a transaction.
// It begins a transaction, executes the function, and either commits on success
// or rolls back on error/panic.
func (m *PgxTransactionManager) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	// Begin transaction
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Ensure rollback on panic
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) // re-throw panic after rollback
		}
	}()

	// Create context with transaction
	txCtx := context.WithValue(ctx, txKey, tx)

	// Execute function
	if err := fn(txCtx); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rollback: %v, original: %w", rbErr, err)
		}
		return err
	}

	// Commit on success
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// ExecuteWithResult executes a function within a transaction and returns a result.
// It begins a transaction, executes the function, and either commits on success
// or rolls back on error/panic.
func (m *PgxTransactionManager) ExecuteWithResult(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	var result interface{}

	err := m.Execute(ctx, func(ctx context.Context) error {
		res, err := fn(ctx)
		if err != nil {
			return err
		}
		result = res
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// FromContext extracts the transaction from the context.
// Returns the transaction and true if found, or nil and false if not found.
// This is used by repositories to use the transaction from the context.
func FromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	return tx, ok
}
