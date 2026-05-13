package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cpuchip/scripture-study/projects/pg-ai-stewards/cmd/stewards-cli/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// runValidateSQL — Batch I.3 (2026-05-12)
//
// Validates a SQL file's syntax by wrapping it in BEGIN/ROLLBACK against
// the live DB. Real Postgres parser; no side effects.
//
// Usage:
//
//	stewards-cli validate-sql --file path/to/file.sql
//	stewards-cli validate-sql < file.sql
//	cat file.sql | stewards-cli validate-sql
//
// Exit 0 on parse OK; exit 1 with diagnostic on parse failure.
func runValidateSQL(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("validate-sql", flag.ExitOnError)
	file := fs.String("file", "", "SQL file to validate; use '-' or omit for stdin")
	quiet := fs.Bool("quiet", false, "suppress 'ok' output on success")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	var sqlBytes []byte
	var err error
	if *file == "" || *file == "-" {
		sqlBytes, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "validate-sql: read stdin: %v\n", err)
			os.Exit(1)
		}
	} else {
		sqlBytes, err = os.ReadFile(*file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "validate-sql: read %s: %v\n", *file, err)
			os.Exit(1)
		}
	}

	sql := strings.TrimSpace(string(sqlBytes))
	if sql == "" {
		fmt.Fprintln(os.Stderr, "validate-sql: empty input")
		os.Exit(1)
	}

	pool, err := db.Connect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "validate-sql: db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := validateSQL(ctx, pool, sql); err != nil {
		fmt.Fprintf(os.Stderr, "validate-sql: %v\n", err)
		os.Exit(1)
	}

	if !*quiet {
		fmt.Println("validate-sql: ok")
	}
}

// validateSQL runs the given SQL inside BEGIN/ROLLBACK using the supplied
// pool. Returns nil if the SQL parses + executes cleanly; returns the
// underlying error otherwise. Shared by runValidateSQL and the
// materialize-writes hook for .sql files in extension/.
func validateSQL(ctx context.Context, pool *pgxpool.Pool, sql string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	// Always rollback. If Exec failed, rollback is a no-op on the aborted txn;
	// if Exec succeeded, rollback undoes the DDL so this stays side-effect-free.
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, sql); err != nil {
		return err
	}
	return nil
}
