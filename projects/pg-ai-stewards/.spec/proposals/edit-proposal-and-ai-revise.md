---
title: Edit + AI-revise for proposed work_items
date: 2026-05-12
status: ratified this session (3 questions); build-ready
parent: substrate-h3-followup-small-items-and-sc-pivot.md
purpose: >
  Close the missing third mode for proposed work_items. Today the
  Ratify / Dispatch / Cancel surface assumes the proposal is either
  good-to-run or worth-cancelling. There's no path for "this is
  close but needs a tweak" — which is the most common case in
  practice. This proposal adds two paths: direct field editing for
  fast tweaks (no AI cost) and AI revise-with-feedback for "tell the
  agent what to change and let it re-propose."
---

# Edit + AI-revise for proposed work_items

## Why this proposal

After H.3-followup landed, the 5 SC AI-literacy work_items appeared
in the UI with Ratify / Dispatch / Cancel buttons. Michael surfaced
the gap immediately: *"How do I update or edit a proposed work? If I
need to tweak it? do I issue a new work item?"*

The current model assumes proposals are either correct (ratify) or
wrong (cancel). In practice, the most common state is **close** —
the binding_question phrasing is off, the scope is too broad, the
pipeline_family_hint should be different. Issuing a new work_item
for these tweaks throws away the agent's rationale + parent linkage
+ project context.

This proposal closes that gap with two distinct affordances and one
new substrate primitive.

## Ratifications (this session)

### Q1 — Revise capability shape
**Ratified:** Both direct edit fields + AI-revise textarea (Recommended).

Two affordances on the existing proposal panel:
- **Direct edit** — inline form to edit `binding_question`,
  `slug`, `pipeline_family_hint` (resolved to pipeline_family),
  `project_association`. No AI involved; just an UPDATE. Fast,
  zero cost.
- **Revise with feedback** — textarea + button. User writes what
  should change. Agent reads the original + parent + feedback and
  emits a revised proposal. Shown to user as a diff for accept/reject.

Full chat panel is deferred to the larger substrate-aware-chat
workstream.

### Q2 — AI-revise substrate mechanism
**Ratified:** New `revise-proposal` pipeline (1 stage, full audit).

A new pipeline family `revise-proposal` with one stage `revise`:
- Reads `parent_work_item_id` (the proposal being revised)
- Reads the parent's parent (the original planning run for context)
- Reads `input.feedback` (user's text)
- Emits strict JSON revision: `{binding_question, rationale,
  slug?, pipeline_family_hint?, project_association?}`
- Tools off; structured output; qwen3.6-plus model (cheap)
- Cost cap default: $0.10 per revise (much smaller than full
  planning pipeline)

Pro of pipeline approach: every revise is an auditable work_item
with cost tracking, retry-with-lessons machinery, sabbath
opportunity. Each revise becomes substrate-visible history.

Con: another pipeline to maintain (mitigated — small, focused).

### Q3 — Revision result handling
**Ratified:** Side-by-side diff with Accept/Reject (Recommended).

When the revise pipeline completes:
- UI fetches the revise work_item's `stage_results.revise.output`
- Renders side-by-side diff of original vs proposed revision
- Two buttons: **Accept** (UPDATE original with revision fields)
  and **Reject** (mark the revise work_item cancelled with reason)
- Multiple pending revisions for the same proposal stack as
  separate diff cards (uncommon, but possible)

Lower commitment than auto-apply; user stays in control.

## Architecture sketch

```
Original proposal (origin=agent_planning, maturity=raw)
  │
  └─ parent_work_item_id ←──┐
                            │
User clicks "Revise with feedback"
  │
  ▼
New work_item created (pipeline=revise-proposal, origin=human)
  • input.feedback = user text
  • parent_work_item_id = original proposal id   ──┐
  • status = in_progress                            │
  │                                                 │
  ▼                                                 │
revise stage runs:                                  │
  • input_template reads original (via parent FK)   │
  • + parent planning context                       │
  • + user feedback                                 │
  • emits JSON revision                             │
  │                                                 │
  ▼                                                 │
review-less; goes straight to maturity=verified     │
status='completed'                                  │
                                                    │
UI polls /api/work-items/{id}/pending-revisions     │
  Returns: list of completed revise-proposal items  │
  where parent_work_item_id = id ◄─────────────────┘
  AND no applied_at / cancelled_at
  │
  ▼
Diff card rendered. User clicks Accept.
  │
  ▼
apply_revision(revise_id) SQL function:
  • Reads stage_results.revise.output
  • UPDATEs original work_item.{binding_question, rationale,
    slug, pipeline_family, project_association}
  • UPDATEs revise work_item with applied_at = now()
  │
  ▼
Original proposal reloads. New values visible. Ratify button still
on it. User can iterate again if needed.
```

## Substrate primitives needed

1. **New pipeline family `revise-proposal`** (1 stage, qwen3.6-plus,
   cost_cap_default_micro=100000)
2. **New SQL function `stewards.apply_revision(uuid)`** — reads
   revise output, validates JSON, UPDATEs original
3. **New column on work_items: `revision_applied_at timestamp`** —
   tracks whether a revise has been accepted (NULL = pending,
   set = accepted)
4. **Optional: `revision_rejected_at timestamp`** — distinguish
   rejected from pending. Or just use status='cancelled' + a
   quarantine_reason.

Decision (stewardship): use `status='cancelled'` for rejected
revises (existing column), add `revision_applied_at` for accepted
ones (new column). Avoids two new columns when one + existing works.

## Backend endpoints

| Endpoint | Action |
|---|---|
| `POST /api/work-items/edit-proposal` | Direct UPDATE of binding_question/slug/pipeline_family/project_association. Restricted to origin=agent_planning AND status≠cancelled. |
| `POST /api/work-items/revise-with-feedback` | Create + dispatch a revise-proposal work_item. Body: `{id, feedback}` |
| `GET /api/work-items/{id}/pending-revisions` | List completed revise-proposal work_items where parent_work_item_id=id AND revision_applied_at IS NULL AND status≠cancelled |
| `POST /api/work-items/apply-revision` | Call apply_revision(revise_id); UPDATE original |
| `POST /api/work-items/reject-revision` | UPDATE revise work_item status=cancelled |

## UI on WorkItemDetail.vue (extends the existing Proposed-work panel)

Three new sub-panels under the existing Ratify/Dispatch/Cancel row:

1. **Edit fields** (collapsible) — inline form with current values
   pre-filled, "Save" button calls edit-proposal endpoint.

2. **Revise with feedback** (collapsible) — textarea + "Revise"
   button. On submit: status spinner; poll pending-revisions every
   3-5s.

3. **Pending revisions** (auto-shown when present) — one diff card
   per pending revision. Side-by-side or above/below; Accept/Reject
   buttons.

## Non-goals (deferred)

- **Multi-turn chat in the work_item context.** Use the larger
  substrate-aware-chat workstream when it lands.
- **Multi-version revision history.** Each revise is independent;
  accepting one auto-rejects others isn't built. If queue fills,
  revisit.
- **Revise on already-running work_items.** Only proposals
  (status=pending, origin=agent_planning) are editable. Cancelled
  / dispatched / completed are immutable.
- **Direct edit on non-proposal work_items.** Edit affordances
  only show for origin=agent_planning to prevent the buttons
  becoming general field-edit levers.

## Build plan

1. SQL: revise-proposal pipeline + apply_revision function + revision_applied_at column + smoke
2. Backend: 5 endpoints + tests
3. Frontend: edit form + revise textarea + pending-revisions polling + diff card
4. E2E: edit one of the SC work_items; revise another with feedback; accept; reject
5. Commit at each phase; journal + summary at end.

## Cost note

Each revise is ~$0.02-0.05 (qwen3.6-plus, ~1500-2500 input tokens
typical, small output). Cap at $0.10 per revise per safety. If a
user wants 5 revisions before ratifying, $0.50 is acceptable.
