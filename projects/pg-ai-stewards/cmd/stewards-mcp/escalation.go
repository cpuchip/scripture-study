// Substrate escalation queue MCP tools — for CLI-mediated Opus boost
// of work_items the OpenCode chain has exhausted.
//
// Phase 4i v1 (2026-05-11). Three tools implementing the consumer side
// of the human-mediated escalation queue (D-EC3 from
// projects/pg-ai-stewards/.spec/proposals/full-agentic-substrate.md):
//
//   - work_item_escalation_list  — read queued items + context
//   - work_item_escalation_claim — atomic claim by claimer
//   - work_item_escalation_resolve — submit success/failure + output
//
// Two consumer paths supported:
//   (a) Stewards-UI button: claim with claimed_by='ui:zen-opus', then
//       dispatch via OpenCode Zen Opus, then resolve with the result.
//   (b) Claude Code CLI: list, claim with claimed_by='cli:claude-code-pro',
//       process locally with Pro subscription, resolve.
//
// State machine: queued → in_progress (claim) → resolved | failed (resolve).

package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerEscalationTools wires up the v1 escalation queue surface.
func registerEscalationTools(srv *mcp.Server, pool *pgxpool.Pool) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "work_item_escalation_list",
		Description: "List work_items currently in the escalation queue (escalation_state='queued'). " +
			"Returns each item's id, slug, pipeline, current_stage, failure context " +
			"(last_failure_reason + diagnosis + failure_count), escalation_attempts, " +
			"and the original input. Use work_item_escalation_claim to take ownership before processing.",
	}, makeEscalationList(pool))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "work_item_escalation_claim",
		Description: "Atomically claim a queued work_item for human-mediated processing. " +
			"Sets escalation_state='in_progress', escalation_claimed_by=<claimer>, " +
			"escalation_claimed_at=now(). Fails with an error if the item is already " +
			"claimed or not in 'queued' state. Returns the full context the claimer " +
			"needs: pipeline + stage + agent_family + composed input prompt + prior " +
			"stage_results. Use work_item_escalation_resolve to submit the result.",
	}, makeEscalationClaim(pool))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "work_item_escalation_resolve",
		Description: "Submit the result of a human-mediated escalation. On success, the " +
			"output is stored in stage_results, escalation_state='resolved', " +
			"model_override cleared, failure_count reset, status='pending' so the " +
			"next steward_tick or work_item_advance can continue the pipeline. " +
			"On failure (no Opus boost worked either), escalation_state='failed' and " +
			"the work_item is quarantined with reason='escalation_failed'.",
	}, makeEscalationResolve(pool))
}

// =====================================================================
// Tool 1: work_item_escalation_list
// =====================================================================

type EscalationListInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"max items returned, default 20, capped at 100"`
}

type EscalationItem struct {
	ID                   string          `json:"id"`
	Slug                 string          `json:"slug,omitempty"`
	PipelineFamily       string          `json:"pipeline_family"`
	CurrentStage         string          `json:"current_stage"`
	FailureCount         int             `json:"failure_count"`
	LastFailureReason    string          `json:"last_failure_reason,omitempty"`
	LastFailureDiagnosis string          `json:"last_failure_diagnosis,omitempty"`
	EscalationAttempts   int             `json:"escalation_attempts"`
	EscalationClaimedBy  string          `json:"escalation_claimed_by,omitempty"`
	Input                json.RawMessage `json:"input"`
	StageResults         json.RawMessage `json:"stage_results"`
	CostMicroDollars     int64           `json:"cost_micro_dollars"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

type EscalationListOutput struct {
	Items []EscalationItem `json:"items"`
	Count int              `json:"count"`
}

func makeEscalationList(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in EscalationListInput,
) (*mcp.CallToolResult, EscalationListOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in EscalationListInput,
	) (*mcp.CallToolResult, EscalationListOutput, error) {
		if in.Limit <= 0 {
			in.Limit = 20
		}
		if in.Limit > 100 {
			in.Limit = 100
		}

		rows, err := pool.Query(ctx, `
			SELECT id::text, coalesce(slug,''), pipeline_family, current_stage,
			       failure_count,
			       coalesce(last_failure_reason,''),
			       coalesce(last_failure_diagnosis,''),
			       escalation_attempts,
			       coalesce(escalation_claimed_by,''),
			       coalesce(input, '{}'::jsonb),
			       coalesce(stage_results, '{}'::jsonb),
			       cost_micro_dollars,
			       created_at, updated_at
			  FROM stewards.work_items
			 WHERE escalation_state = 'queued'
			 ORDER BY updated_at ASC
			 LIMIT $1`, in.Limit)
		if err != nil {
			return toolError("query escalation queue: %v", err), EscalationListOutput{}, nil
		}
		defer rows.Close()

		items := []EscalationItem{}
		for rows.Next() {
			var it EscalationItem
			var inputJSON, stageResultsJSON []byte
			if err := rows.Scan(
				&it.ID, &it.Slug, &it.PipelineFamily, &it.CurrentStage,
				&it.FailureCount,
				&it.LastFailureReason, &it.LastFailureDiagnosis,
				&it.EscalationAttempts, &it.EscalationClaimedBy,
				&inputJSON, &stageResultsJSON,
				&it.CostMicroDollars,
				&it.CreatedAt, &it.UpdatedAt,
			); err != nil {
				return toolError("scan: %v", err), EscalationListOutput{}, nil
			}
			it.Input = inputJSON
			it.StageResults = stageResultsJSON
			items = append(items, it)
		}
		if err := rows.Err(); err != nil {
			return toolError("iterate: %v", err), EscalationListOutput{}, nil
		}

		return nil, EscalationListOutput{Items: items, Count: len(items)}, nil
	}
}

// =====================================================================
// Tool 2: work_item_escalation_claim
// =====================================================================

type EscalationClaimInput struct {
	ID        string `json:"id" jsonschema:"work_item UUID to claim"`
	ClaimedBy string `json:"claimed_by" jsonschema:"identifier of the claimer (e.g. 'ui:zen-opus' or 'cli:claude-code-pro')"`
}

type EscalationClaimOutput struct {
	Claimed              bool            `json:"claimed"`
	Reason               string          `json:"reason,omitempty"`
	WorkItemID           string          `json:"work_item_id"`
	Slug                 string          `json:"slug,omitempty"`
	PipelineFamily       string          `json:"pipeline_family"`
	CurrentStage         string          `json:"current_stage"`
	AgentFamily          string          `json:"agent_family,omitempty"`
	StageDefaultModel    string          `json:"stage_default_model,omitempty"`
	StageDefaultProvider string          `json:"stage_default_provider,omitempty"`
	LastFailureReason    string          `json:"last_failure_reason,omitempty"`
	LastFailureDiagnosis string          `json:"last_failure_diagnosis,omitempty"`
	FailureCount         int             `json:"failure_count"`
	Input                json.RawMessage `json:"input,omitempty"`
	StageResults         json.RawMessage `json:"stage_results,omitempty"`
	StageInputRendered   string          `json:"stage_input_rendered,omitempty"`
	ClaimedBy            string          `json:"claimed_by,omitempty"`
	ClaimedAt            *time.Time      `json:"claimed_at,omitempty"`
}

func makeEscalationClaim(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in EscalationClaimInput,
) (*mcp.CallToolResult, EscalationClaimOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in EscalationClaimInput,
	) (*mcp.CallToolResult, EscalationClaimOutput, error) {
		if in.ID == "" {
			return toolError("id is required"), EscalationClaimOutput{}, nil
		}
		if in.ClaimedBy == "" {
			return toolError("claimed_by is required"), EscalationClaimOutput{}, nil
		}

		out := EscalationClaimOutput{WorkItemID: in.ID}

		// Atomic claim. WHERE escalation_state='queued' is the lock —
		// only one consumer can transition queued→in_progress.
		err := pool.QueryRow(ctx, `
			WITH claimed AS (
				UPDATE stewards.work_items
				   SET escalation_state = 'in_progress',
				       escalation_claimed_by = $2,
				       escalation_claimed_at = now()
				 WHERE id = $1::uuid AND escalation_state = 'queued'
				 RETURNING id, slug, pipeline_family, current_stage,
				           coalesce(last_failure_reason,''),
				           coalesce(last_failure_diagnosis,''),
				           failure_count,
				           coalesce(input, '{}'::jsonb),
				           coalesce(stage_results, '{}'::jsonb),
				           escalation_claimed_by,
				           escalation_claimed_at
			)
			SELECT
			    c.slug, c.pipeline_family, c.current_stage,
			    c.last_failure_reason, c.last_failure_diagnosis,
			    c.failure_count, c.input, c.stage_results,
			    c.escalation_claimed_by, c.escalation_claimed_at,
			    coalesce(stage->>'agent_family', '') AS agent_family,
			    coalesce(stage->>'model', '')        AS stage_model,
			    coalesce(stage->>'provider', '')     AS stage_provider,
			    coalesce(stewards.render_stage_input(c.id), '') AS rendered_input
			  FROM claimed c
			  LEFT JOIN LATERAL stewards.pipeline_stage_lookup(c.pipeline_family, c.current_stage) AS stage ON true`,
			in.ID, in.ClaimedBy,
		).Scan(
			&out.Slug, &out.PipelineFamily, &out.CurrentStage,
			&out.LastFailureReason, &out.LastFailureDiagnosis,
			&out.FailureCount, &out.Input, &out.StageResults,
			&out.ClaimedBy, &out.ClaimedAt,
			&out.AgentFamily, &out.StageDefaultModel, &out.StageDefaultProvider,
			&out.StageInputRendered,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				out.Claimed = false
				out.Reason = "work_item not found OR not in escalation_state='queued' (already claimed?)"
				return nil, out, nil
			}
			return toolError("claim: %v", err), out, nil
		}

		out.Claimed = true
		return nil, out, nil
	}
}

// =====================================================================
// Tool 3: work_item_escalation_resolve
// =====================================================================

type EscalationResolveInput struct {
	ID      string `json:"id" jsonschema:"work_item UUID to resolve"`
	Success bool   `json:"success" jsonschema:"true if the boost produced usable output; false if it failed"`
	Output  string `json:"output,omitempty" jsonschema:"the assistant's output to store as the stage result (success=true) or error context (success=false)"`
	Notes   string `json:"notes,omitempty" jsonschema:"optional human-readable notes added to steward_actions"`
}

type EscalationResolveOutput struct {
	Resolved        bool   `json:"resolved"`
	Reason          string `json:"reason,omitempty"`
	WorkItemID      string `json:"work_item_id"`
	NewState        string `json:"new_escalation_state"`
	NewStatus       string `json:"new_status"`
	StageStored     string `json:"stage_results_key,omitempty"`
	QuarantineFired bool   `json:"quarantined,omitempty"`
}

func makeEscalationResolve(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in EscalationResolveInput,
) (*mcp.CallToolResult, EscalationResolveOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in EscalationResolveInput,
	) (*mcp.CallToolResult, EscalationResolveOutput, error) {
		if in.ID == "" {
			return toolError("id is required"), EscalationResolveOutput{}, nil
		}

		out := EscalationResolveOutput{WorkItemID: in.ID}
		ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		tx, err := pool.BeginTx(ctxTimeout, pgx.TxOptions{})
		if err != nil {
			return toolError("begin tx: %v", err), out, nil
		}
		defer tx.Rollback(ctxTimeout)

		// Verify work_item is currently in_progress (must have been claimed first).
		var currentState, currentStage string
		err = tx.QueryRow(ctxTimeout, `
			SELECT escalation_state, current_stage
			  FROM stewards.work_items
			 WHERE id = $1::uuid
			 FOR UPDATE`,
			in.ID,
		).Scan(&currentState, &currentStage)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				out.Resolved = false
				out.Reason = "work_item not found"
				return nil, out, nil
			}
			return toolError("lock work_item: %v", err), out, nil
		}

		if currentState != "in_progress" {
			out.Resolved = false
			out.Reason = "work_item escalation_state is '" + currentState +
				"', expected 'in_progress' (claim first via work_item_escalation_claim)"
			return nil, out, nil
		}

		if in.Success {
			// Store output in stage_results->current_stage; reset
			// override + failure tracking; transition to 'pending' so
			// downstream (work_item_advance or another steward_tick)
			// can move forward.
			stageResult, _ := json.Marshal(map[string]any{
				"output":         in.Output,
				"source":         "escalation_resolve",
				"completed_at":   time.Now().UTC().Format(time.RFC3339),
				"resolver_notes": in.Notes,
			})

			_, err = tx.Exec(ctxTimeout, `
				UPDATE stewards.work_items
				   SET escalation_state         = 'resolved',
				       escalation_completed_at  = now(),
				       status                   = 'pending',
				       failure_count            = 0,
				       last_failure_reason      = NULL,
				       last_failure_diagnosis   = NULL,
				       model_override           = NULL,
				       provider_override        = NULL,
				       stage_results            = coalesce(stage_results, '{}'::jsonb)
				                                  || jsonb_build_object($2::text, $3::jsonb),
				       updated_at               = now()
				 WHERE id = $1::uuid`,
				in.ID, currentStage, stageResult,
			)
			if err != nil {
				return toolError("update work_item (success path): %v", err), out, nil
			}

			// Audit trail in steward_actions
			_, err = tx.Exec(ctxTimeout, `
				INSERT INTO stewards.steward_actions
				    (work_item_id, observation, diagnosis, action, model_used, details)
				VALUES
				    ($1::uuid,
				     'escalation resolved: ' || coalesce($2,'(no notes)'),
				     'escalated',
				     'escalation_resolved',
				     'human_mediated_opus',
				     jsonb_build_object(
				         'stage', $3::text,
				         'output_chars', length($4::text)))`,
				in.ID, in.Notes, currentStage, in.Output,
			)
			if err != nil {
				return toolError("audit (success): %v", err), out, nil
			}

			out.NewState = "resolved"
			out.NewStatus = "pending"
			out.StageStored = currentStage
		} else {
			// Boost failed too. Quarantine the work_item.
			_, err = tx.Exec(ctxTimeout, `
				UPDATE stewards.work_items
				   SET escalation_state         = 'failed',
				       escalation_completed_at  = now(),
				       quarantined_at           = now(),
				       quarantine_reason        = 'escalation_failed',
				       updated_at               = now()
				 WHERE id = $1::uuid`,
				in.ID,
			)
			if err != nil {
				return toolError("update work_item (failure path): %v", err), out, nil
			}

			_, err = tx.Exec(ctxTimeout, `
				INSERT INTO stewards.steward_actions
				    (work_item_id, observation, diagnosis, action, details)
				VALUES
				    ($1::uuid,
				     'escalation failed: ' || coalesce($2,'(no notes)'),
				     'escalated',
				     'escalation_failed',
				     jsonb_build_object(
				         'stage', $3::text,
				         'quarantine_reason', 'escalation_failed'))`,
				in.ID, in.Notes, currentStage,
			)
			if err != nil {
				return toolError("audit (failure): %v", err), out, nil
			}

			out.NewState = "failed"
			out.NewStatus = "failed"
			out.QuarantineFired = true
		}

		if err := tx.Commit(ctxTimeout); err != nil {
			return toolError("commit: %v", err), out, nil
		}

		out.Resolved = true
		return nil, out, nil
	}
}
