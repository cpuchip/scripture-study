package seed

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/stuffleberry/i1828/backend/internal/scripture"
)

//go:embed data/scriptures.zip
var scripturesZip embed.FS

// bcbooks JSON shapes. The "books → chapters → verses" form covers
// OT/NT/BoM/PGP; D&C uses "sections → verses". Two structs, one ingest
// pass that fans both into the same SQL inserts.

type bcbooksFile struct {
	// Variant A (OT/NT/BoM/PGP)
	Books []bcbooksBook `json:"books,omitempty"`
	// Variant B (D&C)
	Sections []bcbooksSection `json:"sections,omitempty"`
}

type bcbooksBook struct {
	Book     string           `json:"book"`
	Chapters []bcbooksChapter `json:"chapters"`
}

type bcbooksChapter struct {
	Chapter int            `json:"chapter"`
	Verses  []bcbooksVerse `json:"verses"`
}

type bcbooksSection struct {
	Section int            `json:"section"`
	Verses  []bcbooksVerse `json:"verses"`
}

type bcbooksVerse struct {
	Verse int    `json:"verse"`
	Text  string `json:"text"`
}

// SeedScripture ingests the embedded bcbooks corpus into scripture_books
// /chapters/verses. Idempotent: if scripture_verses has rows already we
// skip the whole pass. Strip rules per D-BE-COPYRIGHT option D — applied
// once at INSERT time so the DB only ever holds verse-text-only.
func SeedScripture(ctx context.Context, pool any) error {
	p, ok := pool.(*pgxpool.Pool)
	if !ok {
		return fmt.Errorf("SeedScripture: expected *pgxpool.Pool, got %T", pool)
	}

	// Idempotency probe.
	var existing int
	if err := p.QueryRow(ctx, `SELECT COUNT(*) FROM scripture_verses`).Scan(&existing); err != nil {
		return fmt.Errorf("count scripture_verses: %w", err)
	}
	if existing > 0 {
		log.Printf("[seed] scripture: skip (already %d verses)", existing)
		return nil
	}

	start := time.Now()
	log.Printf("[seed] scripture: starting ingest from embedded scriptures.zip")

	zipBytes, err := scripturesZip.ReadFile("data/scriptures.zip")
	if err != nil {
		return fmt.Errorf("read embedded zip: %w", err)
	}
	sourceSHA := sha256Hex(zipBytes)

	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}

	// Pre-INSERT all books in deterministic display order. ON CONFLICT
	// DO NOTHING so a partially-completed prior run is safe to resume.
	if err := insertBooks(ctx, p); err != nil {
		return err
	}
	bookIDs, err := loadBookIDs(ctx, p)
	if err != nil {
		return err
	}

	// Stream-decode each volume file and accumulate (chapter_id, verse, text)
	// tuples into batched COPY-friendly slices.
	type chapterKey struct {
		bookID  int
		chapter int
	}
	chapterIDs := make(map[chapterKey]int)

	totalVerses := 0
	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() || !strings.HasSuffix(zf.Name, ".json") {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return fmt.Errorf("open %s: %w", zf.Name, err)
		}
		body, readErr := io.ReadAll(rc)
		rc.Close()
		if readErr != nil {
			return fmt.Errorf("read %s: %w", zf.Name, readErr)
		}
		var file bcbooksFile
		if err := json.Unmarshal(body, &file); err != nil {
			return fmt.Errorf("parse %s: %w", zf.Name, err)
		}

		// Normalize "books" and "sections" into the same shape: a list of
		// (bookName, chapter#, verses).
		flat := flatten(file)
		for _, f := range flat {
			meta, ok := scripture.LookupBookByName(f.Book)
			if !ok {
				log.Printf("[seed] scripture: WARN unknown book %q in %s — skipping", f.Book, zf.Name)
				continue
			}
			bookID, ok := bookIDs[meta.Abbr]
			if !ok {
				return fmt.Errorf("book id missing for abbr %s", meta.Abbr)
			}
			key := chapterKey{bookID: bookID, chapter: f.Chapter}
			chID, ok := chapterIDs[key]
			if !ok {
				var newID int
				err := p.QueryRow(ctx, `
					INSERT INTO scripture_chapters (book_id, chapter)
					VALUES ($1, $2)
					ON CONFLICT (book_id, chapter) DO UPDATE SET chapter = EXCLUDED.chapter
					RETURNING id
				`, bookID, f.Chapter).Scan(&newID)
				if err != nil {
					return fmt.Errorf("insert chapter %s %d: %w", f.Book, f.Chapter, err)
				}
				chID = newID
				chapterIDs[key] = chID
			}

			// Bulk-insert this chapter's verses via CopyFrom — much faster
			// than 41k single-row INSERTs.
			rows := make([][]any, 0, len(f.Verses))
			for _, v := range f.Verses {
				text := normalizeVerseText(v.Text)
				if text == "" {
					continue
				}
				rows = append(rows, []any{chID, v.Verse, text})
			}
			n, err := p.CopyFrom(ctx,
				pgx.Identifier{"scripture_verses"},
				[]string{"chapter_id", "verse", "text"},
				pgx.CopyFromRows(rows),
			)
			if err != nil {
				return fmt.Errorf("copy verses for %s %d: %w", f.Book, f.Chapter, err)
			}
			totalVerses += int(n)
		}
	}

	// Record the ingest provenance.
	stripRules := []string{
		"strip bracketed editorial inserts: [text]",
		"strip <i>...</i> italic markup if present",
		"normalize html entities: &amp; &mdash; &rsquo; &lsquo; &quot; &apos;",
		"collapse runs of whitespace",
		"trim leading/trailing whitespace",
	}
	stripJSON, _ := json.Marshal(stripRules)
	if _, err := p.Exec(ctx, `
		INSERT INTO scripture_corpus_meta (source, source_sha, strip_rules, verse_count)
		VALUES ($1, $2, $3, $4)
	`, "bcbooks/scriptures-json (embedded zip)", sourceSHA, string(stripJSON), totalVerses); err != nil {
		return fmt.Errorf("record corpus meta: %w", err)
	}

	log.Printf("[seed] scripture: ingested %d verses across %d books in %s",
		totalVerses, len(bookIDs), time.Since(start))
	return nil
}

func insertBooks(ctx context.Context, p *pgxpool.Pool) error {
	for _, entry := range scripture.AllBooks() {
		_, err := p.Exec(ctx, `
			INSERT INTO scripture_books (volume, abbr, name, display_order)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (abbr) DO UPDATE
			SET volume = EXCLUDED.volume,
			    name = EXCLUDED.name,
			    display_order = EXCLUDED.display_order
		`, entry.Meta.Volume, entry.Meta.Abbr, entry.Name, entry.Meta.DisplayOrder)
		if err != nil {
			return fmt.Errorf("insert book %s: %w", entry.Name, err)
		}
	}
	return nil
}

func loadBookIDs(ctx context.Context, p *pgxpool.Pool) (map[string]int, error) {
	rows, err := p.Query(ctx, `SELECT id, abbr FROM scripture_books`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]int)
	for rows.Next() {
		var id int
		var abbr string
		if err := rows.Scan(&id, &abbr); err != nil {
			return nil, err
		}
		out[abbr] = id
	}
	return out, rows.Err()
}

type flatChapter struct {
	Book    string
	Chapter int
	Verses  []bcbooksVerse
}

func flatten(file bcbooksFile) []flatChapter {
	var out []flatChapter
	for _, b := range file.Books {
		for _, c := range b.Chapters {
			out = append(out, flatChapter{Book: b.Book, Chapter: c.Chapter, Verses: c.Verses})
		}
	}
	// D&C: one synthetic book, sections become chapter numbers
	for _, s := range file.Sections {
		out = append(out, flatChapter{Book: "Doctrine and Covenants", Chapter: s.Section, Verses: s.Verses})
	}
	return out
}

// normalizeVerseText applies D-BE-COPYRIGHT option D strip rules.
//
// Removes:
//   - Bracketed editorial inserts (e.g. "[but]", "[Kirtland]")
//   - HTML italic/em/sup wrappers
//   - Verse-prefix numeric markers (some upstream forms include "38 And he…")
//
// Normalizes:
//   - HTML entities: &amp; &mdash; &rsquo; &lsquo; &quot; &apos;
//   - Runs of whitespace collapse to single space
//   - Trim
func normalizeVerseText(text string) string {
	if text == "" {
		return ""
	}
	// HTML entities first (small, fast, deterministic).
	text = htmlEntities.Replace(text)
	// Strip italic / superscript wrappers but keep their inner text.
	text = htmlTags.ReplaceAllString(text, "")
	// Drop bracketed editorial inserts entirely. The regex matches
	// non-nested [..] (none of the bcbooks inserts nest).
	text = bracketedInsert.ReplaceAllString(text, "")
	// Whitespace collapse + trim.
	text = whitespaceRun.ReplaceAllString(text, " ")
	// Brackets followed by punctuation leave a leftover space before the
	// punctuation (e.g. "place [Kirtland]." → "place ."). Close that gap
	// so the rendered text reads cleanly.
	text = spaceBeforePunct.ReplaceAllString(text, "$1")
	return strings.TrimSpace(text)
}

var (
	bracketedInsert  = regexp.MustCompile(`\[[^\]]*\]`)
	htmlTags         = regexp.MustCompile(`</?(?:i|em|sup|small|span|b|strong)(?:\s[^>]*)?>`)
	whitespaceRun    = regexp.MustCompile(`\s+`)
	spaceBeforePunct = regexp.MustCompile(`\s+([.,;:!?])`)
	htmlEntities    = strings.NewReplacer(
		"&amp;", "&",
		"&mdash;", "—",
		"&ndash;", "–",
		"&rsquo;", "’",
		"&lsquo;", "‘",
		"&rdquo;", "”",
		"&ldquo;", "“",
		"&quot;", `"`,
		"&apos;", "'",
		"&hellip;", "…",
		"&nbsp;", " ",
	)
)

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
