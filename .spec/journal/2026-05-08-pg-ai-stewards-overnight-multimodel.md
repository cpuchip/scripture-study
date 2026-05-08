# pg-ai-stewards overnight: multi-model comparison + HTTP tool path

*2026-05-08, Claude Code Opus 4.7, autonomous session while Michael sleeps.*

## What this session is

Michael went to bed. He authorized three phase chunks in his absence,
gave me stewardship over the Docker container and the code, and left
me with $2-5/day budget for experiments plus local qwen3.6-27b on
the dev box.

Goal: produce evidence that the substrate's voice can be tuned per
model, by running the same binding question (Faith/Hope/Charity
↔ Way/Truth/Life) through three configurations and comparing.

## Approved scope (three numbered phases)

| Phase | What | Status at plan time |
|-------|------|---------------------|
| **3c.3.3** | Importer `model_match` extension + apply kimi-tuned study variant | not started |
| **3c.3.4** | Multi-model FtC/WtL experiment — kimi-k2.6 + kimi-tuned prompt vs qwen3.6-27b + base prompt | not started |
| **3c.4** | gospel-engine-v2 HTTP tool registration via `pg_net` (no bgworker change) | not started — stretch |

Phase numbering rationale: 3c.3.3/.4 sit naturally after 3c.3.2 (the
substrate-produces-first-real-study win). 3c.4 was already reserved
on phases.md for HTTP tool registration. The previously-speculative
"auto-promote work_items into studies" item is bumped to 3c.3.5+.

## Approved choices (Michael, before bed)

1. qwen3.6-27b uses **base prompt** (model-neutral) — authentic look
   at qwen's natural voice for future tuning
2. Token budget **2M for each run** — comfortable on kimi with cache
   pricing, free on qwen
3. Container restart authorized for 3c.4. Stewardship granted. Check
   soak idle before any restart.

## Explicitly deferred to daytime

- **GLM-5.1 / Qwen3.6 Plus on opencode_go** — needs `STEWARDS_PROVIDER_*`
  env additions + container restart in a context that's already
  busy with experiments. Not while unsupervised.
- **Gemini 3.1 Flash/Pro** — Gemini API is not OpenAI-compatible.
  The substrate's bgworker only handles `kind=openai`. Adding a
  `kind=gemini` provider needs new Rust code in the chat dispatch
  path — same path the soak depends on. Eyes-open task, not
  midnight task.

## Phase 3c.3.3 — Importer `model_match` extension

**Why this is first.** `cmd/stewards-cli/internal/importer/agents.go:126`
hardcodes `model_match='*'` in the UPSERT. If I imported
`.stewards/kimi-k2.6/study.agent.md` today, it would *overwrite* the
base study agent for every model. The whole experiment falls apart.

**Plan.**
1. Add `ModelMatch` field to `agentFrontmatter` and `AgentDoc`.
2. Parse it from the YAML frontmatter (we already added that field
   when authoring `.stewards/kimi-k2.6/study.agent.md`).
3. Fall back to `'*'` when absent (keeps existing imports unchanged).
4. UPSERT becomes `(family, model_match)` — already the PK, no schema
   change needed.
5. Tool perms still keyed by `family` alone (the substrate's tool
   resolution doesn't currently fork by model — agent_tool_perms PK
   is `(agent_family, tool_pattern)`). So same perms apply across
   variants. This is correct for now.
6. Build, test against the fixture in `.stewards/kimi-k2.6/`, verify
   in DB that two rows exist for `family='study'`.
7. Run a regression import of `.github/agents/` to confirm the base
   import path is unchanged.
8. Commit.

**Acceptance.**
- `SELECT family, model_match FROM stewards.agents WHERE family='study'`
  returns two rows.
- `study` `*` row's prompt unchanged from prior content.
- `study` `kimi-*` row's prompt matches `.stewards/kimi-k2.6/study.agent.md` body.

## Phase 3c.3.4 — Multi-model FtC/WtL experiment

**Two new work_items, same pipeline (`study-write`), same binding
question, different agent variants and providers.**

### Run #2 — kimi-k2.6 + kimi-tuned prompt

- Provider: `opencode_go`
- Model: `kimi-k2.6`
- Token budget: 2,000,000
- Agent: resolves to `(study, kimi-*)` because model matches `kimi-*`
- Expected: ~17m elapsed, ~600K-1M tokens, $0.30-1.00

### Run #3 — qwen3.6-27b + base prompt

- Provider: `lm_studio`
- Model: `qwen/qwen3.6-27b`
- Token budget: 2,000,000
- Agent: resolves to `(study, *)` because no `qwen-*` variant exists
- Pre-flight: verify `host.docker.internal:1234/v1/models` is alive
  with the expected model loaded
- Expected: longer wall clock (local GPU bound, not API). 30-60 min.

### Output handling

The substrate's `study-write` pipeline writes stage results to
`work_items.stage_results` JSONB. It does NOT write to disk. After
each run completes, I'll dump the final `review` stage's text to:
- `study/.scratch/two-triplets-comparison-2026-05-08/run2-kimi-tuned.md`
- `study/.scratch/two-triplets-comparison-2026-05-08/run3-qwen-base.md`

The original (run #1) is preserved in the substrate's `work_items`
table. I'll also dump it to:
- `study/.scratch/two-triplets-comparison-2026-05-08/run1-original.md`

Then write `comparison.md` in the same folder — three-way diff
focused on whether the kimi-tuned prompt fixed the six signatures
identified in the 2026-05-07 review, and whether qwen exhibits
different signatures worth a future variant.

**Acceptance.**
- Three files in the comparison folder
- A comparison memo that names what changed vs what didn't
- Original work_item #1 untouched (still readable from substrate)

## Phase 3c.4 — gospel-engine-v2 HTTP tools (stretch)

**Critical constraint: no Rust bgworker changes.** The bgworker handles
chat dispatch for the soak; touching it risks the soak. So I'll
implement HTTP tools as pure SQL using `pg_net` (Postgres 18 has it
built into core via the contrib package; the `pgvector/pgvector:pg18`
base does include it, but I need to verify).

**Plan.**
1. Verify `pg_net` is available: `CREATE EXTENSION IF NOT EXISTS pg_net;`
2. Add new `execute_target` value `'http_proxy'` to `tool_defs`.
3. Author SQL helper `stewards.tool_http_dispatch(tool_name text, args jsonb)`
   that:
   - Looks up the tool's HTTP endpoint from a new `tool_defs.http_endpoint` column
   - Constructs the request from args + endpoint
   - Calls `pg_net.http_post` (async)
   - Polls `pg_net._http_response` for the result
   - Returns the parsed body as JSON
4. Register `gospel_search` and `gospel_get` rows in `tool_defs` with
   the engine.ibeco.me endpoint.
5. Grant `gospel_*: allow` to study + stewards-explore agents (NOT
   watchman — no tools by design).
6. Test: manual chat from a session, dispatch a `gospel_get` tool call
   for `John 14:6`, confirm clean roundtrip.
7. If clean, **possibly** dispatch a 4th study run with kimi-tuned
   prompt + gospel tools active to see if quote verification works
   end-to-end. (Run #4 is gravy; not required for tonight's value.)

**Container restart check.** `pg_net` registration is a `CREATE
EXTENSION` — runs inside an active session, no postmaster reload.
Should be zero-downtime. Confirms during step 1.

**Acceptance.**
- `tool_defs` has `gospel_search` and `gospel_get` rows
- A manual `tool_dispatch` call from `psql` returns parsed scripture text
- No bgworker changes, no container restart needed

## Soak interaction

The Watchman soak is running. It uses kimi-k2.6 via `opencode_go` for
its passes. My runs #2 and #3 also use kimi (run #2) and qwen (run
#3). Concurrent dispatch is OK — the bgworker drains the work_queue
in order. Worst case: study runs delay a Watchman pressure pass by
~20 minutes.

If 3c.4 needs a `CREATE EXTENSION pg_net`, that runs in a single
transaction with no locking on `work_queue`. Should be invisible to
the soak.

## Commit checkpoints (Michael's audit trail)

I'll create one commit per phase, plus the plan commit:

| # | What | When |
|---|------|------|
| 0 | This plan + phases.md updates | Now |
| 1 | 3c.3.3 — importer extension + kimi-* variant in DB | After A1+A2 |
| 2 | 3c.3.4 — comparison memo + scratch dumps | After A3-A5 |
| 3 | 3c.4 — gospel-engine HTTP tools (if shipped) | After B1-B2 |
| 4 | Memory + active.md + summary | At end of session |

`git log --oneline main` in the morning will show the order.

---

## Progress log

*(Filled in as I work. Each entry timestamped.)*

### 00:25Z — plan committed

Plan written to this file. About to start 3c.3.3.

### 04:36Z — 3c.3.3 shipped (importer model_match)

`cmd/stewards-cli/internal/importer/agents.go` now reads `model_match`
from frontmatter (defaults to `'*'`). Tool perms only rebuild on
default-variant imports (variants share family-level perms). README.md
and other frontmatter-less files in agent dirs are skipped silently
with verbose-mode info instead of counted as failures.

End-to-end verified:
- `import .stewards/kimi-k2.6` → `study/kimi-*` row at 14325 chars
- regression `.github/agents` re-import: 19 ok / 0 fail, kimi-* preserved
- both rows at `steps=50`
- `agent_tool_perms` count for `study` family unchanged (17)

Commit `46e68ea`.

### 04:55Z — 3c.3.4 dispatched (both runs in flight)

- **Run #2** `5303df09` — `study-write` pipeline (kimi-k2.6 via
  opencode_go), kimi-tuned prompt auto-resolves via `(study, kimi-*)`.
  Slug: `ftc-wtl-meta-v3-kimi-tuned`. Budget 2M.
- **Run #3** `7294ca66` — new `study-write-qwen` pipeline (cloned
  from `study-write` with provider=lm_studio, model=qwen/qwen3.6-27b),
  base prompt resolves via `(study, *)` since qwen-3.6 doesn't match
  `kimi-*`. Slug: `ftc-wtl-meta-v3-qwen-base`. Budget 2M.

Soak idle at dispatch time (last pass 04:20Z, next at ~05:20Z).
Both work_queue chats observed making tool calls within 8s of dispatch.

Original (run #1) review output dumped to scratch as
`run1-original-kimi-base-prompt.md` — 105 lines, the 3c.3.2 artifact
that became `study/two-triplets-one-ascent.md` after Michael's read
and my Opus-4.7 revision pass.

### 04:50Z — 3c.4 deferred to daytime

Neither `pg_net` (Supabase) nor `http` (pramsey) extensions are
available in the `pgvector/pgvector:pg18` base image, and only
`plpgsql` is registered as a procedural language. Both extensions
need source-build against PG18, with uncertain PG18 support upstream
in their stable release branches.

Building an extension from source at midnight against PG18 in
autonomous mode is exactly the kind of "fighting the toolchain" risk
I should not take. Deferred 3c.4 to daytime where Michael can
supervise the Dockerfile change and where I can get a clean answer
on `pg_net` PG18 support without web access.

Alternative path (also deferred): extend the Rust bgworker with a
new `work_kind = tool_http` that uses reqwest directly. This avoids
extension issues but touches the chat-dispatch code path that the
soak depends on — same supervised-daytime caveat applies.

### 05:00Z — bug found: my reimport wiped substrate-internal perms

While checking run #2's progress, found the kimi-tuned agent
**refusing to call substrate study tools** with: *"the
substrate-internal tools you described (`study_search_text`,
`study_get`, `study_similar`, etc.) are not in my available
function namespace. I only have the `skill` tool loaded."*

Root cause: the 3c.2.5 broadcast grant `INSERT INTO agent_tool_perms
... study_*: allow` was a substrate-internal SQL grant **not declared
in any agent's frontmatter `tools:` list**. The importer's
`DELETE FROM agent_tool_perms WHERE agent_family=$1` (3c.3.3 path
for default-variant imports) blew it away, then re-inserted only the
patterns the agent file declares. Result: study tools effectively
de-permitted for every agent family I touched in the regression
re-import.

**Immediate fix.** Re-applied the 3c.2.5 broadcast manually:

```
INSERT INTO stewards.agent_tool_perms (agent_family, tool_pattern, action)
SELECT DISTINCT a.family, 'study_*', 'allow'
  FROM stewards.agents a
 WHERE a.family NOT LIKE 'watchman%'
ON CONFLICT (agent_family, tool_pattern) DO UPDATE SET action = EXCLUDED.action;
```

20 rows restored.

**Root-cause fix.** Added `'study_*'` to the `tools:` frontmatter
of both `.github/agents/study.agent.md` and
`.stewards/kimi-k2.6/study.agent.md`. Now a reimport will
re-grant the perm correctly without depending on the SQL broadcast.

**Architectural note worth carrying forward.** `agent_tool_perms`
currently has no notion of source-of-truth. Frontmatter-declared
patterns and substrate-internal broadcasts both end up as
`(agent_family, tool_pattern, action)` rows with no provenance.
The importer's delete-then-insert blows them all away. Two options
for a more robust design (deferred to daytime):
1. Add a `source text` column (`frontmatter`/`broadcast`/`manual`),
   only delete `source='frontmatter'` on import.
2. Move all canonical broadcasts into agent frontmatter (current
   workaround) — disciplined but loses the "broadcast" semantic.

**Impact on tonight's experiment.** Runs #2 and #3 dispatched
during the perm-broken window. Both will produce studies *without
substrate corpus access*. Run #1 had access. The comparison is now
apples-vs-different-apples: voice-only, not corpus-grounded-voice.
Kept running rather than abort — the voice signal is still real,
and the documented bug is itself a useful artifact. Will note this
clearly in the comparison memo.


