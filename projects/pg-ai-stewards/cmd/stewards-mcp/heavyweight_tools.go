// heavyweight_tools.go — K.5 ratified wrapper tools.
//
// Each wrapper is a thin layer over spawn_subagent: it takes use-case-
// specific parameters, constructs a binding_question + picks a
// pipeline_family, then delegates to the same spawn-and-wait logic.
//
// K.5a (external content): deep_research is the proof-of-pattern
// shipped here. The other K.5a wrappers (summarize_url, audit_files,
// investigate_session) and all K.5b wrappers (summarize_study,
// investigate_study, audit_studies) follow this exact shape — they
// pick a pipeline_family + format a binding_question + call the same
// spawn-and-wait code. See journal 2026-05-14 for build patterns.
//
// Why ship just deep_research first: it reuses research-write (no new
// pipeline needed), so it's instantly useful for the J.3 retry shape
// (heavyweight research into a topic that returns a digest). The other
// 6 wrappers each need a new tightly-scoped pipeline definition; those
// can be added incrementally as use cases emerge.

package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---------------------------------------------------------------------
// deep_research — K.5a wrapper
// ---------------------------------------------------------------------
// Delegates broad multi-source research on a topic to a child running
// the research-write pipeline. Returns a sourced prose digest.

type DeepResearchInput struct {
	Topic        string `json:"topic" jsonschema:"the subject to research (5-20 words; the binding question will be built around this)"`
	Focus        string `json:"focus,omitempty" jsonschema:"optional narrowing focus (e.g. 'safety considerations only' or 'pre-1960 history only')"`
	CostCapMicro int64  `json:"cost_cap_micro,omitempty" jsonschema:"max micro-dollars (default 1500000=$1.50 — research-write is multi-stage and uses heavier models)"`
}

func registerHeavyweightTools(srv *mcp.Server, pool *pgxpool.Pool) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "deep_research",
		Description: "Delegate broad multi-source research to a sub-agent running the research-write pipeline. " +
			"Returns a sourced prose digest (with verbatim URLs / dates / quotes preserved per substrate covenant). " +
			"Use for: topics requiring 3+ web sources, comparison across vendors / studies, historical lineage. " +
			"DO NOT use for: a single URL fetch (use spawn_subagent with summarize-url-style binding instead), " +
			"or work you can answer with one web_search call.",
	}, makeDeepResearch(pool))
}

func makeDeepResearch(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in DeepResearchInput,
) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in DeepResearchInput,
	) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
		if in.Topic == "" {
			return toolError("deep_research: 'topic' is required"), SpawnSubagentOutput{}, nil
		}

		// Construct a research-write-shaped binding question.
		binding := fmt.Sprintf(
			"Research: %s\n\n"+
				"Produce a sourced summary with the key findings, supporting evidence, and any noteworthy disagreements. "+
				"Preserve URLs, dates, names, and direct quotes verbatim for the parent agent's cite chain.",
			in.Topic,
		)
		if in.Focus != "" {
			binding += "\n\nFocus narrowly on: " + in.Focus
		}

		costCap := in.CostCapMicro
		if costCap == 0 {
			costCap = 1_500_000 // $1.50 default for research-write (multi-stage, heavier)
		}

		// Delegate to the same spawn-and-wait logic as spawn_subagent.
		// Reusing makeSpawnSubagent's flow by constructing the input.
		spawnIn := SpawnSubagentInput{
			PipelineFamily:  "research-write",
			BindingQuestion: binding,
			CostCapMicro:    costCap,
		}
		return makeSpawnSubagent(pool)(ctx, req, spawnIn)
	}
}
