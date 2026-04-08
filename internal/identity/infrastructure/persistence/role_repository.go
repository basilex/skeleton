package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepository struct {
	pool *pgxpool.Pool
}

func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

func (r *RoleRepository) Save(ctx context.Context, role *domain.Role) error {
	query := `
		INSERT INTO roles (id, name, description, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT(id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description
	`
	_, err := r.pool.Exec(ctx, query,
		role.ID(),
		role.Name(),
		role.Description(),
		role.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("save role: %w", err)
	}
	return nil
}

func (r *RoleRepository) FindByID(ctx context.Context, id domain.RoleID) (*domain.Role, error) {
	query := `SELECT id, name, description, created_at FROM roles WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	var roleID, name, description string
	var createdAt time.Time

	err := row.Scan(&roleID, &name, &description, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("find role by id: %w", err)
	}

	return r.loadRoleWithPermissions(ctx, roleID, name, description, createdAt)
}

func (r *RoleRepository) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	query := `SELECT id, name, description, created_at FROM roles WHERE name = $1`
	row := r.pool.QueryRow(ctx, query, name)

	var roleID, roleName, description string
	var createdAt time.Time

	err := row.Scan(&roleID, &roleName, &description, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("find role by name: %w", err)
	}

	return r.loadRoleWithPermissions(ctx, roleID, roleName, description, createdAt)
}

func (r *RoleRepository) FindAll(ctx context.Context) ([]*domain.Role, error) {
	query := `SELECT id, name, description, created_at FROM roles ORDER BY name`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all roles: %w", err)
	}
	defer rows.Close()

	roles := make([]*domain.Role, 0)
	for rows.Next() {
		var roleID, name, description string
		var createdAt time.Time

		if err := rows.Scan(&roleID, &name, &description, &createdAt); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}

		role, err := r.loadRoleWithPermissions(ctx, roleID, name, description, createdAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return roles, nil
}

func (r *RoleRepository) FindByIDs(ctx context.Context, ids []domain.RoleID) ([]*domain.Role, error) {
	if len(ids) == 0 {
		return []*domain.Role{}, nil
	}

	query := `SELECT id, name, description, created_at FROM roles WHERE id = ANY($1)`
	pgIDs := make([]string, len(ids))
	for i, id := range ids {
		pgIDs[i] = id.String()
	}

	rows, err := r.pool.Query(ctx, query, pgIDs)
	if err != nil {
		return nil, fmt.Errorf("find roles by ids: %w", err)
	}
	defer rows.Close()

	roles := make([]*domain.Role, 0, len(ids))
	for rows.Next() {
		var roleID, name, description string
		var createdAt time.Time

		if err := rows.Scan(&roleID, &name, &description, &createdAt); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}

		role, err := r.loadRoleWithPermissions(ctx, roleID, name, description, createdAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return roles, nil
}

func (r *RoleRepository) loadRoleWithPermissions(ctx context.Context, roleID, name, description string, createdAt time.Time) (*domain.Role, error) {
	roleIDParsed, err := domain.ParseRoleID(roleID)
	if err != nil {
		return nil, fmt.Errorf("parse role id: %w", err)
	}
	perms, err := r.loadPermissions(ctx, roleIDParsed)
	if err != nil {
		return nil, fmt.Errorf("load permissions: %w", err)
	}
	role, err := domain.NewRoleWithID(roleIDParsed, name, description, perms, createdAt)
	if err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}
	return role, nil
}

func (r *RoleRepository) loadPermissions(ctx context.Context, roleID domain.RoleID) ([]domain.Permission, error) {
	query := `
		SELECT p.name FROM permissions p
		INNER JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1
	`
	rows, err := r.pool.Query(ctx, query, roleID.String())
	if err != nil {
		return nil, fmt.Errorf("load permissions: %w", err)
	}
	defer rows.Close()

	perms := make([]domain.Permission, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		perms = append(perms, domain.Permission(name))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return perms, nil
}
