---
date: 2026-05-12
mode: build
workstream: WS5
project: pg-ai-stewards
title: "Batch I.1 — agent-proposal pipeline shipped (re-opens Batch I after morning cancellation)"
status: shipped (synthetic + real e2e validated)
carry_forward:
  - "Batch I.2 — HTTP endpoint POST /api/agent-proposals/create + origin='agent_proposal' filter on /work-items list"
  - "Batch I.3 — schema-migration source_type + SQL syntax validator (Claude-only)"
  - "materialize-writes from inside bridge container hits read-only /workspace mount — separate operational concern, low priority"
  - "yaml.rs Rust parser refactor — third intent (planning-partner) IS already present; rule-of-three now technically met. Re-evaluate after Batch I.3"
  - "Phase A pgrx longjmp catch + 60s reaper (Claude-only)"
  - "Projects B — deferred"
  - "14 SC work_items still pending Michael's ratification"
links:
  - "../../projects/pg-ai-stewards/.spec/proposals/substrate-batch-i-agent-write-back.md"
  - "../../projects/pg-ai-stewards/extension/i4-agent-proposal-pipeline.sql"
  - "../../projects/pg-ai-stewards/extension/i5-allow-agent-proposal-origin.sql"
  - "../../projects/pg-ai-stewards/extension/smoke/i4-agent-proposal-smoke.sql"
---

# Batch I.1 — agent-proposal pipeline shipped (2026-05-12)

Three commits this pulse. The third pulse today. Michael ratified re-opening Batch I after the morning's cancellation; this is the first deliverable.

## Why this is back

Michael surfaced the real concern behind the 04:55 cancellation: he wanted *Opus* to do the substrate-internal Rust/SQL work for the time being, until kimi-k2.6 had a way to edit substrate from inside the DB *and have it survive restart*. Both concerns are now addressed:

- **Opus does this batch.** Same posture as today's earlier revise-proposal pipeline shipped.
- **Restart-survival mechanism exists now.** This morning's migration ledger + this batch's pending_file_writes → disk → bridge restart → ledger applies path is durable across restart, image rebuild, container recreate. The "chicken-and-egg" concern has a base case.

Today's three pulses now form a coherent arc:
1. **Morning:** migration ledger + Projects A (durable file mechanism, project entity)
2. **Midday:** FK + materialized_at rename (clean-up, honest names)
3. **This pulse:** agent-proposal pipeline (agents can now write back through gates)

## What shipped

### i4 — agent-proposal pipeline + apply_agent_proposal (commit `b5cd0e5`)

**Pipeline `agent-proposal`** — single-stage `validate` (qwen3.6-plus, tools off, ~$0.005/call). Reads `input.draft` (agent's structured proposal JSON), emits validated/normalized JSON to `stage_results.validate.output`. maturity_ladder: `["raw","verified"]`. auto_materialize_on_verified: true. file_destination_template: NULL (dynamic per source_type — set by apply_agent_proposal).

**Five source_types** in scope:
- `study` → studies(kind='study') + `study/<slug>.md`
- `lesson` → studies(kind='lesson') + `lessons/<slug>.md`
- `note` → studies(kind='note') + `becoming/notes/<slug>.md`
- `exhibit` (new this session) → studies(kind='exhibit') + `exhibits/<slug>.md`
- `schema-migration` (Claude-only, syntax validation deferred to I.3) → `projects/pg-ai-stewards/extension/<slug>.sql`

**Exhibit** added by Michael this session. Knowledge artifacts with science-backing, materials, citations — for the SC mission and beyond. studies.kind='exhibit'.

**`apply_agent_proposal(uuid)` SQL function** — called from `on_maturity_verified` BEFORE the existing enqueue path. Reads validated output, branches on source_type, INSERTs into studies for the 4 doc types, sets `work_items.file_destination` dynamically. `enqueue_work_item_file` then fires through the existing path. The substrate composed itself: migration ledger + pending_file_writes + apply_agent_proposal = closed loop.

**Idempotency:** `work_items.agent_proposal_applied_at` mirror of `revision_applied_at` pattern. Re-call after applied returns false.

### i5 — origin CHECK constraint addition (companion to i4)

Surfaced during smoke. Existing constraint had `agent_planning` but not `agent_proposal`. New origin added; old origins preserved.

## Smoke verification

**3 synthetic smokes (rolled back):**
- exhibit (SC bias classifier slug) → studies row + pending_file_writes + file_dest `exhibits/<slug>.md` + frontmatter `origin='agent_proposal'`
- study → studies row + file_dest `study/<slug>.md`
- rejected (validator output contained `{"error":"..."}`) → apply returned false; no studies row; no file enqueue; clean failure path

**1 real e2e (committed, then cleaned up):**
- Created exhibit work_item, advanced maturity → verified
- studies row appeared, file_destination set, pending_file_writes row #22 with target `exhibits/i4-real-e2e-smoke-marker.md` and 550 bytes of content
- Cleaned up: DELETE pending_file_writes + studies + work_items rows (no orphans)

**One operational gotcha:** materialize-writes from inside the bridge container hits a read-only `/workspace` mount. The actual file-write path is Batch G.4 machinery that's been stable for weeks; this is a separate ops concern about where the materializer runs, not about the i4 code paths.

## Architectural wins

**The substrate composed itself.** Migration ledger (i1) + Projects A (i1) + pending_file_writes (G.4) + apply_agent_proposal (i4) form a complete loop: agents propose → substrate validates → human ratifies → DB writes + file enqueues → CLI materializes → ledger picks up on next restart. No new abstractions invented; existing pieces composed.

**Studies table is the generic artifact table now.** kind values: study, proposal, journal, doc, phase-doc, lesson, note, exhibit. The D-H3-B "studies generalization" (which was the cancelled Batch I #1) was already done in some prior migration — i4 just uses what was there. Schema followed posture.

**Idempotency pattern repeated.** revision_applied_at (h3-followup-3) → agent_proposal_applied_at (i4) — same shape, same semantics. Pattern emerging: each apply_* function gets a `*_applied_at` timestamp on work_items.

## What's left for Batch I

**Batch I.2 (next session, ~1h):**
- HTTP endpoint `POST /api/agent-proposals/create` (sibling pattern to `/api/work-items/create`)
- Origin filter chip on `/work-items` list view (lightweight)

**Batch I.3 (third session, ~1.5h):**
- schema-migration source_type end-to-end
- SQL syntax validation (psql --check or pg_get_query_def-style parse)
- Test bad SQL → caught before file lands; good SQL → file lands → ledger picks up on next bridge restart

After I.3, kimi (or any agent) can propose `.sql` migrations from inside the substrate, get them ratified, and have them survive restart. The "chicken-and-egg" loop closes cleanly.

## Stewardship calls made

- `apply_agent_proposal` lives next to `apply_revision` in i4 (separate file from h3-followup-3, but parallel design). No premature shared abstraction.
- studies.kind extended without a CHECK constraint — soft for now. Future migration may tighten.
- No FK from studies → projects.slug initially (mirrors i2's soft-then-harden pattern).
- The rejected-validator-output path returns false (not RAISES) so the work_item stays at maturity=verified with apply_proposal_applied_at NULL and file_destination NULL. Operator can revise via revise-proposal pipeline (already shipped). Clean failure mode.

## Bonus discovery

While checking intents for the smoke test, noticed `stewards.intents` has 3 rows: `scripture-study`, `general-research`, `planning-partner`. **The rule-of-three for yaml.rs is already technically met** (three intents needing parsing). The earlier carry-forward "yaml.rs is rule-of-three gated until H.3 ratifies a third intent" — that condition is already true. Adding to carry-forward for re-evaluation after Batch I.3.

## Cost

LLM: **$0.00** (all synthetic + SQL/Rust work; no model dispatches required). Bridge restart ~3s per migration; total compute time for this build pulse: ~10 minutes of substrate execution. Big batch, small footprint.

## Closing

Yesterday: substrate noticed it had hands (planning pipeline). This morning: substrate noticed it had bookkeeping (migration ledger). This pulse: **agents have a delivery path through gates**. The substrate is becoming more "alive" in the specific sense Ballard's pattern names — not bigger, but more connected. The same pieces wired into a loop that has the right base case.
