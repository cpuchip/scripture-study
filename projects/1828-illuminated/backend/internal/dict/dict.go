// Package dict serves Webster 1828, modern definitions (lazy + write-back
// from Free Dictionary API), and tier metadata.
//
// Endpoints:
//
//	GET /api/dict/1828/{word}      — 1828 entry with archaic-suffix stem fallback
//	GET /api/dict/modern/{word}    — modern entry; lazy fetch + write-back on miss
//	GET /api/dict/tier/{word}      — tier metadata + study cross-refs
//	GET /api/dict/search           — combined tier + class-E reach search
package dict

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/stuffleberry/i1828/backend/internal/httpx"
)

// archaicSuffixes mirrors useWordData.ts's ARCHAIC_SUFFIXES. Server-side
// stem fallback per D-DICT-2 — frontend should stop doing this client-side
// at the phase-5 cutover.
var archaicSuffixes = []string{"eth", "edst", "est", "ing", "ed", "s"}

// Service is the HTTP-level dictionary bundle.
type Service struct {
	pool                *pgxpool.Pool
	modernFetchDailyCap int

	// modernFetcher coalesces concurrent requests for the same word and
	// enforces the global 1-req/sec friendliness cap to Free Dictionary.
	modernFetcher *modernFetcher
}

// New constructs the Service. The pool is held; handlers borrow connections.
func New(pool *pgxpool.Pool, modernFetchDailyCap int) *Service {
	return &Service{
		pool:                pool,
		modernFetchDailyCap: modernFetchDailyCap,
		modernFetcher: &modernFetcher{
			ticker:   time.NewTicker(time.Second),
			inflight: make(map[string]*fetchPromise),
			client:   &http.Client{Timeout: 8 * time.Second},
		},
	}
}

// Register attaches all dictionary routes to mux.
func (s *Service) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/dict/1828/{word}", s.handle1828)
	mux.HandleFunc("GET /api/dict/modern/{word}", s.handleModern)
	mux.HandleFunc("GET /api/dict/tier/{word}", s.handleTier)
	mux.HandleFunc("GET /api/dict/search", s.handleSearch)
}

// ---- response shapes -----------------------------------------------

// Entry mirrors the on-disk shape: {pos: "n.", definitions: ["..."]}.
type Entry struct {
	POS         string   `json:"pos"`
	Definitions []string `json:"definitions"`
}

type Def1828Response struct {
	Word        string  `json:"word"`
	Entries     []Entry `json:"entries"`
	Found       bool    `json:"found"`
	StemMatched *string `json:"stem_matched"`
}

type ModernResponse struct {
	Word    string  `json:"word"`
	Entries []Entry `json:"entries,omitempty"`
	Source  string  `json:"source"`
	Found   bool    `json:"found"`
	Error   *string `json:"error,omitempty"`
}

type TierResponse struct {
	Word          string   `json:"word"`
	Found         bool     `json:"found"`
	Tier          string   `json:"tier,omitempty"`
	StudyTier     *string  `json:"study_tier,omitempty"`
	Studies       []string `json:"studies,omitempty"`
	StudyExcerpts []string `json:"study_excerpts,omitempty"`
	P4Score       *int     `json:"p4_score,omitempty"`
	P4Reasons     []string `json:"p4_reasons,omitempty"`
	Source        string   `json:"source,omitempty"`
}

type DictSearchResult struct {
	Word string `json:"word"`
	Tier string `json:"tier,omitempty"`
}

type DictSearchResponse struct {
	Query          string             `json:"query"`
	TierResults    []DictSearchResult `json:"tier_results"`
	All1828Results []DictSearchResult `json:"all_1828_results"`
}

// ---- handlers ------------------------------------------------------

func (s *Service) handle1828(w http.ResponseWriter, r *http.Request) {
	word := strings.ToLower(strings.TrimSpace(r.PathValue("word")))
	if word == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing_word", "word path segment required")
		return
	}

	// 1) literal lookup
	entries, found, err := s.fetch1828(r.Context(), word)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	if found {
		httpx.WriteJSON(w, http.StatusOK, Def1828Response{Word: word, Entries: entries, Found: true})
		return
	}

	// 2) archaic-suffix stem fallback
	for _, stem := range stemCandidates(word) {
		entries, found, err := s.fetch1828(r.Context(), stem)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}
		if found {
			matched := stem
			httpx.WriteJSON(w, http.StatusOK, Def1828Response{
				Word:        word,
				Entries:     entries,
				Found:       true,
				StemMatched: &matched,
			})
			return
		}
	}

	// 200 (not 404) with found:false — frontend renders "no 1828 entry"
	// gracefully without handling an error code (per scripture-corpus.md §V.1).
	httpx.WriteJSON(w, http.StatusOK, Def1828Response{Word: word, Found: false})
}

func (s *Service) handleModern(w http.ResponseWriter, r *http.Request) {
	word := strings.ToLower(strings.TrimSpace(r.PathValue("word")))
	if word == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing_word", "word path segment required")
		return
	}

	// 1) DB lookup
	row, exists, err := s.fetchModernRow(r.Context(), word)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	if exists {
		// cached 404 (entries NULL, error NULL)
		if !row.HasEntries && row.Error == nil {
			httpx.WriteJSON(w, http.StatusOK, ModernResponse{
				Word:   word,
				Source: "none",
				Found:  false,
			})
			return
		}
		// cached error within 24h: serve the error
		if row.Error != nil && time.Since(row.FetchedAt) < 24*time.Hour {
			errStr := *row.Error
			httpx.WriteJSON(w, http.StatusOK, ModernResponse{
				Word:   word,
				Source: "cache",
				Found:  false,
				Error:  &errStr,
			})
			return
		}
		// happy path: cached entries
		if row.HasEntries {
			httpx.WriteJSON(w, http.StatusOK, ModernResponse{
				Word:    word,
				Entries: row.Entries,
				Source:  "cache",
				Found:   true,
			})
			return
		}
		// stale error: fall through to refetch
	}

	// 2) check daily cap before reaching upstream
	if allowed, err := s.tryConsumeFetchSlot(r.Context()); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	} else if !allowed {
		httpx.WriteJSON(w, http.StatusOK, ModernResponse{
			Word:   word,
			Source: "rate_limited",
			Found:  false,
			Error:  strPtr("daily fetch cap reached; try again tomorrow"),
		})
		return
	}

	// 3) lazy fetch + write back
	result := s.modernFetcher.Fetch(r.Context(), word)
	if err := s.writeModernRow(r.Context(), word, result); err != nil {
		// log but don't fail the request — we have the data in hand
		fmt.Printf("[dict] modern: write-back failed for %s: %v\n", word, err)
	}

	switch result.Kind {
	case fetchKindEntries:
		httpx.WriteJSON(w, http.StatusOK, ModernResponse{
			Word:    word,
			Entries: result.Entries,
			Source:  "fetched",
			Found:   true,
		})
	case fetchKindNotFound:
		httpx.WriteJSON(w, http.StatusOK, ModernResponse{
			Word:   word,
			Source: "fetched",
			Found:  false,
		})
	default:
		errStr := result.Error
		httpx.WriteJSON(w, http.StatusOK, ModernResponse{
			Word:   word,
			Source: "fetched",
			Found:  false,
			Error:  &errStr,
		})
	}
}

func (s *Service) handleTier(w http.ResponseWriter, r *http.Request) {
	word := strings.ToLower(strings.TrimSpace(r.PathValue("word")))
	if word == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing_word", "word path segment required")
		return
	}
	resp, err := s.fetchTier(r.Context(), word)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (s *Service) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	if q == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing_query", "?q is required")
		return
	}
	limit := 40
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	resp, err := s.dictSearch(r.Context(), q, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// ---- queries -------------------------------------------------------

func (s *Service) fetch1828(ctx context.Context, word string) ([]Entry, bool, error) {
	var raw []byte
	err := s.pool.QueryRow(ctx, `SELECT entries FROM webster_1828 WHERE word = $1`, word).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	var entries []Entry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, false, fmt.Errorf("decode 1828 entries: %w", err)
	}
	return entries, true, nil
}

type modernRow struct {
	HasEntries bool
	Entries    []Entry
	FetchedAt  time.Time
	Source     string
	Error      *string
}

func (s *Service) fetchModernRow(ctx context.Context, word string) (modernRow, bool, error) {
	var raw []byte
	var fetched time.Time
	var source string
	var errStr *string
	err := s.pool.QueryRow(ctx, `
		SELECT entries::text, fetched_at, source, error
		FROM modern_defs WHERE word = $1
	`, word).Scan(&raw, &fetched, &source, &errStr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return modernRow{}, false, nil
		}
		return modernRow{}, false, err
	}
	row := modernRow{FetchedAt: fetched, Source: source, Error: errStr}
	if len(raw) > 0 && string(raw) != "null" {
		var entries []Entry
		if err := json.Unmarshal(raw, &entries); err == nil {
			row.Entries = entries
			row.HasEntries = true
		}
	}
	return row, true, nil
}

func (s *Service) writeModernRow(ctx context.Context, word string, result fetchResult) error {
	var entriesArg any
	var errArg any
	switch result.Kind {
	case fetchKindEntries:
		buf, _ := json.Marshal(result.Entries)
		entriesArg = string(buf)
	case fetchKindNotFound:
		// entries NULL, error NULL
	case fetchKindError:
		errArg = result.Error
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO modern_defs (word, entries, fetched_at, source, error)
		VALUES ($1, $2::jsonb, now(), 'free-dictionary-api', $3)
		ON CONFLICT (word) DO UPDATE SET
		  entries    = EXCLUDED.entries,
		  fetched_at = EXCLUDED.fetched_at,
		  source     = EXCLUDED.source,
		  error      = EXCLUDED.error
	`, word, entriesArg, errArg)
	return err
}

func (s *Service) fetchTier(ctx context.Context, word string) (TierResponse, error) {
	var tier, source string
	var studyTier *string
	var studies, excerpts, reasons []byte
	var p4Score *int
	err := s.pool.QueryRow(ctx, `
		SELECT tier, study_tier, studies::text, study_excerpts::text, p4_score, p4_reasons::text, source
		FROM tier_words WHERE word = $1
	`, word).Scan(&tier, &studyTier, &studies, &excerpts, &p4Score, &reasons, &source)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TierResponse{Word: word, Found: false}, nil
		}
		return TierResponse{}, err
	}
	var sArr, eArr, rArr []string
	_ = json.Unmarshal(studies, &sArr)
	_ = json.Unmarshal(excerpts, &eArr)
	_ = json.Unmarshal(reasons, &rArr)
	return TierResponse{
		Word: word, Found: true, Tier: tier, StudyTier: studyTier,
		Studies: sArr, StudyExcerpts: eArr, P4Score: p4Score, P4Reasons: rArr,
		Source: source,
	}, nil
}

func (s *Service) dictSearch(ctx context.Context, q string, limit int) (DictSearchResponse, error) {
	resp := DictSearchResponse{Query: q, TierResults: []DictSearchResult{}, All1828Results: []DictSearchResult{}}

	// Tier results: prefix or trigram against tier_words.
	tierRows, err := s.pool.Query(ctx, `
		SELECT word, tier FROM tier_words
		WHERE word = $1 OR word LIKE $1 || '%' OR word ILIKE '%' || $1 || '%'
		ORDER BY (word = $1) DESC, (word LIKE $1 || '%') DESC, word
		LIMIT $2
	`, q, limit)
	if err != nil {
		return resp, err
	}
	for tierRows.Next() {
		var dr DictSearchResult
		if err := tierRows.Scan(&dr.Word, &dr.Tier); err != nil {
			tierRows.Close()
			return resp, err
		}
		resp.TierResults = append(resp.TierResults, dr)
	}
	tierRows.Close()

	// Class-E reach: prefix against the full 98k 1828 corpus.
	all1828, err := s.pool.Query(ctx, `
		SELECT word FROM webster_1828
		WHERE word LIKE $1 || '%'
		ORDER BY length(word), word
		LIMIT $2
	`, q, limit)
	if err != nil {
		return resp, err
	}
	defer all1828.Close()
	for all1828.Next() {
		var dr DictSearchResult
		if err := all1828.Scan(&dr.Word); err != nil {
			return resp, err
		}
		resp.All1828Results = append(resp.All1828Results, dr)
	}
	return resp, all1828.Err()
}

// ---- daily-cap counter --------------------------------------------

// tryConsumeFetchSlot increments today's counter by 1, returning true
// if the increment stays under modernFetchDailyCap. ON CONFLICT DO
// UPDATE keeps this race-safe across concurrent requests.
func (s *Service) tryConsumeFetchSlot(ctx context.Context) (bool, error) {
	var attempts int
	err := s.pool.QueryRow(ctx, `
		INSERT INTO modern_defs_fetch_log (fetch_date, attempts)
		VALUES (CURRENT_DATE, 1)
		ON CONFLICT (fetch_date) DO UPDATE SET attempts = modern_defs_fetch_log.attempts + 1
		RETURNING attempts
	`).Scan(&attempts)
	if err != nil {
		return false, err
	}
	return attempts <= s.modernFetchDailyCap, nil
}

// ---- stem helpers --------------------------------------------------

func stemCandidates(word string) []string {
	var out []string
	for _, suf := range archaicSuffixes {
		if len(word) > len(suf)+2 && strings.HasSuffix(word, suf) {
			stem := word[:len(word)-len(suf)]
			out = append(out, stem)
			// running → run; obtaining → obtain
			if (suf == "ing" || suf == "ed") && len(stem) >= 3 {
				out = append(out, stem[:len(stem)-1])
			}
			// loveth → love
			if suf == "eth" {
				out = append(out, stem+"e")
			}
		}
	}
	return out
}

func strPtr(s string) *string { return &s }

// ---- fetcher (1 req/sec global cap + singleflight) -----------------

type fetchResultKind int

const (
	fetchKindEntries fetchResultKind = iota
	fetchKindNotFound
	fetchKindError
)

type fetchResult struct {
	Kind    fetchResultKind
	Entries []Entry
	Error   string
}

type fetchPromise struct {
	done   chan struct{}
	result fetchResult
}

type modernFetcher struct {
	mu       sync.Mutex
	ticker   *time.Ticker
	inflight map[string]*fetchPromise
	client   *http.Client
}

// Fetch returns the modern-def entries for word, coalescing duplicate
// concurrent requests and waiting on the 1-req/sec ticker.
func (f *modernFetcher) Fetch(ctx context.Context, word string) fetchResult {
	f.mu.Lock()
	if existing, ok := f.inflight[word]; ok {
		f.mu.Unlock()
		select {
		case <-existing.done:
			return existing.result
		case <-ctx.Done():
			return fetchResult{Kind: fetchKindError, Error: ctx.Err().Error()}
		}
	}
	p := &fetchPromise{done: make(chan struct{})}
	f.inflight[word] = p
	f.mu.Unlock()

	defer func() {
		close(p.done)
		f.mu.Lock()
		delete(f.inflight, word)
		f.mu.Unlock()
	}()

	// Block on the 1-req/sec ticker so we're a polite citizen of the
	// Free Dictionary API.
	select {
	case <-f.ticker.C:
	case <-ctx.Done():
		p.result = fetchResult{Kind: fetchKindError, Error: ctx.Err().Error()}
		return p.result
	}

	p.result = f.doFetch(ctx, word)
	return p.result
}

func (f *modernFetcher) doFetch(ctx context.Context, word string) fetchResult {
	url := "https://api.dictionaryapi.dev/api/v2/entries/en/" + word
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fetchResult{Kind: fetchKindError, Error: err.Error()}
	}
	req.Header.Set("User-Agent", "1828-illuminated.ibeco.me/0.1")
	resp, err := f.client.Do(req)
	if err != nil {
		return fetchResult{Kind: fetchKindError, Error: err.Error()}
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return fetchResult{Kind: fetchKindNotFound}
	}
	if resp.StatusCode != http.StatusOK {
		return fetchResult{Kind: fetchKindError, Error: fmt.Sprintf("upstream %d", resp.StatusCode)}
	}
	// Free Dictionary API returns an array of word-meaning bundles. We
	// flatten into our pos/definitions shape so the response looks like
	// the 1828 shape — frontend renders both identically.
	var raw []struct {
		Meanings []struct {
			PartOfSpeech string `json:"partOfSpeech"`
			Definitions  []struct {
				Definition string `json:"definition"`
			} `json:"definitions"`
		} `json:"meanings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return fetchResult{Kind: fetchKindError, Error: fmt.Sprintf("decode: %v", err)}
	}
	var out []Entry
	for _, bundle := range raw {
		for _, m := range bundle.Meanings {
			entry := Entry{POS: m.PartOfSpeech}
			for _, d := range m.Definitions {
				entry.Definitions = append(entry.Definitions, d.Definition)
			}
			if len(entry.Definitions) > 0 {
				out = append(out, entry)
			}
		}
	}
	if len(out) == 0 {
		return fetchResult{Kind: fetchKindNotFound}
	}
	return fetchResult{Kind: fetchKindEntries, Entries: out}
}
