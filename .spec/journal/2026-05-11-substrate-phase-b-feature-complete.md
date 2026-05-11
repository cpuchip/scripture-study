---
date: 2026-05-11
session_kind: build
workstream: WS5
substrate_phase: B
commits: [263dc3e, 0a0f38d, 1a6c05b]
cost_usd: 0.04
---

# Substrate Phase B — feature complete on dev

## What shipped

Three commits today land Phase B end-to-end. (5a was committed last session as 263dc3e; tonight's pushes are 0a0f38d and 1a6c05b.)

**5a (last session, 263dc3e):** maturity-gate substrate — `work_items.maturity` + `scenarios` + `revision_count` + `spec` + `destination_maturity`. `pipeline_stage_maturity` table seeded for study-write + study-write-qwen (outline→planned, draft→executing, review→verified). `gate_decisions` audit ledger. `gate_prompts` with three templates (evaluate, generate_scenarios, verify). `verify_results` table. Functions: `render_template`, `evaluate_gate`, `apply_gate_decision`, `parse_gate_response`. The smoke test (`verify-5a.sql`) walked all 9 sections — advance, revise×3 with cap auto-surface, surface, audit trail.

**5b (tonight, 0a0f38d):** scenarios + verify — `generate_scenarios(work_item)` enqueues a chat with `_scenarios_gen=true`; `apply_scenarios_result` writes the JSON array to `work_items.scenarios`. `verify_work_item(work_item)` enqueues a chat with `_verify=true` (prompt includes the scenarios); `apply_verify_result` writes to `verify_results` and on failure drops maturity back to `planned` + `status=failed` so the steward picks up a re-execution.

**5c (tonight, 0a0f38d):** tiny fix caught during the e2e smoke. `sessions_kind_check` allowed only `chat|agent|tool|study|dev`. `evaluate_gate` (and the two new 5b functions) insert sessions with `kind='gate'` for clean audit separation between regular agent chats and gate-evaluation chats. Constraint extended to include `'gate'`.

**bgworker auto-fire (tonight, 0a0f38d):** the missing piece that turns Phase B from "schema-and-functions" into "actually autonomous." After a chat completes, the bgworker inspects payload markers `_gate_eval` / `_scenarios_gen` / `_verify` and auto-calls `parse_gate_response` → `apply_gate_decision` / `apply_scenarios_result` / `apply_verify_result`. Errors logged but never propagated — chat is already saved + work_queue is `done`; failed auto-apply leaves the work_item un-transitioned for human re-trigger or hand-apply. Means the human never has to manually call any of the apply_* functions — the gate fires, the LLM responds, the substrate walks the maturity ladder on its own.

**B.4 + B.5 UI surfaces (tonight, 1a6c05b):**
- `NewWork.vue` gains a `destination_maturity` dropdown (default empty = full Ammon-loop to verified, or pick a rung to surface earlier). `new_work.go` carries `DestinationMaturity` in the request struct; UPDATEs after `work_item_create()` since that function doesn't take it.
- `WorkItemDetail.vue` gains 4 new conditional sections: maturity ladder (visual stepper raw→...→verified, current rung emerald-highlighted, destination_maturity ringed in blue, revision count shown if >0), scenarios (acceptance-criteria list), spec (preformatted text), gate decisions audit (action-tinted rows with reasoning + collapsible feedback + raw_response). All four panels invisible when there's nothing to show, so existing pre-Phase-B work_items render unchanged.
- New `/api/work-items/gate-decisions` endpoint backs the audit panel.

## End-to-end test

Synthetic study-write outline work_item with mock outline output (D&C 130:18-19 — intelligence in the resurrection, 5-part structure with Hebrew/Greek word study). Called `evaluate_gate(id)`. Watched the work_queue:

1. Chat 1184 dispatched to qwen3.6-plus
2. Model returned `tool_calls` (decided to research before evaluating)
3. Bgworker enqueued continuation chat 1185, then a tool_dispatch round, etc.
4. Loop ran 1184→1195: 6 chat rounds + 5 tool_dispatch rounds
5. Chat 1194 returned a content message with the JSON gate decision: `action: "revise"`, with substantive reasoning
6. Bgworker auto-fired `parse_gate_response` → `apply_gate_decision`
7. `gate_decisions` row #6 written, `revision_count` 0→1, `status='failed'`
8. Steward retry path picked up the failed item, dispatched chat 1195 to retry the outline stage

The model's critique was real and pointed: "the Hebrew/Greek word study is a category error — D&C 130 was revealed in 1843 English, so the word work should use Webster 1828 for 'principle' and 'intelligence,' not Hebrew binah/daat." Also flagged missing immediate context (April 1843, dinner at the Whites') and lack of differentiation from an existing intelligence study. These are exactly the kind of blind-spot calls the Phase B gate is supposed to catch before drafting starts.

Cost: ~$0.04 for the full e2e (qwen3.6-plus, 6 chat rounds at ~6.4k input avg).

## Surprises

**The model loops through tools before deciding.** Phase B's gate eval reuses the `plan` agent which has the tool catalog wired in. So when given a "should this advance/revise/surface?" prompt, qwen3.6-plus decided to call `study_search_text` to research the existing corpus before evaluating. That's actually good behavior for quality — the critique it produced was sharper because it had checked what was already written. But it 5x'd the cost. Tuning question, not substrate question: should gate-eval prompts disable tools and force a direct JSON return? Probably yes for the binary advance/revise/surface call; probably no for `generate_scenarios` (which benefits from corpus awareness). Note for next session.

**Steward retry surfaced a downstream stage advance.** After the gate said `revise` (status='failed'), the steward picked up the work_item and re-dispatched the outline stage as work_id 1195. That chat completed and the existing Phase 4 work_item_advance machinery moved it to `current_stage='draft'` + `status='awaiting_review'`. This is correct existing behavior in the pipeline progression layer, but it interacts with the gate revise path in a way that's worth thinking about: a `revise` should arguably keep the item ON the same stage, not advance it. The current pattern is: revise → status='failed' → steward retries with potentially-different model → if that retry succeeds, the regular pipeline advances. That's the "model escalation" reading of revise. The alternative (revise → re-dispatch the same stage with feedback prepended, no model change) is closer to what the proposal §V.B described. Worth a follow-up to reconcile.

**The destination_maturity column had to be set via UPDATE.** `work_item_create()` doesn't accept it as a parameter (5a added the column but didn't extend the function signature). New work goes through create-then-update in the API. Could be cleaner with a function-signature change later, but this works.

## Process / covenant

Did 5b before fixing the kind='gate' constraint, so the smoke test caught the bug at first try. Good — that's the inverse-hypothesis pattern (let the failure surface naturally rather than pre-defending against it). The constraint extension ended up as 5c.sql which lets fresh containers get the same definition the live container has.

Used the live-migration pattern for both 5b and 5c: edit the SQL file in the repo, `docker cp` to the container, `psql -f`. Avoids a container restart for each iteration. Phase 5c is in the lib.rs `extension_sql_file!` chain so the next container build gets it natively.

Re-grounded twice mid-session (the 50-tool hook fired). Both times the work was still aligned with the user's "complete all of Phase B" directive. Memory update at session end is per covenant `update_memory` — that's this entry.

## Open

- Tools-disabled gate-eval prompt to cut the per-eval cost ~5×.
- Reconcile the revise path: should it stay on the same stage or hand to the steward retry? The proposal §V.B is more explicit than the current implementation.
- Phase C (Council) — deferred per original ratification, awaits Phase B lived with.
- The `revision_count=0` in gate_decisions row #6 is the count BEFORE the apply incremented it. Worth a column comment.
- Bridge restarted, soak re-enabled at end of session.

## Files touched

Repo:
- `projects/pg-ai-stewards/extension/5b-scenarios-verify.sql` (new)
- `projects/pg-ai-stewards/extension/5c-sessions-gate-kind.sql` (new)
- `projects/pg-ai-stewards/extension/test-gate-e2e.sql` (new — repeatable smoke)
- `projects/pg-ai-stewards/extension/Dockerfile` (COPY adds 5b + 5c)
- `projects/pg-ai-stewards/extension/src/bgworker.rs` (auto-fire path, ~130 lines)
- `projects/pg-ai-stewards/extension/src/lib.rs` (registers 5b + 5c)
- `scripts/stewards-ui/api/new_work.go` (DestinationMaturity field + UPDATE)
- `scripts/stewards-ui/api/work_items.go` (5 new fields in detail + new /api/work-items/gate-decisions handler)
- `scripts/stewards-ui/frontend/src/api.ts` (WorkItemDetail extensions + GateDecision types + workItemGateDecisions wrapper)
- `scripts/stewards-ui/frontend/src/views/NewWork.vue` (destination maturity dropdown)
- `scripts/stewards-ui/frontend/src/views/WorkItemDetail.vue` (4 new conditional panels)

Live containers:
- `pg-ai-stewards-dev`: 5b + 5c live-applied.
- `pg-ai-stewards-ui`: rebuilt + restarted with new image.
- `pg-ai-stewards-bridge`: restarted at end of session.
- Soak: schedule_enabled=true at end of session.
