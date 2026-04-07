// Package persistence provides database repository implementations for the identity domain.
// This package contains SQLite-based repositories for users and roles.
package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/jmoiron/sqlx"
)

// RoleRepository implements the role repository interface using SQL database storage.
// It handles persistence of role entities and their associated permissions.
type RoleRepository struct {
	db *sqlx.DB
}

// NewRoleRepository creates a new role repository with the provided database connection.
func NewRoleRepository(db *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Save persists a role entity to the database. Uses upsert semantics to handle
// both creation and updates.
func (r *RoleRepository) Save(ctx context.Context, role *domain.Role) error {
	query := `
		INSERT INTO roles (id, name, description, created_at)
		VALUES (:id, :name, :description, :created_at)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			description = excluded.description
	`
	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":          string(role.ID()),
		"name":        role.Name(),
		"description": role.Description(),
		"created_at":  role.CreatedAt().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("save role: %w", err)
	}
	return nil
}

// FindByID retrieves a role by its unique identifier.
// Returns domain.ErrRoleNotFound if no matching role exists.
func (r *RoleRepository) FindByID(ctx context.Context, id domain.RoleID) (*domain.Role, error) {
	var row struct {
		ID          string `db:"id"`
		Name        string `db:"name"`
		Description string `db:"description"`
		CreatedAt   string `db:"created_at"`
	}
	err := r.db.GetContext(ctx, &row, `SELECT id, name, description, created_at FROM roles WHERE id = ?`, string(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrRoleNotFound
		}
		return nil, fmt.Errorf("find role by id: %w", err)
	}
	return r.scanRole(ctx, row)
}

// FindByName retrieves a role by its unique name.
// Returns domain.ErrRoleNotFound if no matching role exists.
func (r *RoleRepository) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	var row struct {
		ID          string `db:"id"`
		Name        string `db:"name"`
		Description string `db:"description"`
		CreatedAt   string `db:"created_at"`
	}
	err := r.db.GetContext(ctx, &row, `SELECT id, name, description, created_at FROM roles WHERE name = ?`, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrRoleNotFound
		}
		return nil, fmt.Errorf("find role by name: %w", err)
	}
	return r.scanRole(ctx, row)
}

// FindAll retrieves all roles from the database, ordered by name.
func (r *RoleRepository) FindAll(ctx context.Context) ([]*domain.Role, error) {
	var rows []struct {
		ID          string `db:"id"`
		Name        string `db:"name"`
		Description string `db:"description"`
		CreatedAt   string `db:"created_at"`
	}
	err := r.db.SelectContext(ctx, &rows, `SELECT id, name, description, created_at FROM roles ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("find all roles: %w", err)
	}

	roles := make([]*domain.Role, len(rows))
	for i, row := range rows {
		role, err := r.scanRole(ctx, row)
		if err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles[i] = role
	}
	return roles, nil
}

// FindByIDs retrieves multiple roles by their identifiers.
// Returns an empty slice if no IDs are provided.
func (r *RoleRepository) FindByIDs(ctx context.Context, ids []domain.RoleID) ([]*domain.Role, error) {
	if len(ids) == 0 {
		return []*domain.Role{}, nil
	}
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = string(id)
	}
	query, qArgs, err := sqlx.In(`SELECT id, name, description, created_at FROM roles WHERE id IN (?)`, args...)
	if err != nil {
		return nil, fmt.Errorf("build in query: %w", err)
	}
	query = r.db.Rebind(query)

	var rows []struct {
		ID          string `db:"id"`
		Name        string `db:"name"`
		Description string `db:"description"`
		CreatedAt   string `db:"created_at"`
	}
	if err := r.db.SelectContext(ctx, &rows, query, qArgs...); err != nil {
		return nil, fmt.Errorf("find roles by ids: %w", err)
	}

	roles := make([]*domain.Role, len(rows))
	for i, row := range rows {
		role, err := r.scanRole(ctx, row)
		if err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles[i] = role
	}
	return roles, nil
}

// scanRole converts a database row into a domain Role entity with its permissions.
func (r *RoleRepository) scanRole(ctx context.Context, row struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	CreatedAt   string `db:"created_at"`
}) (*domain.Role, error) {
	perms, err := r.loadPermissions(ctx, domain.RoleID(row.ID))
	if err != nil {
		return nil, fmt.Errorf("load permissions: %w", err)
	}
	role, err := domain.NewRole(row.Name, row.Description, perms)
	if err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}
	return role, nil
}

// loadPermissions retrieves all permissions associated with a role from the database.
func (r *RoleRepository) loadPermissions(ctx context.Context, roleID domain.RoleID) ([]domain.Permission, error) {
	var rows []struct {
		Name string `db:"name"`
	}
	err := r.db.SelectContext(ctx, &rows, `
		SELECT p.name FROM permissions p
		INNER JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = ?
	`, string(roleID))
	if err != nil {
		return nil, fmt.Errorf("load permissions: %w", err)
	}
	perms := make([]domain.Permission, len(rows))
	for i, row := range rows {
		perms[i] = domain.Permission(row.Name)
	}
	return perms, nil
}
