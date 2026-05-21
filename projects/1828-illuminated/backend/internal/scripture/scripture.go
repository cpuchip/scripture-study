// Package scripture serves canon verses, chapters, search, and the
// word-study reverse lookup against the i1828 Postgres tables.
//
// Endpoints (D-BE-COPYRIGHT option D: verse text only, no footnotes/
// headings/study apparatus; the frontend pairs every render with a
// tabbed-iframe breakout to churchofjesuschrist.org):
//
//	GET /api/scripture/:ref                 — single verse or verse range
//	GET /api/scripture/chapter/:ref         — whole chapter
//	GET /api/scripture/search?q=…           — FTS + trigram fallback
//	GET /api/scripture/word-study/:word     — every verse containing the word
package scripture

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/stuffleberry/i1828/backend/internal/httpx"
)

// Service is the HTTP-level handler bundle. It holds a pool reference
// and exposes Register() for the cmd/server router.
type Service struct {
	pool *pgxpool.Pool
}

// New constructs the Service. The pool is held; handlers borrow connections.
func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Register attaches all scripture routes to mux.
func (s *Service) Register(mux *http.ServeMux) {
	// Word-study lives at a fixed prefix so the more-permissive :ref
	// pattern doesn't swallow it.
	mux.HandleFunc("GET /api/scripture/word-study/{word}", s.handleWordStudy)
	mux.HandleFunc("GET /api/scripture/chapter/{ref...}", s.handleChapter)
	mux.HandleFunc("GET /api/scripture/search", s.handleSearch)
	mux.HandleFunc("GET /api/scripture/{ref...}", s.handleGet)
}

// VerseRow is the JSON shape for a single verse in any response.
type VerseRow struct {
	Verse int    `json:"verse"`
	Text  string `json:"text"`
}

// VerseGetResponse covers /api/scripture/:ref.
type VerseGetResponse struct {
	Ref        string     `json:"ref"`
	AbbrRef    string     `json:"abbr_ref"`
	Book       string     `json:"book"`
	Volume     string     `json:"volume"`
	Chapter    int        `json:"chapter"`
	VerseStart int        `json:"verse_start"`
	VerseEnd   int        `json:"verse_end"`
	Verses     []VerseRow `json:"verses"`
}

// ChapterGetResponse covers /api/scripture/chapter/:ref.
type ChapterGetResponse struct {
	Ref     string     `json:"ref"`
	AbbrRef string     `json:"abbr_ref"`
	Book    string     `json:"book"`
	Volume  string     `json:"volume"`
	Chapter int        `json:"chapter"`
	Verses  []VerseRow `json:"verses"`
}

// SearchHit is one ranked search result.
type SearchHit struct {
	Ref     string  `json:"ref"`
	AbbrRef string  `json:"abbr_ref"`
	Book    string  `json:"book"`
	Volume  string  `json:"volume"`
	Chapter int     `json:"chapter"`
	Verse   int     `json:"verse"`
	Text    string  `json:"text"`
	Snippet string  `json:"snippet"`
	Rank    float64 `json:"rank"`
}

// SearchResponse wraps the hits.
type SearchResponse struct {
	Query   string      `json:"query"`
	Mode    string      `json:"mode"` // "fts" | "trigram"
	Results []SearchHit `json:"results"`
}

// WordStudyOccurrence is one verse hit for the word-study reverse lookup.
type WordStudyOccurrence struct {
	Ref     string `json:"ref"`
	AbbrRef string `json:"abbr_ref"`
	Book    string `json:"book"`
	Volume  string `json:"volume"`
	Chapter int    `json:"chapter"`
	Verse   int    `json:"verse"`
	Text    string `json:"text"`
}

type WordStudyResponse struct {
	Word        string                `json:"word"`
	Found       bool                  `json:"found"`
	Occurrences []WordStudyOccurrence `json:"occurrences"`
}

// --- handlers ------------------------------------------------------

func (s *Service) handleGet(w http.ResponseWriter, r *http.Request) {
	raw := r.PathValue("ref")
	ref, err := ParseRef(raw, false)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_ref", err.Error())
		return
	}

	// Whole-chapter ref → redirect-equivalent: return the chapter shape.
	if !ref.HasVerse() {
		s.respondChapter(w, r.Context(), ref)
		return
	}

	verses, err := s.fetchVerseRange(r.Context(), ref)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	if len(verses) == 0 {
		httpx.WriteError(w, http.StatusNotFound, "not_found",
			fmt.Sprintf("no verses found for %s", ref.HumanRef()))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, VerseGetResponse{
		Ref:        ref.HumanRef(),
		AbbrRef:    ref.AbbrRef(),
		Book:       ref.Book,
		Volume:     ref.Meta.Volume,
		Chapter:    ref.Chapter,
		VerseStart: ref.VerseStart,
		VerseEnd:   ref.VerseEnd,
		Verses:     verses,
	})
}

func (s *Service) handleChapter(w http.ResponseWriter, r *http.Request) {
	raw := r.PathValue("ref")
	ref, err := ParseRef(raw, false)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_ref", err.Error())
		return
	}
	s.respondChapter(w, r.Context(), ref)
}

func (s *Service) respondChapter(w http.ResponseWriter, ctx context.Context, ref *Ref) {
	verses, err := s.fetchChapter(ctx, ref)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	if len(verses) == 0 {
		httpx.WriteError(w, http.StatusNotFound, "not_found",
			fmt.Sprintf("chapter not found: %s %d", ref.Book, ref.Chapter))
		return
	}
	chapterRef := &Ref{Raw: ref.Raw, Book: ref.Book, Meta: ref.Meta, Chapter: ref.Chapter}
	httpx.WriteJSON(w, http.StatusOK, ChapterGetResponse{
		Ref:     chapterRef.HumanRef(),
		AbbrRef: chapterRef.AbbrRef(),
		Book:    ref.Book,
		Volume:  ref.Meta.Volume,
		Chapter: ref.Chapter,
		Verses:  verses,
	})
}

func (s *Service) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing_query", "?q is required")
		return
	}
	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	volume := strings.TrimSpace(r.URL.Query().Get("volume"))

	hits, mode, err := s.search(r.Context(), q, limit, volume)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, SearchResponse{
		Query:   q,
		Mode:    mode,
		Results: hits,
	})
}

func (s *Service) handleWordStudy(w http.ResponseWriter, r *http.Request) {
	word := strings.ToLower(strings.TrimSpace(r.PathValue("word")))
	if word == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing_word", "word path segment required")
		return
	}
	occ, err := s.wordStudyOccurrences(r.Context(), word)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, WordStudyResponse{
		Word:        word,
		Found:       len(occ) > 0,
		Occurrences: occ,
	})
}

// --- queries -------------------------------------------------------

func (s *Service) fetchVerseRange(ctx context.Context, ref *Ref) ([]VerseRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.verse, v.text
		FROM scripture_verses v
		JOIN scripture_chapters c ON c.id = v.chapter_id
		JOIN scripture_books    b ON b.id = c.book_id
		WHERE b.abbr = $1 AND c.chapter = $2 AND v.verse BETWEEN $3 AND $4
		ORDER BY v.verse
	`, ref.Meta.Abbr, ref.Chapter, ref.VerseStart, ref.VerseEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []VerseRow
	for rows.Next() {
		var vr VerseRow
		if err := rows.Scan(&vr.Verse, &vr.Text); err != nil {
			return nil, err
		}
		out = append(out, vr)
	}
	return out, rows.Err()
}

func (s *Service) fetchChapter(ctx context.Context, ref *Ref) ([]VerseRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT v.verse, v.text
		FROM scripture_verses v
		JOIN scripture_chapters c ON c.id = v.chapter_id
		JOIN scripture_books    b ON b.id = c.book_id
		WHERE b.abbr = $1 AND c.chapter = $2
		ORDER BY v.verse
	`, ref.Meta.Abbr, ref.Chapter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []VerseRow
	for rows.Next() {
		var vr VerseRow
		if err := rows.Scan(&vr.Verse, &vr.Text); err != nil {
			return nil, err
		}
		out = append(out, vr)
	}
	return out, rows.Err()
}

func (s *Service) search(ctx context.Context, q string, limit int, volume string) ([]SearchHit, string, error) {
	// Short queries (≤3 chars) skip FTS and go straight to trigram.
	if len(strings.TrimSpace(q)) <= 3 {
		hits, err := s.searchTrigram(ctx, q, limit, volume)
		return hits, "trigram", err
	}
	hits, err := s.searchFTS(ctx, q, limit, volume)
	if err != nil {
		// websearch_to_tsquery is forgiving but can reject some inputs;
		// fall back to trigram on any tsquery parse error.
		if isTSQueryError(err) {
			hits, err := s.searchTrigram(ctx, q, limit, volume)
			return hits, "trigram", err
		}
		return nil, "", err
	}
	// If FTS produced nothing, try trigram as a last resort.
	if len(hits) == 0 {
		hits, err := s.searchTrigram(ctx, q, limit, volume)
		return hits, "trigram", err
	}
	return hits, "fts", nil
}

func (s *Service) searchFTS(ctx context.Context, q string, limit int, volume string) ([]SearchHit, error) {
	args := []any{q, limit}
	volFilter := ""
	if volume != "" {
		volFilter = "AND b.volume = $3"
		args = []any{q, limit, volume}
	}
	sql := fmt.Sprintf(`
		SELECT b.abbr, b.name, b.volume, c.chapter, v.verse, v.text,
		       ts_headline('english', v.text, websearch_to_tsquery('english', $1),
		         'StartSel=<em>,StopSel=</em>,MaxFragments=1,MaxWords=30,MinWords=10') AS snippet,
		       ts_rank_cd(v.text_tsv, websearch_to_tsquery('english', $1)) AS rank
		FROM scripture_verses v
		JOIN scripture_chapters c ON c.id = v.chapter_id
		JOIN scripture_books    b ON b.id = c.book_id
		WHERE v.text_tsv @@ websearch_to_tsquery('english', $1) %s
		ORDER BY rank DESC
		LIMIT $2
	`, volFilter)
	return s.scanSearchHits(ctx, sql, args)
}

func (s *Service) searchTrigram(ctx context.Context, q string, limit int, volume string) ([]SearchHit, error) {
	args := []any{q, limit}
	volFilter := ""
	if volume != "" {
		volFilter = "AND b.volume = $3"
		args = []any{q, limit, volume}
	}
	sql := fmt.Sprintf(`
		SELECT b.abbr, b.name, b.volume, c.chapter, v.verse, v.text,
		       v.text AS snippet,
		       similarity(v.text, $1) AS rank
		FROM scripture_verses v
		JOIN scripture_chapters c ON c.id = v.chapter_id
		JOIN scripture_books    b ON b.id = c.book_id
		WHERE v.text ILIKE '%%' || $1 || '%%' %s
		ORDER BY rank DESC
		LIMIT $2
	`, volFilter)
	return s.scanSearchHits(ctx, sql, args)
}

func (s *Service) scanSearchHits(ctx context.Context, sql string, args []any) ([]SearchHit, error) {
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SearchHit
	for rows.Next() {
		var hit SearchHit
		var abbr string
		if err := rows.Scan(&abbr, &hit.Book, &hit.Volume, &hit.Chapter, &hit.Verse,
			&hit.Text, &hit.Snippet, &hit.Rank); err != nil {
			return nil, err
		}
		hit.AbbrRef = fmt.Sprintf("%s/%d:%d", abbr, hit.Chapter, hit.Verse)
		hit.Ref = fmt.Sprintf("%s %d:%d", hit.Book, hit.Chapter, hit.Verse)
		out = append(out, hit)
	}
	return out, rows.Err()
}

func (s *Service) wordStudyOccurrences(ctx context.Context, word string) ([]WordStudyOccurrence, error) {
	// Use ILIKE on the word with word-boundary regex via SIMILAR TO is
	// expensive; tsquery is the right tool. We mirror the search path's
	// archaic-suffix expansion: query both the bare word and its -eth
	// /-est /-edst /-ing /-ed /-s family via tsquery OR.
	tsq := buildArchaicTSQuery(word)
	rows, err := s.pool.Query(ctx, `
		SELECT b.abbr, b.name, b.volume, c.chapter, v.verse, v.text
		FROM scripture_verses v
		JOIN scripture_chapters c ON c.id = v.chapter_id
		JOIN scripture_books    b ON b.id = c.book_id
		WHERE v.text_tsv @@ to_tsquery('english', $1)
		ORDER BY b.display_order, c.chapter, v.verse
		LIMIT 500
	`, tsq)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WordStudyOccurrence
	for rows.Next() {
		var occ WordStudyOccurrence
		var abbr string
		if err := rows.Scan(&abbr, &occ.Book, &occ.Volume, &occ.Chapter, &occ.Verse, &occ.Text); err != nil {
			return nil, err
		}
		occ.AbbrRef = fmt.Sprintf("%s/%d:%d", abbr, occ.Chapter, occ.Verse)
		occ.Ref = fmt.Sprintf("%s %d:%d", occ.Book, occ.Chapter, occ.Verse)
		out = append(out, occ)
	}
	return out, rows.Err()
}

// buildArchaicTSQuery builds a Postgres tsquery that ORs the bare word
// with archaic-suffix variants. Mirrors the frontend's ARCHAIC_SUFFIXES
// list so server-side stem-fallback (D-SC-2 ratification) is shared
// between scripture search and the dictionary handlers.
func buildArchaicTSQuery(word string) string {
	word = sanitizeWord(word)
	if word == "" {
		return ""
	}
	suffixes := []string{"eth", "edst", "est", "ing", "ed", "s"}
	parts := []string{word}
	for _, suf := range suffixes {
		parts = append(parts, word+suf)
	}
	return strings.Join(parts, " | ")
}

// sanitizeWord keeps only [a-z'-] so user input can't smuggle tsquery
// operators (| & ! : <->).
func sanitizeWord(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r == '\'', r == '-':
			b.WriteRune(r)
		}
	}
	return b.String()
}

// isTSQueryError returns true when err is a pgx syntax error for a
// malformed tsquery — those are user-input errors, not infrastructure.
func isTSQueryError(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 42601 = syntax_error, the family pgx uses for malformed tsquery.
		return pgErr.Code == "42601"
	}
	// Fallback: substring match on the error message.
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "syntax error in tsquery")
}
