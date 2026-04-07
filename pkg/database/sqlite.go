// Package database provides database connection utilities and abstractions.
// Currently supports SQLite with optimized settings for application use.
package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// NewSQLite creates a new SQLite database connection with optimized settings.
// It configures WAL mode, foreign keys, and connection pooling appropriate for SQLite.
// Returns a sqlx.DB instance or an error if the connection cannot be established.
func NewSQLite(dbPath string) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=ON&_busy_timeout=5000", dbPath)
	db, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}
