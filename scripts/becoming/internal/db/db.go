// Package db provides database operations for the Becoming app.
package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

// DB wraps the SQLite database connection.
type DB struct {
	*sql.DB
	path string
}

// Open opens or creates the database at the specified path.
func Open(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	dsn := fmt.Sprintf("%s?_foreign_keys=on&_journal_mode=WAL", path)
	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	db := &DB{DB: sqlDB, path: path}

	if err := db.initSchema(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("initializing schema: %w", err)
	}

	return db, nil
}

func (db *DB) initSchema() error {
	if _, err := db.Exec(schemaSQL); err != nil {
		return err
	}
	return db.runMigrations()
}

func (db *DB) runMigrations() error {
	// Rename "exercise" type to "tracker"
	_, err := db.Exec(`UPDATE practices SET type = 'tracker' WHERE type = 'exercise'`)
	return err
}

// Path returns the database file path.
func (db *DB) Path() string {
	return db.path
}
