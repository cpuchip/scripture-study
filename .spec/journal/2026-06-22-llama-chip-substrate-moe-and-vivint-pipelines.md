# 2026-06-22 — llama-chip UI, substrate-on-llama-chip, MoE default, and the vivint pipeline design

A very long session (spanned 06-21 evening → 06-22), several arcs, all in the
local-inference + substrate space. Michael drove throughout; full stewardship.

## What was done

**llama-chip — 6 UI/feature asks (all shipped + pushed `cpuchip/llama-chip`):**
1. Installed the **tufte** Claude skill (`.claude/skills/tufte/` from aref-vc/tufte-claude-skill).
2. **Tufte charts** in the UI — solid single-hue GPU bars (red accent >85%, no gradient), tok/s
   sparkline = thin reference line + accent dot on the latest value.
3. **Streaming chat** (SSE `stream:true`, live content+reasoning+tok/s).
4. **`diverse` fix** — it loaded distinct models fine but both were qwen-4b; made it genuinely
   diverse (qwen-4b + gemma-12b) + short model names in the table.
5. **Parallel** input in the load form.
6. **Progress bars + errors** (loading bar/elapsed, inline crash errors).
7. **guess-max-context** — wrote `internal/gguf` (reads real attention dims) so the estimate is
   sane (file-size heuristic gave 4.5M for a 4B); conservative, capped at trained ctx.

**Substrate wired onto llama-chip** — kept the `flexllama` provider name (zero churn);
`oss/.env STEWARDS_PROVIDER_FLEXLLAMA_BASE_URL=http://host.docker.internal:8090/v1`; recreate `pg`
(data-safe) loads it. Proven end-to-end (a probe spiked the 4090 + updated model_capability).
★ The substrate dispatches DIRECTLY (no `/api/ensure`) → the dance must be **pre-loaded**, and it
holds **both 4090s (~43GB)**.

**UI rig control** on the substrate dashboard (Michael declined auto-start; wanted a button) —
`cmd/stewards-ui/api/rig.go`: `GET /api/rig/state`, `POST /api/rig/{autonomy,brain-on,brain-off}`;
llama-chip `/api/unload-all`. brain-on = load dance-moe + resume; brain-off = pause + unload (free
GPUs for games). Dashboard.vue panel (buttons + autonomy/up badges + GPU bars). Verified e2e.

**Model bake-offs → `projects/llama-chip/docs/model-insights.md`:**
- gemma-26b-a4b (MoE) vs 12b dense: MoE ~2.7× faster (161 vs 58 tok/s), equal-or-better quality.
- qwen-35b-a3b (MoE) vs 27b dense: MoE ~3.9× faster (183 vs 47), comparable-or-crisper.
- Role-fit CONFIRMED: **gemma = ingest/gather** (tightest briefs), **qwen = reason/critic** (depth).
- All models confabulate historical specifics → quote-gate stays mandatory.

**ctx_size semantics — fixed a real bug.** llama-chip's `--ctx-size` is the **TOTAL** (per-slot =
ctx_size/parallel), standard llama.cpp — the code comment said "per-slot" (wrong; verified by VRAM).
FlexLLama's config was per-slot + multiplied, so porting the dance 1:1 silently **halved** gemma to
106k/slot. Fixed: dance gemma `ctx_size 425984` = the intended 213k/slot.

**Parallel/concurrency benchmark:** parallel-2 = ~1.6× aggregate per card (continuous batching);
both cards 4-concurrent = 488 tok/s = **0.94× the summed single-card** → true dual-card parallelism.
qwen-35b-a3b holds its full **256k single-slot** on one 4090 (GQA → cheap KV).

**★ MoE swap DEPLOYED (data-only substrate change):** new llama-chip **`dance-moe`** preset
(qwen-35b-a3b @ 120k/slot par2 + gemma-26b-a4b @ 213k/slot par2). Substrate reason/critic →
`qwen3.6-35b-a3b` (p0), dense 27b (p1 fallback), kimi (p2) — pure `model_capability` + `model_aliases`
rows, ZERO bgworker code (overlays `flexllama-models.sql` + `role-aliases.sql`). UI brain-on →
dance-moe. Proven end-to-end (probe → qwen-35b-a3b, healthy, 0 restarts, GPU0 478 MiB cushion held).
Autonomy resumed, guard clean.

**Stage taxonomy + vivint pipeline design → workspace `docs/stage-taxonomy-and-vivint-pipelines.md`:**
- **Finding: the engine is already data-driven.** A stage = `{name, next, model(role), input_template,
  tool_groups, auto_advance}`; the bgworker does NOT branch on stage name. New pipeline = a row.
- Answers Michael's "do we need more generic?" → no, the architecture is generic; the gap is a
  documented **palette** + the occasional new **tool**. Discipline: new things land as
  tool/tool_group/prompt/pipeline-row, never a `stage=="x"` branch.
- **Vivint reframed (his correction — "fault" was a typo) to a knowledge assembly line:**
  **gather → organize(graph) → analyze/act.** gather+organize are domain-agnostic ("the info part of
  the brain", shared with book/video/article digesters); analyze is domain-specific. Grounded vs the
  existing graph (353 nodes, 186 typed edges, 19 edge_kinds, 2473 engram embeds): gather built,
  organize built-but-ambient (engram_extract + memory-tend, not a deliberate stage), analyze thin+new
  (graph_recall read + doc + prompt).
- **Design intent captured (§6.5):** GATHER = all-sources, incremental (aware of existing, ADD only),
  category-gap-aware (missing categories / more witnesses / new categories), directed by analyze
  feedback, and **TIME-AWARE** (subject evolves; issues resolve/go stale → stamp recency, re-validate,
  age-out — a general info-brain property). ANALYZE = do differently / products fill gaps / too far /
  not far enough. Two build mechanics flagged: analyze→gather feedback loop + node freshness in the graph.

## Surprises / lessons

- **The MoE pattern holds for both families** — few-active-params MoE beats the dense sibling ~3-4× on
  speed at equal quality. The single-card MoE is ~160-183 tok/s, not the ~45 I'd cited (that was under
  parallel-2). Updated `reference_local_moe_for_substrate`.
- **ctx_size port bug** — FlexLLama (per-slot, multiplied) vs llama-chip (total, passed through) differ;
  the 1:1 port halved gemma. Watch config semantics across tools.
- **Stale-guard re-trips** — after each pause→resume, the watchman auto-paused on 5 "consecutive
  failures" that were STALE scheduled-pipeline failures from when the rig was down. Fix each time:
  cancel the failed autonomous runs newer than the last success (they re-fire on schedule) → consec=0.
- **The engine is more generic than it felt** — almost no stage-name code. The recent stage-building
  work (doc-construction, tool_groups, role-aliases) is what made it data-driven.

## Carry-forward (for next session)

- **★ llama-chip is a MANUAL process** (this session's bg launch, PID changes each restart) — it dies
  when the Claude Code session/terminal closes or on reboot. The substrate depends on it + the dance-moe
  loaded. To persist: Michael starts it himself (`cd projects/llama-chip && ./llama-chip.exe serve
  --config config.json`) then UI "Start brain" (→ dance-moe), OR the future `--install-startup` task.
  Before gaming/closing: UI **Free GPUs** (pauses autonomy + unloads) so the brain isn't dispatching
  into a dead rig.
- **GPU0 cushion ~478 MiB at 120k/slot** — held under a single probe; if 2 concurrent reason dispatches
  OOM, drop qwen to ~110k/slot.
- **Next build: the vivint pipelines** — build order in §6: (1) organize-as-a-stage keystone (test on a
  book/article we already gather), (2) vivint analyze pipeline (file_private), (3) backfill digesters
  onto shared gather+organize. The prompts (what's a fault / actionable / which edge kinds) are
  Michael's domain input.
- Deferred still: `--install-startup`, refine guess-context KV for sliding-window/compressed-KV.

## State at close

Substrate autonomy RUNNING on dance-moe (qwen-35b-a3b reason/critic + gemma-26b-a4b ingest), guard
clean, GPU0 478 / GPU1 936 MiB free. llama-chip serving on :8090 (manual process). All work
committed+pushed (llama-chip, oss, workspace overlays + docs).
