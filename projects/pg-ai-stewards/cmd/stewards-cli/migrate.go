// stewards-cli migrate — substrate migration runner
//
// Reads extension/*.sql files in lexical order, checks
// stewards.schema_migrations for what's already applied, runs the
// rest in a transaction each (one tx per file). Records each applied
// migration with its sha256 so future re-runs are no-ops.
//
// Design decisions (per substrate-migration-ledger-and-projects.md):
//   - Up-only migrations (no rollback).
//   - Lexical filename ordering.
//   - sha256 catches drift: if a file changes after being recorded,
//     migrator warns and skips (does NOT re-run with new content).
//   - Each migration runs in its own transaction. Failure rolls
//     back that file only; subsequent files still attempt.
//   - On startup the bridge entrypoint invokes this command so the
//     substrate is always current after a docker restart.
//
// Usage:
//
//	stewards-cli migrate [--repo-root PATH] [--dry-run] [--target NAME]

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cpuchip/scripture-study/projects/pg-ai-stewards/cmd/stewards-cli/internal/db"
)

func runMigrate(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	repoRoot := fs.String("repo-root", ".", "repo root containing projects/pg-ai-stewards/extension/")
	dryRun := fs.Bool("dry-run", false, "print what would be applied without touching the DB")
	target := fs.String("target", "", "stop after applying the migration with this name (without .sql)")
	listOnly := fs.Bool("list", false, "list all migrations + their applied state and exit")
	backfill := fs.Bool("backfill", false, "record all existing .sql files as 'applied' without executing them (one-time op for pre-ledger DBs)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	rootAbs, err := filepath.Abs(*repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: --repo-root: %v\n", err)
		os.Exit(1)
	}
	migrationsDir := filepath.Join(rootAbs, "projects", "pg-ai-stewards", "extension")
	info, err := os.Stat(migrationsDir)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "migrate: not a directory: %s\n", migrationsDir)
		os.Exit(1)
	}

	// List .sql files in lexical order, excluding subdirectories.
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: read dir: %v\n", err)
		os.Exit(1)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)
	if len(files) == 0 {
		fmt.Println("migrate: no .sql files found in", migrationsDir)
		return
	}

	pool, err := db.Connect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Ensure schema_migrations exists before we query it. This makes
	// the migrator bootstrappable: on a brand-new DB, the table
	// doesn't exist yet, but the ledger schema file (which creates
	// it) is itself one of the migrations. We special-case its
	// existence check.
	var hasTable bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			 WHERE table_schema = 'stewards' AND table_name = 'schema_migrations'
		)`).Scan(&hasTable); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: check schema_migrations: %v\n", err)
		os.Exit(1)
	}

	// Pull current state (empty map if table doesn't exist yet).
	applied := map[string]string{} // name -> sha256
	if hasTable {
		rows, err := pool.Query(ctx, `SELECT name, sha256 FROM stewards.schema_migrations`)
		if err != nil {
			fmt.Fprintf(os.Stderr, "migrate: query: %v\n", err)
			os.Exit(1)
		}
		for rows.Next() {
			var n, s string
			if err := rows.Scan(&n, &s); err != nil {
				continue
			}
			applied[n] = s
		}
		rows.Close()
	}

	// Decision matrix per file:
	//   - not in applied + table exists + dry-run → "would apply"
	//   - not in applied + table exists + real run → apply + record
	//   - in applied + matching sha → "skip (applied)"
	//   - in applied + different sha → "DRIFT WARNING, skip"
	//
	// Special case: the ledger schema file itself. When the table
	// doesn't exist yet, we apply it then mark every applied file
	// (including itself) as recorded. From then on it's a normal
	// migration.
	var (
		toApply []string
		warnings int
		skipped int
		applied2 int
	)

	for _, name := range files {
		full := filepath.Join(migrationsDir, name)
		body, err := os.ReadFile(full)
		if err != nil {
			fmt.Fprintf(os.Stderr, "migrate: read %s: %v\n", name, err)
			os.Exit(1)
		}
		sum := sha256.Sum256(body)
		shaHex := hex.EncodeToString(sum[:])
		baseName := strings.TrimSuffix(name, ".sql")

		// Listing mode prints state + continues.
		if *listOnly {
			state := "PENDING"
			if recorded, ok := applied[baseName]; ok {
				if recorded == shaHex {
					state = "applied"
				} else {
					state = "DRIFT (file changed since record)"
				}
			}
			fmt.Printf("  %-50s %s\n", baseName, state)
			continue
		}

		if recorded, ok := applied[baseName]; ok {
			if recorded != shaHex {
				fmt.Fprintf(os.Stderr,
					"migrate: WARNING %s — file sha256 differs from recorded; skipping (manual review needed)\n",
					baseName)
				warnings++
			}
			skipped++
			continue
		}

		toApply = append(toApply, name)
	}

	if *listOnly {
		fmt.Printf("\n%d total; %d applied, %d drift warnings\n",
			len(files), len(files)-len(toApply)-warnings, warnings)
		return
	}

	if *backfill {
		// One-time recovery for pre-ledger DBs. Record every file
		// as already applied without executing it. h-ledger-1
		// must have already been applied separately (it created
		// the table) — guard against it not being there.
		if !hasTable {
			fmt.Fprintln(os.Stderr,
				"migrate --backfill: schema_migrations does not exist; "+
					"apply h-ledger-1 first via psql -f, then re-run --backfill")
			os.Exit(1)
		}
		fmt.Printf("migrate --backfill: recording %d file(s) as already applied\n", len(toApply))
		var recorded int
		for _, name := range toApply {
			full := filepath.Join(migrationsDir, name)
			body, err := os.ReadFile(full)
			if err != nil {
				fmt.Fprintf(os.Stderr, "migrate: read %s: %v\n", name, err)
				continue
			}
			sum := sha256.Sum256(body)
			shaHex := hex.EncodeToString(sum[:])
			baseName := strings.TrimSuffix(name, ".sql")
			if _, err := pool.Exec(ctx, `
				INSERT INTO stewards.schema_migrations (name, sha256, notes)
				VALUES ($1, $2, 'backfilled')
				ON CONFLICT (name) DO NOTHING`,
				baseName, shaHex); err != nil {
				fmt.Fprintf(os.Stderr, "migrate: record %s: %v\n", name, err)
				continue
			}
			recorded++
		}
		fmt.Printf("migrate --backfill: done. %d recorded.\n", recorded)
		return
	}

	if len(toApply) == 0 {
		fmt.Printf("migrate: substrate is current (%d files; %d applied; %d skipped; %d drift)\n",
			len(files), len(files)-skipped, skipped, warnings)
		return
	}

	if *dryRun {
		fmt.Printf("migrate: %d pending migration(s) (dry-run):\n", len(toApply))
		for _, n := range toApply {
			fmt.Printf("  would apply: %s\n", n)
		}
		return
	}

	fmt.Printf("migrate: applying %d migration(s)\n", len(toApply))
	for _, name := range toApply {
		full := filepath.Join(migrationsDir, name)
		body, err := os.ReadFile(full)
		if err != nil {
			fmt.Fprintf(os.Stderr, "migrate: read %s: %v\n", name, err)
			os.Exit(1)
		}
		sum := sha256.Sum256(body)
		shaHex := hex.EncodeToString(sum[:])
		baseName := strings.TrimSuffix(name, ".sql")

		tx, err := pool.Begin(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "migrate: BEGIN %s: %v\n", name, err)
			os.Exit(1)
		}
		if _, err := tx.Exec(ctx, string(body)); err != nil {
			tx.Rollback(ctx)
			fmt.Fprintf(os.Stderr, "migrate: FAILED %s: %v\n", name, err)
			os.Exit(1)
		}
		// Record. If schema_migrations didn't exist at start, the
		// very first applied file (h-ledger-1) creates it; this
		// INSERT then succeeds.
		if _, err := tx.Exec(ctx, `
			INSERT INTO stewards.schema_migrations (name, sha256, notes)
			VALUES ($1, $2, 'auto')
			ON CONFLICT (name) DO NOTHING`,
			baseName, shaHex); err != nil {
			tx.Rollback(ctx)
			fmt.Fprintf(os.Stderr, "migrate: record %s: %v\n", name, err)
			os.Exit(1)
		}
		if err := tx.Commit(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "migrate: COMMIT %s: %v\n", name, err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ %s\n", baseName)
		applied2++

		if *target != "" && baseName == *target {
			fmt.Printf("migrate: stopped at target=%s\n", *target)
			break
		}
	}

	fmt.Printf("migrate: done. %d applied, %d skipped, %d warning(s)\n",
		applied2, skipped, warnings)
	if warnings > 0 {
		os.Exit(2) // non-zero so CI / startup catches drift
	}
}
