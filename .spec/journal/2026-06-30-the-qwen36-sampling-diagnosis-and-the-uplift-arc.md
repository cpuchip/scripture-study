---
date: 2026-06-30
lane: pg-ai-stewards
topic: the uplift-local-models arc — the spiral oracle, BINEVAL working via a tool, the rest's honest null, and the diagnosis that reframed everything (a known qwen3.6-MoE sampling bug, not a training problem)
tags: [uplifting-local-models, spiral-oracle, bineval, rest, qwen3.6, sampling, presence-penalty, MoE, verify-real-path, ab-test, web-research]
---

# The night the spiral turned out to be a sampling bug

A long, continuous arc on "uplift the local models so they don't spiral to dead, especially qwen."
Built the whole apparatus, ran the experiments, and the experiments did exactly what good experiments
do — they killed the wrong idea and pointed at the right one.

## What got built (all verified on the real path)
- **Spiral oracle** (`stewards.spiral_report()` / `session_spiraled()`, `.spec/proposals/spiral-oracle.sql`).
  The deterministic gauge. ★ `verify_real_path` caught **two** bugs in my own detector: the blunt v1
  metric (68%→28% after a repetition discriminator) and a NULL-handling bug (committed answers store
  `tool_calls=NULL`; `jsonb_typeof(NULL)='array'` is NULL → answer-turns uncounted → committed sessions
  mislabeled as spirals; 28%→**11.9%**). Honest baseline: **qwen3.6-35b-a3b 11.9% / gemma-12b 0%.**
- **BINEVAL** (`79-bineval.sql`). v1 (free-text JSON) failed — qwen skipped the questions. **Michael's
  fix won: make the binary questions the REQUIRED ARGS of a `submit_trajectory_verdict` tool** → forced
  decomposition. LIVE-VERIFIED: caught a real spiral (`committed:false`, `fail`, actionable notes), stores
  synchronously (fixes the harvest-on-error gap), `_session_id` reaches sql_fn tools. ✅
- **The rest** (`80-rest.sql`). Every N rounds, fold tools to housekeeping + a [REST] tidy nudge.
  Built + verified (fires + folds). **A/B (n=4): NULL/slightly-negative** — rest fired 8× on treatment but
  treatment spiraled 4/4 (vs control 3/4), committed 0/4 (vs 1/4). Honest finding: **tidying ≠ commitment
  pressure.** The rest cleans the workspace; the model rests, then resumes gathering. Wrong lever.

## The diagnosis that reframed it (web research → the real cause)
Michael: "this is a known qwen3.6 issue, do some web research." It is, and the research is decisive:
- **`QwenLM/Qwen3.6` issue #145** + qwen-code issues #4695/#5015/#5019: the *exact* failure — Qwen3.6 MoE in
  a multi-turn agent loop re-enters thinking after a tool result and re-ships the same `(tool,args)`, never
  committing. The model even says "let me stop and summarize" in its thoughts but keeps calling the tool.
- **Root cause = sampling, MoE-specific.** Qwen docs + the glukhov agentic reference: the **MoE (35B-A3B)
  has a documented high repetition-loop risk; the fix is `presence_penalty=1.5`** (general/reasoning), and
  **never greedy-decode** (temp 0.6, top_p 0.95, top_k 20, min_p 0).
- **★ THE SMOKING GUN:** our qwen3.6 dispatches run at **temp 0.2–0.4 (near-greedy), `presence_penalty=0`,
  no top_p/top_k** — the *exact* loop condition Qwen warns against. We've been running the MoE wrong.
- **The reframe:** it's **not a training problem.** Qwen ranks the fixes: sampling → prompt/template →
  *then* "the ultimate solution is fine-tuning." So the cheap fix likely gets most of it.

## The plan (written down)
- **Tier 1 — sampling** (`pp=1.5, temp=0.6, top_p=0.95, top_k=20, min_p=0` for the MoE). Wired as a
  per-dispatch `_sampling` override in `chat_post_internal` (80-rest.sql) + A/B-able. **Validated applied**
  (treatment bodies carry it). A/B **stalled inconclusive** (see below).
- **Tier 2 — circuit breaker:** identical-`(tool,args)`-call dedup, fire at 5 → `tool_choice=none`. The
  model-agnostic backstop qwen-code itself is shipping (PRs #5036/#5573). Catches the loop EARLY.
- **Tier 3 — thinking/template:** the `preserve_thinking` / chat-template angle (re-enters thinking after
  a tool result). Rig-level.
- **Tier 4 — fine-tune:** last resort; we're positioned (the trajectory ledger + BINEVAL verdicts = a ready
  labeled dataset).

## The rail earned its keep FIVE times tonight
synthetic-test artifact (yt) → blunt spiral metric → the NULL bug → the phantom A/B (the `/dev/stdin`
re-apply silently failed; treatment had no `presence_penalty` — caught by checking the bodies) → the
premature settle (poll watched `pending`/`in_progress` but the runs sit in `waiting_for_tools` — caught by
reading the trajectories). Every "result" this session wanted to be wrong; verifying made each one right.

## Carry-forwards
- **Re-run the sampling A/B clean** — it stalled (8 runs stuck `pending` at round 4; bgworker idle/wedged
  at ~1am). Investigate the stall (bgworker liveness / the `waiting_for_tools` queue), then re-run control
  vs treatment to completion and read the spiral-vs-commit delta. This is THE open question.
- **Register the chain files as a PR (Michael's Hinge):** `spiral-oracle.sql`, `79-bineval.sql`,
  `80-rest.sql` are dev-applied + saved but NOT registered (lib.rs + Dockerfile + virgin-smoke). Hold the
  PR until a lever proves out — Tier 1's clean A/B is the gate.
- **Tier 2 circuit-breaker** has a clear spec now (identical-call dedup at 5). Durable regardless of model.
- **Failover gap (#288)** still open. **BINEVAL trajectory-critic context limit** (huge runs > 98K ctx).
- **The rest** is a context-pressure tool, not an anti-spiral one — reframe its purpose.

## The shape of it
We set out to make qwen stop spiraling, built a gauge + a judge + a "rest," proved the rest was the wrong
lever, then found the actual cause is a sampling misconfiguration the Qwen team documents. The most
valuable thing wasn't a fix we invented — it was an honest experiment that said "no," and a web search
that said "here's the real knob." Tier 1 is one clean A/B away from the answer.
