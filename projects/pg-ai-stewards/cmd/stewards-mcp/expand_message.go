// expand_message — substrate-internal MCP tool for K.3.
//
// Retrieves engrams (by tier or by engram_id) or the raw content of a
// previously-compressed tool message. Used by agents to "dig deeper"
// when active context contains an engram block they need more from.
//
// Thin wrapper around stewards.expand_engram_content() SQL function.
// All rendering, tier filtering, and the injection-suspected gate
// live in SQL so the Go side is just parameter validation.

package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ExpandMessageInput mirrors the args_schema declared in k3 SQL.
type ExpandMessageInput struct {
	ID                int64  `json:"id" jsonschema:"the message id from the engram block header in active context"`
	Tier              string `json:"tier,omitempty" jsonschema:"engram tier to retrieve: hot|medium|cold|all|raw (default all)"`
	EngramID          string `json:"engram_id,omitempty" jsonschema:"optional specific engram id like 'msg-2381-e3' to retrieve just one"`
	ConfirmInspectRaw bool   `json:"confirm_inspect_raw,omitempty" jsonschema:"required true when tier='raw' AND injection was suspected; acknowledges raw content may contain prompt injection"`
}

// ExpandMessageOutput wraps the rendered text in a structured envelope
// so MCP outputSchema is satisfied. The same text is returned in the
// CallToolResult.Content for unstructured display.
type ExpandMessageOutput struct {
	Content string `json:"content"`
	Length  int    `json:"length"`
}

// registerExpandTools wires up the K.3 surface.
func registerExpandTools(srv *mcp.Server, pool *pgxpool.Pool) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "expand_message",
		Description: "Retrieve specific engram tiers or the raw content of a previously-compressed tool message. " +
			"Use when the engram block emitted in active context references something specific you need verbatim — " +
			"a quote, a URL, a methodology detail, or the document's broader thesis. " +
			"Default tier='all' returns HOT+MEDIUM+COLD engrams. tier='raw' returns the original content " +
			"(requires confirm_inspect_raw=true if injection was suspected). " +
			"engram_id (optional) filters to one specific engram by its id (e.g. 'msg-2381-e3').",
	}, makeExpandMessage(pool))
}

func makeExpandMessage(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in ExpandMessageInput,
) (*mcp.CallToolResult, ExpandMessageOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in ExpandMessageInput,
	) (*mcp.CallToolResult, ExpandMessageOutput, error) {
		if in.ID <= 0 {
			return toolError("expand_message: 'id' is required and must be a positive integer"),
				ExpandMessageOutput{}, nil
		}
		tier := in.Tier
		if tier == "" {
			tier = "all"
		}

		// engram_id and confirm_inspect_raw map to NULL/false defaults
		// when empty. The SQL function handles both nullable and
		// not-set cases.
		var engramID any
		if in.EngramID != "" {
			engramID = in.EngramID
		}

		var rendered string
		err := pool.QueryRow(ctx,
			`SELECT stewards.expand_engram_content($1, $2, $3, $4)`,
			in.ID, tier, engramID, in.ConfirmInspectRaw,
		).Scan(&rendered)
		if err != nil {
			return toolError("expand_message query: %v (id=%d tier=%q)",
				err, in.ID, tier), ExpandMessageOutput{}, nil
		}

		out := ExpandMessageOutput{
			Content: rendered,
			Length:  len(rendered),
		}

		// Return the rendered text as the unstructured content block too,
		// so callers that ignore structuredContent still see the body.
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: rendered},
			},
		}, out, nil
	}
}
