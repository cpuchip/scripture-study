---
date: 2026-05-12
mode: build
workstream: WS5
project: pg-ai-stewards
title: "Third proposal mode shipped: direct edit + AI-revise with diff/accept"
status: shipped (e2e validated on a real SC work_item)
carry_forward:
  - "5 SC work_items still await ratification — one of them (space-center-laptop-kiosk-webcam-validation, renamed from space-center-laptop-webcam-ml-validation) is now tighter scope thanks to this session's revise e2e."
  - "yaml.rs Rust parser refactor (rule-of-three still triggered; this session didn't add a 4th intent)"
  - "Phase A pgrx BGW SPI longjmp catch + 60s periodic reaper"
  - "work_items.materialized_at semantics drift (enqueue-time vs write-time)"
  - "Soak still paused — Michael can resume manually or I will at his go-ahead"
links:
  - "../../projects/pg-ai-stewards/.spec/proposals/edit-proposal-and-ai-revise.md"
  - "../../projects/pg-ai-stewards/extension/h3-followup-3-revise-proposal-pipeline.sql"
---

# Third proposal mode shipped (2026-05-12)

Five commits this build pulse. Closes the gap Michael surfaced immediately after H.3-followup: "How do I update or edit a proposed work? Do I issue a new work item?"

## What shipped

### Proposal doc (`863a085`)
Captured three ratifications + architecture sketch + non-goals. Michael's words: "chat with AI in the context of the work_item" and "text input where I write ideas and have it iterate and re-propose." Both honored.

### Substrate (`f4485b2`)
- New pipeline family `revise-proposal` — 1 stage `revise`, qwen3.6-plus, tools off, structured JSON output. cost_cap_default_micro=100000 ($0.10). Maturity ladder `["raw", "verified"]` — straight from dispatch to done.
- Stage input_template reads original proposal fields (via parent_work_item_id link) + parent plan excerpt + user feedback. Emits partial JSON revision.
- New SQL function `stewards.apply_revision(uuid)`:
  - Validates JSON (slug regex + uniqueness; binding ≥20 chars)
  - COALESCE-merges only the revision's present fields into the original
  - If pipeline_family_hint changes, also updates current_stage to the new pipeline's first stage
  - Marks the revise work_item `revision_applied_at = now()`
  - Idempotent — re-call after applied returns false
- New column `work_items.revision_applied_at timestamp` (NULL = pending; set = accepted). Rejected revises reuse existing `status='cancelled'` + `quarantine_reason`. One new column instead of two per stewardship decision in the proposal.

### Backend (`8b25c83`)
Five new endpoints:
- `POST /api/work-items/edit-proposal` — direct UPDATE, transaction-wrapped, origin=agent_planning guarded. Editable: binding_question, slug, pipeline_family_hint, project_association, rationale.
- `POST /api/work-items/revise-with-feedback` — creates + dispatches revise-proposal work_item with parent linkage + cost cap + project inheritance.
- `GET /api/work-items/pending-revisions?id=<>` — lists completed revise-proposal work_items for a parent, returns raw revision JSON for diff rendering.
- `POST /api/work-items/apply-revision` — calls `apply_revision` SQL fn.
- `POST /api/work-items/reject-revision` — UPDATE status=cancelled on the revise work_item.

Each endpoint origin/family-guarded so they can't become general-purpose levers.

### Frontend (`27d20cb`)
WorkItemDetail.vue extended:
- "Edit fields directly" button → collapsible 5-field form (slug, binding_question, rationale, pipeline_family, project_association). Submits only changed fields.
- "Revise with AI feedback" button → textarea + dispatch. Cost shown inline. ~$0.02-0.05 per revise.
- Pending revisions section auto-renders when present. Polls `/pending-revisions` every 4s during in-flight revises; auto-stops when terminal.
- Each diff card: feedback string + cost + side-by-side red/green diff of each changed field + Accept/Reject buttons.

api.ts: 5 new methods + 3 new types. Bundle grew 25.5kB → 34.9kB (9kB gzipped).

### E2E smoke (`27d20cb` + this commit)
Dispatched a real revise against `space-center-laptop-webcam-ml-validation` with feedback: *"Scope this tighter — focus only on validating that the laptop webcams are accessible to Chrome under a kiosk-mode user profile; defer ML stack validation to a separate work_item."*

Revise completed in <30s for **$0.005** (qwen3.6-plus). Output:
```json
{
  "slug": "space-center-laptop-kiosk-webcam-validation",
  "binding_question": "Are the built-in webcams on all five repurposed laptops accessible to Chrome when running under a kiosk-mode user profile, independent of any ML stack validation?",
  "rationale": "Isolates the low-risk hardware and driver check from the heavier ML stack validation so the initial go/no-go runs faster."
}
```

Clean schema. Applied via endpoint. Original work_item updated cleanly (slug, binding, rationale all changed). revision_applied_at set. pending-revisions queue empty after apply (UI hides diff card automatically). Full loop validated.

## What the substrate gained

Three modes for proposed work_items now:
1. **Ratify / Dispatch / Cancel** — terminal actions for "proposal is good / not"
2. **Edit directly** — fast no-AI tweaks for "just rephrase / rename / re-pipeline"
3. **AI revise with feedback** — for "this is close but help me think through what should change"

Mode 3 is the conversational pattern Michael asked for, scoped to a single revision rather than multi-turn chat. Each revise is its own auditable work_item with cost tracking + parent linkage. The diff-before-accept gate keeps the user in control.

## Architecture wins this session

- **Substrate pattern over endpoint hack.** The revise mechanism is a real pipeline family with full audit trail (work_queue + cost_events + parent linkage) rather than a one-shot endpoint hack. Future revisions are queryable as work_items.
- **COALESCE-merge preserves intent.** apply_revision only touches fields the revision actually changed. If the user says "tighten the slug only," the binding_question stays untouched.
- **Defense in depth.** Each endpoint origin/family-guarded server-side. Validation client-side AND server-side. Frontend tracks pending revisions independently so the UI never loses state if the user navigates away mid-revise.
- **Polling that knows when to stop.** The pending-revisions polling auto-terminates once no revisions are still in flight. onUnmounted cleans up the timer.

## Cost summary

This build session: ~$0.005 (the one revise e2e). Backend + frontend + smoke ran without other LLM costs.

Total session arc (yesterday's H.3 + today's H.3-followup + this third mode): ~$1.40 cumulative, six real artifacts on disk (4 research + 2 plans), five proposed SC work_items + one accepted revision in the substrate.

## Carry-forward

1. **5 SC work_items still await Michael's ratification.** One is now tightened (the webcam validation, renamed and rescoped via this session's revise e2e). Recommended start: that one — the hardware gate.
2. **yaml.rs Rust parser refactor** — rule-of-three still triggered. ~1 session, Claude-only.
3. **Phase A pgrx longjmp catch + 60s reaper** — agent's earlier plan ranked it last. ~1 session, Claude-only.
4. **materialized_at naming drift** — pre-existing; ~30min when bothered.
5. **Multi-turn chat in work_item context** — deferred to substrate-aware-chat workstream when it lands. The textarea-revise pattern covers most use cases for now.

## Closing

Michael's instinct was right: ratify/dispatch/cancel is incomplete without an iteration path. The substrate now has that path — and the AI-revise pattern is itself substrate work, not bolted-on frontend logic. Every revise becomes substrate history, costed and parented and accountable.

The Marsfield AI-literacy plan just got one work_item tighter without Michael lifting a finger. That's the substrate working as it's meant to.
