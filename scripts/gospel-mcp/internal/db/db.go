// Package db provides database operations for the Gospel MCP server.
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
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	// Open with foreign keys enabled and WAL mode for better concurrency
	dsn := fmt.Sprintf("%s?_foreign_keys=on&_journal_mode=WAL", path)
	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	db := &DB{DB: sqlDB, path: path}

	// Initialize schema
	if err := db.initSchema(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("initializing schema: %w", err)
	}

	return db, nil
}

// initSchema creates tables if they don't exist.
func (db *DB) initSchema() error {
	_, err := db.Exec(schemaSQL)
	return err
}

// Path returns the database file path.
func (db *DB) Path() string {
	return db.path
}

// Reset drops all data and recreates the schema (for full reindex).
func (db *DB) Reset() error {
	// Drop all tables in reverse dependency order
	tables := []string{
		"index_metadata",
		"cross_references",
		"books_fts",
		"books",
		"manuals_fts",
		"manuals",
		"talks_fts",
		"talks",
		"chapters",
		"scriptures_fts",
		"scriptures",
		"schema_version",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
		if err != nil {
			return fmt.Errorf("dropping table %s: %w", table, err)
		}
	}

	// Recreate schema
	return db.initSchema()
}

// Stats returns statistics about indexed content.
type Stats struct {
	Scriptures int64
	Chapters   int64
	Talks      int64
	Manuals    int64
	Books      int64
	CrossRefs  int64
}

// GetStats returns counts of indexed content.
func (db *DB) GetStats() (*Stats, error) {
	stats := &Stats{}

	queries := []struct {
		table string
		dest  *int64
	}{
		{"scriptures", &stats.Scriptures},
		{"chapters", &stats.Chapters},
		{"talks", &stats.Talks},
		{"manuals", &stats.Manuals},
		{"books", &stats.Books},
		{"cross_references", &stats.CrossRefs},
	}

	for _, q := range queries {
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", q.table)).Scan(q.dest)
		if err != nil {
			return nil, fmt.Errorf("counting %s: %w", q.table, err)
		}
	}

	return stats, nil
}
