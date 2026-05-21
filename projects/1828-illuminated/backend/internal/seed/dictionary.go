package seed

import "context"

// Phase 3 will populate these. The phase 1 build needed callable
// symbols so RunAll could chain.

// SeedWebster1828 loads the full 98k 1828 dictionary corpus (D-DICT-1)
// from data/webster1828.json.gz into the webster_1828 table.
func SeedWebster1828(ctx context.Context, pool any) error { return nil }

// SeedTierWords loads the curated tier-words.json (~853 words with
// study cross-refs and P4 scoring) into the tier_words table.
func SeedTierWords(ctx context.Context, pool any) error { return nil }

// SeedModernDefs primes modern_defs with the build-time pre-fetched
// definitions; the lazy fetcher (phase 3) grows it from there.
func SeedModernDefs(ctx context.Context, pool any) error { return nil }
