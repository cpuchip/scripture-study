// Package seed loads the bundled scripture / dictionary / tier-metadata
// JSON corpora into the i1828 database on first boot. Phase 1 ships an
// empty RunAll that no-ops; phases 2-3 add the actual ingest paths.
//
// All seeders share two properties:
//
//   - Idempotent: re-running them against a populated DB is a fast skip,
//     not a destructive reload. We check whether the target table is empty
//     (or use ON CONFLICT DO NOTHING) before bulk-inserting.
//   - Safe-mid-migration: if Postgres restarts mid-ingest, the next boot
//     resumes cleanly. Verses already inserted stay; missing chapters
//     get backfilled.
//
// Seed data files live under backend/internal/seed/data/. They're copied
// in by hand (or by a build-time make target) from their canonical
// sources elsewhere in the workspace:
//
//	scriptures.zip                  ← external_context/scriptures-mcp/internal/scripture/data/
//	webster1828.json.gz             ← scripts/webster-mcp/data/
//	tier-words.json                 ← projects/1828-illuminated/frontend/src/data/
//	definitions-modern.seed.json    ← projects/1828-illuminated/frontend/src/data/definitions-modern.json (renamed)
package seed

import "context"

// RunAll fans out to the per-corpus seeders. Order matters: scripture
// books need to exist before verse_highlights_cache can reference verses;
// tier_words need to exist before verse-highlight precomputation makes
// sense. Failure of any one seeder aborts the rest.
func RunAll(ctx context.Context, pool any) error {
	if err := SeedScripture(ctx, pool); err != nil {
		return err
	}
	if err := SeedWebster1828(ctx, pool); err != nil {
		return err
	}
	if err := SeedTierWords(ctx, pool); err != nil {
		return err
	}
	if err := SeedModernDefs(ctx, pool); err != nil {
		return err
	}
	return nil
}

// Per-corpus seeders are stubs until phases 2-3 fill them in.

func SeedScripture(ctx context.Context, pool any) error    { return nil }
func SeedWebster1828(ctx context.Context, pool any) error  { return nil }
func SeedTierWords(ctx context.Context, pool any) error    { return nil }
func SeedModernDefs(ctx context.Context, pool any) error   { return nil }
