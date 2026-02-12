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

//go:embed auth_schema.sql
var authSchemaSQL string

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
		return fmt.Errorf("main schema: %w", err)
	}
	if _, err := db.Exec(authSchemaSQL); err != nil {
		return fmt.Errorf("auth schema: %w", err)
	}
	return db.runMigrations()
}

func (db *DB) runMigrations() error {
	// Rename "exercise" type to "tracker"
	if _, err := db.Exec(`UPDATE practices SET type = 'tracker' WHERE type = 'exercise'`); err != nil {
		return err
	}

	// Migration: add user_id columns for multi-user support
	if err := db.migrateAddUserID(); err != nil {
		return fmt.Errorf("adding user_id columns: %w", err)
	}

	// Seed default reflection prompts for user 1 (dev user)
	if err := db.SeedPrompts(1); err != nil {
		return fmt.Errorf("seeding prompts: %w", err)
	}

	return nil
}

// migrateAddUserID adds user_id columns to tables that don't have them yet.
func (db *DB) migrateAddUserID() error {
	tables := []string{"practices", "tasks", "notes", "prompts", "pillars"}
	for _, table := range tables {
		if !db.hasColumn(table, "user_id") {
			// SQLite doesn't allow REFERENCES with non-NULL default in ALTER TABLE.
			// Foreign key is defined in schema.sql for new databases; migration just adds the column.
			_, err := db.Exec(fmt.Sprintf(
				`ALTER TABLE %s ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1`, table))
			if err != nil {
				return fmt.Errorf("adding user_id to %s: %w", table, err)
			}
		}
	}

	// Reflections needs special handling: UNIQUE(date) → UNIQUE(user_id, date)
	if !db.hasColumn("reflections", "user_id") {
		if err := db.migrateReflectionsTable(); err != nil {
			return fmt.Errorf("migrating reflections: %w", err)
		}
	}

	// Create user_id indexes (idempotent — runs after columns exist)
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_practices_user ON practices(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_user ON tasks(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_notes_user ON notes(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_prompts_user ON prompts(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_reflections_user ON reflections(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pillars_user ON pillars(user_id)`,
	}
	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			return fmt.Errorf("creating index: %w", err)
		}
	}

	return nil
}

// hasColumn checks if a table has a specific column.
func (db *DB) hasColumn(table, column string) bool {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt *string
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false
		}
		if name == column {
			return true
		}
	}
	return false
}

// migrateReflectionsTable recreates the reflections table with user_id and UNIQUE(user_id, date).
func (db *DB) migrateReflectionsTable() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmts := []string{
		`CREATE TABLE reflections_new (
			id          INTEGER PRIMARY KEY,
			user_id     INTEGER NOT NULL DEFAULT 1 REFERENCES users(id),
			date        DATE NOT NULL,
			prompt_id   INTEGER REFERENCES prompts(id) ON DELETE SET NULL,
			prompt_text TEXT,
			content     TEXT NOT NULL,
			mood        INTEGER,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, date)
		)`,
		`INSERT INTO reflections_new (id, user_id, date, prompt_id, prompt_text, content, mood, created_at, updated_at)
		 SELECT id, 1, date, prompt_id, prompt_text, content, mood, created_at, updated_at FROM reflections`,
		`DROP TABLE reflections`,
		`ALTER TABLE reflections_new RENAME TO reflections`,
		`CREATE INDEX IF NOT EXISTS idx_reflections_date ON reflections(date)`,
		`CREATE INDEX IF NOT EXISTS idx_reflections_user ON reflections(user_id)`,
	}

	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("migrating reflections: %w", err)
		}
	}

	return tx.Commit()
}

// Path returns the database file path.
func (db *DB) Path() string {
	return db.path
}
