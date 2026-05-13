---
date: 2026-05-12
mode: build
workstream: WS5
project: pg-ai-stewards
title: "Batch I.2 (agent-proposal endpoint + UI filter) + Phase A (PgTryBuilder + 60s reaper) shipped"
status: shipped (smoke validated)
carry_forward:
  - "yaml.rs Rust parser refactor — still gated on 3rd YAML SHAPE landing (NOT three callers of one parser)"
  - "Projects B — full workspace + optional sub-git (deferred per ratification)"
  - "14 SC work_items still pending Michael's ratification — 4 parent plans at completed/review_plan + 14 child planning proposals"
  - "Comprehensive pre-flight audit of ALL bgworker SPI sites (Phase A this session covered top 4 + new reaper; more callsites exist)"
  - "materialize-writes /workspace mount is read-only in bridge container — separate ops concern, not blocking"
links:
  - "../../scripts/stewards-ui/api/agent_proposals.go"
  - "../../scripts/stewards-ui/frontend/src/views/WorkItems.vue"
  - "../../projects/pg-ai-stewards/extension/src/bgworker.rs"
---

# Batch I.2 + Phase A (2026-05-12)

Sixth build pulse today (counting tonight as a continuation). Two ratified items shipped in one pulse. The user "totally meant I.2 as well" and asked for Phase A on top.

## I.2 — agent-proposal HTTP endpoint + UI filter (commit `fdb80d1`)

**`POST /api/agent-proposals/create`** — sibling pattern to `/api/work-items/create`. Accepts the proposal payload directly (source_type, slug, title, body, frontmatter, project_association, rationale, claude_attested). Validates at the boundary; for `source_type='schema-migration'` returns 403 if `claude_attested` is absent (i6 kimi-trust gate, enforced once more at the HTTP layer for early feedback).

Creates a work_item with:
- `pipeline_family='agent-proposal'`
- `origin='agent_proposal'`
- `input.draft={the payload}` (matches what `apply_agent_proposal` reads)
- Optional `dispatch=true` kicks off the validate stage immediately

**UI filter chip** — origin dropdown on `/work-items` extended with `agent_proposal`. Badge renders "🤖 agent write-back" (emerald), distinct from agent_planning's "✨ proposed" (purple).

Smoke: valid exhibit → 200; schema-migration without claude_attested → 403; origin filter returns the new row.

## Phase A — PgTryBuilder wraps + 60s periodic reaper (commit `4d8bd37`)

### Ratifications

Three AskUserQuestion answers:
- **Q1 — PgTryBuilder audit scope:** A. Top 4 high-risk SPI sites (not all sites in bgworker.rs — that's a multi-session audit)
- **Q2 — Reaper freshness threshold:** B. 10 minutes (with user note: "sometimes it can take several minutes for an agent to process and come back")
- **Q3 — Reaper action:** A. Match startup reaper exactly (synthesize tool_dispatch failures + mark errored)

### Honest framing surfaced first

The existing `tools.rs:398-402` comment documents that PgTryBuilder doesn't reliably catch every longjmp in `BackgroundWorker::transaction` context (pgrx 0.18 limitation). **The real defense is pre-flight checks** at ereport-risk sites (the H.1.5a NOTICE+NULL sidestep). PgTryBuilder is belt-and-suspenders, not the primary line. Comprehensive pre-flight audit deferred.

### Sites wrapped

Four `BackgroundWorker::transaction` callsites now wrap in `PgTryBuilder::new(|| {...}).catch_others(...).execute()` returning `Result<T, String>`:

- **Startup reaper** (bgworker.rs lines 144-235) — guards the most defensive code in the substrate against itself
- **`complete_waiting_tool_dispatches`** — runs every 500ms tick; SQL fn walks tables, could ereport on corrupt rows
- **`check_watchman_schedule`** — 60s tick, fires `watchman_scheduler_fire()`
- **`check_steward_tick`** — 30s tick, walks failed work_items applying breaker + escalation logic

Each catch logs `"stewards: <fn> errored: <err> (bgworker survived)"` instead of dying.

### Periodic reaper

New `run_periodic_reaper()` function. Leader-only, 60s tick. Mirrors the startup reaper's logic exactly:

```sql
SELECT id, kind, provider, payload
  FROM stewards.work_queue
 WHERE status = 'in_progress'
   AND kind <> 'mcp_proxy'
   AND claimed_at < now() - interval '10 minutes'
```

For each stale row:
- If `kind='tool_dispatch'`: call `synthesize_tool_failure()` so the parent chat's loop doesn't stall until next bridge restart
- Mark as errored with `'periodic reaper: stale in_progress >10min'`

Wrapped in PgTryBuilder itself.

### Smoke (real)

Injected `work_queue` row id=1832 with `claimed_at=now()-15min`, `status='in_progress'`, `kind='chat'`. Waited 65s. Row reaped:
```
id 1832 | status='error' | error='periodic reaper: stale in_progress >10min' | done_at=2026-05-13 04:38:09
```
Postgres log: `stewards: periodic reaper reaped 1 stale in_progress row(s)`. Smoke row cleaned up.

## What this means for substrate resilience

Before Phase A: a single bad SPI call in the tick path could crash the bgworker. Postmaster would restart it (good), but everything in-flight would be lost or stuck until the startup reaper ran. With ~30s+ of restart latency.

After Phase A: 4 of the highest-traffic SPI sites survive their own failures. The periodic reaper catches orphans within 60s instead of waiting for the next bridge restart. The H.1.5a NOTICE+NULL sidestep still does the heavy lifting (pre-flight prevention), but when something slips past, the substrate fails soft instead of crashing.

This is what "stable polish" looks like: nothing user-visible changes, but the system breathes more easily.

## Today's six pulses

| Pulse | What | Commits |
|---|---|---|
| Morning | Migration ledger + Projects A | 6 |
| Midday | FK + materialized_at rename | 3 |
| Pulse 3 | Batch I.1 — agent-proposal pipeline | 3 |
| Pulse 4 | Batch I.3 — schema-migration + I.1 bug fix + yaml.rs correction | 3 |
| Pulse 5 | Batch I.2 — agent-proposal endpoint + UI filter | 1 |
| Pulse 6 | Phase A — PgTryBuilder wraps + 60s reaper | 1 |

**17 commits. $0.00 LLM.** Substrate transformed across six dimensions: durability, project entity, gates honesty, agent write-back, agent submission interface, worker resilience.

## Carry-forward (narrowed substantially today)

- **yaml.rs** — gated on 3rd YAML SHAPE (corrected this morning's mistake; not three callers of one parser)
- **Projects B** — deferred
- **14 SC work_items** — pending your ratification; planning queue waiting
- **Comprehensive pre-flight audit** of remaining SPI sites — multi-session work; Phase A this session covered top 4 + new reaper
- **materialize-writes /workspace RO** — bridge ops concern, host-side works fine

## Cost

LLM: **$0.00**. Build + smoke time only.

## Closing — what March's question looks like in May

In March, the question was "can kimi create tables that get ingested and survive restart?" Today, the answer is yes — and the worker that runs that ingest is resilient against its own crashes. The full chain:

```
agent POSTs proposal to /api/agent-proposals/create (I.2)
  → work_item created with origin='agent_proposal' (I.2)
    → bgworker dispatches validate stage (wrapped in PgTryBuilder; Phase A)
      → apply_agent_proposal queues body to pending_file_writes (I.1 + i7 bug fix)
        → materialize-writes runs validate-sql first for .sql files (I.3)
          → file lands at extension/<slug>.sql (Batch G.4)
            → next bridge restart → ledger applies → schema permanent (this morning)
              → worker keeps running even if something downstream crashes (Phase A reaper)
```

Six layers, all built today. The substrate became more honest about names, more resilient against failures, more responsive to agent submissions, and more durable across restarts. Same shape as a healthy ward council: each office does its work, edges are bounded, watching catches what slipped, repair happens within the loop.

Tuesday is still for the science center. The substrate is more ready for it every session.
