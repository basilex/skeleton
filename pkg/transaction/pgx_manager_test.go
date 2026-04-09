package transaction_test

import (
	"context"
	"errors"
	"testing"

	"github.com/basilex/skeleton/pkg/testutil"
	"github.com/basilex/skeleton/pkg/transaction"
	"github.com/stretchr/testify/require"
)

func TestPgxTransactionManager_Execute_Commit(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	tm := transaction.NewPgxTransactionManager(pool)
	ctx := context.Background()

	// Create test table
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS test_transactions (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	// Test successful transaction
	err = tm.Execute(ctx, func(ctx context.Context) error {
		// Get transaction from context
		tx, ok := transaction.FromContext(ctx)
		require.True(t, ok, "transaction should be in context")

		// Execute within transaction
		_, err := tx.Exec(ctx, "INSERT INTO test_transactions (value) VALUES ($1)", "test-value")
		return err
	})
	require.NoError(t, err, "transaction should commit successfully")

	// Verify data was committed
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM test_transactions WHERE value = $1", "test-value").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "row should be committed")
}

func TestPgxTransactionManager_Execute_Rollback(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	tm := transaction.NewPgxTransactionManager(pool)
	ctx := context.Background()

	// Create test table
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS test_transactions (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	// Test failed transaction
	testErr := errors.New("test error")
	err = tm.Execute(ctx, func(ctx context.Context) error {
		// Get transaction from context
		tx, ok := transaction.FromContext(ctx)
		require.True(t, ok, "transaction should be in context")

		// Execute within transaction
		_, err := tx.Exec(ctx, "INSERT INTO test_transactions (value) VALUES ($1)", "should-rollback")
		require.NoError(t, err)

		// Return error to trigger rollback
		return testErr
	})
	require.Error(t, err, "transaction should return error")
	require.Equal(t, testErr, err, "should return original error")

	// Verify data was rolled back
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM test_transactions WHERE value = $1", "should-rollback").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count, "row should be rolled back")
}

func TestPgxTransactionManager_Execute_MultipleOperations(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	tm := transaction.NewPgxTransactionManager(pool)
	ctx := context.Background()

	// Create test table
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS test_transactions (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	// Test multiple operations in transaction
	err = tm.Execute(ctx, func(ctx context.Context) error {
		tx, ok := transaction.FromContext(ctx)
		require.True(t, ok)

		// First operation
		_, err := tx.Exec(ctx, "INSERT INTO test_transactions (value) VALUES ($1)", "first")
		if err != nil {
			return err
		}

		// Second operation
		_, err = tx.Exec(ctx, "INSERT INTO test_transactions (value) VALUES ($1)", "second")
		if err != nil {
			return err
		}

		// Third operation
		_, err = tx.Exec(ctx, "INSERT INTO test_transactions (value) VALUES ($1)", "third")
		return err
	})
	require.NoError(t, err, "multiple operations should commit")

	// Verify all operations committed
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM test_transactions").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 3, count, "all rows should be committed")
}

func TestPgxTransactionManager_Execute_Panic(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	tm := transaction.NewPgxTransactionManager(pool)
	ctx := context.Background()

	// Create test table
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS test_transactions (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	// Test panic in transaction
	defer func() {
		r := recover()
		require.NotNil(t, r, "panic should be re-thrown")
		require.Equal(t, "test panic", r)

		// Verify data was rolled back
		var count int
		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM test_transactions WHERE value = $1", "before-panic").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 0, count, "panicked transaction should roll back")
	}()

	err = tm.Execute(ctx, func(ctx context.Context) error {
		tx, ok := transaction.FromContext(ctx)
		require.True(t, ok)

		_, err := tx.Exec(ctx, "INSERT INTO test_transactions (value) VALUES ($1)", "before-panic")
		require.NoError(t, err)

		panic("test panic")
	})

	require.Fail(t, "should not reach here")
}

func TestPgxTransactionManager_ExecuteWithResult_Success(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	tm := transaction.NewPgxTransactionManager(pool)
	ctx := context.Background()

	// Test ExecuteWithResult
	result, err := tm.ExecuteWithResult(ctx, func(ctx context.Context) (interface{}, error) {
		return "test-result", nil
	})
	require.NoError(t, err)
	require.Equal(t, "test-result", result)
}

func TestPgxTransactionManager_ExecuteWithResult_Error(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	tm := transaction.NewPgxTransactionManager(pool)
	ctx := context.Background()

	// Test ExecuteWithResult with error
	testErr := errors.New("test error")
	result, err := tm.ExecuteWithResult(ctx, func(ctx context.Context) (interface{}, error) {
		return nil, testErr
	})
	require.Error(t, err)
	require.Equal(t, testErr, err)
	require.Nil(t, result)
}

func TestFromContext_NoTransaction(t *testing.T) {
	ctx := context.Background()

	tx, ok := transaction.FromContext(ctx)
	require.False(t, ok, "should return false when no transaction in context")
	require.Nil(t, tx, "transaction should be nil")
}
