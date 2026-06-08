package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// defaultDSN points at the dev substrate's host-mapped Postgres port (55433 →
// the pg-ai-stewards-dev container's 5432). The cockpit runs on the host, not
// inside the bridge, so it uses the published port — unlike stewards-cli, which
// often runs in-container against 5432.
const defaultDSN = "postgres://stewards:stewards@localhost:55433/stewards?sslmode=disable"

func dsn() string {
	if d := os.Getenv("STEWARDS_DSN"); d != "" {
		return d
	}
	return defaultDSN
}

// connect opens a pgxpool to the substrate. DSN parse/build errors are returned
// redacted — pgx echoes the full connection string (password included) in those
// messages. Ping errors reference host:port only and are safe to surface.
func connect(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn())
	if err != nil {
		return nil, errors.New("invalid STEWARDS_DSN (redacted to avoid leaking credentials)")
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.New("could not build connection pool (DSN redacted)")
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("cannot reach the substrate at %s:%d — is pg-ai-stewards-dev up? (%w)",
			cfg.ConnConfig.Host, cfg.ConnConfig.Port, err)
	}
	return pool, nil
}

// mustConnect connects or exits with a clear message. Every verb is read-only,
// so a failed connect is the only fatal infra error.
func mustConnect(ctx context.Context) *pgxpool.Pool {
	pool, err := connect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	return pool
}

// fail prints "context: err" to stderr and exits 1.
func fail(what string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", what, err)
	os.Exit(1)
}
