// heavyweight_tools.go — K.5 / L.4 / L.5 / L.6 wrapper tools.
//
// Each L.6 wrapper is a thin layer over spawn_subagent: it takes
// use-case-specific parameters, constructs a binding_question + picks
// the dedicated pipeline_family declared in the L.6 SQL, then delegates
// to the same spawn-and-wait logic.
//
// L.4 (mark_engram_important) and L.5 (re_extract_engrams) are direct
// SQL fn wrappers — they do not spawn sub-agents.
//
// K.5a (external content): deep_research is the proof-of-pattern.
// L.6 fills out the rest of the wrapper set:
//   summarize_url, audit_files, investigate_session,
//   summarize_study, investigate_study, audit_studies.

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---------------------------------------------------------------------
// deep_research — K.5a wrapper
// ---------------------------------------------------------------------

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
			"DO NOT use for: a single URL fetch (use summarize_url instead), or work you can answer with one web_search call.",
	}, makeDeepResearch(pool))

	// L.6 wrappers.
	mcp.AddTool(srv, &mcp.Tool{
		Name: "summarize_url",
		Description: "Fetch a single URL and return a focused engram-shaped digest. " +
			"Delegates to a sub-agent restricted to fetch_url + expand_message ONLY. " +
			"Use for: pulling content from one specific page where you need the substance but don't want " +
			"the full document in your active context.",
	}, makeSummarizeURL(pool))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "audit_files",
		Description: "Read files matching a glob and answer a question about them. " +
			"Delegates to a sub-agent restricted to fs_read / fs_search / fs_list + expand_message ONLY. " +
			"Use for: 'do any of these files do X?', 'which file references Y?', surveys across a directory.",
	}, makeAuditFiles(pool))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "investigate_session",
		Description: "Inspect a session's history and answer a question about it. " +
			"Delegates to a sub-agent restricted to work_item_show / work_item_list + expand_message ONLY. " +
			"Use for: 'what did session X conclude about Y?', 'why did stage Z fail?'.",
	}, makeInvestigateSession(pool))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "summarize_study",
		Description: "Read a substrate study by slug and return a focused digest. " +
			"Delegates to a sub-agent restricted to study_get + expand_message ONLY. " +
			"Use for: pulling a known study's substance without the full text in active context.",
	}, makeSummarizeStudy(pool))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "investigate_study",
		Description: "Search the studies corpus and synthesize what it knows about a topic. " +
			"Delegates to a sub-agent restricted to study_search / study_get / study_similar + expand_message. " +
			"Use for: 'what has the corpus said about X?', cross-study syntheses, finding adjacent material.",
	}, makeInvestigateStudy(pool))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "audit_studies",
		Description: "Audit the studies corpus against a quality / completeness question. " +
			"Delegates to a sub-agent restricted to study_search / study_get + expand_message. " +
			"Use for: 'which studies on X lack a Becoming section?', 'are any studies contradicting Y?'.",
	}, makeAuditStudies(pool))

	// L.4 / L.5 direct SQL fn wrappers.
	mcp.AddTool(srv, &mcp.Tool{
		Name: "mark_engram_important",
		Description: "Flag a specific engram (by message_id + engram_id) as is_important. " +
			"Important engrams are anchored at HOT through context pressure — they survive all pressure " +
			"thresholds except crisis, and even then they emit first. " +
			"Use when an engram contains a quote, URL, date, or claim you'll cite later and can't afford to lose. " +
			"Pass important=false to clear the flag.",
	}, makeMarkEngramImportant(pool))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "re_extract_engrams",
		Description: "Re-extract engrams for a tool message with a different binding question. " +
			"Use when the existing engrams (tuned to the original binding) miss material relevant to your " +
			"current focus. The old engrams are archived in engrams._history; a fresh extraction runs with " +
			"the new binding. Cost-capped at $0.10 per re-extraction by default.",
	}, makeReExtractEngrams(pool))

	// L.1.1.12 — corpus access for the judge surface.
	mcp.AddTool(srv, &mcp.Tool{
		Name: "read_corpus_parents",
		Description: "Read parent chunks from an indexed corpus on a [CORPUS-INDEXED] tool message. " +
			"Use after the L.1.1.8 judge surface presents you with a corpus — paginate with parent_ord_start + count. " +
			"Mark anything precious with mark_engram_important once you find it.",
	}, makeReadCorpusParents(pool))
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
			costCap = 1_500_000
		}

		spawnIn := SpawnSubagentInput{
			PipelineFamily:  "research-write",
			BindingQuestion: binding,
			CostCapMicro:    costCap,
		}
		return makeSpawnSubagent(pool)(ctx, req, spawnIn)
	}
}

// ---------------------------------------------------------------------
// L.6 wrappers — each delegates to its dedicated subagent pipeline.
// ---------------------------------------------------------------------

type SummarizeURLInput struct {
	URL          string `json:"url" jsonschema:"the URL to summarize"`
	Focus        string `json:"focus,omitempty" jsonschema:"optional focus to narrow the summary"`
	CostCapMicro int64  `json:"cost_cap_micro,omitempty" jsonschema:"max micro-dollars (default 300000=$0.30)"`
}

func makeSummarizeURL(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in SummarizeURLInput,
) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in SummarizeURLInput,
	) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
		if in.URL == "" {
			return toolError("summarize_url: 'url' is required"), SpawnSubagentOutput{}, nil
		}
		binding := fmt.Sprintf("Summarize the content at this URL: %s\n\n"+
			"Use fetch_url to retrieve it. Produce a focused engram-shaped digest preserving the cite chain "+
			"(title, source URL, key dates / names / quotes verbatim).", in.URL)
		if in.Focus != "" {
			binding += "\n\nFocus: " + in.Focus
		}
		costCap := in.CostCapMicro
		if costCap == 0 {
			costCap = 300_000
		}
		return makeSpawnSubagent(pool)(ctx, req, SpawnSubagentInput{
			PipelineFamily:  "subagent-url-summary",
			BindingQuestion: binding,
			CostCapMicro:    costCap,
		})
	}
}

type AuditFilesInput struct {
	Glob         string `json:"glob" jsonschema:"file glob pattern (e.g. .spec/journal/*.md)"`
	Question     string `json:"question" jsonschema:"the question to answer about matching files"`
	CostCapMicro int64  `json:"cost_cap_micro,omitempty" jsonschema:"max micro-dollars (default 500000=$0.50)"`
}

func makeAuditFiles(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in AuditFilesInput,
) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in AuditFilesInput,
	) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
		if in.Glob == "" {
			return toolError("audit_files: 'glob' is required"), SpawnSubagentOutput{}, nil
		}
		if in.Question == "" {
			return toolError("audit_files: 'question' is required"), SpawnSubagentOutput{}, nil
		}
		binding := fmt.Sprintf("Audit files matching glob: %s\n\nQuestion: %s\n\n"+
			"Use fs_search / fs_list to enumerate, fs_read to inspect. Produce a per-file findings table.",
			in.Glob, in.Question)
		costCap := in.CostCapMicro
		if costCap == 0 {
			costCap = 500_000
		}
		return makeSpawnSubagent(pool)(ctx, req, SpawnSubagentInput{
			PipelineFamily:  "subagent-files-audit",
			BindingQuestion: binding,
			CostCapMicro:    costCap,
		})
	}
}

type InvestigateSessionInput struct {
	SessionID    string `json:"session_id" jsonschema:"the session id to investigate (e.g. wi--abc123--gather)"`
	Question     string `json:"question" jsonschema:"the question to answer"`
	CostCapMicro int64  `json:"cost_cap_micro,omitempty" jsonschema:"max micro-dollars (default 400000=$0.40)"`
}

func makeInvestigateSession(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in InvestigateSessionInput,
) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in InvestigateSessionInput,
	) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
		if in.SessionID == "" {
			return toolError("investigate_session: 'session_id' is required"), SpawnSubagentOutput{}, nil
		}
		if in.Question == "" {
			return toolError("investigate_session: 'question' is required"), SpawnSubagentOutput{}, nil
		}
		binding := fmt.Sprintf("Investigate session %s.\n\nQuestion: %s\n\n"+
			"Use work_item_list to find the parent work_item, work_item_show to inspect history. "+
			"Cite specific message ids and stage names supporting your answer.",
			in.SessionID, in.Question)
		costCap := in.CostCapMicro
		if costCap == 0 {
			costCap = 400_000
		}
		return makeSpawnSubagent(pool)(ctx, req, SpawnSubagentInput{
			PipelineFamily:  "subagent-session-investigate",
			BindingQuestion: binding,
			CostCapMicro:    costCap,
		})
	}
}

type SummarizeStudyInput struct {
	Slug         string `json:"slug" jsonschema:"the study slug"`
	Focus        string `json:"focus,omitempty" jsonschema:"optional focus"`
	CostCapMicro int64  `json:"cost_cap_micro,omitempty" jsonschema:"max micro-dollars (default 300000=$0.30)"`
}

func makeSummarizeStudy(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in SummarizeStudyInput,
) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in SummarizeStudyInput,
	) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
		if in.Slug == "" {
			return toolError("summarize_study: 'slug' is required"), SpawnSubagentOutput{}, nil
		}
		binding := fmt.Sprintf("Summarize the substrate study with slug: %s\n\n"+
			"Use study_get to read it. Preserve key quotes verbatim with attribution.", in.Slug)
		if in.Focus != "" {
			binding += "\n\nFocus: " + in.Focus
		}
		costCap := in.CostCapMicro
		if costCap == 0 {
			costCap = 300_000
		}
		return makeSpawnSubagent(pool)(ctx, req, SpawnSubagentInput{
			PipelineFamily:  "subagent-study-summary",
			BindingQuestion: binding,
			CostCapMicro:    costCap,
		})
	}
}

type InvestigateStudyInput struct {
	Query        string `json:"query" jsonschema:"search query"`
	Focus        string `json:"focus,omitempty" jsonschema:"optional focus"`
	CostCapMicro int64  `json:"cost_cap_micro,omitempty" jsonschema:"max micro-dollars (default 600000=$0.60)"`
}

func makeInvestigateStudy(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in InvestigateStudyInput,
) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in InvestigateStudyInput,
	) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
		if in.Query == "" {
			return toolError("investigate_study: 'query' is required"), SpawnSubagentOutput{}, nil
		}
		binding := fmt.Sprintf("Investigate the studies corpus for: %s\n\n"+
			"Use study_search to find relevant studies, study_get to read them, study_similar to surface "+
			"adjacent material. Synthesize what the corpus knows; cite study slugs.", in.Query)
		if in.Focus != "" {
			binding += "\n\nFocus: " + in.Focus
		}
		costCap := in.CostCapMicro
		if costCap == 0 {
			costCap = 600_000
		}
		return makeSpawnSubagent(pool)(ctx, req, SpawnSubagentInput{
			PipelineFamily:  "subagent-study-investigate",
			BindingQuestion: binding,
			CostCapMicro:    costCap,
		})
	}
}

type AuditStudiesInput struct {
	Query        string `json:"query" jsonschema:"search query to find studies to audit"`
	Question     string `json:"question" jsonschema:"the audit question"`
	CostCapMicro int64  `json:"cost_cap_micro,omitempty" jsonschema:"max micro-dollars (default 600000=$0.60)"`
}

func makeAuditStudies(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in AuditStudiesInput,
) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in AuditStudiesInput,
	) (*mcp.CallToolResult, SpawnSubagentOutput, error) {
		if in.Query == "" {
			return toolError("audit_studies: 'query' is required"), SpawnSubagentOutput{}, nil
		}
		if in.Question == "" {
			return toolError("audit_studies: 'question' is required"), SpawnSubagentOutput{}, nil
		}
		binding := fmt.Sprintf("Audit studies matching: %s\n\nAudit question: %s\n\n"+
			"Use study_search to find candidates, study_get to inspect. Produce a per-study finding table.",
			in.Query, in.Question)
		costCap := in.CostCapMicro
		if costCap == 0 {
			costCap = 600_000
		}
		return makeSpawnSubagent(pool)(ctx, req, SpawnSubagentInput{
			PipelineFamily:  "subagent-studies-audit",
			BindingQuestion: binding,
			CostCapMicro:    costCap,
		})
	}
}

// ---------------------------------------------------------------------
// L.4 — mark_engram_important (direct SQL fn wrapper)
// ---------------------------------------------------------------------

type MarkEngramImportantInput struct {
	MessageID int64  `json:"message_id" jsonschema:"the message id from the engram block header in active context"`
	EngramID  string `json:"engram_id" jsonschema:"the engram's id (e.g. 'msg-2381-e3') from the engram you want to mark"`
	Important *bool  `json:"important,omitempty" jsonschema:"true to mark important (default), false to clear the flag"`
}

type MarkEngramImportantOutput struct {
	MessageID    int64  `json:"message_id"`
	EngramID     string `json:"engram_id"`
	IsImportant  bool   `json:"is_important"`
	TotalEngrams int    `json:"total_engrams"`
}

func makeMarkEngramImportant(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in MarkEngramImportantInput,
) (*mcp.CallToolResult, MarkEngramImportantOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in MarkEngramImportantInput,
	) (*mcp.CallToolResult, MarkEngramImportantOutput, error) {
		if in.MessageID <= 0 {
			return toolError("mark_engram_important: 'message_id' is required and must be positive"),
				MarkEngramImportantOutput{}, nil
		}
		if in.EngramID == "" {
			return toolError("mark_engram_important: 'engram_id' is required"),
				MarkEngramImportantOutput{}, nil
		}
		important := true
		if in.Important != nil {
			important = *in.Important
		}

		var raw []byte
		err := pool.QueryRow(ctx,
			`SELECT stewards.mark_engram_important($1, $2, $3)`,
			in.MessageID, in.EngramID, important,
		).Scan(&raw)
		if err != nil {
			return toolError("mark_engram_important query: %v (message_id=%d engram_id=%q)",
				err, in.MessageID, in.EngramID), MarkEngramImportantOutput{}, nil
		}

		var out MarkEngramImportantOutput
		if err := json.Unmarshal(raw, &out); err != nil {
			return toolError("mark_engram_important decode: %v", err), MarkEngramImportantOutput{}, nil
		}
		return nil, out, nil
	}
}

// ---------------------------------------------------------------------
// L.5 — re_extract_engrams (direct SQL fn wrapper)
// ---------------------------------------------------------------------

type ReExtractEngramsInput struct {
	MessageID          int64  `json:"message_id" jsonschema:"the message id whose engrams should be re-extracted"`
	NewBindingQuestion string `json:"new_binding_question" jsonschema:"the new binding question to focus extraction on"`
	CostCapMicro       int64  `json:"cost_cap_micro,omitempty" jsonschema:"max micro-dollars (default 100000=$0.10)"`
}

type ReExtractEngramsOutput struct {
	WorkQueueID int64 `json:"work_queue_id"`
}

func makeReExtractEngrams(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in ReExtractEngramsInput,
) (*mcp.CallToolResult, ReExtractEngramsOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in ReExtractEngramsInput,
	) (*mcp.CallToolResult, ReExtractEngramsOutput, error) {
		if in.MessageID <= 0 {
			return toolError("re_extract_engrams: 'message_id' is required and must be positive"),
				ReExtractEngramsOutput{}, nil
		}
		if in.NewBindingQuestion == "" {
			return toolError("re_extract_engrams: 'new_binding_question' is required"),
				ReExtractEngramsOutput{}, nil
		}
		costCap := in.CostCapMicro
		if costCap == 0 {
			costCap = 100_000
		}

		var wqID int64
		err := pool.QueryRow(ctx,
			`SELECT stewards.re_extract_engrams($1, $2, $3)`,
			in.MessageID, in.NewBindingQuestion, costCap,
		).Scan(&wqID)
		if err != nil {
			return toolError("re_extract_engrams query: %v (message_id=%d)",
				err, in.MessageID), ReExtractEngramsOutput{}, nil
		}

		out := ReExtractEngramsOutput{WorkQueueID: wqID}
		body := fmt.Sprintf("re_extract_engrams queued. message_id=%d work_queue_id=%d. "+
			"Old engrams archived to engrams._history; fresh extraction will populate items[] "+
			"tuned to the new binding. Track via work_queue id=%d.",
			in.MessageID, wqID, wqID)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: body}},
		}, out, nil
	}
}

// ---------------------------------------------------------------------
// L.1.1.12 — read_corpus_parents (paginated overflow read)
// ---------------------------------------------------------------------

type ReadCorpusParentsInput struct {
	MessageID        int64 `json:"message_id" jsonschema:"the message id from the [CORPUS-INDEXED] surface header"`
	ParentOrdStart   int   `json:"parent_ord_start,omitempty" jsonschema:"first parent ordinal to return (default 0)"`
	Count            int   `json:"count,omitempty" jsonschema:"how many parents to return this call (default 4)"`
	MaxCharsPerPart  int   `json:"max_chars_per_part,omitempty" jsonschema:"char cap per parent (default 14000)"`
}

type ReadCorpusParentsHit struct {
	ParentOrdinal int    `json:"parent_ordinal"`
	ByteSize      int    `json:"byte_size"`
	Content       string `json:"content"`
	HasMore       bool   `json:"has_more"`
}

type ReadCorpusParentsOutput struct {
	Parents []ReadCorpusParentsHit `json:"parents"`
	Count   int                    `json:"count"`
	HasMore bool                   `json:"has_more"`
}

func makeReadCorpusParents(pool *pgxpool.Pool) func(
	ctx context.Context, req *mcp.CallToolRequest, in ReadCorpusParentsInput,
) (*mcp.CallToolResult, ReadCorpusParentsOutput, error) {
	return func(
		ctx context.Context, req *mcp.CallToolRequest, in ReadCorpusParentsInput,
	) (*mcp.CallToolResult, ReadCorpusParentsOutput, error) {
		if in.MessageID <= 0 {
			return toolError("read_corpus_parents: 'message_id' is required and must be positive"),
				ReadCorpusParentsOutput{}, nil
		}
		if in.Count <= 0 {
			in.Count = 4
		}
		if in.MaxCharsPerPart <= 0 {
			in.MaxCharsPerPart = 14000
		}

		rows, err := pool.Query(ctx,
			`SELECT parent_ordinal, byte_size, content, has_more
			   FROM stewards.read_corpus_parents($1, $2, $3, $4)`,
			in.MessageID, in.ParentOrdStart, in.Count, in.MaxCharsPerPart,
		)
		if err != nil {
			return toolError("read_corpus_parents query: %v (message_id=%d)",
				err, in.MessageID), ReadCorpusParentsOutput{}, nil
		}
		defer rows.Close()

		var hits []ReadCorpusParentsHit
		hasMore := false
		for rows.Next() {
			var h ReadCorpusParentsHit
			if err := rows.Scan(&h.ParentOrdinal, &h.ByteSize, &h.Content, &h.HasMore); err != nil {
				return toolError("read_corpus_parents scan: %v", err),
					ReadCorpusParentsOutput{}, nil
			}
			hits = append(hits, h)
			hasMore = h.HasMore
		}
		if err := rows.Err(); err != nil {
			return toolError("read_corpus_parents rows: %v", err),
				ReadCorpusParentsOutput{}, nil
		}

		out := ReadCorpusParentsOutput{
			Parents: hits,
			Count:   len(hits),
			HasMore: hasMore,
		}
		return nil, out, nil
	}
}
