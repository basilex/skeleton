// Package testutil provides testing utilities for the application.
// It includes testcontainers-based PostgreSQL setup for integration tests.
package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer holds the test container configuration.
type PostgresContainer struct {
	testcontainers.Container
	ConnectionString string
}

// SetupPostgres creates a PostgreSQL test container for integration testing.
// It returns configured pgxpool.Pool ready for use in tests.
// The container and pool are automatically cleaned up when the test completes.
//
// Example:
//
//	func TestUserRepository_Save(t *testing.T) {
//		pool := testutil.SetupPostgres(t)
//		repo := persistence.NewUserRepository(pool)
//		// use repo...
//	}
func SetupPostgres(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_USER":     "testuser",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "failed to start postgres container")

	// Get connection details
	host, err := container.Host(ctx)
	require.NoError(t, err, "failed to get container host")

	mappedPort, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err, "failed to get mapped port")

	connectionString := fmt.Sprintf(
		"postgres://testuser:testpass@%s:%s/testdb?sslmode=disable",
		host, mappedPort.Port(),
	)

	// Create connection pool
	poolConfig, err := pgxpool.ParseConfig(connectionString)
	require.NoError(t, err, "failed to parse connection string")

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	require.NoError(t, err, "failed to create connection pool")

	// Wait for database to be ready
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	require.NoError(t, pool.Ping(ctx), "failed to ping database")

	// Cleanup
	t.Cleanup(func() {
		pool.Close()
		if err := container.Terminate(ctx); err != nil {
			t.Logf("warning: failed to terminate container: %v", err)
		}
	})

	return pool
}

// RunMigrations executes database migrations for tests.
// This should be called after SetupPostgres to create required tables.
func RunMigrations(t *testing.T, pool *pgxpool.Pool, schema string) {
	ctx := context.Background()

	_, err := pool.Exec(ctx, schema)
	require.NoError(t, err, "failed to run migrations")
}
