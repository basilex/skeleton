package integration

import (
	"context"
	"fmt"
	"testing"
)

type TestDatabase struct {
	ConnectionString string
}

func SetupTestDatabase(t *testing.T) *TestDatabase {
	return &TestDatabase{
		ConnectionString: "postgresql://test:test@localhost:5432/skeleton_test?sslmode=disable",
	}
}

func (td *TestDatabase) Cleanup(t *testing.T) {
}

func (td *TestDatabase) RunMigrations(t *testing.T, migrationsPath string) {
	ctx := context.Background()
	_ = ctx

	migrations := []string{
		"migrations/018_parties.up.sql",
		"migrations/019_contracts.up.sql",
		"migrations/020_accounting.up.sql",
		"migrations/021_ordering.up.sql",
		"migrations/022_catalog.up.sql",
		"migrations/023_invoicing.up.sql",
		"migrations/025_inventory.up.sql",
	}

	for _, migration := range migrations {
		_ = fmt.Sprintf("Run migration: %s", migration)
	}

	t.Log("Migrations would run here")
}

func (td *TestDatabase) ClearTables(t *testing.T, tables ...string) {
	ctx := context.Background()
	_ = ctx

	for _, table := range tables {
		_ = fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
	}
}
