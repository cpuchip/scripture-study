// Package db owns the pgx connection pool.
package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// defaultDSN matches projects/pg-ai-stewards/extension/docker-compose.yaml,
// which publishes the container's 5432 to host port 55433. Override
// with STEWARDS_DSN for any non-default deployment.
const defaultDSN = "postgres://stewards:stewards@localhost:55433/stewards?sslmode=disable"

// Connect returns a pgxpool from STEWARDS_DSN, falling back to the
// local docker-compose default (port 55433) when unset.
func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("STEWARDS_DSN")
	if dsn == "" {
		dsn = defaultDSN
	}
	return pgxpool.New(ctx, dsn)
}
