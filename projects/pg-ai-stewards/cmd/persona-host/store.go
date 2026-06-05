package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

// Store is the persona-host's data layer over the substrate's Postgres. It owns
// only the persona_host schema; it never touches the stewards extension's tables.
type Store struct {
	pool *pgxpool.Pool
}

// OpenStore connects to the substrate Postgres and verifies the connection.
//
// DSN parse/construction errors are deliberately NOT wrapped — pgx echoes the
// full connection string (including the password) in those messages, so we
// return a redacted error instead. Ping/connect errors reference host:port only
// and are safe to surface.
func OpenStore(ctx context.Context, dsn string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, errors.New("invalid DSN (redacted to avoid leaking credentials)")
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.New("could not build pool from DSN (redacted)")
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping persona_host db: %w", err)
	}
	return &Store{pool: pool}, nil
}

// Close releases the connection pool.
func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// Migrate applies the embedded persona_host schema. It is idempotent and runs on
// every boot. pgx executes a no-argument multi-statement string via the simple
// protocol, so the whole script applies in one call.
func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply persona_host schema: %w", err)
	}
	return nil
}

// EnsureSigningKey returns the singleton signing keypair PEMs, generating and
// persisting a new keypair (via gen) if none exists. Race-safe under concurrent
// first-boot: the INSERT is ON CONFLICT DO NOTHING on the id=1 primary key and we
// re-SELECT, so every caller converges on the same stored key. The private PEM
// transits this function but is never logged.
func (s *Store) EnsureSigningKey(ctx context.Context, gen func() (privPEM, pubPEM string, err error)) (privPEM, pubPEM string, err error) {
	privPEM, pubPEM, err = s.selectSigningKey(ctx)
	if err == nil {
		return privPEM, pubPEM, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", "", err
	}
	np, pub, gerr := gen()
	if gerr != nil {
		return "", "", gerr
	}
	if _, err = s.pool.Exec(ctx,
		`INSERT INTO persona_host.signing_key (id, private_key_pem, public_key_pem)
		 VALUES (1, $1, $2) ON CONFLICT (id) DO NOTHING`, np, pub); err != nil {
		return "", "", fmt.Errorf("insert signing key: %w", err)
	}
	// Re-select: our row, or the race winner's.
	return s.selectSigningKey(ctx)
}

func (s *Store) selectSigningKey(ctx context.Context) (privPEM, pubPEM string, err error) {
	err = s.pool.QueryRow(ctx,
		`SELECT private_key_pem, public_key_pem FROM persona_host.signing_key WHERE id = 1`).
		Scan(&privPEM, &pubPEM)
	return privPEM, pubPEM, err
}

// tableNames returns the persona_host tables present, sorted — used by the smoke
// to prove the migration landed.
func (s *Store) tableNames(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'persona_host'
		ORDER BY table_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
