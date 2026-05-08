// Tool handlers for the stewards-mcp sidecar.
//
// Each handler is a thin wrapper: validate inputs → run a single SQL
// query against the substrate → marshal the result. The substrate's
// own SQL functions enforce semantics (FTS, line pagination, etc.); we
// just expose them through the MCP interface.

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// toolError builds the model-visible error result.
//
// Per the MCP spec (and per the SDK's protocol.go comment block on
// CallToolResult.IsError), tool-execution failures are returned as a
// `CallToolResult` with `IsError: true` plus a text content block
// describing what went wrong. JSON-RPC errors are reserved for protocol
// violations (unknown method, malformed params) which the SDK handles
// for us. Mixing them up means the model sees DB outages as
// unrecoverable system errors and stops trying.
func toolError(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(format, args...)},
		},
	}
}

// registerStudyTools wires up the v1+v1.1 (Phase 3e.1, 3e.1.1) tool surface:
// study_search, study_get, study_similar, study_citations.
func registerStudyTools(srv *mcp.Server, pool *pgxpool.Pool) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "study_search",
		Description: "Full-text search the substrate's studies corpus. " +
			"Returns matching slugs, titles, kinds, snippets, and ranks. " +
			"Filter by kinds (e.g. ['study','journal','proposal']) to narrow " +
			"to a specific document type. Use study_get afterward to read a " +
			"matched document by slug.",
	}, makeStudySearch(pool))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "study_get",
		Description: "Read a substrate study by slug, with optional line-range " +
			"pagination for large documents. Returns the body, frontmatter, " +
			"file path, and metadata as a single JSON object. Use study_search " +
			"first to find slugs by topic.",
	}, makeStudyGet(pool))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "study_similar",
		Description: "Find studies similar to a given slug, via the substrate's " +
			"precomputed embedding edges. Returns related slugs with similarity " +
			"scores and edge direction (in/out/both). Useful after study_get to " +
			"surface adjacent material the author may have cross-referenced.",
	}, makeStudySimilar(pool))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "study_citations",
		Description: "List the canonical sources (scriptures, talks, etc.) that a " +
			"given study cites. Returns cited URIs grouped by kind, with anchor " +
			"text and citation count per URI. The URIs are resolvable via " +
			"gospel-engine-v2 (path semantics like 'eng/scriptures/bofm/mosiah/18.md#11').",
	}, makeStudyCitations(pool))
}

// ---------------------------------------------------------------------
// study_search
// ---------------------------------------------------------------------

// StudySearchInput mirrors stewards.study_search_text(text, text[], int).
//
// jsonschema struct tags are description-only per jsonschema-go's For
// documentation. The library reserves WORD= prefixes for future syntax,
// so do not write 'description=foo,minimum=1' — that violates the
// future-compatibility rule. Use plain prose. Constraints (min, max,
// enum) require manual *Schema construction; the substrate's own SQL
// functions already enforce reasonable bounds, so we don't bother.
type StudySearchInput struct {
	Query string   `json:"query" jsonschema:"natural-language search text (websearch_to_tsquery semantics)"`
	Kinds []string `json:"kinds,omitempty" jsonschema:"optional filter on document kinds (study journal proposal phase-doc doc); empty matches all"`
	Limit int      `json:"limit,omitempty" jsonschema:"max results, default 10, capped at 100"`
}

// StudySearchHit is one row returned by stewards.study_search_text.
type StudySearchHit struct {
	Slug    string  `json:"slug"`
	Kind    string  `json:"kind"`
	Title   string  `json:"title"`
	Snippet string  `json:"snippet"`
	Rank    float32 `json:"rank"`
}

// StudySearchOutput is the structured envelope. We wrap in a `results`
// field rather than returning the array directly because MCP outputSchema
// expects an object at the top level.
type StudySearchOutput struct {
	Results []StudySearchHit `json:"results"`
	Count   int              `json:"count"`
}

func makeStudySearch(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in StudySearchInput,
) (*mcp.CallToolResult, StudySearchOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in StudySearchInput,
	) (*mcp.CallToolResult, StudySearchOutput, error) {
		if in.Query == "" {
			return toolError("study_search: 'query' is required and must be non-empty"),
				StudySearchOutput{}, nil
		}
		if in.Limit <= 0 {
			in.Limit = 10
		}
		// Pass nil/empty array for kinds when caller didn't filter; the
		// substrate fn already treats an empty array as "no filter".
		kinds := in.Kinds
		if kinds == nil {
			kinds = []string{}
		}

		rows, err := pool.Query(ctx,
			"SELECT slug, kind, title, snippet, rank "+
				"FROM stewards.study_search_text($1, $2, $3)",
			in.Query, kinds, in.Limit)
		if err != nil {
			return toolError("study_search query: %v", err),
				StudySearchOutput{}, nil
		}
		defer rows.Close()

		var results []StudySearchHit
		for rows.Next() {
			var h StudySearchHit
			if err := rows.Scan(&h.Slug, &h.Kind, &h.Title, &h.Snippet, &h.Rank); err != nil {
				return toolError("study_search scan: %v", err),
					StudySearchOutput{}, nil
			}
			results = append(results, h)
		}
		if err := rows.Err(); err != nil {
			return toolError("study_search rows: %v", err),
				StudySearchOutput{}, nil
		}

		out := StudySearchOutput{Results: results, Count: len(results)}
		// Returning (nil, out, nil) lets the SDK build the standard
		// {content: [{type: text, text: <JSON>}], structuredContent: out,
		// isError: false} envelope.
		return nil, out, nil
	}
}

// ---------------------------------------------------------------------
// study_get
// ---------------------------------------------------------------------

// StudyGetInput mirrors stewards.study_get(text, bool, int, int, int).
// The line-pagination defaults (offset=0, count=200, max_chars=20000)
// match the substrate fn's own defaults; callers only need to provide
// slug for the common case.
type StudyGetInput struct {
	Slug       string `json:"slug" jsonschema:"substrate study slug (kebab-case e.g. way-truth-life or substrate--ftc-wtl-meta-v3-kimi-tuned)"`
	LineOffset int    `json:"line_offset,omitempty" jsonschema:"0-indexed line to start at, default 0"`
	LineCount  int    `json:"line_count,omitempty" jsonschema:"max body lines, default 200, capped at 2000"`
	MaxChars   int    `json:"max_chars,omitempty" jsonschema:"hard cap on body characters returned, default 20000, capped at 200000"`
}

// StudyGetOutput is the substrate fn's jsonb return value, decoded.
// We use map[string]any so the shape passes through whatever the
// substrate decided to include without us having to mirror every key.
type StudyGetOutput map[string]any

func makeStudyGet(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in StudyGetInput,
) (*mcp.CallToolResult, StudyGetOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in StudyGetInput,
	) (*mcp.CallToolResult, StudyGetOutput, error) {
		if in.Slug == "" {
			return toolError("study_get: 'slug' is required"), nil, nil
		}
		if in.LineCount == 0 {
			in.LineCount = 200
		}
		if in.MaxChars == 0 {
			in.MaxChars = 20000
		}

		var raw []byte
		err := pool.QueryRow(ctx,
			"SELECT stewards.study_get($1, $2, $3, $4, $5)",
			in.Slug, true /* include_body */, in.LineOffset, in.LineCount, in.MaxChars,
		).Scan(&raw)
		if err != nil {
			return toolError("study_get query: %v (slug=%q)", err, in.Slug), nil, nil
		}

		var out StudyGetOutput
		if err := json.Unmarshal(raw, &out); err != nil {
			return toolError("study_get decode: %v", err), nil, nil
		}
		// The substrate fn returns NULL when the slug doesn't exist.
		// pgx scans NULL jsonb into raw=nil → Unmarshal succeeds with
		// out=nil. len() on a nil map returns 0, so this check covers
		// both the truly-empty and the not-found cases.
		if len(out) == 0 {
			return toolError("study_get: no study with slug %q", in.Slug), nil, nil
		}

		return nil, out, nil
	}
}

// ---------------------------------------------------------------------
// study_similar
// ---------------------------------------------------------------------

// StudySimilarInput mirrors stewards.study_similar(text, int).
type StudySimilarInput struct {
	Slug  string `json:"slug" jsonschema:"substrate study slug to find neighbors of"`
	Limit int    `json:"limit,omitempty" jsonschema:"max neighbors returned, default 10, capped at 100"`
}

// StudySimilarHit is one row from stewards.study_similar.
// `direction` is one of 'in' (cited by slug), 'out' (slug cites it),
// or 'both' (mutual). Score is cosine similarity in [0, 1].
type StudySimilarHit struct {
	Slug      string  `json:"slug"`
	Title     string  `json:"title"`
	FilePath  string  `json:"file_path"`
	Score     float64 `json:"score"`
	Direction string  `json:"direction"`
}

type StudySimilarOutput struct {
	Results []StudySimilarHit `json:"results"`
	Count   int               `json:"count"`
}

func makeStudySimilar(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in StudySimilarInput,
) (*mcp.CallToolResult, StudySimilarOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in StudySimilarInput,
	) (*mcp.CallToolResult, StudySimilarOutput, error) {
		if in.Slug == "" {
			return toolError("study_similar: 'slug' is required"),
				StudySimilarOutput{}, nil
		}
		if in.Limit <= 0 {
			in.Limit = 10
		}

		rows, err := pool.Query(ctx,
			"SELECT slug, title, file_path, score, direction "+
				"FROM stewards.study_similar($1, $2)",
			in.Slug, in.Limit)
		if err != nil {
			return toolError("study_similar query: %v (slug=%q)", err, in.Slug),
				StudySimilarOutput{}, nil
		}
		defer rows.Close()

		var results []StudySimilarHit
		for rows.Next() {
			var h StudySimilarHit
			if err := rows.Scan(&h.Slug, &h.Title, &h.FilePath, &h.Score, &h.Direction); err != nil {
				return toolError("study_similar scan: %v", err),
					StudySimilarOutput{}, nil
			}
			results = append(results, h)
		}
		if err := rows.Err(); err != nil {
			return toolError("study_similar rows: %v", err),
				StudySimilarOutput{}, nil
		}

		return nil, StudySimilarOutput{Results: results, Count: len(results)}, nil
	}
}

// ---------------------------------------------------------------------
// study_citations
// ---------------------------------------------------------------------

// StudyCitationsInput mirrors stewards.study_citations(text).
type StudyCitationsInput struct {
	Slug string `json:"slug" jsonschema:"substrate study slug to list citations for"`
}

// StudyCitation is one row from stewards.study_citations.
// study_slug is repeated per row (the substrate fn could in principle
// be reused for graph walks across multiple studies, but for now it's
// always the input slug). cited_kind is e.g. 'scripture', 'talk',
// 'manual'. anchor_text is the displayed link text. citation_count is
// how many times this URI is cited within the source study.
type StudyCitation struct {
	StudySlug     string `json:"study_slug"`
	CitedURI      string `json:"cited_uri"`
	CitedKind     string `json:"cited_kind"`
	AnchorText    string `json:"anchor_text"`
	CitationCount int    `json:"citation_count"`
}

type StudyCitationsOutput struct {
	Citations []StudyCitation `json:"citations"`
	Count     int             `json:"count"`
}

func makeStudyCitations(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in StudyCitationsInput,
) (*mcp.CallToolResult, StudyCitationsOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in StudyCitationsInput,
	) (*mcp.CallToolResult, StudyCitationsOutput, error) {
		if in.Slug == "" {
			return toolError("study_citations: 'slug' is required"),
				StudyCitationsOutput{}, nil
		}

		rows, err := pool.Query(ctx,
			"SELECT study_slug, cited_uri, cited_kind, anchor_text, citation_count "+
				"FROM stewards.study_citations($1)",
			in.Slug)
		if err != nil {
			return toolError("study_citations query: %v (slug=%q)", err, in.Slug),
				StudyCitationsOutput{}, nil
		}
		defer rows.Close()

		var results []StudyCitation
		for rows.Next() {
			var c StudyCitation
			if err := rows.Scan(&c.StudySlug, &c.CitedURI, &c.CitedKind, &c.AnchorText, &c.CitationCount); err != nil {
				return toolError("study_citations scan: %v", err),
					StudyCitationsOutput{}, nil
			}
			results = append(results, c)
		}
		if err := rows.Err(); err != nil {
			return toolError("study_citations rows: %v", err),
				StudyCitationsOutput{}, nil
		}

		return nil, StudyCitationsOutput{Citations: results, Count: len(results)}, nil
	}
}
