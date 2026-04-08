package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	action := flag.String("action", "up", "migration action: up, down, status")
	flag.Parse()

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

	dir, err := filepath.Abs("./migrations")
	if err != nil {
		log.Fatalf("resolve migrations path: %v", err)
	}

	if err := ensureMigrationTable(ctx, pool); err != nil {
		log.Fatalf("ensure migration table: %v", err)
	}

	switch *action {
	case "up":
		if err := runMigrationsUp(ctx, pool, dir); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
		log.Println("✓ migrations applied successfully")
	case "down":
		if err := runMigrationsDown(ctx, pool, dir); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
		log.Println("✓ migrations rolled back successfully")
	case "status":
		version, dirty, err := getMigrationStatus(ctx, pool)
		if err != nil {
			log.Fatalf("get migration status: %v", err)
		}
		fmt.Printf("Current version: %d (dirty: %v)\n", version, dirty)
	default:
		log.Fatalf("unknown action: %s (use: up, down, status)", *action)
	}
}

func ensureMigrationTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			dirty   BOOLEAN NOT NULL DEFAULT false,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func getMigrationStatus(ctx context.Context, pool *pgxpool.Pool) (version int, dirty bool, err error) {
	err = pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(version), 0), COALESCE(BOOL_OR(dirty), false)
		FROM schema_migrations
	`).Scan(&version, &dirty)
	return
}

func runMigrationsUp(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	current, _, err := getMigrationStatus(ctx, pool)
	if err != nil {
		return fmt.Errorf("get current version: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	applied := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		var version int
		fmt.Sscanf(name, "%d_", &version)
		if version <= current {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}

		// Start transaction
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin transaction: %w", err)
		}

		// Execute migration
		if _, err := tx.Exec(ctx, string(content)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("exec %s: %w", name, err)
		}

		// Record migration
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version, dirty) VALUES ($1, false)`,
			version,
		); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("record migration %d: %w", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %d: %w", version, err)
		}

		log.Printf("✓ applied migration: %s", name)
		applied++
	}

	if applied == 0 {
		log.Println("no pending migrations")
	}

	return nil
}

func runMigrationsDown(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	current, _, err := getMigrationStatus(ctx, pool)
	if err != nil {
		return fmt.Errorf("get current version: %w", err)
	}
	if current == 0 {
		log.Println("no migrations to roll back")
		return nil
	}

	name := fmt.Sprintf("%03d_", current)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), name) && strings.HasSuffix(entry.Name(), ".down.sql") {
			content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				return fmt.Errorf("read %s: %w", entry.Name(), err)
			}

			// Start transaction
			tx, err := pool.Begin(ctx)
			if err != nil {
				return fmt.Errorf("begin transaction: %w", err)
			}

			// Execute down migration
			if _, err := tx.Exec(ctx, string(content)); err != nil {
				tx.Rollback(ctx)
				return fmt.Errorf("exec %s: %w", entry.Name(), err)
			}

			// Remove migration record
			if _, err := tx.Exec(ctx,
				`DELETE FROM schema_migrations WHERE version = $1`,
				current,
			); err != nil {
				tx.Rollback(ctx)
				return fmt.Errorf("remove migration record: %w", err)
			}

			if err := tx.Commit(ctx); err != nil {
				return fmt.Errorf("commit rollback: %w", err)
			}

			log.Printf("✓ rolled back migration: %s", entry.Name())
			return nil
		}
	}

	return fmt.Errorf("down migration for version %d not found", current)
}
