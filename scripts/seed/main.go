package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/basilex/skeleton/pkg/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/skeleton?sslmode=disable"
	}

	ctx := context.Background()

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("parse database url: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	seedRoles(ctx, pool)
	seedPermissions(ctx, pool)
	seedRolePermissions(ctx, pool)
	seedAdminUser(ctx, pool)

	log.Println("✓ seed completed successfully")
}

func seedRoles(ctx context.Context, pool *pgxpool.Pool) {
	roles := []struct {
		id, name, desc string
	}{
		{uuid.NewV7().String(), "super_admin", "Full access to everything"},
		{uuid.NewV7().String(), "admin", "Admin access"},
		{uuid.NewV7().String(), "viewer", "Read-only access"},
	}

	for _, r := range roles {
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM roles WHERE name = $1)`,
			r.name,
		).Scan(&exists)
		if err != nil {
			log.Printf("check role %s: %v", r.name, err)
			continue
		}
		if exists {
			log.Printf("role %s already exists, skipping", r.name)
			continue
		}

		_, err = pool.Exec(ctx,
			`INSERT INTO roles (id, name, description, created_at) VALUES ($1, $2, $3, $4)`,
			r.id, r.name, r.desc, time.Now(),
		)
		if err != nil {
			log.Printf("insert role %s: %v", r.name, err)
			continue
		}
		log.Printf("✓ created role: %s", r.name)
	}
}

func seedPermissions(ctx context.Context, pool *pgxpool.Pool) {
	perms := []struct {
		name, resource, action string
	}{
		{"users:read", "users", "read"},
		{"users:write", "users", "write"},
		{"users:delete", "users", "delete"},
		{"roles:read", "roles", "read"},
		{"roles:manage", "roles", "manage"},
		{"files:read", "files", "read"},
		{"files:write", "files", "write"},
		{"files:delete", "files", "delete"},
		{"audit:read", "audit", "read"},
		{"notifications:read", "notifications", "read"},
		{"notifications:write", "notifications", "write"},
	}

	for _, p := range perms {
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM permissions WHERE name = $1)`,
			p.name,
		).Scan(&exists)
		if err != nil {
			log.Printf("check permission %s: %v", p.name, err)
			continue
		}
		if exists {
			continue
		}

		_, err = pool.Exec(ctx,
			`INSERT INTO permissions (name, description) VALUES ($1, $2)`,
			p.name, p.name,
		)
		if err != nil {
			log.Printf("insert permission %s: %v", p.name, err)
			continue
		}
		log.Printf("✓ created permission: %s", p.name)
	}
}

func seedRolePermissions(ctx context.Context, pool *pgxpool.Pool) {
	var superAdminID, adminID, viewerID string
	pool.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'super_admin'`).Scan(&superAdminID)
	pool.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'admin'`).Scan(&adminID)
	pool.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'viewer'`).Scan(&viewerID)

	// Get permission IDs
	permIDs := make(map[string]int)
	for _, name := range []string{"users:read", "users:write", "users:delete", "roles:read", "roles:manage", "files:read", "files:write", "files:delete", "audit:read", "notifications:read", "notifications:write"} {
		var id int
		if err := pool.QueryRow(ctx, `SELECT id FROM permissions WHERE name = $1`, name).Scan(&id); err == nil {
			permIDs[name] = id
		}
	}

	// Super admin has all permissions (wildcard in code)
	if superAdminID != "" {
		log.Printf("super_admin has wildcard permission (*:*)")
	}

	// Admin permissions
	adminPerms := []string{"users:read", "users:write", "roles:read", "roles:manage", "files:read", "files:write", "audit:read", "notifications:read", "notifications:write"}
	for _, permName := range adminPerms {
		if permID, ok := permIDs[permName]; ok && adminID != "" {
			insertRolePerm(ctx, pool, adminID, permID)
		}
	}

	// Viewer permissions
	viewerPerms := []string{"users:read", "files:read", "audit:read", "notifications:read"}
	for _, permName := range viewerPerms {
		if permID, ok := permIDs[permName]; ok && viewerID != "" {
			insertRolePerm(ctx, pool, viewerID, permID)
		}
	}
}

func insertRolePerm(ctx context.Context, pool *pgxpool.Pool, roleID string, permissionID int) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM role_permissions WHERE role_id = $1 AND permission_id = $2)`,
		roleID, permissionID,
	).Scan(&exists)
	if err != nil {
		log.Printf("check role_permission: %v", err)
		return
	}
	if exists {
		return
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`,
		roleID, permissionID,
	)
	if err != nil {
		log.Printf("insert role_permission: %v", err)
	}
}

func seedAdminUser(ctx context.Context, pool *pgxpool.Pool) {
	email := "admin@skeleton.local"
	var userID string
	var exists bool

	// Check if user exists
	err := pool.QueryRow(ctx,
		`SELECT id FROM users WHERE email = $1`,
		email,
	).Scan(&userID)
	if err != nil && err != pgx.ErrNoRows {
		log.Printf("check admin user: %v", err)
		return
	}

	if userID != "" {
		log.Println("admin user already exists, checking role assignment")
	} else {
		// Create admin user
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Admin1234!"), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("hash password: %v", err)
			return
		}

		userID = uuid.NewV7().String()
		now := time.Now()

		_, err = pool.Exec(ctx,
			`INSERT INTO users (id, email, password_hash, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			userID, email, string(hashedPassword), true, now, now,
		)
		if err != nil {
			log.Printf("insert admin user: %v", err)
			return
		}
		log.Printf("✓ created admin user: %s", email)
	}

	// Assign super_admin role
	var superAdminID string
	err = pool.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'super_admin'`).Scan(&superAdminID)
	if err != nil {
		log.Printf("get super_admin role: %v", err)
		return
	}

	// Check if role already assigned
	err = pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id = $1 AND role_id = $2)`,
		userID, superAdminID,
	).Scan(&exists)
	if err != nil {
		log.Printf("check role assignment: %v", err)
		return
	}

	if exists {
		log.Println("super_admin role already assigned")
		return
	}

	now := time.Now()
	_, err = pool.Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id, created_at) VALUES ($1, $2, $3)`,
		userID, superAdminID, now,
	)
	if err != nil {
		log.Printf("assign super_admin role: %v", err)
		return
	}
	log.Println("✓ assigned super_admin role to admin user")
}
