// panel_redline — fan a document-redline mandate across a panel of models.
//
// Batch R.4 (2026-05-30). The generative analog of start_brainstorm: instead
// of "critique a binding question," it's "here is a document — each of you
// propose concrete edits." Thin wrapper over stewards.start_panel_redline,
// which reads the document server-side (R.2, so the panel needs no fs access),
// fans out one redline child per model (model+provider resolved, tools-off,
// 32k output, auto-scaled cost cap), and returns the parent. Mirrors the
// brainstorm.go read-back pattern.

package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerRedlineTools(srv *mcp.Server, pool *pgxpool.Pool) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "panel_redline",
		Description: "Fan a document-redline mandate across a PANEL of models. Each model receives the SAME " +
			"document (read server-side from the repo — .md/.markdown/.txt only) plus your mandate, and returns " +
			"location-anchored edit proposals (current snippet + proposed replacement + rationale + a " +
			"touches-quote/doctrine flag for human verification). Off-disk: proposals only — nothing is written " +
			"to the source. Pass `document` as a repo-relative path or a single-dir glob " +
			"(e.g. projects/scripture-book/src/chapters/*.md). Returns the parent + N children; monitor with " +
			"work_item_show, then condense the reports yourself. Use for diverse concrete edits across a manuscript/" +
			"spec/doc — NOT abstract critique (use start_brainstorm for that).",
	}, makePanelRedline(pool))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "panel_redline_condense",
		Description: "Optional: merge a completed panel_redline's N reports into ONE ranked, deduplicated proposal " +
			"menu via a chosen model (tools-off). Preserves every touches-quote/doctrine flag and notes per-edit " +
			"consensus (k of N panelists). Use AFTER the panel children finish (work_item_show shows them verified) " +
			"if you'd rather the substrate merge them than do it yourself. Returns the condense child id; read its " +
			"final assistant message for the menu.",
	}, makePanelRedlineCondense(pool))
}

type PanelRedlineInput struct {
	Document             string   `json:"document" jsonschema:"repo-relative path or single-dir filename glob of the document(s) to redline (.md/.markdown/.txt only)"`
	Mandate              string   `json:"mandate" jsonschema:"what edits to propose (e.g. 'tighten prose, cut filler' or 'find weak engineering parallels'). The model proposes; the human applies."`
	Models               []string `json:"models" jsonschema:"the panel of model names, e.g. [\"kimi-k2.6\",\"glm-5.1\",\"qwen3.6-plus\",\"gemini-2.5-flash\",\"deepseek-v4-flash\"]. Each runs the same document. Provider is resolved per model."`
	Destination          string   `json:"destination,omitempty" jsonschema:"optional index file path (defaults to study/.scratch/redline-<slug>-index.md)"`
	MaxTokens            int      `json:"max_tokens,omitempty" jsonschema:"per-model output ceiling, per API call (default 32000)"`
	CostCapPerModelMicro int64    `json:"cost_cap_per_model_micro,omitempty" jsonschema:"override per-model cost cap in micro-dollars (default: auto-scaled from document size + max_tokens)"`
	ProjectAssociation   string   `json:"project_association,omitempty" jsonschema:"optional project slug for tagging"`
	Slug                 string   `json:"slug,omitempty" jsonschema:"optional parent slug (defaults to redline-YYYYMMDD-HHMMSS)"`
}

type PanelRedlineChild struct {
	ID            string `json:"id"`
	Slug          string `json:"slug"`
	ModelOverride string `json:"model_override,omitempty"`
	CostCapMicro  int64  `json:"cost_cap_micro"`
}

type PanelRedlineOutput struct {
	ParentID     string              `json:"parent_id"`
	Slug         string              `json:"slug"`
	Destination  string              `json:"destination"`
	Children     []PanelRedlineChild `json:"children"`
	AggregatorID string              `json:"aggregator_id,omitempty"`
	Notes        string              `json:"notes,omitempty"`
}

func makePanelRedline(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in PanelRedlineInput,
) (*mcp.CallToolResult, PanelRedlineOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in PanelRedlineInput,
	) (*mcp.CallToolResult, PanelRedlineOutput, error) {
		if strings.TrimSpace(in.Document) == "" {
			return toolError("panel_redline: 'document' is required (repo-relative path or glob)"), PanelRedlineOutput{}, nil
		}
		if strings.TrimSpace(in.Mandate) == "" {
			return toolError("panel_redline: 'mandate' is required (what edits do you want?)"), PanelRedlineOutput{}, nil
		}
		if len(in.Models) == 0 {
			return toolError("panel_redline: 'models' must list at least one model"), PanelRedlineOutput{}, nil
		}
		maxTokens := in.MaxTokens
		if maxTokens <= 0 {
			maxTokens = 32000
		}

		var costCapArg any
		if in.CostCapPerModelMicro > 0 {
			costCapArg = in.CostCapPerModelMicro
		}

		var parentID string
		err := pool.QueryRow(ctx, `
			SELECT stewards.start_panel_redline(
				p_document                 := $1,
				p_mandate                  := $2,
				p_models                   := $3::text[],
				p_destination              := $4,
				p_actor                    := $5,
				p_slug                     := $6,
				p_max_tokens               := $7,
				p_cost_cap_per_model_micro := $8,
				p_project_association      := $9
			)::text`,
			in.Document, in.Mandate, in.Models,
			nullableString(in.Destination), "michael", nullableString(in.Slug),
			maxTokens, costCapArg, nullableString(in.ProjectAssociation),
		).Scan(&parentID)
		if err != nil {
			return toolError("panel_redline: start_panel_redline failed: %v", err), PanelRedlineOutput{}, nil
		}

		out := PanelRedlineOutput{ParentID: parentID}

		// Parent slug + the spawned children (and the aggregator).
		_ = pool.QueryRow(ctx, `SELECT slug FROM stewards.work_items WHERE id = $1::uuid`, parentID).Scan(&out.Slug)

		rows, err := pool.Query(ctx, `
			SELECT id::text, slug, COALESCE(model_override, ''), cost_cap_micro
			  FROM stewards.work_items
			 WHERE parent_work_item_id = $1::uuid
			   AND pipeline_family = 'redline'
			 ORDER BY slug`, parentID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var c PanelRedlineChild
				if err := rows.Scan(&c.ID, &c.Slug, &c.ModelOverride, &c.CostCapMicro); err == nil {
					out.Children = append(out.Children, c)
				}
			}
		}

		_ = pool.QueryRow(ctx, `
			SELECT id::text, COALESCE(input->>'destination', '')
			  FROM stewards.work_items
			 WHERE parent_work_item_id = $1::uuid
			   AND pipeline_family = 'aggregate-children'
			 LIMIT 1`, parentID).Scan(&out.AggregatorID, &out.Destination)

		out.Notes = fmt.Sprintf("Panel dispatched to %d model(s). Each child redlines the document tools-off at "+
			"max_tokens=%d. Monitor with work_item_show on the child ids; read each child's final assistant "+
			"message for its redline report, then condense. Proposals only — nothing is written to the source.",
			len(out.Children), maxTokens)

		return nil, out, nil
	}
}

// ---------------------------------------------------------------------
// panel_redline_condense — optional substrate-side ranked merge (R.5)
// ---------------------------------------------------------------------

type PanelRedlineCondenseInput struct {
	ParentID      string `json:"parent_id" jsonschema:"the parent work_item id returned by panel_redline"`
	CondenseModel string `json:"condense_model" jsonschema:"the model to merge the reports (e.g. gemini-2.5-flash, kimi-k2.6)"`
	MaxTokens     int    `json:"max_tokens,omitempty" jsonschema:"output ceiling for the merge (default 32000)"`
	CostCapMicro  int64  `json:"cost_cap_micro,omitempty" jsonschema:"override cost cap in micro-dollars (default: auto-scaled)"`
}

type PanelRedlineCondenseOutput struct {
	CondenseChildID string `json:"condense_child_id"`
	Notes           string `json:"notes,omitempty"`
}

func makePanelRedlineCondense(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in PanelRedlineCondenseInput,
) (*mcp.CallToolResult, PanelRedlineCondenseOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in PanelRedlineCondenseInput,
	) (*mcp.CallToolResult, PanelRedlineCondenseOutput, error) {
		if strings.TrimSpace(in.ParentID) == "" {
			return toolError("panel_redline_condense: 'parent_id' is required"), PanelRedlineCondenseOutput{}, nil
		}
		if strings.TrimSpace(in.CondenseModel) == "" {
			return toolError("panel_redline_condense: 'condense_model' is required"), PanelRedlineCondenseOutput{}, nil
		}
		maxTokens := in.MaxTokens
		if maxTokens <= 0 {
			maxTokens = 32000
		}
		var costCapArg any
		if in.CostCapMicro > 0 {
			costCapArg = in.CostCapMicro
		}

		var childID string
		err := pool.QueryRow(ctx,
			`SELECT stewards.panel_redline_condense($1::uuid, $2, $3, $4)::text`,
			in.ParentID, in.CondenseModel, maxTokens, costCapArg,
		).Scan(&childID)
		if err != nil {
			return toolError("panel_redline_condense: failed: %v", err), PanelRedlineCondenseOutput{}, nil
		}

		return nil, PanelRedlineCondenseOutput{
			CondenseChildID: childID,
			Notes: fmt.Sprintf("Condense dispatched via %s. Read the final assistant message on work_item %s "+
				"for the ranked merged menu (preserves every touches-quote/doctrine flag).", in.CondenseModel, childID),
		}, nil
	}
}
