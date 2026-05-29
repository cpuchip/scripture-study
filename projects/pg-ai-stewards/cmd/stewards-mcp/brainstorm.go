// Brainstorm MCP tool — wraps stewards.start_brainstorm() so Claude Code
// can dispatch a multi-lens brainstorm without dropping into psql.
//
// Backs the J.8 + J.9 SQL work landed 2026-05-29:
//   - 12 lenses available (existing 4 + 8 new from J.9)
//   - per-lens model override via p_models
//   - per-call lens subset via p_lenses
//   - 4-layer dispatch fallback chain (J.8.a)
//
// The tool is a thin wrapper over the SQL function — it pre-validates
// the lens list (so callers get a helpful tool-level error rather than
// a SQL EXCEPTION), defaults p_destination to a scratch path keyed by
// slug + timestamp, and returns the parent_id plus the spawned child
// summary so the caller can monitor without a follow-up call.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerBrainstormTools wires the start_brainstorm wrapper into the
// MCP server. Single tool — see decisions ratified 2026-05-29.
func registerBrainstormTools(srv *mcp.Server, pool *pgxpool.Pool) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "start_brainstorm",
		Description: "Dispatch a multi-lens brainstorm across the pg-ai-stewards substrate. " +
			"12 lens techniques available (SCAMPER, Six Hats, Crazy 8s, Reverse, " +
			"Mind Mapping, Brainwriting, Starbursting/5W1H, Disney Method, " +
			"Storyboarding, TRIZ, Forced Analogy, Worst Possible Idea). " +
			"Each runs in parallel and a synthesis aggregator combines them. " +
			"Default lens subset is the original 4 (scamper, six-hats, crazy8s, " +
			"reverse) for backward compat. Pass `lenses` to pick a subset of the 12. " +
			"Pass `models` to override per-lens model (object keyed by short lens " +
			"name, value is either a model string or {model, provider} object). " +
			"Returns parent_id + spawned children + manifest summary.",
	}, makeStartBrainstorm(pool))
}

// ---------------------------------------------------------------------
// start_brainstorm
// ---------------------------------------------------------------------

type StartBrainstormInput struct {
	BindingQuestion       string          `json:"binding_question" jsonschema:"the question the brainstorm should answer (required, 1-3 sentences)"`
	Destination           string          `json:"destination,omitempty" jsonschema:"path the aggregator index materializes to (defaults to study/.scratch/brainstorm-{slug}.md)"`
	Lenses                []string        `json:"lenses,omitempty" jsonschema:"subset of lens short names: scamper, six-hats, crazy8s, reverse, mind-mapping, brainwriting, starbursting, disney, storyboarding, triz, forced-analogy, worst-idea. Default = first 4 (today's behavior)."`
	Models                json.RawMessage `json:"models,omitempty" jsonschema:"per-lens model overrides as a JSON object keyed by short lens name. Values: model string ('opus-4.7') or {model, provider} object. Omitted lenses use the J.8.a fallback chain."`
	ProjectAssociation    string          `json:"project_association,omitempty" jsonschema:"optional project slug for tagging"`
	Slug                  string          `json:"slug,omitempty" jsonschema:"optional parent work_item slug (defaults to brainstorm-YYYYMMDD-HHMMSS)"`
	Actor                 string          `json:"actor,omitempty" jsonschema:"actor recorded on the work_item (default 'michael')"`
	CostCapPerLensMicro   int64           `json:"cost_cap_per_lens_micro,omitempty" jsonschema:"micro-dollar cost cap per child lens dispatch (default 200000 = $0.20)"`
}

type BrainstormChildSummary struct {
	ID             string  `json:"id"`
	Slug           string  `json:"slug"`
	PipelineFamily string  `json:"pipeline_family"`
	ModelOverride  *string `json:"model_override,omitempty"`
}

type StartBrainstormOutput struct {
	ParentID       string                   `json:"parent_id"`
	Slug           string                   `json:"slug"`
	Destination    string                   `json:"destination"`
	LensesUsed     []string                 `json:"lenses_used"`
	Children       []BrainstormChildSummary `json:"children"`
	AggregatorID   string                   `json:"aggregator_id"`
	Notes          string                   `json:"notes,omitempty"`
}

func makeStartBrainstorm(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in StartBrainstormInput,
) (*mcp.CallToolResult, StartBrainstormOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in StartBrainstormInput,
	) (*mcp.CallToolResult, StartBrainstormOutput, error) {
		if strings.TrimSpace(in.BindingQuestion) == "" {
			return toolError("start_brainstorm: 'binding_question' is required"),
				StartBrainstormOutput{}, nil
		}

		// Default actor — matches the SQL function's default.
		if in.Actor == "" {
			in.Actor = "michael"
		}

		// Default slug — matches the SQL function's pattern but resolve
		// here so we can use the same value to default `destination` AND
		// to return it to the caller.
		slug := in.Slug
		if slug == "" {
			slug = "brainstorm-" + time.Now().UTC().Format("20060102-150405")
		}

		// Default destination — substrate path under study/.scratch/. The
		// autonomous materializer (am1, commit 767386a 2026-05-22) writes
		// the file when the aggregator verifies; no manual git work needed.
		destination := in.Destination
		if destination == "" {
			destination = "study/.scratch/" + slug + ".md"
		}

		// Default lens list — let the SQL function's DEFAULT kick in by
		// passing NULL. This avoids drift between the Go default and the
		// SQL default; the source of truth is the SQL signature.
		var lensesArg any
		if len(in.Lenses) > 0 {
			// Pre-validate lens names at the tool layer for a friendlier
			// error than the SQL EXCEPTION. We pull the available lens
			// names from the pipelines table.
			available, err := loadAvailableLenses(ctx, pool)
			if err != nil {
				return toolError("start_brainstorm: could not load available lenses: %v", err),
					StartBrainstormOutput{}, nil
			}
			unknown := []string{}
			for _, lens := range in.Lenses {
				if _, ok := available[lens]; !ok {
					unknown = append(unknown, lens)
				}
			}
			if len(unknown) > 0 {
				availableList := make([]string, 0, len(available))
				for name := range available {
					availableList = append(availableList, name)
				}
				return toolError(
					"start_brainstorm: unknown lens name(s): %v. Available lenses: %v",
					unknown, availableList,
				), StartBrainstormOutput{}, nil
			}
			lensesArg = in.Lenses
		}

		// Default models — pass NULL so the SQL fallback chain handles
		// every layer. If caller supplied models, pass the raw JSON.
		var modelsArg any
		if len(in.Models) > 0 {
			// Sanity-validate that it's actually JSON. Don't try to
			// validate keys against lens names — that would couple us to
			// the SQL function's lens list. The SQL function silently
			// ignores unknown keys (forward-compat for future J.10+ lens
			// expansions); the MCP layer should do the same.
			var probe any
			if err := json.Unmarshal(in.Models, &probe); err != nil {
				return toolError("start_brainstorm: 'models' is not valid JSON: %v", err),
					StartBrainstormOutput{}, nil
			}
			modelsArg = string(in.Models)
		}

		// Default cost cap — pass nil so SQL default (200000 = $0.20) wins.
		var costCapArg any
		if in.CostCapPerLensMicro > 0 {
			costCapArg = in.CostCapPerLensMicro
		}

		// Project association — pass nil if not set so SQL DEFAULT NULL kicks in.
		var projectArg any
		if in.ProjectAssociation != "" {
			projectArg = in.ProjectAssociation
		}

		// Call the SQL function. Note: passing nil for lenses argument lets
		// the SQL DEFAULT kick in, preserving today's behavior.
		var parentID string
		err := pool.QueryRow(ctx, `
			SELECT stewards.start_brainstorm(
				p_binding_question        := $1,
				p_destination             := $2,
				p_project_association     := $3,
				p_actor                   := $4,
				p_slug                    := $5,
				p_cost_cap_per_lens_micro := $6,
				p_models                  := $7::jsonb,
				p_lenses                  := $8::text[]
			)::text`,
			in.BindingQuestion,
			destination,
			projectArg,
			in.Actor,
			slug,
			costCapArg,
			modelsArg,
			lensesArg,
		).Scan(&parentID)
		if err != nil {
			return toolError("start_brainstorm: SQL call failed: %v", err),
				StartBrainstormOutput{}, nil
		}

		// Read back the parent's manifest + spawned children for the
		// caller's convenience. The manifest is in stage_results.decompose.output.
		var manifestRaw string
		err = pool.QueryRow(ctx, `
			SELECT stage_results->'decompose'->>'output'
			  FROM stewards.work_items
			 WHERE id = $1::uuid`, parentID).Scan(&manifestRaw)
		if err != nil {
			// Non-fatal — parent created, just couldn't read manifest back.
			return nil, StartBrainstormOutput{
				ParentID:    parentID,
				Slug:        slug,
				Destination: destination,
				Notes:       fmt.Sprintf("parent created; manifest read-back failed: %v", err),
			}, nil
		}

		// Pull the child work_items the trigger spawned.
		rows, err := pool.Query(ctx, `
			SELECT id::text, slug, pipeline_family, model_override
			  FROM stewards.work_items
			 WHERE parent_work_item_id = $1::uuid
			   AND pipeline_family LIKE 'brainstorm-%'
			 ORDER BY slug`, parentID)
		if err != nil {
			return nil, StartBrainstormOutput{
				ParentID:    parentID,
				Slug:        slug,
				Destination: destination,
				Notes:       fmt.Sprintf("parent created; child read-back failed: %v", err),
			}, nil
		}
		defer rows.Close()

		var children []BrainstormChildSummary
		lensesUsed := []string{}
		for rows.Next() {
			var c BrainstormChildSummary
			if err := rows.Scan(&c.ID, &c.Slug, &c.PipelineFamily, &c.ModelOverride); err != nil {
				return toolError("start_brainstorm: child row scan: %v", err),
					StartBrainstormOutput{}, nil
			}
			children = append(children, c)
			lensesUsed = append(lensesUsed, strings.TrimPrefix(c.PipelineFamily, "brainstorm-"))
		}

		// Pull the aggregator child too — it has a different pipeline_family.
		var aggregatorID string
		_ = pool.QueryRow(ctx, `
			SELECT id::text
			  FROM stewards.work_items
			 WHERE parent_work_item_id = $1::uuid
			   AND pipeline_family = 'aggregate-children'
			 LIMIT 1`, parentID).Scan(&aggregatorID)

		return nil, StartBrainstormOutput{
			ParentID:     parentID,
			Slug:         slug,
			Destination:  destination,
			LensesUsed:   lensesUsed,
			Children:     children,
			AggregatorID: aggregatorID,
		}, nil
	}
}

// loadAvailableLenses returns the set of short lens names (e.g. "scamper",
// "mind-mapping") currently registered in stewards.pipelines. Used for
// caller-friendly validation before dispatching to the SQL function.
func loadAvailableLenses(ctx context.Context, pool *pgxpool.Pool) (map[string]struct{}, error) {
	rows, err := pool.Query(ctx, `
		SELECT regexp_replace(family, '^brainstorm-', '')
		  FROM stewards.pipelines
		 WHERE family LIKE 'brainstorm-%'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]struct{}{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out[name] = struct{}{}
	}
	return out, rows.Err()
}
