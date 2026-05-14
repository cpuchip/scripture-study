// spawn_subagent — substrate-internal MCP tool for K.4.
//
// Delegates verbose multi-turn work to a child work_item that runs in
// its own context. Synchronously waits for the child to reach a terminal
// state (verified / failed / cancelled), then returns the child's last
// assistant message as a prose digest to the parent.
//
// Ratification: sync sub-agent (block until child terminates).
// Cost protection: every spawn gets a cost_cap_micro (default $0.50).
// Depth protection: enforced by the SQL function (parent linkage) +
//   per-child cost cap. A future enhancement could explicitly walk the
//   parent_work_item_id chain to enforce a depth-2 limit; deferred.

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SpawnSubagentInput struct {
	PipelineFamily     string `json:"pipeline_family" jsonschema:"which pipeline the sub-agent runs (e.g. research-write study-write echo-test)"`
	BindingQuestion    string `json:"binding_question" jsonschema:"the specific question the sub-agent answers"`
	CostCapMicro       int64  `json:"cost_cap_micro,omitempty" jsonschema:"max micro-dollars (default 500000=$0.50)"`
	ProjectAssociation string `json:"project_association,omitempty" jsonschema:"optional project slug; inherits from parent if not set"`
	Slug               string `json:"slug,omitempty" jsonschema:"optional slug; auto-generated if absent"`
}

type SpawnSubagentOutput struct {
	ChildWorkItemID string `json:"child_work_item_id"`
	Slug            string `json:"slug"`
	Status          string `json:"status"`
	Maturity        string `json:"maturity"`
	Digest          string `json:"digest"`
	CostMicro       int64  `json:"cost_micro_dollars"`
	WallTimeSeconds int64  `json:"wall_time_seconds"`
}

// Sync-wait config. The hard ceiling matches the Phase-A periodic
// reaper threshold (15min); poll interval is 5s so we don't hammer
// the DB on long extractions.
const (
	spawnSubagentMaxWaitSeconds  = 1200 // 20 minutes ceiling per sub-agent
	spawnSubagentPollIntervalSec = 5
)

func registerSpawnSubagentTools(srv *mcp.Server, pool *pgxpool.Pool) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "spawn_subagent",
		Description: "Delegate verbose / multi-turn work to a child agent that runs in its own isolated context. " +
			"The child uses up to its own context budget exploring the binding_question; you only see the digest it returns. " +
			"Use for: deep research across multiple sources, audits over many files, surveys of related sessions. " +
			"DO NOT use for: a single cheap tool call (overhead exceeds savings), or work that needs to read/write your active state.",
	}, makeSpawnSubagent(pool))
}

func makeSpawnSubagent(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in SpawnSubagentInput,
) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in SpawnSubagentInput,
	) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
		if in.PipelineFamily == "" {
			return toolError("spawn_subagent: 'pipeline_family' is required"),
				SpawnSubagentOutput{}, nil
		}
		if in.BindingQuestion == "" {
			return toolError("spawn_subagent: 'binding_question' is required"),
				SpawnSubagentOutput{}, nil
		}

		costCap := in.CostCapMicro
		if costCap == 0 {
			costCap = 500_000 // $0.50 default
		}

		// 1. Create + dispatch the child via SQL.
		var childID string
		if err := pool.QueryRow(ctx,
			`SELECT stewards.spawn_subagent_create($1, $2, NULL, $3, $4, $5, 'subagent')::text`,
			in.PipelineFamily, in.BindingQuestion, costCap,
			nullableString(in.ProjectAssociation), nullableString(in.Slug),
		).Scan(&childID); err != nil {
			return toolError("spawn_subagent_create: %v", err),
				SpawnSubagentOutput{}, nil
		}

		// 2. Synchronously poll the child until terminal.
		start := time.Now()
		deadline := start.Add(time.Duration(spawnSubagentMaxWaitSeconds) * time.Second)
		var (
			status   string
			maturity string
			slug     string
			costMicro int64
		)
		for {
			err := pool.QueryRow(ctx,
				`SELECT status, maturity, slug, cost_micro_dollars
				   FROM stewards.work_items WHERE id = $1::uuid`,
				childID,
			).Scan(&status, &maturity, &slug, &costMicro)
			if err != nil {
				return toolError("spawn_subagent poll: %v (child=%s)", err, childID),
					SpawnSubagentOutput{}, nil
			}

			// Terminal status: failed/cancelled OR maturity=verified.
			if status == "failed" || status == "cancelled" {
				return finalize(pool, ctx, childID, status, maturity, slug, costMicro, start, true)
			}
			if maturity == "verified" {
				return finalize(pool, ctx, childID, status, maturity, slug, costMicro, start, false)
			}

			if time.Now().After(deadline) {
				return toolError(
					"spawn_subagent: child %s timed out after %ds (status=%s maturity=%s). "+
						"Inspect via work_item_show. The child may still complete async.",
					childID, spawnSubagentMaxWaitSeconds, status, maturity,
				), SpawnSubagentOutput{}, nil
			}

			select {
			case <-ctx.Done():
				return toolError("spawn_subagent: context cancelled while waiting on child %s", childID),
					SpawnSubagentOutput{}, nil
			case <-time.After(time.Duration(spawnSubagentPollIntervalSec) * time.Second):
				// continue polling
			}
		}
	}
}

// finalize fetches the child's last assistant message as the digest and
// returns the result.
func finalize(
	pool *pgxpool.Pool, ctx context.Context,
	childID, status, maturity, slug string, costMicro int64,
	start time.Time, isFailure bool,
) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
	// Pull the last assistant message from the child's session_ids.
	var digest string
	err := pool.QueryRow(ctx, `
		SELECT coalesce(m.content, '')
		  FROM stewards.work_items wi
		  JOIN stewards.messages m
		    ON m.session_id = ANY(wi.session_ids)
		 WHERE wi.id = $1::uuid
		   AND m.role = 'assistant'
		   AND coalesce(m.content,'') <> ''
		 ORDER BY m.created_at DESC, m.id DESC
		 LIMIT 1`, childID,
	).Scan(&digest)
	if err != nil {
		digest = fmt.Sprintf("(sub-agent %s reached terminal status=%s maturity=%s but no assistant message was found)",
			childID, status, maturity)
	}

	wallSec := int64(time.Since(start).Seconds())
	out := SpawnSubagentOutput{
		ChildWorkItemID: childID,
		Slug:            slug,
		Status:          status,
		Maturity:        maturity,
		Digest:          digest,
		CostMicro:       costMicro,
		WallTimeSeconds: wallSec,
	}

	// Build the model-facing markdown digest.
	header := fmt.Sprintf("[spawn_subagent %s complete in %ds, cost=$%.4f, status=%s, maturity=%s]",
		slug, wallSec, float64(costMicro)/1_000_000.0, status, maturity)

	body := digest
	if isFailure {
		body = "⚠️ Sub-agent did not reach verified maturity. Partial output below.\n\n" + body
	}
	body += fmt.Sprintf("\n\n(more available via expand_message on this sub-agent's session, "+
		"work_item_show(id=%s) for full state)", childID)

	rendered := header + "\n\n" + body
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: rendered},
		},
	}, out, nil
}

// nullableString returns nil for an empty string so the SQL parameter
// is NULL rather than ''.
func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
