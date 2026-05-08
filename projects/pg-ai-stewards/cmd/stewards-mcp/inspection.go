// Substrate inspection tools — read-only views into work_items and
// watchman_passes for triage from Claude Code without `docker exec`.
//
// Phase 3e.4 v1 (2026-05-08). All four tools are read-only:
//   - work_item_list(pipeline?, status?, limit?)
//   - work_item_show(id_or_slug)
//   - watchman_passes_list(limit?)
//   - watchman_pass_show(pass_id)
//
// Write-mutating tools (work_item_create, work_item_dispatch,
// watchman_pass_now) are deferred to a future phase pending scope
// decision. The substrate's token_budget per work_item bounds blast
// radius once they ship.

package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerInspectionTools wires up the v1 (Phase 3e.4) read-only
// inspection surface for work_items and watchman_passes.
func registerInspectionTools(srv *mcp.Server, pool *pgxpool.Pool) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "work_item_list",
		Description: "List recent work_items in the substrate, optionally " +
			"filtered by pipeline_family and/or status. Returns id, slug, " +
			"pipeline, current_stage, status, token totals, actor, and " +
			"timestamps. Use work_item_show afterward to read stage_results " +
			"or full input for one item.",
	}, makeWorkItemList(pool))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "work_item_show",
		Description: "Show full details for a single work_item by UUID or slug. " +
			"Returns all metadata plus stage_results (the per-stage JSONB output " +
			"of the pipeline) and input (the original binding question or args).",
	}, makeWorkItemShow(pool))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "watchman_passes_list",
		Description: "List recent Watchman passes with status, timing, doc counts, " +
			"token totals, and verdict_counts (clean/drift/done/superseded/skipped). " +
			"Use watchman_pass_show afterward to read the per-doc verdicts and " +
			"reasoning for one pass.",
	}, makeWatchmanPassesList(pool))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "watchman_pass_show",
		Description: "Show one Watchman pass plus the per-doc verdicts it produced. " +
			"Each verdict includes study_id, verdict (clean/drift/done/superseded/skipped), " +
			"reasoning, model, and tokens. Findings (drift details) are surfaced separately.",
	}, makeWatchmanPassShow(pool))
}

// ---------------------------------------------------------------------
// work_item_list
// ---------------------------------------------------------------------

type WorkItemListInput struct {
	Pipeline string `json:"pipeline,omitempty" jsonschema:"optional pipeline_family filter (e.g. study-write study-write-qwen echo-test)"`
	Status   string `json:"status,omitempty" jsonschema:"optional status filter (pending in_progress completed failed cancelled awaiting_review)"`
	Limit    int    `json:"limit,omitempty" jsonschema:"max items returned, default 20, capped at 100"`
}

type WorkItemSummary struct {
	ID             string     `json:"id"`
	Slug           string     `json:"slug,omitempty"`
	PipelineFamily string     `json:"pipeline_family"`
	CurrentStage   string     `json:"current_stage"`
	Status         string     `json:"status"`
	TokensIn       int        `json:"tokens_in"`
	TokensOut      int        `json:"tokens_out"`
	TokenBudget    int        `json:"token_budget"`
	Actor          string     `json:"actor"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

type WorkItemListOutput struct {
	Items []WorkItemSummary `json:"items"`
	Count int               `json:"count"`
}

func makeWorkItemList(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in WorkItemListInput,
) (*mcp.CallToolResult, WorkItemListOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in WorkItemListInput,
	) (*mcp.CallToolResult, WorkItemListOutput, error) {
		if in.Limit <= 0 {
			in.Limit = 20
		}
		if in.Limit > 100 {
			in.Limit = 100
		}

		// Build query with optional filters. We use COALESCE-on-empty-string
		// rather than dynamic SQL fragments to keep the call site one shape
		// for pgx parameter binding.
		query := `
			SELECT id::text, COALESCE(slug,''), pipeline_family, current_stage,
			       status, tokens_in, tokens_out, token_budget,
			       actor, created_at, updated_at, completed_at
			  FROM stewards.work_items
			 WHERE ($1 = '' OR pipeline_family = $1)
			   AND ($2 = '' OR status = $2)
			 ORDER BY created_at DESC
			 LIMIT $3`

		rows, err := pool.Query(ctx, query, in.Pipeline, in.Status, in.Limit)
		if err != nil {
			return toolError("work_item_list query: %v", err),
				WorkItemListOutput{}, nil
		}
		defer rows.Close()

		var items []WorkItemSummary
		for rows.Next() {
			var s WorkItemSummary
			if err := rows.Scan(&s.ID, &s.Slug, &s.PipelineFamily, &s.CurrentStage,
				&s.Status, &s.TokensIn, &s.TokensOut, &s.TokenBudget,
				&s.Actor, &s.CreatedAt, &s.UpdatedAt, &s.CompletedAt); err != nil {
				return toolError("work_item_list scan: %v", err),
					WorkItemListOutput{}, nil
			}
			items = append(items, s)
		}
		if err := rows.Err(); err != nil {
			return toolError("work_item_list rows: %v", err),
				WorkItemListOutput{}, nil
		}

		return nil, WorkItemListOutput{Items: items, Count: len(items)}, nil
	}
}

// ---------------------------------------------------------------------
// work_item_show
// ---------------------------------------------------------------------

type WorkItemShowInput struct {
	IDOrSlug string `json:"id_or_slug" jsonschema:"work_item UUID or slug"`
}

// WorkItemDetail is map[string]any so callers see whatever fields the
// substrate decides to expose, including the rich JSONB columns
// (stage_results, input).
type WorkItemDetail map[string]any

func makeWorkItemShow(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in WorkItemShowInput,
) (*mcp.CallToolResult, WorkItemDetail, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in WorkItemShowInput,
	) (*mcp.CallToolResult, WorkItemDetail, error) {
		if in.IDOrSlug == "" {
			return toolError("work_item_show: 'id_or_slug' is required"), nil, nil
		}

		// Try id first (uuid), fall back to slug. row_to_json builds the
		// envelope; we ::jsonb explicitly so pgx scans into bytes for
		// json.Unmarshal.
		query := `
			SELECT to_jsonb(wi)::text
			  FROM stewards.work_items wi
			 WHERE wi.id::text = $1 OR wi.slug = $1
			 LIMIT 1`

		var raw string
		err := pool.QueryRow(ctx, query, in.IDOrSlug).Scan(&raw)
		if err != nil {
			return toolError("work_item_show query: %v (id_or_slug=%q)",
				err, in.IDOrSlug), nil, nil
		}

		var out WorkItemDetail
		if err := json.Unmarshal([]byte(raw), &out); err != nil {
			return toolError("work_item_show decode: %v", err), nil, nil
		}
		if len(out) == 0 {
			return toolError("work_item_show: no work_item with id_or_slug %q",
				in.IDOrSlug), nil, nil
		}
		return nil, out, nil
	}
}

// ---------------------------------------------------------------------
// watchman_passes_list
// ---------------------------------------------------------------------

type WatchmanPassesListInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"max passes returned, default 10, capped at 50"`
}

type WatchmanPassSummary struct {
	PassID          string     `json:"pass_id"`
	Status          string     `json:"status"`
	Trigger         string     `json:"trigger"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	Provider        string     `json:"provider"`
	Model           string     `json:"model"`
	AgentFamily     string     `json:"agent_family"`
	DocCountPlanned int        `json:"doc_count_planned"`
	DocCountDone    int        `json:"doc_count_done"`
	TokensIn        int        `json:"tokens_in"`
	TokensOut       int        `json:"tokens_out"`
	TokenBudget     int        `json:"token_budget"`
	BudgetStopped   bool       `json:"budget_stopped"`
	VerdictCounts   any        `json:"verdict_counts"`
}

type WatchmanPassesListOutput struct {
	Passes []WatchmanPassSummary `json:"passes"`
	Count  int                   `json:"count"`
}

func makeWatchmanPassesList(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in WatchmanPassesListInput,
) (*mcp.CallToolResult, WatchmanPassesListOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in WatchmanPassesListInput,
	) (*mcp.CallToolResult, WatchmanPassesListOutput, error) {
		if in.Limit <= 0 {
			in.Limit = 10
		}
		if in.Limit > 50 {
			in.Limit = 50
		}

		rows, err := pool.Query(ctx, `
			SELECT pass_id, status, trigger, started_at, finished_at,
			       provider, model, agent_family,
			       doc_count_planned, doc_count_done,
			       tokens_in, tokens_out, token_budget, budget_stopped,
			       verdict_counts
			  FROM stewards.watchman_passes
			 ORDER BY started_at DESC
			 LIMIT $1`, in.Limit)
		if err != nil {
			return toolError("watchman_passes_list query: %v", err),
				WatchmanPassesListOutput{}, nil
		}
		defer rows.Close()

		var passes []WatchmanPassSummary
		for rows.Next() {
			var p WatchmanPassSummary
			var vc []byte
			if err := rows.Scan(&p.PassID, &p.Status, &p.Trigger,
				&p.StartedAt, &p.FinishedAt,
				&p.Provider, &p.Model, &p.AgentFamily,
				&p.DocCountPlanned, &p.DocCountDone,
				&p.TokensIn, &p.TokensOut, &p.TokenBudget, &p.BudgetStopped,
				&vc); err != nil {
				return toolError("watchman_passes_list scan: %v", err),
					WatchmanPassesListOutput{}, nil
			}
			// Decode verdict_counts jsonb into a generic map.
			var vcMap any
			_ = json.Unmarshal(vc, &vcMap)
			p.VerdictCounts = vcMap
			passes = append(passes, p)
		}
		if err := rows.Err(); err != nil {
			return toolError("watchman_passes_list rows: %v", err),
				WatchmanPassesListOutput{}, nil
		}

		return nil, WatchmanPassesListOutput{Passes: passes, Count: len(passes)}, nil
	}
}

// ---------------------------------------------------------------------
// watchman_pass_show
// ---------------------------------------------------------------------

type WatchmanPassShowInput struct {
	PassID string `json:"pass_id" jsonschema:"pass_id like watchman-20260508T142906Z-50d7be"`
}

type WatchmanVerdict struct {
	StudyID   string    `json:"study_id"`
	Verdict   string    `json:"verdict"`
	Reasoning string    `json:"reasoning"`
	Model     string    `json:"model"`
	TokensIn  int       `json:"tokens_in"`
	TokensOut int       `json:"tokens_out"`
	Actor     string    `json:"actor"`
	CreatedAt time.Time `json:"created_at"`
}

type WatchmanPassShowOutput struct {
	Pass     WatchmanPassSummary `json:"pass"`
	Verdicts []WatchmanVerdict   `json:"verdicts"`
}

func makeWatchmanPassShow(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in WatchmanPassShowInput,
) (*mcp.CallToolResult, WatchmanPassShowOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in WatchmanPassShowInput,
	) (*mcp.CallToolResult, WatchmanPassShowOutput, error) {
		if in.PassID == "" {
			return toolError("watchman_pass_show: 'pass_id' is required"),
				WatchmanPassShowOutput{}, nil
		}

		// Pass header
		var p WatchmanPassSummary
		var vc []byte
		err := pool.QueryRow(ctx, `
			SELECT pass_id, status, trigger, started_at, finished_at,
			       provider, model, agent_family,
			       doc_count_planned, doc_count_done,
			       tokens_in, tokens_out, token_budget, budget_stopped,
			       verdict_counts
			  FROM stewards.watchman_passes WHERE pass_id = $1`, in.PassID).Scan(
			&p.PassID, &p.Status, &p.Trigger, &p.StartedAt, &p.FinishedAt,
			&p.Provider, &p.Model, &p.AgentFamily,
			&p.DocCountPlanned, &p.DocCountDone,
			&p.TokensIn, &p.TokensOut, &p.TokenBudget, &p.BudgetStopped,
			&vc)
		if err != nil {
			return toolError("watchman_pass_show pass query: %v (pass_id=%q)",
				err, in.PassID), WatchmanPassShowOutput{}, nil
		}
		var vcMap any
		_ = json.Unmarshal(vc, &vcMap)
		p.VerdictCounts = vcMap

		// Verdicts
		rows, err := pool.Query(ctx, `
			SELECT study_id, verdict, COALESCE(reasoning,''), COALESCE(model,''),
			       COALESCE(tokens_in,0), COALESCE(tokens_out,0),
			       actor, created_at
			  FROM stewards.verdicts
			 WHERE pass_id = $1
			 ORDER BY created_at`, in.PassID)
		if err != nil {
			return toolError("watchman_pass_show verdicts query: %v", err),
				WatchmanPassShowOutput{}, nil
		}
		defer rows.Close()

		var verdicts []WatchmanVerdict
		for rows.Next() {
			var v WatchmanVerdict
			if err := rows.Scan(&v.StudyID, &v.Verdict, &v.Reasoning, &v.Model,
				&v.TokensIn, &v.TokensOut, &v.Actor, &v.CreatedAt); err != nil {
				return toolError("watchman_pass_show verdicts scan: %v", err),
					WatchmanPassShowOutput{}, nil
			}
			verdicts = append(verdicts, v)
		}
		if err := rows.Err(); err != nil {
			return toolError("watchman_pass_show verdicts rows: %v", err),
				WatchmanPassShowOutput{}, nil
		}

		return nil, WatchmanPassShowOutput{Pass: p, Verdicts: verdicts}, nil
	}
}

