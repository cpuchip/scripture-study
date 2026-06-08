---
date: 2026-06-08
title: Migration drift reconciled + CT2.2 render shipped (the k2/l13 regression catch)
workstream: pg-ai-stewards
mode: dev
tags: [ct2, migration-ledger, drift, inverse-hypothesis, regression, mosiah-4-27]
---

# Morning: clear the drift, then ship CT2.2

Michael woke, cleared the substrate restart ("I'm not using it"), and asked me to
proceed with CT2.2 — and to review/resolve the migration drift I'd flagged. Two
clean wins came out of it, plus a regression I caught before it could bite.

## 1. Migration-ledger drift — reviewed + reconciled (0 pending / 0 drift)

The bridge entrypoint runs `stewards-cli migrate` under `set -e` with no
`|| true`, so a migrate **exit-2 blocks the bridge from starting** — confirmed by
reading `bridge-entrypoint.sh`. So the drift HAD to be fixed before any restart.

The drift was 4 files (`4a-cost-tracking`, `h1-1-general-research-intent`,
`j10-provider-models-pricing`, `j11-provider-spend-caps`) whose ledger sha was
stale — all from legit commits (catalog prune, intent extension, pricing adds,
gemini cap) that were live-applied without re-recording. I verified the live DB
already had the content (gemini cap = the file's $18M, kimi pricing matches, the
research intent exists), so I **re-recorded the shas without re-applying**. Key
reason NOT to re-apply: `4a`/`j10`'s `model_pricing` INSERT omits `effective_at`
(defaults `now()`), so `ON CONFLICT` never fires and each re-apply adds duplicate
rows (that's why there were already two kimi-k2.6 rows). Then `migrate` recorded
the 6 genuinely-pending files (`cv12`, `r7/r8/r9`, `ct2-1`, `r10` — all already
live-applied, idempotent no-ops). Result: **216 files, 0 pending, 0 drift.**

## 2. CT2.2 — the render honors the state model (PURE SQL, no restart)

The spec said CT2.2 was a "Rust rebuild." It was wrong about the render layer:
`compose_messages` and `compose_system_prompt` are **plpgsql**, called by the
dispatch loop, so a `CREATE OR REPLACE` takes effect next dispatch — no rebuild,
no restart. CT2.2 became a safe, reversible, live-applyable SQL change.

Shipped (`ct2-2-context-render.sql`, commit d985b1c): an `agents.context_tools_enabled`
opt-in flag (default false), the pressure-line formatter, and the gated
`compose_messages` render — `[ctx:handle]` prefixes on addressable messages,
`context_state` honored (pinned=raw/exempt, compressed=engram, muted=tombstone),
**locked handles stripped** (the §4 breaker by absence), and the §5 pressure line
on the system message. Smoke (tools-on family + states on a real session):
pinned→handle+full, compressed(locked)→handle stripped, muted(locked)→
`[context muted]`, verbatim torso→handles, user msgs→no handle, pressure line
present. Tools-off → byte-identical to l13.

## 3. ★ The regression I caught (the lesson)

My first CT2.2 draft based `compose_messages` on **k2** (the original). But the
function had evolved k2→k6→k7→k8→k9→l1→**l13** (injection defense, provider
reasoning-strip rules, pressure-aware engram rendering, effective-budget cascade).
Basing on k2 silently **reverted all of that.**

What caught it: an **inverse-hypothesis hash check** — md5 the tools-off render
for a normal family before and after the apply. It MUST be identical (the gate is
off for that family). It wasn't (`c21b449e` → `24eed5bb`). That one check turned a
silent, live, behavior-changing regression into a 2-minute fix: re-applied l13
(hash returned to `c21b449e`), then rebuilt CT2.2 on the l13 base.

**Lesson (durable):** a `CREATE OR REPLACE` of a multiply-evolved function must
start from the LIVE/latest definition (`pg_get_functiondef` or the last migration
that touched it), never an old migration file. And: when a change is gated to be
a no-op on some path, *prove* it with a before/after hash of that path — don't
trust the gate by inspection.

## 4. CT2.3 — scoped, and it IS the Rust+restart step

CT2.3 (expose the levers so a dispatched agent can self-manage) needs a small
Rust change: `exec_sql_fn_tool` (tools.rs) runs `SELECT fn($1)` with only the
model's args — no session — so a lever can't resolve `[ctx:handle]`→message_id.
Plan: inject `_session_id` into the sql_fn args (backward-compatible — the 4
existing sql_fn tools ignore extra keys) + handle-resolution wrappers + tool_defs
+ grants + rebuild pg + restart. Held for fresh budget (Mosiah 4:27) — Michael
cleared restarts and the ledger is clean, so it's ready to run.

## Carry-forward
- CT2.3 (above) → then CT2.4 (A/B). §7 unratified.
- Root unpushed: the overnight 4 + this morning's CT2.2 + memory commits (Michael pushes root).
- The dup model_pricing rows (effective_at=now() trap on 4a/j10 re-apply) are a
  pre-existing wart — harmless (cost calc uses latest effective_at), not fixed.
