# Proposal: Brain Project UX Improvements

*Created: 2026-04-14*
*Status: Phase 1-2 Shipped (2026-04-14)*
*Type: Quick plan — 4 frontend/backend improvements*
*Research: [.spec/scratch/brain-project-ux/main.md](../scratch/brain-project-ux/main.md)*

---

## Binding Problem

Real usage of the project + commission flow revealed 4 UX gaps: the create project form is missing workspace fields (requiring a second edit), dialog positioning/text-color bugs make dialogs hard to use, the commission-on-create flow requires two sequential dialogs, and there's no visibility into which models and phases consumed premium credits during a commission.

---

## Phase 1: Create Project — Full Fields (Frontend only)

**Problem:** Creating a project only captures name/description/emoji/init-instructions. Workspace type, workspace path, GitHub repo, and repo visibility are only available in the edit form — forcing users to create → immediately edit.

**Fix:** Add workspace fields to the create form in `ProjectsView.vue`, matching the edit form's conditional field logic from `ProjectDetailView.vue`.

**Fields to add:**
- `workspace_type` — select: integrated (default), subfolder, external
- `workspace_path` — conditional on type ≠ integrated
- `github_repo` — conditional on type = external
- `repo_visibility` — conditional on type = external (default: private)

**Not adding to create:** `status` (defaults to "active") and `context_file` (set during init or edit). These are edit-only concerns.

**API already supports it:** `api.createProject()` accepts all these fields. Backend `CreateProject` handler already processes them. Pure frontend change.

**AI initialization uses Sonnet 4.6** (`config.PipelineSmartModel = "claude-sonnet-4.6"`). This is already correct — no change needed. After creating a project with workspace fields, the "Initialize" button on the project detail page triggers agent-driven scaffolding via Sonnet.

**Verification:**
- [ ] Create project with "External" workspace type → workspace_path, github_repo, repo_visibility fields appear
- [ ] Create project with "Subfolder" type → workspace_path field appears, github fields hidden
- [ ] Create project with "Integrated" type → no workspace fields shown
- [ ] Created project shows correct workspace fields on edit form
- [ ] `vue-tsc --noEmit` passes

---

## Phase 2: Dialog Styling Fixes — Text Color + Centering

**Problem:** Both the New Entry dialog and Commission dialog have:
1. **Black text on dark background** — the `<dialog>` element has browser-default text color (black), making text in textareas and labels invisible against `bg-gray-950`
2. **Potential centering issues** — `<dialog>` has browser-default `position: absolute; margin: auto` that can fight with `position: fixed; flex`

**Root cause:** HTML `<dialog>` elements have a UA stylesheet that sets `color: CanvasText` (typically black) and `position: absolute`. Our CSS uses `fixed inset-0 flex items-center justify-center` on the dialog, but the UA styles may partially override.

**Fix for both dialogs (`ProjectDetailView.vue` New Entry dialog + `CommissionDialog.vue`):**
1. Add `text-gray-100` to the `<dialog>` element (override UA text color)
2. Add `m-0` to the `<dialog>` element (override UA `margin: auto`)
3. Add `w-full h-full` to ensure the dialog fills viewport for flex centering
4. Alternatively: replace `<dialog>` with a plain `<div role="dialog">` to avoid UA stylesheet fights entirely

**Preferred approach:** Replace `<dialog :open>` with `<div v-if class="fixed inset-0 z-40 flex items-center justify-center"` + `role="dialog"` + `aria-modal="true"`. This avoids all UA stylesheet conflicts and is the pattern most Vue apps use. The `<dialog>` element's benefits (focustrap, ::backdrop) aren't being used since we're manually implementing the backdrop anyway.

**Verification:**
- [ ] New Entry dialog: centered on screen, white/light text in textarea, readable labels
- [ ] Commission dialog: centered on screen, white/light text in intent textarea, readable labels
- [ ] Both dialogs close on backdrop click
- [ ] Both dialogs close on Escape key
- [ ] `vue-tsc --noEmit` passes

---

## Phase 3: Inline Commission Fields in New Entry Dialog

**Problem:** When "📜 Commission steward immediately" is checked and the entry is created, a *second* dialog (CommissionDialog) opens for the user to fill in intent + options. This is jarring — two dialogs in sequence.

**Fix:** When the checkbox is checked, expand commission fields inline within the New Entry dialog:
- `commissionIntent` textarea — "What should the steward accomplish?" (required when checked)
- Collapsible "Advanced options" (same as CommissionDialog): authority, model, budget
- **Pre-fill `commissionIntent`** from the entry body text (since the entry description is often the commission intent too)

**On submit:** Create entry → set project → create commission — all in one action. If commission creation fails, the entry still exists (created first) but show an error toast.

**Updated flow:**
```
[x] Commission steward immediately

  What should the steward accomplish?
  [textarea — pre-filled from entry body]

  ▸ Advanced options
    Authority: [Advance & Execute ▼]
    Model:     [Claude Opus 4.6 ▼]
    Budget:    [50]

[Cancel] [Create & Commission]
```

Button text changes from "Create" to "Create & Commission" when checkbox is checked.

**Verification:**
- [ ] Unchecked: dialog works as before (just entry creation)
- [ ] Checked: commission fields appear inline, intent pre-filled from body
- [ ] Submit creates entry + commission in one flow
- [ ] Commission appears in project commission list after creation
- [ ] Error in commission creation still saves the entry, shows toast
- [ ] CommissionDialog still works independently from EntryDetailView (don't break the standalone dialog)

---

## Phase 4: Commission Cost Breakdown

**Problem:** Commission tracking shows total `cost_used` but no breakdown by model, phase, or cost type. The `commission_decisions` table records `stage` and `cost` per decision but not which model was used or whether the cost was pipeline work vs steward evaluation.

**Current schema:**
```sql
commission_decisions (
  id, commission_id, timestamp, entry_id, 
  stage,      -- "research", "plan", "spec", "execute", "verify"
  action,     -- "advance", "revise", "fail", etc.
  reasoning,  -- text
  cost        -- float (premium requests for this decision)
)
```

### Phase 4a: Backend — Add model + type to decisions

**Schema migration:** Add two columns to `commission_decisions`:
```sql
ALTER TABLE commission_decisions ADD COLUMN model TEXT NOT NULL DEFAULT '';
ALTER TABLE commission_decisions ADD COLUMN cost_type TEXT NOT NULL DEFAULT 'pipeline';
```

**cost_type values:**
- `pipeline` — actual pipeline work (research, plan, execute)
- `eval` — steward gate evaluation 
- `retry` — retried operation
- `verify` — scenario verification pass

**Update `recordDecision()`:** Accept model and cost_type parameters. Thread through from all call sites in `commission.go`.

**Update `CommissionDecision` struct:** Add `Model string` and `CostType string` fields.

### Phase 4b: Frontend — Cost breakdown display

**Where:** Commission status panel (currently shown in `EntryDetailView.vue` and `CommissionDialog.vue` on commissioned entries).

**Display:** A collapsible cost breakdown section showing:
```
💰 Cost: 39.0 / 50.0 premium requests

  By Phase:
    research   3.3  (1× Opus eval + 1× Haiku pipeline)
    plan       6.0  (1× Opus eval + 1× Sonnet pipeline)
    spec       6.0  (1× Opus eval + 1× Opus scenarios)  
    execute    9.0  (1× Opus eval + 1× Opus execute)
    verify     9.0  (1× Opus eval + 1× Opus verify)
    
  By Model:
    Claude Opus 4.6    33.0  (11 calls × 3.0)
    Claude Sonnet 4     3.0  (3 calls × 1.0)
    Claude Haiku 4.5    3.0  (9 calls × 0.33)

  Decision Log (12 decisions):
    [timestamp] research  advance  3.33  haiku  pipeline
    [timestamp] research  advance  3.00  opus   eval
    ...
```

**Computed from existing `decisions` array** — just needs model + cost_type fields from 4a.

### Phase 4c: Steward overhead tracking (if needed)

Currently, all steward costs (gate evaluation, retries, verification) are tracked *within* the commission's `cost_used` and decision log. If the steward is restarted, it resumes from where it left off — the existing decisions remain. New decisions after restart are just additional rows.

**No separate "steward overhead" tracking needed** — the `cost_type` field distinguishes pipeline work from steward evaluation. A restart doesn't lose data since decisions are persisted per-step.

**Verification:**
- [ ] New columns created via migration, existing data unaffected (defaults to '' and 'pipeline')
- [ ] All `recordDecision` call sites pass model and cost_type
- [ ] Commission decisions API includes model and cost_type in response
- [ ] Frontend shows cost breakdown by phase and by model
- [ ] Decision log shows model + type per decision
- [ ] All Go tests pass
- [ ] `vue-tsc --noEmit` passes

---

## Execution Summary

| Phase | Scope | Effort | Dependencies |
|-------|-------|--------|-------------|
| 1 | Frontend only (ProjectsView.vue) | ~30 min | None |
| 2 | Frontend only (2 files) | ~20 min | None |
| 3 | Frontend only (ProjectDetailView.vue) | ~45 min | Phase 2 (styling must be right first) |
| 4a | Backend (migration + commission.go + types) | ~30 min | None |
| 4b | Frontend (EntryDetailView.vue) | ~45 min | Phase 4a |

**Phases 1, 2, and 4a are independent** — can be built in any order.
Phase 3 depends on Phase 2 (dialog styling). Phase 4b depends on 4a (schema).

**Recommended build order:** 2 → 1 → 3 → 4a → 4b

---

## Recommendation

**Build.** All four items are real friction points discovered through actual usage. Small scope, high impact. Phases 1-3 are pure frontend fixes. Phase 4 is a modest backend change + frontend display. Total: roughly 3 hours of work across 2-3 sessions.
