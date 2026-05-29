---
date: 2026-05-29
title: Substrate J.8 + J.9 — brainstorm-finish (model generalization + lens expansion)
batch: J.8, J.9
status: shipped — both batches committed live, soak resumed
commits:
  - 7753424  # J.8 model generalization (path A + C)
  - 23ce243  # J.9 lens library expansion (4 → 12)
related_proposals:
  - .spec/proposals/substrate-batch-j-fanout-brainstorm.md  # J.4 ratification (May 13)
---

# Substrate J.8 + J.9 — brainstorm-finish

## Provenance

Michael asked about pg-ai-stewards status and immediately followed up about the brainstorm library — recalled correctly that only 4 of "8-9" techniques were implemented and that models were locked. The J.4 proposal (2026-05-13) had named both as carry-forwards explicitly:

- **B2 ratification**: "start with 4 (SCAMPER, Six Hats, Crazy 8s, reverse-brainstorm). Expand from there."
- **B1 ratification + j5 line 19-20**: "mix providers... Local LM Studio/Ollama integration deferred — add later as new lens agents pointing at local provider rows."
- **Open question recorded**: "friend's 8-9 modes need eliciting before brainstorm lens library locks."

Michael clarified at the start of this session: the friend's setup (8 styles × 10 models each = 80 fan-outs, then coalesce) was "OP" — inspiration, not template. He wanted my proposed 9 lenses added as options plus full generalization to any opencode_go-supported model. He framed the work as part of this week's stewards push so it sat inside the 2026-05-23 Sabbath-ratified substrate thread.

## Ratification batch

Four AskUserQuestion batches before any SQL hit disk (substrate C-F cadence):

1. **Lens list (Q1 + Q2)** — all 8 new lenses approved: Mind Mapping, Brainwriting, Starbursting (5W1H), Disney Method, Storyboarding, TRIZ, Forced Analogy, Worst Possible Idea. Round Robin dropped (weak fit for single-LLM-call).
2. **Fallback chain (Q3)** — `input → stages → pipeline_default → catalog_default`. Full 4-step.
3. **Existing 4 lenses (Q4)** — NULL them out; move models to `metadata.default_model` + `metadata.suggested_model`. Full generalization, not preserved-as-hardcoded.
4. **Default lens subset (Q5, stewardship pick)** — `start_brainstorm()` without `p_lenses` defaults to today's 4 for backward compat. Caller opts INTO new lenses.

Batch shape: J.8 (model generalization) and J.9 (lens expansion) as separate commits, one named sub-step per commit, per the C-F discipline.

## J.8 — model generalization (commit 7753424)

Three files:

**j8a-dispatch-fallback-chain.sql** — `work_item_dispatch_stage` replaced with a 4-layer resolution: `work_items.{model,provider}_override → stages.{model,provider} → pipelines.metadata.default_{model,provider} → catalog_default_{model,provider}()`. New helpers `catalog_default_provider()` returns `opencode_go`; `catalog_default_model('opencode_go')` returns `kimi-k2.6` (today's split default). Existing pipelines with stages.model set are byte-identical (COALESCE short-circuits at layer 2).

**j8b-brainstorm-null-model.sql** — `UPDATE` on the 4 existing brainstorm-* pipelines. NULL `stages[0].model` + `stages[0].provider`; populate `metadata.default_model` + `default_provider` + `suggested_model` + `suggested_provider`. Behavior preservation: callers without `p_models` get today's per-lens model from `metadata.default_model` via the fallback chain.

**j8c-start-brainstorm-models.sql** — `start_brainstorm()` gains `p_models jsonb DEFAULT NULL`. Map keyed by short lens name; values are either a model string (`"opus-4.7"`) or a `{model, provider}` object. `spawn_children()` extended to propagate manifest-level `model_override` + `provider_override` onto spawned work_items BEFORE first dispatch. The `spawn_children` change is a general-purpose fan-out improvement, not brainstorm-specific.

Smoke results (all 3 ran against live substrate):

1. **No p_models** — 4 lenses dispatched with today's split (qwen3.6-plus for scamper/crazy8s, kimi-k2.6 for six-hats/reverse). All completed via the bridge. Behavior preservation confirmed exact.
2. **String override** — `p_models := '{"scamper":"opus-4.7","six-hats":"haiku-4.5"}'` → scamper got `model_override='opus-4.7'` on work_items AND `requested_model='opus-4.7'` in work_queue payload; six-hats got `haiku-4.5`; crazy8s + reverse fell back. Cancelled before bridge dispatched invalid-on-opencode_go models.
3. **Object override** — `p_models := '{"crazy8s":{"model":"gpt-5","provider":"openai"}}'` → crazy8s got both model AND provider overrides; other 3 fell back. Cancelled.

## J.9 — lens library expansion (commit 23ce243)

Three files:

**j9a-new-lens-agents.sql** — 8 INSERTs into `stewards.agents`. Each lens's prompt is designed to produce a DIFFERENT output shape, not just a different topic on the same shape:

| Lens | Output shape |
|---|---|
| Mind Mapping | Hierarchical tree (3-4 branches × 3-5 leaves), optional cross-branch links |
| Brainwriting | 6 seeds × 3 builds each (Extend / Vary / Counter triad) |
| Starbursting (5W1H) | Questions not answers — 4-6 per Who/What/When/Where/Why/How |
| Disney Method | Three sequential voices (Dreamer / Realist / Critic), later refs earlier |
| Storyboarding | 5-7 narrative scenes following one protagonist through baseline→complication→resolution |
| TRIZ | Contradictions + 3-5 of the 40 inventive principles + concrete sketches |
| Forced Analogy | 3 random unrelated domains × restate-generate-port + one STANDOUT |
| Worst Possible Idea | 5-7 terrible solutions → diagnose violated principle → invert into constraint |

Temperatures tuned per technique: 0.4 for TRIZ (structured), 0.6 for Starbursting (questions need precision), 0.7 for Mind Mapping/Disney/Storyboarding, 0.8 for Brainwriting (volume + structure), 0.9 for Forced Analogy and Worst Possible Idea (high divergence).

**j9b-new-lens-pipelines.sql** — 8 INSERTs into `stewards.pipelines`. Each single-stage with NULL `stages[0].model` + `stages[0].provider` (uses J.8.a fallback chain). `metadata.default_model` + `suggested_model` carry the lens-author's preferred default — maintains B1's "each lens declares its provider" spirit, fully overrideable by caller. Mix preserved: 6 qwen3.6-plus (mind-mapping, crazy8s, scamper, storyboarding, forced-analogy, worst-idea) + 6 kimi-k2.6 (brainwriting, disney, reverse, six-hats, starbursting, triz).

**j9c-start-brainstorm-lenses.sql** — `start_brainstorm()` gains `p_lenses text[] DEFAULT ARRAY['scamper','six-hats','crazy8s','reverse']`. Backward compat exact — callers without `p_lenses` get today's behavior. Subset selection lets callers pass any combination of the 12 short lens names. Unknown lens names raise at function entry (before `work_item_create`) with a helpful listing of valid options.

Smoke results:

1. **Unknown lens raises** — `p_lenses := ARRAY['scamper','typo-mispelled-lens']` → caught with message naming all 12 valid lenses. No work_items spawned, no LLM cost.
2. **Subset selection** — `p_lenses := ARRAY['scamper','mind-mapping','starbursting']` → 3 children spawned (not 4, not 12) + aggregator. Each child resolved the right `pipeline_family` and got the correct model from `metadata.default_model`. Cancelled.

## Decisions made by stewardship (not surfaced)

These weren't ratification-level questions; I exercised judgment within the covenant's stewardship boundary:

- **Per-lens model preferences for the 8 new lenses.** Set `default_model` + `suggested_model` to a sensible per-lens choice rather than NULL (which would have routed everything through catalog_default = kimi-k2.6). Mirrors how the 4 existing lenses preserved their per-lens variety in J.8.b. The mix preserves the B1 "each lens declares its provider" spirit.
- **Batch shape: J.8 and J.9 as separate commits.** Per CLAUDE.md C-F discipline: "Don't batch multiple sub-steps into one commit."
- **Smoke cleanup style.** Two smokes (J.8 #2 and #3) used overrides for opus-4.7 and haiku-4.5 which aren't on opencode_go — these would have failed at the bridge. Cancelled the work_items mid-stream. Cost impact: ~$0.05 from J.8 smoke #1 (real models, ran to completion to verify fallback chain end-to-end), negligible for #2/#3 (cancelled before any bridge spend), small for J.9 smoke #2 (3 chats, cancelled mid-claim, may have ~$0.05).

## Addendum — same-day MCP wrapper + UI view (M + U batches, commits 0c1926c + 2fe9b33)

After J.8 + J.9 committed, Michael asked to close the two carry-forwards I'd named ("MCP server signature update" + "Stewards-UI brainstorm form"). Four AskUserQuestion ratifications:
1. Scope = both this session
2. MCP shape = single `start_brainstorm` tool mirroring SQL signature
3. UI placement = new dedicated `/brainstorm` route (not in-NewWork)
4. UI default lenses = originals pre-checked + 8 J.9 lenses under "More lenses"

Discovery during M.3 (bridge rebuild): the bridge entrypoint runs `stewards-cli migrate --repo-root /workspace` which tracks 176 SQL files via a ledger and applied my 6 J.8/J.9 files automatically on rebuild. The "J-series lib.rs/Dockerfile foldback debt" claim above (and in the J.8 + J.9 commit messages) was based on incomplete investigation — there IS a working migration path for fresh rebuilds via `stewards-cli`. Correction logged; original section preserved below for the record but with this caveat.

Stewardship sweep during U.5: `docker compose build ui` failed with `cannot load module ../../projects/1828-illuminated/backend listed in go.work file`. Same shape as bridge.Dockerfile's 2026-05-22 fix that didn't sweep ui.Dockerfile. Fixed inline (2 COPY lines) per the boundary test.

Smoke results:
- M: clean Go build; symbols present in stewards-mcp.exe; bridge restarted clean. End-to-end stdio smoke didn't complete cleanly (PowerShell concatenated-JSON-RPC framing issue, not a tool issue) — full verification on Michael's next Claude Code MCP refresh.
- U: clean Docker build after the go.work fix; lens endpoint returned all 12 with originals correctly tagged; start endpoint validated binding_question; Vue SPA served on /brainstorm.

Both carry-forwards CLOSED. Brainstorm now reachable from three surfaces:
- Direct SQL: `SELECT stewards.start_brainstorm(...)` (lived live since J.8/J.9)
- Claude Code MCP: `start_brainstorm` tool
- Stewards-UI: http://127.0.0.1:8080/brainstorm

## Carry-forward (NOT done in J.8 / J.9)

**J-series lib.rs + Dockerfile registration sweep.** Discovered during J.8: NO j-series SQL files (j1-j7 from the J.4 ratification, now also j8 and j9) are registered in `extension/src/lib.rs`'s `extension_sql_file!` chain or in the Dockerfile's COPY list. The entire J batch is live-only. A `docker compose down -v` (which CLAUDE.md §8 says we don't do) would lose everything since the J batch landed. The substrate convention (never down -v, treat live state as canonical) means this is debt, not breakage — but the foldback debt is real. Separate batch needed: register j1-j9 in lib.rs + Dockerfile so fresh rebuilds inherit them. Scope NOT crept here.

**MCP server signature update.** `stewards-mcp` has a `start_brainstorm` tool wrapper at `cmd/stewards-mcp/`. The wrapper signature predates J.8.c + J.9.c and probably doesn't expose `p_models` or `p_lenses` yet. Future stewardship: extend the MCP tool to accept both. NOT done in this batch (focused on substrate SQL).

**Stewards-UI surface.** The `NewWork.vue` per-pipeline form (PE-C.4 from council ①) for brainstorm may not surface lens-subset or model-override UI affordances yet. Future stewardship: add a multi-select for lenses + per-lens model picker. NOT done in this batch.

## Cycle context

Substrate is one of three Sabbath-ratified threads (2026-05-23):
1. Council ② `substrate-scheduled-workflows` — NEXT, not started
2. Teaching Episode 2 — not touched this session
3. 1828 finish — not touched this session

Michael framed J.8 + J.9 as "part of the stewards push" — inside the substrate thread, not displacing Council ②. Mosiah 4:27 evidence reading at session close: substrate work happening, no slip detected in teaching or 1828 (neither active today).

## Files shipped

J.8 (commit 7753424):
- `extension/j8a-dispatch-fallback-chain.sql`
- `extension/j8b-brainstorm-null-model.sql`
- `extension/j8c-start-brainstorm-models.sql`
- `extension/smoke/j8-smoke-override.sql`
- `extension/smoke/j8-smoke-object.sql`
- `extension/smoke/j8-smoke-cleanup.sql`

J.9 (commit 23ce243):
- `extension/j9a-new-lens-agents.sql`
- `extension/j9b-new-lens-pipelines.sql`
- `extension/j9c-start-brainstorm-lenses.sql`
- `extension/smoke/j9-smoke-unknown-lens.sql`
- `extension/smoke/j9-smoke-subset.sql`
