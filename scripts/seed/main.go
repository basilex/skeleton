package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/basilex/skeleton/pkg/uuid"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/skeleton.db"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	seedRoles(ctx, db)
	seedPermissions(ctx, db)
	seedRolePermissions(ctx, db)
	seedAdminUser(ctx, db)

	log.Println("seed completed")
}

func seedRoles(ctx context.Context, db *sql.DB) {
	roles := []struct {
		id, name, desc string
	}{
		{uuid.NewV7().String(), "super_admin", "Full access to everything"},
		{uuid.NewV7().String(), "admin", "Admin access"},
		{uuid.NewV7().String(), "viewer", "Read-only access"},
	}

	for _, r := range roles {
		var exists bool
		err := db.QueryRowContext(ctx, `SELECT COUNT(*) > 0 FROM roles WHERE name = ?`, r.name).Scan(&exists)
		if err != nil {
			log.Printf("check role %s: %v", r.name, err)
			continue
		}
		if exists {
			log.Printf("role %s already exists, skipping", r.name)
			continue
		}
		_, err = db.ExecContext(ctx,
			`INSERT INTO roles (id, name, description, created_at) VALUES (?, ?, ?, ?)`,
			r.id, r.name, r.desc, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			log.Printf("insert role %s: %v", r.name, err)
			continue
		}
		log.Printf("created role: %s", r.name)
	}
}

func seedPermissions(ctx context.Context, db *sql.DB) {
	perms := []struct {
		id, name, resource, action string
	}{
		{uuid.NewV7().String(), "users:read", "users", "read"},
		{uuid.NewV7().String(), "users:write", "users", "write"},
		{uuid.NewV7().String(), "roles:read", "roles", "read"},
		{uuid.NewV7().String(), "roles:manage", "roles", "manage"},
		{uuid.NewV7().String(), "audit:read", "audit", "read"},
	}

	for _, p := range perms {
		var exists bool
		err := db.QueryRowContext(ctx, `SELECT COUNT(*) > 0 FROM permissions WHERE name = ?`, p.name).Scan(&exists)
		if err != nil {
			log.Printf("check permission %s: %v", p.name, err)
			continue
		}
		if exists {
			log.Printf("permission %s already exists, skipping", p.name)
			continue
		}
		_, err = db.ExecContext(ctx,
			`INSERT INTO permissions (id, name, resource, action) VALUES (?, ?, ?, ?)`,
			p.id, p.name, p.resource, p.action)
		if err != nil {
			log.Printf("insert permission %s: %v", p.name, err)
			continue
		}
		log.Printf("created permission: %s", p.name)
	}
}

func seedRolePermissions(ctx context.Context, db *sql.DB) {
	var superAdminID, adminID, viewerID string
	db.QueryRowContext(ctx, `SELECT id FROM roles WHERE name = 'super_admin'`).Scan(&superAdminID)
	db.QueryRowContext(ctx, `SELECT id FROM roles WHERE name = 'admin'`).Scan(&adminID)
	db.QueryRowContext(ctx, `SELECT id FROM roles WHERE name = 'viewer'`).Scan(&viewerID)

	var usersRead, usersWrite, rolesRead, rolesManage, auditRead string
	db.QueryRowContext(ctx, `SELECT id FROM permissions WHERE name = 'users:read'`).Scan(&usersRead)
	db.QueryRowContext(ctx, `SELECT id FROM permissions WHERE name = 'users:write'`).Scan(&usersWrite)
	db.QueryRowContext(ctx, `SELECT id FROM permissions WHERE name = 'roles:read'`).Scan(&rolesRead)
	db.QueryRowContext(ctx, `SELECT id FROM permissions WHERE name = 'roles:manage'`).Scan(&rolesManage)
	db.QueryRowContext(ctx, `SELECT id FROM permissions WHERE name = 'audit:read'`).Scan(&auditRead)

	if superAdminID != "" {
		log.Printf("super_admin has *:* wildcard (handled in code)")
	}

	if adminID != "" && usersRead != "" {
		insertRolePerm(ctx, db, adminID, usersRead)
	}
	if adminID != "" && usersWrite != "" {
		insertRolePerm(ctx, db, adminID, usersWrite)
	}
	if adminID != "" && rolesRead != "" {
		insertRolePerm(ctx, db, adminID, rolesRead)
	}
	if adminID != "" && rolesManage != "" {
		insertRolePerm(ctx, db, adminID, rolesManage)
	}
	if adminID != "" && auditRead != "" {
		insertRolePerm(ctx, db, adminID, auditRead)
	}
	if viewerID != "" && usersRead != "" {
		insertRolePerm(ctx, db, viewerID, usersRead)
	}
	if viewerID != "" && auditRead != "" {
		insertRolePerm(ctx, db, viewerID, auditRead)
	}
}

func insertRolePerm(ctx context.Context, db *sql.DB, roleID, permID string) {
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) > 0 FROM role_permissions WHERE role_id = ? AND permission_id = ?`,
		roleID, permID).Scan(&exists)
	if err != nil {
		log.Printf("check role_permission: %v", err)
		return
	}
	if exists {
		return
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)`,
		roleID, permID)
	if err != nil {
		log.Printf("insert role_permission: %v", err)
	}
}

func seedAdminUser(ctx context.Context, db *sql.DB) {
	email := "admin@skeleton.local"
	var exists bool
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) > 0 FROM users WHERE email = ?`, email).Scan(&exists)
	if err != nil {
		log.Printf("check admin user: %v", err)
		return
	}
	if exists {
		log.Println("admin user already exists, skipping")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Admin1234!"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("hash password: %v", err)
		return
	}

	userID := uuid.NewV7().String()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.ExecContext(ctx,
		`INSERT INTO users (id, email, password_hash, is_active, created_at, updated_at) VALUES (?, ?, ?, 1, ?, ?)`,
		userID, email, string(hashedPassword), now, now)
	if err != nil {
		log.Printf("insert admin user: %v", err)
		return
	}
	log.Printf("created admin user: %s", email)

	var superAdminID string
	db.QueryRowContext(ctx, `SELECT id FROM roles WHERE name = 'super_admin'`).Scan(&superAdminID)
	if superAdminID != "" {
		_, err = db.ExecContext(ctx,
			`INSERT INTO user_roles (user_id, role_id, assigned_at) VALUES (?, ?, ?)`,
			userID, superAdminID, now)
		if err != nil {
			log.Printf("assign super_admin role: %v", err)
			return
		}
		log.Println("assigned super_admin role to admin user")
	}
}
