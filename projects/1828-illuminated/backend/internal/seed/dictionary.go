package seed

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed data/webster1828.json.gz
var webster1828Gz []byte

//go:embed data/tier-words.json
var tierWordsJSON []byte

//go:embed data/manual-additions.json
var manualAdditionsJSON []byte

//go:embed data/definitions-modern.json
var modernDefsJSON []byte

// --- 1828 corpus --------------------------------------------------

// webster1828Entry mirrors the JSON shape in webster1828.json.gz —
// one entry per (word, pos) tuple; multiple entries with the same word
// merge on ingest into a single row keyed by word.
type webster1828Entry struct {
	POS         string   `json:"pos"`
	Word        string   `json:"word"`
	Definitions []string `json:"definitions"`
}

// SeedWebster1828 ingests the genuine 1828 corpus (D-DICT-1).
// Idempotent via fingerprint: skips only when webster_1828 has rows AND the
// stored sha256 of the embedded corpus matches. When the bundled data
// changes (e.g. the 2026-06-09 swap from the mislabeled 1913 corpus to the
// genuine 1828), the table is truncated and re-ingested. Uses CopyFrom
// after grouping entries by lowercased headword.
func SeedWebster1828(ctx context.Context, pool any) error {
	p, ok := pool.(*pgxpool.Pool)
	if !ok {
		return fmt.Errorf("SeedWebster1828: expected *pgxpool.Pool, got %T", pool)
	}

	sum := sha256.Sum256(webster1828Gz)
	fingerprint := hex.EncodeToString(sum[:])

	var existing int
	if err := p.QueryRow(ctx, `SELECT COUNT(*) FROM webster_1828`).Scan(&existing); err != nil {
		return fmt.Errorf("count webster_1828: %w", err)
	}
	if existing > 0 {
		var stored string
		err := p.QueryRow(ctx,
			`SELECT sha256 FROM seed_fingerprints WHERE corpus = 'webster_1828'`).Scan(&stored)
		if err == nil && stored == fingerprint {
			log.Printf("[seed] webster_1828: skip (already %d entries, fingerprint match)", existing)
			return nil
		}
		log.Printf("[seed] webster_1828: embedded corpus changed (rows=%d, stored=%.12q) — re-ingesting", existing, stored)
		if _, err := p.Exec(ctx, `TRUNCATE webster_1828`); err != nil {
			return fmt.Errorf("truncate webster_1828: %w", err)
		}
	}

	start := time.Now()
	log.Printf("[seed] webster_1828: gunzipping + parsing %d bytes", len(webster1828Gz))

	gz, err := gzip.NewReader(bytes.NewReader(webster1828Gz))
	if err != nil {
		return fmt.Errorf("gunzip: %w", err)
	}
	defer gz.Close()
	raw, err := io.ReadAll(gz)
	if err != nil {
		return fmt.Errorf("read gunzipped: %w", err)
	}

	var entries []webster1828Entry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return fmt.Errorf("parse webster1828: %w", err)
	}

	// Group by lowercased headword. The on-disk JSON has multiple entries
	// for words with multiple senses (e.g. "lay" as verb + noun) — we
	// store them as an array under one row so the API returns all senses
	// in one shot.
	grouped := make(map[string][]map[string]any, len(entries))
	for _, e := range entries {
		key := strings.ToLower(strings.TrimSpace(e.Word))
		if key == "" {
			continue
		}
		grouped[key] = append(grouped[key], map[string]any{
			"pos":         e.POS,
			"definitions": e.Definitions,
		})
	}

	// Build CopyFrom rows. We also persist a tiny source_offsets shape
	// so an audit knows where in the original corpus each row came from.
	rows := make([][]any, 0, len(grouped))
	for word, entryGroup := range grouped {
		entryJSON, err := json.Marshal(entryGroup)
		if err != nil {
			return fmt.Errorf("marshal entries for %q: %w", word, err)
		}
		offJSON, _ := json.Marshal(map[string]any{
			"source":      "webster1828.json.gz",
			"entry_count": len(entryGroup),
		})
		rows = append(rows, []any{word, string(entryJSON), string(offJSON)})
	}

	n, err := p.CopyFrom(ctx,
		pgx.Identifier{"webster_1828"},
		[]string{"word", "entries", "source_offsets"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("copy webster_1828: %w", err)
	}

	if _, err := p.Exec(ctx, `
		INSERT INTO seed_fingerprints (corpus, sha256, updated_at)
		VALUES ('webster_1828', $1, now())
		ON CONFLICT (corpus) DO UPDATE SET sha256 = EXCLUDED.sha256, updated_at = now()
	`, fingerprint); err != nil {
		return fmt.Errorf("record webster_1828 fingerprint: %w", err)
	}

	log.Printf("[seed] webster_1828: inserted %d distinct headwords in %s", n, time.Since(start))
	return nil
}

// --- Tier words ---------------------------------------------------

type tierWordsFile struct {
	Words []tierWordEntry `json:"words"`
}

type tierWordEntry struct {
	Word          string   `json:"word"`
	Tier          string   `json:"tier"`
	StudyTier     *string  `json:"study_tier"`
	Studies       []string `json:"studies"`
	StudyExcerpts []string `json:"study_excerpts"`
	P4Score       *int     `json:"p4_score"`
	P4Reasons     []string `json:"p4_reasons"`
}

type manualAdditionsFile struct {
	Additions []tierWordEntry `json:"additions"`
}

func SeedTierWords(ctx context.Context, pool any) error {
	p, ok := pool.(*pgxpool.Pool)
	if !ok {
		return fmt.Errorf("SeedTierWords: expected *pgxpool.Pool, got %T", pool)
	}

	var existing int
	if err := p.QueryRow(ctx, `SELECT COUNT(*) FROM tier_words`).Scan(&existing); err != nil {
		return fmt.Errorf("count tier_words: %w", err)
	}
	if existing > 0 {
		log.Printf("[seed] tier_words: skip (already %d entries)", existing)
		return nil
	}

	start := time.Now()
	var auto tierWordsFile
	if err := json.Unmarshal(tierWordsJSON, &auto); err != nil {
		return fmt.Errorf("parse tier-words.json: %w", err)
	}
	var manual manualAdditionsFile
	if err := json.Unmarshal(manualAdditionsJSON, &manual); err != nil {
		return fmt.Errorf("parse manual-additions.json: %w", err)
	}

	if err := upsertTierWords(ctx, p, auto.Words, "auto"); err != nil {
		return err
	}
	if err := upsertTierWords(ctx, p, manual.Additions, "manual"); err != nil {
		return err
	}

	var total int
	_ = p.QueryRow(ctx, `SELECT COUNT(*) FROM tier_words`).Scan(&total)
	log.Printf("[seed] tier_words: %d entries (auto=%d, manual=%d) in %s",
		total, len(auto.Words), len(manual.Additions), time.Since(start))
	return nil
}

func upsertTierWords(ctx context.Context, p *pgxpool.Pool, words []tierWordEntry, source string) error {
	for _, w := range words {
		studiesJSON, _ := json.Marshal(w.Studies)
		if string(studiesJSON) == "null" {
			studiesJSON = []byte("[]")
		}
		excerptsJSON, _ := json.Marshal(w.StudyExcerpts)
		if string(excerptsJSON) == "null" {
			excerptsJSON = []byte("[]")
		}
		reasonsJSON, _ := json.Marshal(w.P4Reasons)
		if string(reasonsJSON) == "null" {
			reasonsJSON = []byte("[]")
		}
		_, err := p.Exec(ctx, `
			INSERT INTO tier_words (word, tier, study_tier, studies, study_excerpts, p4_score, p4_reasons, source)
			VALUES ($1, $2, $3, $4::jsonb, $5::jsonb, $6, $7::jsonb, $8)
			ON CONFLICT (word) DO UPDATE SET
			  tier = EXCLUDED.tier,
			  study_tier = EXCLUDED.study_tier,
			  studies = EXCLUDED.studies,
			  study_excerpts = EXCLUDED.study_excerpts,
			  p4_score = EXCLUDED.p4_score,
			  p4_reasons = EXCLUDED.p4_reasons,
			  source = EXCLUDED.source
		`, strings.ToLower(w.Word), w.Tier, w.StudyTier,
			string(studiesJSON), string(excerptsJSON), w.P4Score, string(reasonsJSON), source)
		if err != nil {
			return fmt.Errorf("upsert tier_words[%s]: %w", w.Word, err)
		}
	}
	return nil
}

// --- Modern defs seed --------------------------------------------

type modernDefsFile struct {
	Definitions map[string]json.RawMessage `json:"definitions"`
}

// SeedModernDefs primes modern_defs with the build-time pre-fetched
// definitions JSON. The on-disk shape stores some entries as null
// (the "looked up + 404" signal); we translate that into entries IS
// NULL AND error IS NULL in the DB row.
func SeedModernDefs(ctx context.Context, pool any) error {
	p, ok := pool.(*pgxpool.Pool)
	if !ok {
		return fmt.Errorf("SeedModernDefs: expected *pgxpool.Pool, got %T", pool)
	}

	var existing int
	if err := p.QueryRow(ctx, `SELECT COUNT(*) FROM modern_defs`).Scan(&existing); err != nil {
		return fmt.Errorf("count modern_defs: %w", err)
	}
	if existing > 0 {
		log.Printf("[seed] modern_defs: skip (already %d entries)", existing)
		return nil
	}

	start := time.Now()
	var file modernDefsFile
	if err := json.Unmarshal(modernDefsJSON, &file); err != nil {
		return fmt.Errorf("parse definitions-modern.json: %w", err)
	}

	var foundCount, nullCount int
	for word, raw := range file.Definitions {
		word = strings.ToLower(strings.TrimSpace(word))
		if word == "" {
			continue
		}
		// raw is either "null" (cached 404) or an object {entries: [...]}.
		if isJSONNull(raw) {
			if _, err := p.Exec(ctx, `
				INSERT INTO modern_defs (word, entries, source, error)
				VALUES ($1, NULL, 'seed-pre-fetched', NULL)
				ON CONFLICT (word) DO NOTHING
			`, word); err != nil {
				return fmt.Errorf("insert modern_defs null %q: %w", word, err)
			}
			nullCount++
			continue
		}
		// The on-disk shape is {entries: [...], error?: "..."}.
		var rec struct {
			Entries json.RawMessage `json:"entries"`
			Error   *string         `json:"error"`
		}
		if err := json.Unmarshal(raw, &rec); err != nil {
			log.Printf("[seed] modern_defs: WARN bad shape for %q: %v", word, err)
			continue
		}
		var entriesArg any
		if len(rec.Entries) > 0 && !isJSONNull(rec.Entries) {
			entriesArg = string(rec.Entries)
		}
		if _, err := p.Exec(ctx, `
			INSERT INTO modern_defs (word, entries, source, error)
			VALUES ($1, $2::jsonb, 'seed-pre-fetched', $3)
			ON CONFLICT (word) DO NOTHING
		`, word, entriesArg, rec.Error); err != nil {
			return fmt.Errorf("insert modern_defs %q: %w", word, err)
		}
		foundCount++
	}
	log.Printf("[seed] modern_defs: %d found + %d null in %s", foundCount, nullCount, time.Since(start))
	return nil
}

func isJSONNull(b json.RawMessage) bool {
	s := strings.TrimSpace(string(b))
	return s == "null" || s == ""
}
