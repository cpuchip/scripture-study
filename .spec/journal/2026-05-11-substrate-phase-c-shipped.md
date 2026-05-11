---
date: 2026-05-11
session_kind: build
workstream: WS5
substrate_phase: C
commits: [f3ab677, c47d6a5, e9e0cb9, 12ae5ed, e0efdb7, c7f5253, 0f54bbe, 3082332]
cost_usd: 0.006
---

# Substrate Phase C — shipped end-to-end on dev in one session

## What happened

Michael said "Lets blaze through All of phase C with git commits at natural steps, testing along the way." We did exactly that — eight commits in build order, each with a smoke test, all in one session. Phase C as ratified and sub-specced this morning is now live.

## What shipped

**C.1 (f3ab677)** — `stewards.intents` (slug-unique) + `stewards.covenants` (scope-keyed with `covenants_active_scope` partial unique index for at-most-one-active-per-scope) + `work_items.intent_id` FK (nullable). YAML provenance fields (`source_file`, `source_yaml_sha`) baked in.

**C.2 (c47d6a5)** — Three pg_extern helpers in new `src/yaml.rs`:
- `yaml_sha256(text)` — sha-comparison for idempotent re-seeds
- `parse_yaml_intent(text)` — normalizes `intent.yaml` shape (purpose + values map + constraints map) into a single jsonb with `values_hierarchy` as ordered array; constraints land in same array tagged `kind='constraint'`
- `parse_yaml_covenant(text)` — same for `.spec/covenant.yaml` (human + agent commits → ordered arrays of `{key, description, why}`; preserves teaching_extension nested)

Plus `seed_intents_from_yaml` + `seed_covenant_from_yaml` SQL composers — both idempotent (no-op if sha unchanged), covenant version atomically deactivates prior active row in scope before inserting.

**C.3 (e9e0cb9)** — `scripts/git-hooks/pre-commit` re-seeds substrate on staged change to `intent.yaml` or `.spec/covenant.yaml`. Docker cp + psql via `docker exec`. Skips gracefully when container is down — never blocks a commit.

**C.4 (12ae5ed)** — `compose_system_prompt` extended to prepend two new blocks:
- `=== Active Covenant ===` (always, when scope='global' has an active row): human commits + agent commits as bullets, plus optional `council_moment` paragraph
- `=== Intent ===` (only when session_id resolves to a work_item with intent_id): slug, purpose, beneficiary, values_hierarchy with constraint badges, non_goals, scripture_anchor

Followed by `=== Agent ===` then existing agent + instructions + skills blocks. Sessions without a work_item get covenant only. ~600 tokens added per dispatch with both blocks.

**C.5 (e0efdb7)** — Backfilled 17 existing work_items with default scripture-study intent, then `ALTER COLUMN intent_id SET NOT NULL`. Per D-C3, every work_item now requires explicit intent.

Stewardship moment: discovered during smoke that `work_item_create()` would now fail because it didn't set `intent_id` — watchman + every existing caller would break. Same-fix-same-shape: extended signature with optional `p_intent_id uuid` (default NULL → resolves to scripture-study). Legacy callers stay working; new callers pass explicit. Required `DROP FUNCTION + CREATE OR REPLACE` since the parameter changed signature (Phase 4b lesson).

**C.6 (c7f5253)** — Three pieces:
1. `bgworker.rs`: when `payload.tools_disabled=true`, strip `tools` from body before POST
2. New `gate_prompts.covenant_check` template (free-form per D-C4); revised `evaluate` template references intent + covenant + reminds model not to call tools
3. `evaluate_gate` + `generate_scenarios` + `verify_work_item` all set `tools_disabled=true` on payload

End-to-end verified: previous gate-eval (Phase B) cost 11 chats over ~70s ≈ $0.04. New tools-off gate-eval: **1 chat in 16s, $0.0056** — about 7× cost reduction (better than the 5× estimate from the Phase B 2026-05-11 lesson). Auto-fire still triggered (`action='surface'`, gate_decisions row #7).

The 'surface' decision was correct: I'd reset the test work_item's outline stage_results to a stale value, and the model — without tools to research — refused to advance an empty/error output. Sharper refusal than the prior tools-on round which went researching.

**C.7 (0f54bbe)** — Backend Go endpoints:
- `GET /api/intents/list` (with work_item_count per intent)
- `GET /api/intents/get?id=|slug=`
- `POST /api/intents/create` (inline-create from NewWork)
- `GET /api/covenants/active?scope=` (default global)
- `GET /api/covenants/list`

Smoke verified: `scripture-study` intent shows `work_item_count=18` post-C.5 backfill; active covenant returns `ratified_by=both` with full text from `.spec/covenant.yaml`.

**C.8 (3082332)** — Vue surfaces:
- `NewWork.vue` gains required intent picker (dropdown) + "+ new" button opening inline create-intent modal. Submit disabled until intent picked. `new_work.go` relays intent_id as 6th positional arg to `work_item_create`.
- `Intents.vue` — list view with expandable detail panels (purpose, values_hierarchy with constraint badges + severity, non_goals, scripture_anchor, source_file, work_item_count)
- `Covenants.vue` — active global covenant rendered as side-by-side human/agent commitment cards (each with expandable "why"), plus when_broken/recovery/council_moment sections, expandable teaching_extension
- `router.ts` + `App.vue` nav extended with a separator + Intents + Covenant entries

End-to-end smoke: `/intents` and `/covenants` serve 200; POST `/api/work-items/create` with `intent_id` returns work_item with intent_slug='scripture-study' resolved through the FK.

## Surprises

**Lesson #3 bit twice in one session.** After both C.2 and C.6 image rebuilds, the new `pg_extern` functions weren't auto-installed — extension at version 0.2.0 still, so CREATE EXTENSION was a no-op. Worked around manually creating the functions with `CREATE OR REPLACE FUNCTION ... AS '$libdir/pg_ai_stewards', '<name>_wrapper'`. Lesson #3 still hurts. The right long-term fix is an extension version bump strategy with proper upgrade scripts, but that's its own substrate-y project — the live-migration pattern is good enough for dev iteration.

**Tools-off cost reduction came in BETTER than the 5× estimate.** Predicted 5× from the Phase B lesson; measured 7× (11 chats → 1 chat). That's because the Phase B test happened to dispatch through `plan` agent which has `gospel_search` and corpus tools, which the model heavily exercised. Without tools the model gives a sharper, more decisive answer in one round.

**The 'surface' decision was actually better than the prior 'revise' decision.** With tools, the model researched the corpus and produced detailed critique ("Hebrew/Greek word study is a category error" etc.) — sharp content. Without tools, the model just refused to evaluate stale stage_results and surfaced — sharp judgment. Both are useful in different situations. Tools-off gates are the right call for binary decisions where research is overhead; tools-on might be worth keeping for high-signal evaluation work where the model needs corpus context.

**Stewardship caught its own bug.** C.5's NOT NULL would have broken every legacy caller of `work_item_create()`. The sub-spec didn't call this out — it was a discovery moment during smoke. Caught it, fixed it inline, no detour to the user. Per `exercise_stewardship` covenant: "would Michael, if asked in advance, say 'yes, obviously do that'? If yes, do it and tell them." Yes — and reported in the commit message.

## Process / covenant

Eight commits, each with a smoke test. The cadence held end-to-end. No commits without a verifying smoke. No piling work into a single commit.

Cost for the session: $0.006 (one tools-off gate-eval test). All other work was schema + code, no LLM dispatches.

Soak paused at session start, re-enabled at session end. Bridge stopped before the pg restarts (3 of them this session: C.2 image rebuild, C.6 image rebuild, plus down/up cycles). Restarted at session end.

## Open / carry-forward

- **Lesson #3 needs a real fix** (extension version bump strategy) — keep flagging until we do it.
- **The covenant_check template hasn't been wired to anything yet.** It's seeded into gate_prompts but no SQL function dispatches it. That's a Phase D task (Atonement + Sabbath also use the tools-off + JSON-output pattern; covenant_check joins them as one of three new dispatch families).
- **Token cost of compose_system_prompt injection** — predicted 600 tokens/dispatch; not measured yet on a real workload. Cost panel will surface it. If it becomes painful, add `compose_system_prompt(skip_covenant=true)` for stage chats that don't need re-stating.
- **YAML edits don't trigger full re-seed of work_items.** If you edit values_hierarchy, existing work_items still reference the same intent_id; the intent's row gets updated in place (UPDATE on slug conflict). Work_items pick up the new values on next dispatch via compose_system_prompt's fresh query.
- **Phase D is next** when Michael wants to keep going. Sub-spec already drafted (`phase-d-design.md`) — Atonement + Sabbath + Consecration with lessons mirroring gate_decisions ledger shape.

## Files touched

Repo:
- `projects/pg-ai-stewards/extension/5d-intents-covenants.sql` (new)
- `projects/pg-ai-stewards/extension/5d2-seed-fns.sql` (new)
- `projects/pg-ai-stewards/extension/5d3-compose-with-intent.sql` (new)
- `projects/pg-ai-stewards/extension/5d4-backfill-intent.sql` (new)
- `projects/pg-ai-stewards/extension/5d5-tools-off-and-templates.sql` (new)
- `projects/pg-ai-stewards/extension/Cargo.toml` (serde_yaml + sha2 + hex deps)
- `projects/pg-ai-stewards/extension/src/yaml.rs` (new — 3 pg_extern helpers)
- `projects/pg-ai-stewards/extension/src/lib.rs` (5 new extension_sql_file! + mod yaml)
- `projects/pg-ai-stewards/extension/src/bgworker.rs` (tools_disabled strip)
- `projects/pg-ai-stewards/extension/Dockerfile` (5 new SQL files in COPY)
- `scripts/git-hooks/pre-commit` (new)
- `scripts/git-hooks/README.md` (new)
- `.git/hooks/pre-commit` (copy installed)
- `scripts/stewards-ui/api/intents.go` (new)
- `scripts/stewards-ui/api/covenants.go` (new)
- `scripts/stewards-ui/api/api.go` (Register)
- `scripts/stewards-ui/api/new_work.go` (intent_id passthrough)
- `scripts/stewards-ui/frontend/src/api.ts` (intent + covenant types + wrappers)
- `scripts/stewards-ui/frontend/src/views/NewWork.vue` (intent picker + create modal)
- `scripts/stewards-ui/frontend/src/views/Intents.vue` (new)
- `scripts/stewards-ui/frontend/src/views/Covenants.vue` (new)
- `scripts/stewards-ui/frontend/src/router.ts` (2 new routes)
- `scripts/stewards-ui/frontend/src/App.vue` (nav)

Live containers:
- `pg-ai-stewards-dev`: 5 new SQL files live-applied; pg_extern functions manually CREATE'd to work around Lesson #3.
- `pg-ai-stewards-ui`: rebuilt + restarted with new image (C.8 surfaces).
- `pg-ai-stewards-bridge`: restarted at session end.
- Soak: schedule_enabled=true at session end.
