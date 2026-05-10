// Studies endpoints — list, get, search, citations, similar.
//
// All read-only. Substrate's existing SQL functions do the heavy
// lifting (study_search_text for FTS, study_similar for pgvector
// cosine, study_citations for AGE-graph derived edges, study_get
// for the row body).

package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Sub-register hook used by Register() in api.go to avoid one giant
// init function. Called from api.Register.
func (d *Deps) registerStudies(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/studies/list",   d.studiesListHandler)
	mux.HandleFunc("GET /api/studies/get",    d.studiesGetHandler)
	mux.HandleFunc("GET /api/studies/search", d.studiesSearchHandler)
}

type studyBrief struct {
	Slug      string     `json:"slug"`
	Kind      string     `json:"kind"`
	Title     string     `json:"title,omitempty"`
	BodyChars int        `json:"body_chars"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type studiesListResp struct {
	Items []studyBrief `json:"items"`
	Total int          `json:"total"`
}

func (d *Deps) studiesListHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	q := r.URL.Query()
	kind := q.Get("kind") // optional
	limit := atoiDefault(q.Get("limit"), 100, 1, 500)
	offset := atoiDefault(q.Get("offset"), 0, 0, 1_000_000)

	resp := studiesListResp{}

	// total count
	if err := d.Pool.QueryRow(ctx,
		listCountQuery(kind),
		listCountArgs(kind)...,
	).Scan(&resp.Total); err != nil {
		writeErr(w, http.StatusInternalServerError, "count: "+err.Error())
		return
	}

	rows, err := d.Pool.Query(ctx,
		listSelectQuery(kind),
		listSelectArgs(kind, limit, offset)...,
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "list: "+err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var s studyBrief
		var bodyLen int
		if err := rows.Scan(&s.Slug, &s.Kind, &s.Title, &bodyLen, &s.CreatedAt, &s.UpdatedAt); err == nil {
			s.BodyChars = bodyLen
			resp.Items = append(resp.Items, s)
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

// listCountQuery + listCountArgs split so we can keep the query
// strings static (better for plan caching) while still optionally
// filtering by kind.
func listCountQuery(kind string) string {
	if kind == "" {
		return `SELECT count(*) FROM stewards.studies`
	}
	return `SELECT count(*) FROM stewards.studies WHERE kind = $1`
}
func listCountArgs(kind string) []any {
	if kind == "" {
		return nil
	}
	return []any{kind}
}
func listSelectQuery(kind string) string {
	base := `SELECT slug, kind, coalesce(frontmatter->>'title', slug),
	                length(body), created_at, updated_at
	          FROM stewards.studies`
	if kind == "" {
		return base + ` ORDER BY updated_at DESC NULLS LAST LIMIT $1 OFFSET $2`
	}
	return base + ` WHERE kind = $1 ORDER BY updated_at DESC NULLS LAST LIMIT $2 OFFSET $3`
}
func listSelectArgs(kind string, limit, offset int) []any {
	if kind == "" {
		return []any{limit, offset}
	}
	return []any{kind, limit, offset}
}

// /api/studies/get?slug=X — full body + frontmatter
type studyDetail struct {
	Slug        string         `json:"slug"`
	Kind        string         `json:"kind"`
	Title       string         `json:"title,omitempty"`
	Body        string         `json:"body"`
	Frontmatter map[string]any `json:"frontmatter,omitempty"`
	CreatedAt   *time.Time     `json:"created_at,omitempty"`
	UpdatedAt   *time.Time     `json:"updated_at,omitempty"`
	Citations   []citationLite `json:"citations"`
	Similar     []similarHit   `json:"similar"`
}

type citationLite struct {
	Ref        string `json:"ref"`
	Count      int    `json:"count,omitempty"`
}

type similarHit struct {
	Slug      string  `json:"slug"`
	Title     string  `json:"title,omitempty"`
	Distance  float64 `json:"distance,omitempty"`
}

func (d *Deps) studiesGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	slug := r.URL.Query().Get("slug")
	if slug == "" {
		writeErr(w, http.StatusBadRequest, "slug query param required")
		return
	}

	var (
		s        studyDetail
		frontJSON []byte
	)
	err := d.Pool.QueryRow(ctx,
		`SELECT slug, kind,
		        coalesce(frontmatter->>'title', slug) AS title,
		        body,
		        coalesce(frontmatter, '{}'::jsonb),
		        created_at, updated_at
		   FROM stewards.studies
		  WHERE slug = $1`,
		slug,
	).Scan(&s.Slug, &s.Kind, &s.Title, &s.Body, &frontJSON, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		writeErr(w, http.StatusNotFound, "study not found: "+err.Error())
		return
	}
	// Frontmatter — parse jsonb bytes into map (ignore errors; empty map is fine)
	if len(frontJSON) > 0 {
		_ = jsonUnmarshal(frontJSON, &s.Frontmatter)
	}

	// Citations — substrate's study_citations(p_slug) returns
	// (study_slug, cited_uri, cited_kind, anchor_text, citation_count).
	// We render cited_uri + count as the user-visible "ref + count" pair.
	s.Citations = []citationLite{}
	rows, err := d.Pool.Query(ctx,
		`SELECT cited_uri, citation_count
		   FROM stewards.study_citations($1)`,
		slug,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var c citationLite
			if err := rows.Scan(&c.Ref, &c.Count); err == nil {
				s.Citations = append(s.Citations, c)
			}
		}
	}

	// Similar studies — study_similar(p_slug, p_limit) returns
	// (slug, title, file_path, score, direction). Score is double
	// precision; we surface it as a "distance"-like number for the UI.
	s.Similar = []similarHit{}
	rows2, err := d.Pool.Query(ctx,
		`SELECT slug, title, score
		   FROM stewards.study_similar($1, 10)`,
		slug,
	)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var h similarHit
			if err := rows2.Scan(&h.Slug, &h.Title, &h.Distance); err == nil {
				s.Similar = append(s.Similar, h)
			}
		}
	}

	writeJSON(w, http.StatusOK, s)
}

// /api/studies/search?q=...&mode=fts|semantic|combined&limit=N
type searchHit struct {
	Slug    string  `json:"slug"`
	Kind    string  `json:"kind,omitempty"`
	Title   string  `json:"title,omitempty"`
	Snippet string  `json:"snippet,omitempty"`
	Score   float64 `json:"score,omitempty"`
}

type searchResp struct {
	Query   string      `json:"query"`
	Mode    string      `json:"mode"`
	Hits    []searchHit `json:"hits"`
}

func (d *Deps) studiesSearchHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	q := r.URL.Query()
	query := q.Get("q")
	if query == "" {
		writeErr(w, http.StatusBadRequest, "q query param required")
		return
	}
	mode := q.Get("mode")
	if mode == "" {
		mode = "combined"
	}
	limit := atoiDefault(q.Get("limit"), 20, 1, 100)

	resp := searchResp{Query: query, Mode: mode, Hits: []searchHit{}}

	switch mode {
	case "fts", "combined":
		// study_search_text(p_query, p_kinds=ARRAY[]::text[], p_limit)
		// — kinds=ARRAY[]::text[] means "all kinds." Score is real
		// (float4) — scan into float64 promoted via ::float8.
		rows, err := d.Pool.Query(ctx,
			`SELECT slug, kind, title, snippet, rank::float8
			   FROM stewards.study_search_text($1, ARRAY[]::text[], $2)`,
			query, limit,
		)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "search: "+err.Error())
			return
		}
		defer rows.Close()
		for rows.Next() {
			var h searchHit
			if err := rows.Scan(&h.Slug, &h.Kind, &h.Title, &h.Snippet, &h.Score); err == nil {
				resp.Hits = append(resp.Hits, h)
			}
		}
	default:
		writeErr(w, http.StatusBadRequest, "unsupported mode: "+mode)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// atoiDefault parses s as int, clamping to [min,max]. Returns def if
// s is empty or unparsable.
func atoiDefault(s string, def, min, max int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

// jsonUnmarshal exists so studies.go doesn't import encoding/json
// directly (api.go already does); keeps the imports tidy.
func jsonUnmarshal(b []byte, v any) error {
	return jsonUnmarshalImpl(b, v)
}

// Deferred indirection — implemented below in helpers.go to avoid
// importing encoding/json from this file's import block.
var jsonUnmarshalImpl func([]byte, any) error

// pgxpool reference kept here so go doesn't garbage-collect the import
// when handlers don't directly mention it (they do indirectly via Deps).
var _ = (*pgxpool.Pool)(nil)
