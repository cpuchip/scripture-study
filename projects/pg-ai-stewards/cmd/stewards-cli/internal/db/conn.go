// Package db owns the pgx connection pool.
package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultDSN = "postgres://stewards:stewards@localhost:5432/stewards?sslmode=disable"

// Connect returns a pgxpool from STEWARDS_DSN, falling back to the
// local docker-compose default when unset.
func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("STEWARDS_DSN")
	if dsn == "" {
		dsn = defaultDSN
	}
	return pgxpool.New(ctx, dsn)
}
