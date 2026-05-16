// consult_subagent — substrate-internal MCP tool for ES.3.s3.
//
// The companion to spawn_subagent. spawn_subagent creates a child and
// returns its digest; the sub-agent's session then PERSISTS. consult_
// subagent sends that sub-agent a NEW question in the context it
// already built — a report you file once becomes a steward you can
// send back.
//
// The judge that compiled a brief from an oversized fetch is the main
// caller's target: consult_subagent re-reads the SAME document on a new
// angle, without re-fetching it. Generalizes to any sub-agent session.
//
// Synchronous, mirroring spawn_subagent: block until the re-engagement
// chat reaches a terminal state, then return the sub-agent's answer.

package main

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ConsultSubagentInput struct {
	Target   string `json:"target" jsonschema:"session_id (e.g. judge-5636) or a spawned work_item uuid"`
	Question string `json:"question" jsonschema:"the new question for the sub-agent — answered from the context it already holds"`
}

type ConsultSubagentOutput struct {
	SessionID       string `json:"session_id"`
	ReaskIndex      int    `json:"reask_index"`
	Status          string `json:"status"`
	Answer          string `json:"answer"`
	WallTimeSeconds int64  `json:"wall_time_seconds"`
}

// Sync-wait config. A re-ask is one LLM call; 15min is a generous
// ceiling (a 1M-token judge re-read is the slow case).
const (
	consultMaxWaitSeconds  = 900
	consultPollIntervalSec = 5
)

// Standard UUID shape — distinguishes a work_item id from a session id.
var consultUUIDRe = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func registerConsultSubagentTools(srv *mcp.Server, pool *pgxpool.Pool) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "consult_subagent",
		Description: "Re-engage a sub-agent you (or the substrate) already spawned — send it a NEW question " +
			"in the context it already built. For a judge that compiled a brief from an oversized fetch, " +
			"this re-reads the SAME document on a new angle without re-fetching it. You only see the answer. " +
			"Pass target = a session_id (a judge brief names its session as judge-<id>) or a sub-agent work_item uuid.",
	}, makeConsultSubagent(pool))
}

func makeConsultSubagent(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in ConsultSubagentInput,
) (*mcp.CallToolResult, ConsultSubagentOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in ConsultSubagentInput,
	) (*mcp.CallToolResult, ConsultSubagentOutput, error) {
		if in.Target == "" {
			return toolError("consult_subagent: 'target' is required"),
				ConsultSubagentOutput{}, nil
		}
		if in.Question == "" {
			return toolError("consult_subagent: 'question' is required"),
				ConsultSubagentOutput{}, nil
		}

		// Resolve target -> session_id. A uuid is a work_item; resolve
		// to its most recent session. Otherwise target IS the session.
		sessionID := in.Target
		if consultUUIDRe.MatchString(in.Target) {
			var resolved string
			err := pool.QueryRow(ctx,
				`SELECT session_ids[array_length(session_ids, 1)]
				   FROM stewards.work_items WHERE id = $1::uuid`,
				in.Target,
			).Scan(&resolved)
			if err != nil {
				return toolError("consult_subagent: work_item %s not found or has no session (%v)",
					in.Target, err), ConsultSubagentOutput{}, nil
			}
			if resolved == "" {
				return toolError("consult_subagent: work_item %s has no sessions to consult", in.Target),
					ConsultSubagentOutput{}, nil
			}
			sessionID = resolved
		}

		// 1. Dispatch the re-engagement chat via SQL.
		var (
			chatWQ     int64
			reaskIndex int
		)
		if err := pool.QueryRow(ctx,
			`SELECT stewards.consult_subagent_dispatch($1, $2)`,
			sessionID, in.Question,
		).Scan(&chatWQ); err != nil {
			return toolError("consult_subagent_dispatch: %v", err),
				ConsultSubagentOutput{}, nil
		}
		_ = pool.QueryRow(ctx,
			`SELECT COALESCE((payload->>'_consult_reask_index')::int, 0)
			   FROM stewards.work_queue WHERE id = $1`, chatWQ,
		).Scan(&reaskIndex)

		// 2. Synchronously poll the chat row until terminal.
		start := time.Now()
		deadline := start.Add(time.Duration(consultMaxWaitSeconds) * time.Second)
		var status string
		for {
			if err := pool.QueryRow(ctx,
				`SELECT status FROM stewards.work_queue WHERE id = $1`, chatWQ,
			).Scan(&status); err != nil {
				return toolError("consult_subagent poll: %v (chat wq=%d)", err, chatWQ),
					ConsultSubagentOutput{}, nil
			}
			if status == "done" || status == "error" {
				break
			}
			if time.Now().After(deadline) {
				return toolError(
					"consult_subagent: re-engagement of %s timed out after %ds (status=%s). "+
						"It may still complete async — inspect work_queue id=%d.",
					sessionID, consultMaxWaitSeconds, status, chatWQ,
				), ConsultSubagentOutput{}, nil
			}
			select {
			case <-ctx.Done():
				return toolError("consult_subagent: context cancelled while waiting on %s", sessionID),
					ConsultSubagentOutput{}, nil
			case <-time.After(time.Duration(consultPollIntervalSec) * time.Second):
			}
		}

		// 3. Pull the sub-agent's answer — newest assistant message.
		var answer string
		err := pool.QueryRow(ctx, `
			SELECT COALESCE(content, '')
			  FROM stewards.messages
			 WHERE session_id = $1 AND role = 'assistant'
			 ORDER BY id DESC LIMIT 1`, sessionID,
		).Scan(&answer)
		if err != nil || answer == "" {
			answer = "(re-engagement reached status=" + status +
				" but no assistant answer was found in session " + sessionID + ")"
		}

		wallSec := int64(time.Since(start).Seconds())
		out := ConsultSubagentOutput{
			SessionID:       sessionID,
			ReaskIndex:      reaskIndex,
			Status:          status,
			Answer:          answer,
			WallTimeSeconds: wallSec,
		}

		header := fmt.Sprintf("[consult_subagent %s — re-ask #%d, %ds, status=%s]",
			sessionID, reaskIndex, wallSec, status)
		body := answer
		if status == "error" {
			body = "⚠️ The sub-agent did not complete cleanly. Partial/diagnostic output:\n\n" + body
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: header + "\n\n" + body}},
		}, out, nil
	}
}
