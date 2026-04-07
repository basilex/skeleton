package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

func main() {
	action := flag.String("action", "up", "migration action: up, down")
	flag.Parse()

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/skeleton.db"
	}

	dir, err := filepath.Abs("./migrations")
	if err != nil {
		log.Fatalf("resolve migrations path: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := ensureMigrationTable(db); err != nil {
		log.Fatalf("ensure migration table: %v", err)
	}

	switch *action {
	case "up":
		if err := runMigrationsUp(db, dir); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
		log.Println("migrations applied")
	case "down":
		if err := runMigrationsDown(db, dir); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
		log.Println("migrations rolled back")
	default:
		log.Fatalf("unknown action: %s", *action)
	}
}

func ensureMigrationTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		dirty   INTEGER NOT NULL DEFAULT 0
	)`)
	return err
}

func currentVersion(db *sql.DB) (int, error) {
	var v int
	err := db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations WHERE dirty = 0`).Scan(&v)
	return v, err
}

func runMigrationsUp(db *sql.DB, dir string) error {
	current, err := currentVersion(db)
	if err != nil {
		return fmt.Errorf("get current version: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

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

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("exec %s: %w", name, err)
		}

		if _, err := db.Exec(`INSERT INTO schema_migrations (version, dirty) VALUES (?, 0)`, version); err != nil {
			return fmt.Errorf("record migration %d: %w", version, err)
		}

		log.Printf("applied migration: %s", name)
	}

	return nil
}

func runMigrationsDown(db *sql.DB, dir string) error {
	current, err := currentVersion(db)
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
			if _, err := db.Exec(string(content)); err != nil {
				return fmt.Errorf("exec %s: %w", entry.Name(), err)
			}
			if _, err := db.Exec(`DELETE FROM schema_migrations WHERE version = ?`, current); err != nil {
				return fmt.Errorf("remove migration record: %w", err)
			}
			log.Printf("rolled back migration: %s", entry.Name())
			return nil
		}
	}

	return fmt.Errorf("down migration for version %d not found", current)
}
