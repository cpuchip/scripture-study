---
date: 2026-05-30
title: Batch R — multi-model document redline (panel) on the substrate
workstream: WS5
session_type: dev (decisions-up-front, gated phased commits)
status: R.1–R.6 SHIPPED + live-verified; MCP binary swap pending
---

# Batch R — the substrate can now redline a document with a model panel

## Where it started

The `projects/scripture-book` 3rd-draft attempt tried to use `start_brainstorm`
to give a 7-model panel the book manuscript + a redline mandate. **All failed to
read the book** — 2 of 7 produced generic advice, 5 returned empty (burned their
budget on failed `fs_search` loops). The scripture-book session wrote a clean
consumer spec (`.spec/proposals/substrate-multimodel-document-redline.md`) and
Michael pointed this stew session at it.

## Root cause (confirmed against the code, not the hypothesis)

The `fs-read` MCP server enforces ONE global allow-list (`mcp_servers.args`):
`.spec/journal/*`, `.spec/proposals/*`, `.mind/*`, `docs/**`, `projects/space-center/**`.
`projects/scripture-book/src/chapters/` isn't in it, and `sandbox.go` rejects
absolute paths — so the lenses' `/workspace/...` attempts failed on path-form AND
allow-list. **Correction to the spec:** `audit_files` is NOT a privileged fs path
— it uses the same allow-list, so it would have failed too. There was no
mechanism to hand a document to a model outside the allow-list. That was the gap.

## What shipped (5 decisions, 6 gated phases)

Decisions ratified up front (AskUserQuestion): **D-RL1** dedicated `redline`
pipeline (not extend brainstorm); **D-RL2** server-side document injection;
**D-RL3** condense = both (default orchestrator, optional substrate); **D-RL4**
verification gate + off-disk; **D-RL5** 32k per-call max_tokens.

- **R.1** `redline` pipeline + `panel-redline` agent — location-anchored edit
  mandate with the verification gate baked in (never alter a quote, flag
  touches-quote/doctrine, proposals only), `tools_disabled` so the panel can't
  even reach a quote to "fix" it. 7 defense-in-depth perm denies.
- **R.2** pg `/workspace:ro` mount + `read_workspace_doc(path|glob)` via
  `pg_read_file` — doc-extension-only gate (refuses `.env`/secrets/traversal).
  **D-RL2 was revised here:** the ratified "bgworker reads /workspace" was
  impossible — the bgworker runs in the pg container, which had NO /workspace
  mount (only the bridge did). The pg-side `pg_read_file` path is lighter (a
  compose mount + recreate, no Rust rebuild) and Michael confirmed it.
- **R.3** per-call `max_tokens` (32k) + input-scoped `tools_disabled` in
  `work_item_dispatch_stage`. **Latent bug surfaced (NOT fixed):** 10 pipelines
  (planning, research-*, yt-*, revise-proposal, thummim-define, agent-proposal)
  declare `stage.tools_disabled=true` on review/synthesize stages but this fn
  never propagated it — they've run WITH tools, a real cost leak the C.6 lesson
  was about. Fixing it flips tools off across the live soak, so R.3 reads
  `tools_disabled` from INPUT only (scopes to redline); the stage-level
  propagation is left for Michael to ratify (see carry-forward).
- **R.4** `start_panel_redline()` + `panel_redline` MCP tool — reads the doc,
  injects it into each child binding_question, builds a `decompose-fanout`
  manifest (reusing `spawn_children`) with one `redline` child per model:
  model_override + provider resolved from `model_pricing` (gemini → google_gemini)
  + `input_extra{tools_disabled, max_tokens}` + auto-scaled cost cap.
- **R.5** `panel_redline_condense()` + `redline-condense` agent/pipeline + MCP
  tool — optional substrate merge of the N reports into one ranked menu
  (rank by value × consensus, preserve every flag). Default stays orchestrator.
- **R.6** auto-verify: redline children finished `completed` but stayed
  `maturity=raw` (j6 auto-verifies brainstorm + aggregate, not redline), so the
  index aggregator dangled. Extended `on_one_shot_pipeline_completed` to qualify
  `redline%`.

## Live end-to-end (real, < $0.02)

3-model panel (deepseek-v4-flash, mimo-v2.5, qwen3.6-plus) on
`06_bilateral_covenant.md` (8948 chars). All 3 produced substantial
location-anchored reports (7–10KB) quoting real chapter text. **The verification
gate worked:** the D&C 82:10 quote was flagged `Touches quote/doctrine: yes` and
preserved verbatim. `panel_redline_condense` (deepseek) merged the 3 into one
ranked menu with per-edit consensus (k of N) and every flag preserved. Children
auto-verified, the index aggregator fired + materialized. **The original
"can't read the doc → empty" failure is gone** (inverse hypothesis satisfied).

This is the substrate-native replacement for the `agy-cli` stopgap skill ("until
the substrate redline pipeline lands").

## Carry-forward

- **MCP binary swap pending.** `panel_redline` / `panel_redline_condense` (and
  the earlier `list_models`/`list_connectors`) are in the binary code but the
  live `bin/stewards-mcp.exe` is MCP-locked. Needs a `/mcp` disconnect →
  `go build -o ../../bin/stewards-mcp.exe .` → reconnect. The SQL path
  (`start_panel_redline`) is live and was used for R.6, so the feature works
  today via psql; the swap just makes it callable from Claude Code.
- **The tools_disabled latent bug (10 pipelines)** — ratify whether to globally
  honor `stage.tools_disabled` (cheaper, honors declared intent, but changes live
  soak output for research-write/yt-gospel/planning review stages). Separate
  decision; R.3 scoped around it.
- **Manifest doc-replication** — the document is stored N+1× (manifest + each
  child input). Fine for chapters; for whole-book × many models it's ~1MB jsonb.
  A future optimization stores the doc once + references it.
- **Whole-book redline** not yet run live (only one chapter); the read path
  globs all 21 chapters (147KB) fine — the open question is per-model context
  limits + cost at full size, which `max_tokens` + cost_cap already bound.
- **R.5 metadata-driven auto-verify** — j6's deferred "drive from a flag instead
  of an explicit pipeline_family list" is now 3 patterns overdue.

## Pace note

Another large build, Michael-driven (ratified all 5 decisions + "do all R
points"). Soak paused for the build, resumed at close. 7 commits, zero rollbacks,
the C–F cadence held (smoke before every commit; transactional rollback smokes
kept spend at ~$0.02 total for the whole batch).
