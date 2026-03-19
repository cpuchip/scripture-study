// Package db provides database operations for the Becoming app.
package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib" // register "pgx" driver
	"github.com/pressly/goose/v3"
)

// DB wraps the database connection.
type DB struct {
	conn *sql.DB
	path string
}

// Open opens a PostgreSQL database connection.
func Open(dsn string) (*DB, error) {
	if !strings.HasPrefix(dsn, "postgres://") && !strings.HasPrefix(dsn, "postgresql://") {
		return nil, fmt.Errorf("invalid database URL: must start with postgres:// or postgresql://")
	}

	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening postgres: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}

	db := &DB{conn: sqlDB, path: dsn}
	if err := db.runMigrations(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	return db, nil
}

// --- Query helpers with automatic placeholder rebinding ---

// rebind converts ? placeholders to $1, $2, ... for PostgreSQL.
func rebind(query string) string {
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
	return db.conn.Exec(rebind(query), args...)
}

// Query runs a query with auto-rebinding.
func (db *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return db.conn.Query(rebind(query), args...)
}

// QueryRow runs a single-row query with auto-rebinding.
func (db *DB) QueryRow(query string, args ...any) *sql.Row {
	return db.conn.QueryRow(rebind(query), args...)
}

// InsertReturningID executes an INSERT and returns the new row's id.
func (db *DB) InsertReturningID(query string, args ...any) (int64, error) {
	query = rebind(query) + " RETURNING id"
	var id int64
	err := db.conn.QueryRow(query, args...).Scan(&id)
	return id, err
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Path returns the database connection string.
func (db *DB) Path() string {
	return db.path
}

// --- Transaction helpers ---

// Tx wraps sql.Tx with placeholder rebinding.
type Tx struct {
	tx *sql.Tx
}

func (db *DB) Begin() (*Tx, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

func (t *Tx) Exec(query string, args ...any) (sql.Result, error) {
	return t.tx.Exec(rebind(query), args...)
}

func (t *Tx) Query(query string, args ...any) (*sql.Rows, error) {
	return t.tx.Query(rebind(query), args...)
}

func (t *Tx) QueryRow(query string, args ...any) *sql.Row {
	return t.tx.QueryRow(rebind(query), args...)
}

func (t *Tx) Commit() error   { return t.tx.Commit() }
func (t *Tx) Rollback() error { return t.tx.Rollback() }

// --- SQL helpers ---

// JSONExtract returns a SQL expression that extracts a text value from a JSON column.
func (db *DB) JSONExtract(column, key string) string {
	return column + "::json->>'" + key + "'"
}

// DateCast returns a SQL expression that extracts the date part of a timestamp.
func (db *DB) DateCast(expr string) string {
	return expr + "::date"
}

// DateText returns a SQL expression that produces a text string in 'YYYY-MM-DD' format.
func (db *DB) DateText(expr string) string {
	return "TO_CHAR(" + expr + "::date, 'YYYY-MM-DD')"
}

// --- PostgreSQL schema initialization via goose migrations ---

//go:embed migrations/postgres/*.sql
var postgresMigrations embed.FS

func (db *DB) runMigrations() error {
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
