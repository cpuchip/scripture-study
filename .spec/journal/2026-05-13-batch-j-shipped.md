---
date: 2026-05-13
mode: build
workstream: WS5 (substrate) + WS9 (science center)
project: pg-ai-stewards
title: "Batch J shipped — fan-out + brainstorm primitives + UI hierarchy"
status: shipped (J.1, J.2, J.4, J.5 verified end-to-end; J.3 partial — 2 of 6 children produced publishable briefs)
carry_forward:
  - "research-write token-limit failures (Kimi 262K input limit) on 2 of 6 J.3 children — surfaced in gather stage when 5 rounds of web search return rich content. Substrate-level issue; not blocking J.3 partial completion."
  - "Brainstorm-lens output landed in single combined file via aggregator; individual lens outputs only viewable via work_item detail page. Future: per-lens file_destinations so each lens output is its own artifact."
  - "Aggregator's 'Children' table column 'Link' points to the index file itself for lens children (no per-lens file_destination). Minor UI polish for later."
  - "Local-model (LM Studio / Ollama) integration for brainstorm lenses deferred. Pattern in place: add agent + pipeline pointing at local provider row."
  - "stewards-cli migrate doesn't handle re-runs of partially-applied files cleanly — surfaced when j6 was applied via psql -f then re-attempted via migrate. Fix: add DROP TRIGGER IF EXISTS for new triggers before CREATE TRIGGER."
links:
  - "../proposals/substrate-batch-j-fanout-brainstorm.md"
  - "../../projects/pg-ai-stewards/extension/j1-fanout-machinery.sql"
  - "../../projects/pg-ai-stewards/extension/j2-aggregate-auto-verify.sql"
  - "../../projects/pg-ai-stewards/extension/j3-spawn-sets-aggregator-destination.sql"
  - "../../projects/pg-ai-stewards/extension/j4-spawn-supports-child-file-destination.sql"
  - "../../projects/pg-ai-stewards/extension/j5-brainstorm-lenses.sql"
  - "../../projects/pg-ai-stewards/extension/j6-brainstorm-lens-auto-verify.sql"
  - "../../scripts/stewards-ui/api/work_items.go"
  - "../../scripts/stewards-ui/frontend/src/views/WorkItems.vue"
---

# Batch J Shipped — 2026-05-13

Plan ratified earlier today (12 decisions in 3 AskUserQuestion batches), all five sub-phases shipped in one continuous build session. Michael's framing — "we need more legs here, help me build the museum of my dreams" — landed across the substrate as two new primitive shapes and the UI affordances that make them usable.

## What shipped (in commit order)

| Commit | Sub-phase | Substrate / UI changes |
|---|---|---|
| `377b12f` | J.2 | `j1-fanout-machinery.sql` + `j2-aggregate-auto-verify.sql` + `j3-spawn-sets-aggregator-destination.sql` — decompose-fanout + aggregate-children pipelines, spawn_children() SQL fn, on_maturity_verified branches, aggregator auto-verify |
| `c6ce6e0` | J.1 | UI: status group dropdown (open / done / all), tree rendering with indent + expand/collapse, parent-link badge. API: open/done virtual statuses |
| `f6e81b9` | J.4 | `j5-brainstorm-lenses.sql` + `j6-brainstorm-lens-auto-verify.sql` — 4 lens agents (SCAMPER, Six Hats, Crazy 8s, Reverse) + 4 single-stage pipelines + start_brainstorm() entry point. Trigger generalized to all one-shot pipelines |
| `23ccfd0` | J.3 | `j4-spawn-supports-child-file-destination.sql` — per-child file_destination + science-center exhibits fanout script |

**Total LLM cost across the day:** ~$0.05 (J.2 smoke) + $0 (J.1 UI) + $0.09 (J.5 brainstorm) + J.3-in-flight (currently ~$1.50, ETA ~$3-5).

## The framework move

The substrate gained two reusable shapes today:

**Fan-out** (`decompose-fanout` + `aggregate-children`): one binding question → N child work_items → roll-up index/digest. Each child gets its own pipeline, its own cost cap, its own quality review. The aggregator dispatches event-driven when all (non-failed) siblings verify. Per-child file_destination means each artifact lands at its own path.

**Brainstorm** (4 lens pipelines + `start_brainstorm()` SQL function): NOT a new pipeline shape — implemented as a SPECIAL CASE of fan-out. The decompose stage is pre-populated (deterministic, always the same 4 lens children + synthesis aggregator). Brainstorm composes from fan-out primitives. This is the "more legs" point — once fan-out exists, brainstorm is ~250 lines of SQL.

The aha moment: **brainstorm-then-fanout chaining is just two linked work_items**. A brainstorm produces a ranked candidate list at one file; a follow-on fan-out can take K of those candidates as its decompose manifest. No third "brainstorm-then-fanout" pipeline shape needed.

## Smoke results

### J.2 — Synthetic 2-child echo-test fanout (commit pulse)
Hand-crafted parent, echo-test children, verified all branches of on_maturity_verified fire correctly. Aggregator wrote 575 bytes of clean index markdown via pending_file_writes → materialize-writes → disk. ~$0.05.

### J.5 — Real biology-exhibits brainstorm
End-to-end via `SELECT stewards.start_brainstorm('What biology-focused interactive exhibits...', 'projects/space-center/brainstorm/biology-exhibits-candidates.md', 'space-center', 'michael', '...', 250000)`. All 4 lenses ran in parallel (queued behind J.3's 6 children — bgworker has 4 parallel slots), aggregator synthesized with cross-cutting analysis.

Notable: the synthesis genuinely added value beyond the 4 lens outputs. The aggregator detected that Winogradsky columns appeared in 3 of 4 lenses (convergence signal), surfaced 6 design principles from the Reverse session as gates for the other lenses, and flagged the rural-Missouri-maintenance constraint that both Six-Hats BLACK and Reverse independently identified. The output is 4036 bytes of usable markdown at `projects/space-center/brainstorm/biology-exhibits-candidates.md`. Total cost: ~$0.09.

### J.3 — Science-center exhibits fanout (final)
Hand-crafted 6-child manifest from the 8218aa77 survey. Each child runs research-write with a ~250-word binding question requesting the 6-field exhibit-brief structure (Story / Application / Demo / Science / History / Build). Each child writes to its own `projects/space-center/exhibits/<slug>.md`.

**Final state: 2 of 6 children verified + aggregator wrote a 4KB index.** 4 children failed at various stages:
- crystal-radio: gather, token limit (376K requested, 262K cap)
- bacteriopolis-winogradsky: gather, token limit (373K vs 262K)
- cs-unplugged: review, Moonshot 400 (token limit)
- indicating-electrolysis: review (after slow MSDS fetch unblocked); cost $0.42

2 children verified and wrote substantial briefs:
- rural-electrification-webster-coop.md — 17KB, $0.81. Includes specific Webster Electric 1946 charter date, 2010 smart meter deployment, partnership concept note for WEC sponsorship, ~$314 build cost.
- symmetry-polyhedra.md — 16KB, $0.49. Genuinely publishable quality: Marcus Hagler narrative from Versailles MO, Osage/Gasconade River bridge references, Klein's 1872 Erlangen Program with arXiv citation, California Math Show 1995 with named directors and funding source. Full markdown citations throughout.

**Bug surfaced + fixed during J.3** (j7-failed-sibling-also-triggers-aggregator.sql):
When the LAST sibling fails (rather than verifies), the original on_maturity_verified branch B never fired because of the early-return on maturity check. So the aggregator stayed pending forever even though all siblings were terminal. Required a manual `work_item_dispatch_stage` to unblock. Fix: extract sibling-count + dispatch logic into idempotent helper `check_and_dispatch_fanout_aggregator(parent_id)`; add a second trigger `on_child_status_terminal` that fires on status='failed'|'cancelled' and calls the same helper. Now the chain converges whether the last sibling succeeds OR fails.

**Aggregator quality issue** (carry-forward, not blocking): The aggregator's index README links to cs-unplugged and indicating-electrolysis as if they completed, when actually they failed. The aggregator read each child's stage_results via work_item_show but didn't gate the link on maturity=verified. Future enhancement: aggregator prompt should explicitly check each child's terminal status before linking.

**Slow tool-call hung the chain temporarily**: A `fetch_url_raw` call to a Universal Indicator MSDS site hung for 11+ minutes before I manually cancelled it. The periodic reaper (Phase A from yesterday) excludes `kind='mcp_proxy'` by design (since chat dispatches wait synchronously on tool results), so this case isn't covered. Carry-forward: consider per-tool timeouts at the bridge layer, or a separate reaper that catches stuck mcp_proxy older than N minutes.

**Total J.3 cost: ~$2.95** across all 6 children (including failures) + $0.05 aggregator. Within budget.

## Issues surfaced

**Research-write token-limit failures.** Gather stage with 5 rounds of search+fetch consistently breaks Kimi K2.6's 262K context window on topics that produce rich source material. Fix options for follow-up: limit gather to 3 rounds; switch gather to a 1M-context model; or trim each fetched source more aggressively before feeding to synthesize.

**Pre-commit hook + orphaned pending_file_writes.** During J.2 smoke, a row in pending_file_writes with `target_path='{{input.destination}}'` (the bogus template before j3 fix) got materialized by the pre-commit hook, writing a file literally named `{{input.destination}}` in the repo root. Cleaned up manually. Worth noting: smoke cleanup should always match the bogus template paths too.

**migrate non-idempotent on partially-applied files.** j6 was applied via `psql -f` (committed via pgsql tx), then re-attempted via `stewards-cli migrate` (separate tx) which hit duplicate-trigger error on CREATE TRIGGER. Fix: add `DROP TRIGGER IF EXISTS` before every CREATE TRIGGER in migration files. Will be discipline going forward.

**Trigger fires from retroactive UPDATE.** The j6 retroactive `UPDATE … SET maturity='verified' WHERE pipeline_family LIKE 'brainstorm-%' AND status='completed' AND maturity <> 'verified'` flipped 4 already-completed brainstorm rows, which fired on_maturity_verified's aggregator-dispatch branch immediately. That's the intended behavior — surfaced here because the J.5 chain had been blocked by the auto-verify gap and the retroactive UPDATE unblocked it.

## What the substrate looks like now

Six new SQL files in `projects/pg-ai-stewards/extension/`:
- `j1-fanout-machinery.sql` — pipelines + spawn_children + on_maturity_verified extensions
- `j2-aggregate-auto-verify.sql` — initial aggregator auto-verify trigger (superseded by j6)
- `j3-spawn-sets-aggregator-destination.sql` — bug fix: aggregator file_destination set at spawn
- `j4-spawn-supports-child-file-destination.sql` — per-child file_destination
- `j5-brainstorm-lenses.sql` — 4 agents + 4 pipelines + start_brainstorm()
- `j6-brainstorm-lens-auto-verify.sql` — generalized auto-verify for all one-shot pipelines

All 6 recorded in `stewards.schema_migrations` via `stewards-cli migrate`. Future image rebuilds apply them automatically.

UI changes:
- `scripts/stewards-ui/api/work_items.go` — `?status=open|done` virtual groups
- `scripts/stewards-ui/frontend/src/views/WorkItems.vue` — tree render + status-group dropdown + parent-link badge

## What this enables next

Tuesday — Science Center day. The brainstorm + fanout primitives are ready. Michael can:
- `SELECT stewards.start_brainstorm(...)` to spin up a 4-lens brainstorm on any question
- Hand-craft a fanout manifest from any survey to produce N briefs in parallel
- See the parent → children tree directly in the work-items UI with the open filter showing what's in flight

The 14 SC work_items still pending ratification become easier to triage with the tree view (they're all children of 4 parent plans).

## Cost summary

| Activity | Cost |
|---|---|
| J.2 smoke (2-child echo-test fanout) | $0.05 |
| J.5 brainstorm (4-lens biology) | $0.09 |
| J.3 exhibits fanout (in flight) | ~$1.50 currently, ETA $3-5 |
| **Day total** | **~$1.65 (current), $3-5 (projected)** |

## What's next (carry-forward)

- Resolve research-write token-limit gather failures (substrate)
- Per-lens file_destination so each lens output is its own artifact (UI/aggregator polish)
- Local-model integration for brainstorm lenses (LM Studio / Ollama provider rows + new lens pipelines)
- Friend's 8-9 brainstorming modes — Michael to share async; integrate as additional lens agents
- yaml.rs (gated on 3rd YAML shape — unchanged)
- 14 SC work_items pending ratification (unchanged)
- Resume soak at session close
