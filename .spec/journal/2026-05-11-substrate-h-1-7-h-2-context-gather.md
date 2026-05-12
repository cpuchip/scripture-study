---
date: 2026-05-11
mode: build
workstream: WS5
project: pg-ai-stewards
title: "Substrate H.1.7 + H.2 — substrate self-awareness via fs-read + context_gather stage"
status: shipped (design proven; bridge bug carry-forward blocks full e2e)
carry_forward:
  - "Bridge session-staleness bug: after ~14 intensive fs-read calls (H.1.7) or ~4 (H.2 retry), the bridge's persistent fs-read session is invalidated by the 60s call-timeout but the work_queue mcp_proxy row stays in_progress instead of flipping to 'error'. The bridge needs deadline-handler code that writes status='error' to the queue row on CallTool timeout. Same shape as H.1.5a soft-fail pattern, different code path."
  - "Tool-name normalization gap: kimi-k2.6 sometimes calls `study_search_text` when the registered tool is `study_search`. Likely model hallucination, not bridge bug. Worth surfacing the actual tool name list more clearly in the prompt context."
  - "H.3 build (planning pipeline family) ratified scope; five §IV.6 open questions in the scope proposal await answers before code. See substrate-batch-h-context-gather-and-planning.md §IV.6 (Q-H3.1..Q-H3.5)."
  - "Per-pipeline-scoped fs-read remains a future extension — currently fs-read uses one global allowlist. H.3 planning pipelines will need per-pipeline scope (e.g., space-center planning gets /projects/space-center/* added). Mechanism is on the substrate runway, not yet shipped."
links:
  - "../../projects/pg-ai-stewards/.spec/proposals/substrate-batch-h-context-gather-and-planning.md"
  - "../../projects/pg-ai-stewards/cmd/fs-read-mcp/"
  - "../../projects/pg-ai-stewards/extension/h1-7a-fs-read-pg-stewards-mcp-seed.sql"
  - "../../projects/pg-ai-stewards/extension/h1-7b-research-grants-and-template.sql"
  - "../../projects/pg-ai-stewards/extension/h2-context-gather-stage.sql"
---

# Substrate H.1.7 + H.2 — the substrate becomes self-aware (2026-05-11 evening)

## What shipped

Five commits walked the substrate from "research agent does external search" to "research agent first reads what we've already built, then does external search." That's the cap-rock for compounding knowledge — every future research run benefits from every prior research run, journal entry, proposal, and mind-file insight.

### H.1.7 — the research agent gets eyes on the substrate
- **`6695cd0`** — Scope proposal (substrate-batch-h-context-gather-and-planning.md). Captures four ratifications (D-H1.7-A validation case, D-H3-A dual output, D-H3-B studies generalization timing, D-H3-C per-pipeline fs-read scope) + §IV design for H.3 with five open questions for next-session ratification.
- **`94950d3`** — New Go MCP server at `projects/pg-ai-stewards/cmd/fs-read-mcp/` with three tools (`fs_list`, `fs_read`, `fs_search`) over a path-scoped view of the repo. Sandbox uses `path.Match` not `filepath.Match` — unit tests caught a real cross-platform bug where `filepath.Match` on Windows treats `\` as separator, letting `*` match across `/` (sandbox bypass). Also fixed pre-existing bridge.Dockerfile gap (stewards-ui go.mod stub missing since e99b67b).
- **`227fba4`** — Two SQL files: `h1-7a` registers fs-read + pg-ai-stewards as bridge-spawnable MCP servers; `h1-7b` grants 7 new tools to the research agent (fs_*, work_item_list/show, watchman_pass_show/passes_list) and rewrites research-write's gather input_template with a "CONSULT PRIOR WORK FIRST" section + bumps round budget 4→8.
- **`1f91d41`** — Validation smoke + carry-forward bridge bug. The substrate-reflective binding question (D-H1.7-A) dispatched and the research agent autonomously read prior journals, proposals, the 11-cycle guide, and last week's roundup — exactly the pattern we designed for.

### H.2 — context_gather as research-write's new first stage
- **`545b542`** — Prepended context_gather as research-write's first stage. Uses cheaper qwen3.6-plus for structured prior-work consultation; gather stays on kimi-k2.6 for external research. Pipeline is now `context_gather → gather → synthesize → review`. The briefing flows to gather via `{{stage_results.context_gather.output}}`. Gather's template dropped the H.1.7 prior-work section and lowered round budget 8→5. No new substrate machinery — stage shape is identical to existing stages; `pipeline_first_stage_name` auto-picks `stages->0`.

## What's validated end-to-end

**H.1.7 validation (D-H1.7-A binding question):** kimi-k2.6 dispatched with the substrate-reflective binding question. Across 4 rounds in gather, it issued 14+ tool calls:
- `fs_read .spec/proposals/pg-ai-stewards-11-cycle-review.md` (the 11-cycle guide)
- `fs_read .spec/proposals/memory-research-bundle.md`
- `fs_read .spec/proposals/gospel-engine-v3-proxy-pointer.md`
- `fs_read .spec/proposals/lightrag-investigation.md`
- `fs_read .spec/journal/2026-05-11-substrate-batch-h-research-write.md`
- `fs_read .spec/journal/2026-05-11-substrate-h-1-5-1-6-completion.md`
- `fs_read research/ai-tools-weekly-2026-05-11.md`
- `fs_search` across journals + proposals + .mind for `agent.memory|RAG|context.gathering|multi.agent|compounding.knowledge`
- `study_search` query for the topic

**H.2 validation (science-museum-exhibits binding question):** qwen3.6-plus on context_gather issued the correct shape — `study_search` + 2x `fs_search` — before the bridge session-staleness bug recurred. Wiring is proven; full e2e gated on bridge fix.

## The bridge session-staleness bug (load-bearing carry-forward)

After ~14 intensive fs-read calls (H.1.7) or ~3-4 (H.2 retry), the bridge's persistent fs-read MCP session goes stale. The bridge logs `context deadline exceeded (session invalidated)` but the work_queue `mcp_proxy` row stays in `in_progress` indefinitely instead of flipping to `error`. Manual queue cleanup unblocked both runs.

The fix is well-defined: the bridge's CallTool deadline-handler needs to UPDATE the work_queue row to status='error' when a call times out. Same shape as H.1.5a's soft-fail pattern, different code path. Should be a short patch in `bridge_run.go`. Not done this session — wanted to ship H.1.7/H.2/H.3 design first per the user's "as much of H as you have clear enough."

## Architecture wins this session

- **The agent now reads the substrate's own state.** This is the structural prerequisite for compounding knowledge. The research agent isn't blind to what we've built anymore.
- **The compose volume mount of repo-root at /workspace** opens the door for future fs-read scopes without docker-compose changes. New scope = new mcp_servers row + restart. No image rebuild needed.
- **Sandbox correctness was caught by tests, not the live run.** `path.Match` vs `filepath.Match` would have been a Windows-only sandbox bypass — unit tests caught it before deployment.
- **Two model tiers in one pipeline.** context_gather uses qwen3.6-plus ($0.50/$3) for structured retrieval; gather uses kimi-k2.6 ($0.95/$4) for synthesis-shaping research. Per-stage model selection is the substrate primitive that makes this efficient.

## What this enables

The substrate is now positioned for the user's "agents that work in the DB and improve/expound" vision. The pieces:
1. **Read access** — H.1.7 lands it; agents can consult prior journals, proposals, studies, work_items
2. **First-class context-gather stage** — H.2 lands it; every pipeline can begin with situational awareness
3. **Write-back rung on trust ladder** — Batch I (future); agent-proposed studies/notes route through gate machinery
4. **Planning pipeline family** — H.3 (next session); first pipeline whose output is plan + proposed work_items

H.3 design is captured in the scope proposal with five open §IV.6 questions awaiting ratification. The build is ~1 session of work once those are answered.

## Cost summary

Two cancelled validation runs (~$0.08 + ~$0.006), three production-grade SQL migrations applied live, one Go MCP server built and bridge-image deployed, five commits. Session total: <$0.10 in LLM costs, plus container rebuilds.

## Closing — the "agents that improve/expound" arc

Michael's framing at the start of this session ("Pinky and the Brain — but for the science center") captured the right north star. The substrate's job isn't just to produce one good piece of research; it's to *get smarter at producing research over time*. H.1.7 + H.2 are the substrate noticing that it has a memory and using it. That's compounding. The next rung is the substrate noticing that it has hands — writing back what it learns. Batch I.

Carry-forwards above. World domination postponed until Tuesday. Science center is what we're really building.
