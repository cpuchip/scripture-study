// Package seed loads the bundled scripture / dictionary / tier-metadata
// JSON corpora into the i1828 database on first boot.
//
// All seeders share two properties:
//
//   - Idempotent: re-running them against a populated DB is a fast skip
//     (or a no-op via ON CONFLICT DO NOTHING), not a destructive reload.
//   - Safe-mid-migration: if Postgres restarts mid-ingest, the next boot
//     resumes cleanly. Already-inserted rows stay; missing rows backfill.
//
// Seed data files live under backend/internal/seed/data/:
//
//	scriptures.zip                 ← external_context/scriptures-mcp/internal/scripture/data/scriptures.zip
//	webster1828.json.gz            ← scripts/webster-mcp/data/webster1828.json.gz
//	tier-words.json                ← projects/1828-illuminated/frontend/src/data/tier-words.json
//	definitions-modern.json        ← projects/1828-illuminated/frontend/src/data/definitions-modern.json
//
// Per-corpus seeders live in sibling files (scriptures.go, dictionary.go).
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
