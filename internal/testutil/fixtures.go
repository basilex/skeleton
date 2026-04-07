package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/identity/infrastructure/persistence"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func NewTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err, "open test db")

	err = runMigrations(db)
	require.NoError(t, err, "run migrations")

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func runMigrations(db *sqlx.DB) error {
	dir, err := filepath.Abs("../../migrations")
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("exec %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func CreateTestRole(t *testing.T, repo *persistence.RoleRepository, perms ...string) *domain.Role {
	t.Helper()

	rolePerms := make([]domain.Permission, len(perms))
	for i, p := range perms {
		perm, err := domain.NewPermission(p)
		require.NoError(t, err)
		rolePerms[i] = perm
	}

	role, err := domain.NewRole("test_role", "Test role", rolePerms)
	require.NoError(t, err)

	ctx := context.Background()
	err = repo.Save(ctx, role)
	require.NoError(t, err)

	return role
}

func CreateTestUser(t *testing.T, userRepo *persistence.UserRepository) *domain.User {
	t.Helper()

	email, err := domain.NewEmail("test@example.com")
	require.NoError(t, err)

	passwordHash, err := domain.NewPasswordHash("TestPass1234!")
	require.NoError(t, err)

	user, err := domain.NewUser(email, passwordHash)
	require.NoError(t, err)

	ctx := context.Background()
	err = userRepo.Save(ctx, user)
	require.NoError(t, err)

	return user
}
