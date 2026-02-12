// Package db provides database operations for the Becoming app.
package db

import (
	"database/sql"
	"embed"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib" // register "pgx" driver
	_ "github.com/mattn/go-sqlite3"    // register "sqlite3" driver
	"github.com/pressly/goose/v3"
)

//go:embed schema.sql
var schemaSQL string

//go:embed auth_schema.sql
var authSchemaSQL string

// Driver identifies the database backend.
const (
	DriverSQLite   = "sqlite3"
	DriverPostgres = "pgx"
)

// DB wraps the database connection with driver-aware helpers.
type DB struct {
	conn   *sql.DB
	driver string
	path   string
}

// Driver returns the current database driver name.
func (db *DB) Driver() string { return db.driver }

// IsPostgres returns true if using PostgreSQL.
func (db *DB) IsPostgres() bool { return db.driver == DriverPostgres }

// Open opens the database. If dsn starts with "postgres://" or "postgresql://",
// it uses PostgreSQL (pgx); otherwise, it treats the dsn as a SQLite file path.
func Open(dsn string) (*DB, error) {
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		return openPostgres(dsn)
	}
	return openSQLite(dsn)
}

func openSQLite(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	connStr := fmt.Sprintf("%s?_foreign_keys=on&_journal_mode=WAL", path)
	sqlDB, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	db := &DB{conn: sqlDB, driver: DriverSQLite, path: path}
	if err := db.initSQLiteSchema(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("initializing schema: %w", err)
	}
	return db, nil
}

func openPostgres(dsn string) (*DB, error) {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening postgres: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}

	db := &DB{conn: sqlDB, driver: DriverPostgres, path: dsn}
	if err := db.initPostgresSchema(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("initializing postgres schema: %w", err)
	}
	return db, nil
}

// --- Query helpers with automatic placeholder rebinding ---

// rebind converts ? placeholders to $1, $2, ... for PostgreSQL.
// For SQLite, it returns the query unchanged.
func (db *DB) rebind(query string) string {
	if db.driver == DriverSQLite {
		return query
	}
	var b strings.Builder
	n := 0
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			n++
			fmt.Fprintf(&b, "$%d", n)
		} else {
			b.WriteByte(query[i])
		}
	}
	return b.String()
}

// Exec executes a query with auto-rebinding.
func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	return db.conn.Exec(db.rebind(query), args...)
}

// Query runs a query with auto-rebinding.
func (db *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return db.conn.Query(db.rebind(query), args...)
}

// QueryRow runs a single-row query with auto-rebinding.
func (db *DB) QueryRow(query string, args ...any) *sql.Row {
	return db.conn.QueryRow(db.rebind(query), args...)
}

// InsertReturningID executes an INSERT and returns the new row's id.
// For SQLite: uses result.LastInsertId().
// For PostgreSQL: appends RETURNING id and scans the result.
func (db *DB) InsertReturningID(query string, args ...any) (int64, error) {
	if db.driver == DriverPostgres {
		query = db.rebind(query) + " RETURNING id"
		var id int64
		err := db.conn.QueryRow(query, args...).Scan(&id)
		return id, err
	}
	res, err := db.conn.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// --- Transaction helpers ---

// Tx wraps sql.Tx with driver-aware placeholder rebinding.
type Tx struct {
	tx     *sql.Tx
	driver string
}

func (db *DB) Begin() (*Tx, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, driver: db.driver}, nil
}

func (t *Tx) rebind(query string) string {
	if t.driver == DriverSQLite {
		return query
	}
	var b strings.Builder
	n := 0
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			n++
			fmt.Fprintf(&b, "$%d", n)
		} else {
			b.WriteByte(query[i])
		}
	}
	return b.String()
}

func (t *Tx) Exec(query string, args ...any) (sql.Result, error) {
	return t.tx.Exec(t.rebind(query), args...)
}

func (t *Tx) Query(query string, args ...any) (*sql.Rows, error) {
	return t.tx.Query(t.rebind(query), args...)
}

func (t *Tx) QueryRow(query string, args ...any) *sql.Row {
	return t.tx.QueryRow(t.rebind(query), args...)
}

func (t *Tx) Commit() error   { return t.tx.Commit() }
func (t *Tx) Rollback() error { return t.tx.Rollback() }

// --- SQL dialect helpers ---

// JSONExtract returns a SQL expression that extracts a text value from a JSON column.
// For SQLite:    json_extract(column, '$.key')
// For PostgreSQL: column::json->>'key'
func (db *DB) JSONExtract(column, key string) string {
	if db.driver == DriverPostgres {
		return column + "::json->>'" + key + "'"
	}
	return "json_extract(" + column + ", '$." + key + "')"
}

// --- SQLite schema initialization (CREATE TABLE IF NOT EXISTS + ad-hoc migrations) ---

func (db *DB) initSQLiteSchema() error {
	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("main schema: %w", err)
	}
	if _, err := db.Exec(authSchemaSQL); err != nil {
		return fmt.Errorf("auth schema: %w", err)
	}
	return db.runSQLiteMigrations()
}

func (db *DB) runSQLiteMigrations() error {
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

// hasColumn checks if a table has a specific column (SQLite only, uses PRAGMA).
func (db *DB) hasColumn(table, column string) bool {
	rows, err := db.conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
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

// Path returns the database file path (or connection string for PostgreSQL).
func (db *DB) Path() string {
	return db.path
}

// --- PostgreSQL schema initialization via goose migrations ---

//go:embed migrations/postgres/*.sql
var postgresMigrations embed.FS

func (db *DB) initPostgresSchema() error {
	goose.SetBaseFS(postgresMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("setting goose dialect: %w", err)
	}
	if err := goose.Up(db.conn, "migrations/postgres"); err != nil {
		return fmt.Errorf("running postgres migrations: %w", err)
	}
	log.Printf("PostgreSQL migrations applied successfully")
	return nil
}
