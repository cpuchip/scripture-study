# Scratch: Brain Project UX Improvements

*Created: 2026-04-14*

## Binding Problem

The project creation and commission flows have UX gaps exposed by real usage: missing fields on create, styling bugs in dialogs, and no cost breakdown visibility for commissions.

---

## Inventory Findings

### 1. Create Project Dialog (ProjectsView.vue)

**Current fields:** emoji, name, description, init_instructions
**Missing vs Edit form (ProjectDetailView.vue lines 529-615):**
- `status` — select (active/paused/archived) — probably fine to default to "active" on create
- `context_file` — text input for a custom context file path
- `workspace_type` — select: integrated / subfolder / external
- `workspace_path` — conditional on type (subfolder path or external path)
- `github_repo` — only if type === "external"
- `repo_visibility` — only if type === "external" (private/public)

**Create payload already supports these fields** in api.ts:
```typescript
createProject(data: { name: string; description?: string; emoji?: string; 
  workspace_type?: string; workspace_path?: string; github_repo?: string; 
  repo_visibility?: string; init_instructions?: string })
```

**Create form is inline** (slides open in page, not a dialog), edit is also inline on the project detail page. Both use the same color scheme (bg-gray-900/950).

**AI initialization:** After project creation, there's an "Initialize" button on ProjectDetailView that calls `POST /api/projects/{id}/scaffold`. This uses the Copilot SDK agent. Current model for initialization: need to check.

### 2. New Entry Dialog (ProjectDetailView.vue lines 1272-1309)

**Positioning:** Uses `<dialog>` with `fixed inset-0 z-40 flex items-center justify-center` — this SHOULD center it.
**Actual issue from screenshot:** The dialog appears LEFT-aligned, NOT centered. The `<dialog>` element has default browser styles that may override the flex centering. The text "to fix font color" appears as BLACK text on dark blue background — very hard to read.

**Root cause of text color:** The `<textarea>` and other inputs use `bg-gray-950 border-gray-700` but the `<dialog>` element itself may inherit default browser text color (black). The outer `<dialog>` doesn't have explicit `text-white` or `text-gray-100`.

**Root cause of positioning:** The `<dialog>` element in HTML has default styles (`position: absolute`, `margin: auto`) that fight with `position: fixed`. The `items-center justify-center` on the dialog itself may not work because the dialog's `open` attribute causes default positioning.

### 3. Commission Dialog (CommissionDialog.vue)

**Same structural issue as New Entry:** Uses `<dialog :open>` with `fixed inset-0 z-40 flex items-center justify-center`. Will have the same centering and text color issues.

**From screenshot (image 4):** Commission intent textarea shows BLACK text on dark background — same bug.

**Both dialogs need:** 
- Add `text-gray-100` to the outer `<dialog>` element
- Fix positioning: likely need to NOT use `<dialog>` as the positioning container (use a wrapper `<div>` instead), or override dialog defaults

### 4. Merge Commission Fields into New Entry Dialog

**Current flow when checkbox is checked:**
1. Entry is created  
2. CommissionDialog opens automatically  
3. User fills commission intent + advanced options  
4. Commission is created

**Proposed flow:** 
1. When checkbox is checked, commission fields expand inline in the New Entry dialog
2. Both entry + commission are created in one submit
3. No second dialog

**Commission fields to inline:**
- intent (textarea) — required when commission is checked
- Advanced options (collapsible): authority, model, budget

### 5. Commission Cost Breakdown

**Current state:**
- `commission_decisions` table has `stage`, `action`, `cost` per decision
- But no `model` field per decision — it only uses the commission's model
- The decisions DO have cost per stage (research=0.33 from Haiku, gate eval=3.0 from Opus, etc.)  
- But the per-decision model is inferred from the commission's model, not recorded explicitly
- Steward restart costs: nothing tracks steward overhead separately from entry costs. `addCommissionCost` tracks pipeline costs. If the steward retries, those retries add to the commission cost via the same decisions table.

**What's missing for a good cost breakdown UI:**
- `model` column on `commission_decisions` — which model was used for this specific call
- Per-phase subtotals (sum decisions by stage)
- Distinguish between pipeline costs (the actual work) and steward overhead (gate evaluation, retries, verification)
- `type` or `category` on decisions: "pipeline" vs "steward_eval" vs "retry" vs "verification"

**Current CommissionDecision schema:**
```sql
commission_id, timestamp, entry_id, stage, action, reasoning, cost
```

**Needed additions:**
```sql
model TEXT NOT NULL DEFAULT ''   -- which model was used
type  TEXT NOT NULL DEFAULT 'pipeline'  -- pipeline|eval|retry|verify
```

---

## Summary of Issues

| # | Issue | Scope | Effort |
|---|-------|-------|--------|
| 1 | Create project missing workspace fields | Frontend only | Small |
| 2 | New Entry dialog: not centered, black text | Frontend CSS fix | Small |
| 3 | Commission dialog: same CSS issues + inline in entry | Frontend CSS + restructure | Medium |
| 4 | Commission cost breakdown (model + type per decision) | Backend schema + frontend | Medium |
