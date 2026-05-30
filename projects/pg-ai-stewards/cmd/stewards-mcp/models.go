// Model + connector catalog tools — read-only views into which models the
// substrate can dispatch and the state of each provider connector.
//
// Batch M.3 (2026-05-29). Born from the brainstorm run that picked
// qwen3.7-max (gateway-rejected) and glm-5 (streams empty) with no way for
// the agent to know in advance. These tools surface the M.1 capability
// registry + M.2 dispatch reality + J.11 spend caps:
//
//   - list_models(provider?, only_usable?, limit?)
//   - list_connectors()
//
// Both read from DB-resident state (stewards.model_catalog, model_capability,
// provider_spend_caps) — NOT providers_loaded(), whose in-memory registry is
// only populated in the bgworker process, not a plain SQL backend.

package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerModelTools wires up the M.3 read-only model + connector surface.
func registerModelTools(srv *mcp.Server, pool *pgxpool.Pool) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "list_models",
		Description: "List models the substrate knows about, with pricing and " +
			"dispatchability. Each row: provider, model, input/output micro-$ per " +
			"Mtok, usable (false = dispatch substitutes it — e.g. glm-5 streams " +
			"empty, qwen3.7-max is gateway-rejected), supports_streaming, " +
			"last_probed_at, probe_detail, probed_via (seed/manual/auto-probe/unprobed). " +
			"Filter by provider and/or only_usable. Use this before assigning a model " +
			"to a brainstorm lens so you don't pick one that gets substituted away.",
	}, makeListModels(pool))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "list_connectors",
		Description: "List provider connectors and their state: enforced spend cap " +
			"(micro-$), spend-since-refill, remaining, whether the cap is currently " +
			"exceeded (dispatch refused), and model counts (total / usable / unusable). " +
			"This is the budget + capability health of each provider the substrate can " +
			"dispatch to (e.g. google_gemini's prepaid cap, opencode_go's model set).",
	}, makeListConnectors(pool))
}

// ---------------------------------------------------------------------
// list_models
// ---------------------------------------------------------------------

type ListModelsInput struct {
	Provider   string `json:"provider,omitempty" jsonschema:"optional provider filter (e.g. opencode_go, google_gemini)"`
	OnlyUsable bool   `json:"only_usable,omitempty" jsonschema:"if true, return only models the substrate can dispatch (usable=true)"`
	Limit      int    `json:"limit,omitempty" jsonschema:"max models returned, default 100, capped at 500"`
}

type ModelRow struct {
	Provider          string     `json:"provider"`
	Model             string     `json:"model"`
	InputMicroPerMtok int64      `json:"input_micro_per_mtok"`
	OutputMicroPerMtok int64     `json:"output_micro_per_mtok"`
	Usable            bool       `json:"usable"`
	SupportsStreaming *bool      `json:"supports_streaming,omitempty"`
	LastProbedAt      *time.Time `json:"last_probed_at,omitempty"`
	ProbeDetail       string     `json:"probe_detail,omitempty"`
	ProbedVia         string     `json:"probed_via"`
	PricingNotes      string     `json:"pricing_notes,omitempty"`
}

type ListModelsOutput struct {
	Models []ModelRow `json:"models"`
	Count  int        `json:"count"`
}

func makeListModels(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in ListModelsInput,
) (*mcp.CallToolResult, ListModelsOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in ListModelsInput,
	) (*mcp.CallToolResult, ListModelsOutput, error) {
		if in.Limit <= 0 {
			in.Limit = 100
		}
		if in.Limit > 500 {
			in.Limit = 500
		}

		rows, err := pool.Query(ctx, `
			SELECT provider, model,
			       input_micro_per_mtok, output_micro_per_mtok,
			       usable, supports_streaming, last_probed_at,
			       COALESCE(probe_detail, ''), probed_via,
			       COALESCE(pricing_notes, '')
			  FROM stewards.model_catalog
			 WHERE ($1 = '' OR provider = $1)
			   AND (NOT $2 OR usable)
			 ORDER BY provider, NOT usable, model
			 LIMIT $3`, in.Provider, in.OnlyUsable, in.Limit)
		if err != nil {
			return toolError("list_models query: %v", err), ListModelsOutput{}, nil
		}
		defer rows.Close()

		var out []ModelRow
		for rows.Next() {
			var m ModelRow
			if err := rows.Scan(&m.Provider, &m.Model,
				&m.InputMicroPerMtok, &m.OutputMicroPerMtok,
				&m.Usable, &m.SupportsStreaming, &m.LastProbedAt,
				&m.ProbeDetail, &m.ProbedVia, &m.PricingNotes); err != nil {
				return toolError("list_models scan: %v", err), ListModelsOutput{}, nil
			}
			out = append(out, m)
		}
		if err := rows.Err(); err != nil {
			return toolError("list_models rows: %v", err), ListModelsOutput{}, nil
		}

		return nil, ListModelsOutput{Models: out, Count: len(out)}, nil
	}
}

// ---------------------------------------------------------------------
// list_connectors
// ---------------------------------------------------------------------

type ListConnectorsInput struct{}

type ConnectorRow struct {
	Provider       string `json:"provider"`
	Enforced       bool   `json:"enforced"`
	CapMicro       *int64 `json:"cap_micro,omitempty"`
	SpentMicro     int64  `json:"spent_since_refill_micro"`
	RemainingMicro *int64 `json:"remaining_micro,omitempty"`
	CapExceeded    bool   `json:"cap_exceeded"`
	ModelCount     int    `json:"model_count"`
	UsableCount    int    `json:"usable_count"`
	UnusableCount  int    `json:"unusable_count"`
}

type ListConnectorsOutput struct {
	Connectors []ConnectorRow `json:"connectors"`
	Count      int            `json:"count"`
}

func makeListConnectors(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in ListConnectorsInput,
) (*mcp.CallToolResult, ListConnectorsOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in ListConnectorsInput,
	) (*mcp.CallToolResult, ListConnectorsOutput, error) {
		// Base set = providers with priced models; left-joined to their cap
		// row. spend-since + cap-exceeded come from the J.11 functions so the
		// numbers match what the dispatcher actually gates on.
		rows, err := pool.Query(ctx, `
			SELECT c.provider,
			       COALESCE(sc.enforced, false)                  AS enforced,
			       sc.cap_micro,
			       stewards.provider_spend_since(c.provider)     AS spent_micro,
			       CASE WHEN sc.cap_micro IS NOT NULL
			            THEN sc.cap_micro - stewards.provider_spend_since(c.provider)
			       END                                           AS remaining_micro,
			       stewards.provider_cap_exceeded(c.provider)    AS cap_exceeded,
			       c.model_count, c.usable_count, c.unusable_count
			  FROM (
			      SELECT provider,
			             count(*)                          AS model_count,
			             count(*) FILTER (WHERE usable)     AS usable_count,
			             count(*) FILTER (WHERE NOT usable) AS unusable_count
			        FROM stewards.model_catalog
			       GROUP BY provider
			  ) c
			  LEFT JOIN stewards.provider_spend_caps sc ON sc.provider = c.provider
			 ORDER BY c.provider`)
		if err != nil {
			return toolError("list_connectors query: %v", err), ListConnectorsOutput{}, nil
		}
		defer rows.Close()

		var out []ConnectorRow
		for rows.Next() {
			var c ConnectorRow
			if err := rows.Scan(&c.Provider, &c.Enforced, &c.CapMicro,
				&c.SpentMicro, &c.RemainingMicro, &c.CapExceeded,
				&c.ModelCount, &c.UsableCount, &c.UnusableCount); err != nil {
				return toolError("list_connectors scan: %v", err), ListConnectorsOutput{}, nil
			}
			out = append(out, c)
		}
		if err := rows.Err(); err != nil {
			return toolError("list_connectors rows: %v", err), ListConnectorsOutput{}, nil
		}

		return nil, ListConnectorsOutput{Connectors: out, Count: len(out)}, nil
	}
}
