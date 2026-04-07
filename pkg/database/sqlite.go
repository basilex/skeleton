package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

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
