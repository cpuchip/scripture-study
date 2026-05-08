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

// registerStudyTools wires up the v1 (Phase 3e.1) tool surface:
// study_search and study_get.
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
}

// ---------------------------------------------------------------------
// study_search
// ---------------------------------------------------------------------

// StudySearchInput mirrors stewards.study_search_text(text, text[], int).
type StudySearchInput struct {
	Query string   `json:"query" jsonschema:"natural-language search text (websearch_to_tsquery semantics)"`
	Kinds []string `json:"kinds,omitempty" jsonschema:"optional filter on document kinds (e.g. study journal proposal phase-doc doc); empty matches all"`
	Limit int      `json:"limit,omitempty" jsonschema:"max results (default 10),minimum=1,maximum=100"`
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
	LineOffset int    `json:"line_offset,omitempty" jsonschema:"0-indexed line to start at (default 0),minimum=0"`
	LineCount  int    `json:"line_count,omitempty" jsonschema:"max body lines (default 200),minimum=1,maximum=2000"`
	MaxChars   int    `json:"max_chars,omitempty" jsonschema:"hard cap on body characters returned (default 20000),minimum=100,maximum=200000"`
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
		// out=nil. Detect and surface as a model-visible error.
		if out == nil || len(out) == 0 {
			return toolError("study_get: no study with slug %q", in.Slug), nil, nil
		}

		return nil, out, nil
	}
}
