# Commission UI

**Status:** Phase 1-2 Shipped
**Created:** 2026-04-11
**Binding problem:** The steward commission backend is fully wired (6 API endpoints, goroutine state machine, gate evaluation, budget tracking, pause/resume/revoke), but the frontend has zero commission support. Michael can't commission a steward without manual curl calls. The delegation mechanism itself requires API expertise to invoke — which defeats the point of delegation.

**Research:** [.spec/scratch/commission-ui/main.md](../scratch/commission-ui/main.md)

---

## Success Criteria

1. Michael can create a new entry tied to a project and commission a steward in a single flow
2. Michael can commission any existing entry from its detail view or the project board
3. Commission defaults (model, authority, budget) are sensible — most commissions should require only an intent sentence and a click
4. Active commissions are visible on entry cards (badge) and entry detail (status + controls)
5. Michael can pause, resume, and revoke commissions from the UI
6. Commission decisions (the steward's judgment log) are visible

---

## Constraints

- Follow existing dialog pattern: `<Teleport to="body"><dialog :open>` with backdrop, `bg-gray-900 border-gray-700 rounded-xl p-6`, max-width
- Follow existing toast pattern: local `ref()` + `setTimeout(3000)`, fixed top-right
- Follow existing button patterns: colored primary action + gray Cancel, loading flag disables during submission
- All new API functions go in `api.ts` following existing patterns (`fetch` + JSON)
- TypeScript strict — new interfaces for Commission and CommissionDecision
- No new npm dependencies
- No changes to Go backend (all 6 endpoints already exist)

---

## Phase 1: Foundation (Types + API + Commission Dialog)

**Deliverable:** Reusable commission dialog component and TypeScript layer.

### 1a. TypeScript interfaces in `api.ts`

```typescript
interface Commission {
  id: string
  entry_id: string
  project_id: number | null
  intent: string
  scope: string
  authority: string
  model: string
  max_cost: number
  cost_used: number
  status: string // "active" | "paused" | "completed" | "revoked" | "failed"
  started_at: string
  expires_at: string
  created_at: string
  decisions: CommissionDecision[]
}

interface CommissionDecision {
  id: number
  commission_id: string
  timestamp: string
  entry_id: string
  stage: string
  action: string
  reasoning: string
  cost: number
}
```

### 1b. API functions in `api.ts`

```typescript
createCommission(entryId: string, intent: string, authority?: string, model?: string, maxCost?: number): Promise<Commission>
listCommissions(): Promise<Commission[]>
getCommission(id: string): Promise<Commission>
pauseCommission(id: string): Promise<void>
resumeCommission(id: string): Promise<void>
revokeCommission(id: string): Promise<void>
```

POST body for create: `{ entry_id, intent, authority, model, max_cost }`. Only `entry_id` is required by backend; defaults apply for rest.

### 1c. CommissionDialog component

**Location:** `src/components/CommissionDialog.vue`

**Props:**
- `open: boolean` — controls visibility
- `entryId: string` — which entry to commission
- `entryTitle: string` — display in dialog header

**Emits:**
- `close` — user cancelled or dialog should close
- `commissioned(commission: Commission)` — commission created successfully

**Layout — progressive disclosure:**

```
┌─────────────────────────────────────────┐
│  📜 Commission Steward                  │
│  for: "LCARS Dashboard Display"         │
│                                         │
│  What should the steward accomplish?    │
│  ┌─────────────────────────────────┐    │
│  │ Build the clock display with... │    │
│  └─────────────────────────────────┘    │
│                                         │
│  ▸ Advanced options                     │
│  ┌─────────────────────────────────┐    │
│  │ Authority: [Advance & Execute ▾]│    │
│  │ Model:     [claude-opus-4.6   ▾]│    │
│  │ Budget:    [50] premium requests│    │
│  └─────────────────────────────────┘    │
│                                         │
│  [Commission]              [Cancel]     │
└─────────────────────────────────────────┘
```

- **Intent textarea** — always visible, required. Placeholder: "What should the steward accomplish?"
- **Advanced options** — collapsed by default (click to expand). Contains:
  - **Authority** — `<select>` with two options: "Advance & Execute" (default, value `advance_and_execute`) and "Advance Only" (value `advance_only`). Advance Only means the steward researches/plans/specs but stops before execution.
  - **Model** — `<select>` with options: `claude-opus-4.6` (default), `claude-sonnet-4` (cheaper, less capable). Values match backend model strings.
  - **Budget** — `<input type="number">` defaulting to 50. Label: "premium requests".
- **Commission button** — primary action (amber/gold), disabled while submitting
- **Cancel button** — gray, closes dialog

**Behavior:**
1. User fills intent (required)
2. Optionally adjusts advanced settings
3. Clicks "Commission" → calls `api.createCommission()` 
4. On success: emits `commissioned`, shows toast, closes
5. On error: shows error toast, keeps dialog open

### 1d. Verification

- [x] TypeScript compiles with no errors
- [ ] Dialog opens, submits, and closes correctly
- [ ] Commission appears in backend (verify via `GET /api/commissions`)
- [ ] Toast shows on success/failure

**Status: SHIPPED Apr 11.** Types, API functions, and CommissionDialog component all implemented.

---

## Phase 2: Commission from Entry Detail + Project Board

**Deliverable:** Commission buttons on entry detail view and project board entry cards.

### 2a. EntryDetailView — commission button

Add a "📜 Commission" button in the pipeline gates section, visible when:
- Entry is NOT a notebook entry
- Entry does NOT have an active commission already
- Entry maturity is `raw`, `researched`, `planned`, or `specced` (not already executing/verified/complete)

The button opens the CommissionDialog with the entry's ID and title.

When commission is created: show a status section replacing the pipeline gates:

```
┌─────────────────────────────────────────┐
│  📜 Commission Active                   │
│  Intent: "Build the clock display..."   │
│  Authority: Advance & Execute           │
│  Budget: 2.3 / 50 premium requests      │
│  Status: active                         │
│                                         │
│  [⏸ Pause]  [⏹ Revoke]                 │
│                                         │
│  Decision Log                           │
│  ├ research → advance (0.33) "Entry..." │
│  ├ plan → advance (1.0) "Spec meets..." │
│  └ execute → advance (1.0) "Code..."    │
└─────────────────────────────────────────┘
```

**Commission status indicators:**
- `active` — green badge, pause + revoke buttons
- `paused` — amber badge, resume + revoke buttons
- `completed` — gray badge, no controls
- `revoked` — red badge, no controls
- `failed` — red badge, no controls

**Loading commission data:** On mount (or when entry loads), call `api.listCommissions()` and find the one matching `entry_id`. Or add a query param filter to the backend — but the simpler approach is to fetch all and filter client-side (commission count will be low).

### 2b. ProjectDetailView — commission button on cards

Add "📜 Commission" to the entry card action buttons (alongside existing Advance/Revise/Execute/Verify/Cancel/Complete). Show when the same conditions as 2a apply.

Opens CommissionDialog with the card's entry ID and title.

### 2c. Commission badge on entry cards

When an entry has an active or paused commission, show a small badge on the card:
- `📜` icon with status color (green for active, amber for paused)
- Prevents accidentally starting manual pipeline actions on commissioned entries

**Guard:** When a commission is active, hide the manual Advance/Execute/Revise buttons on that entry. The steward is in charge. Show "Steward commissioned — [Revoke] to take manual control" instead.

### 2d. Verification

- [x] Commission button appears on qualifying entries
- [x] Commission dialog works from both ProjectDetailView and EntryDetailView
- [x] Active commission shows status, budget, controls
- [x] Decision log displays correctly
- [x] Manual pipeline buttons hidden when commission is active
- [ ] Pause/resume/revoke buttons work (needs live testing)

**Status: SHIPPED Apr 11.** EntryDetailView: commission status panel with pause/resume/revoke + decision log, commission button in pipeline gates. ProjectDetailView: commission badge on board cards + list items, commission button, guard hiding manual buttons when commissioned.

---

## Phase 3: Create Entry in Project

**Deliverable:** "New Entry" button on ProjectDetailView that creates an entry already assigned to the project.

### 3a. "New Entry" button on project board

Add a `+ New Entry` button at the top of the Inbox column (or in the project header). Opens a dialog:

```
┌─────────────────────────────────────────┐
│  + New Entry                            │
│  Project: Space Center                  │
│                                         │
│  What needs to happen?                  │
│  ┌─────────────────────────────────┐    │
│  │ Build a starship bridge simul..│    │
│  └─────────────────────────────────┘    │
│                                         │
│  ☐ Commission steward immediately       │
│                                         │
│  [Create]                  [Cancel]     │
└─────────────────────────────────────────┘
```

- **Textarea** — body text, required. Title auto-derived from first 60 chars (same as CaptureView pattern).
- **Commission checkbox** — when checked, after creating the entry and assigning it to the project, opens the CommissionDialog (or a combined flow). Intent defaults to the entry body text.
- **Create button** — calls `api.createEntry()`, then `api.setEntryProject(entryId, projectId)`, then optionally opens CommissionDialog.

### 3b. Combined create + commission flow

When the "Commission steward immediately" checkbox is checked, the flow becomes:

1. Create entry → get entry ID
2. Assign to project → `setEntryProject(id, projectId)`
3. Auto-open CommissionDialog with the new entry ID, intent pre-populated from body text
4. User confirms commission → steward starts

This is two dialogs in sequence (create → commission), not one giant form. Keeps each dialog focused.

### 3c. Verification

- "New Entry" button visible on project board
- Entry is created and assigned to the correct project
- Entry appears in Inbox column immediately (via WebSocket `entry.created` event or manual refresh)
- Commission checkbox triggers CommissionDialog after creation
- Full flow: type body → create → commission dialog → confirm → steward starts

---

## Phase 4: Commission List (Future)

**Not in initial build.** When commission count grows, add a Commissions tab or view showing all commissions with status, cost, and controls. For now, commissions are accessible from individual entries.

---

## Phasing Summary

| Phase | Scope | Standalone value? |
|-------|-------|-------------------|
| 1 | Types + API + CommissionDialog component | Yes — dialog is reusable, API layer complete |
| 2 | Commission triggers on entry detail + project board + status display | Yes — can commission and monitor entries |
| 3 | Create entry in project + commission shortcut | Yes — streamlines the project workflow |
| 4 | Commission list view | Deferred |

Each phase delivers value independently. Phase 1 is small enough for one session. Phase 2 is the core value. Phase 3 is convenience.

---

## Costs & Risks

**Implementation cost:** ~2-3 sessions across phases 1-3. Pure frontend work — no backend changes needed.

**Maintenance cost:** Low. CommissionDialog is a single reusable component. API functions are thin wrappers.

**Risks:**
- Commission state polling: no WebSocket events for commission status changes yet. Initial implementation can use polling on entry detail load. If this feels sluggish, add `commission.updated` WebSocket events in a future pass.
- Budget display accuracy depends on `cost_used` being updated in real-time by the steward. This is already handled by the backend commission loop.
- Edge case: what happens if Michael commissions an entry and then navigates away? The commission runs in the backend goroutine — this is fine. Status is persisted. No frontend presence required.

---

## Recommendation

**Build.** This is the missing link between a fully wired delegation backend and Michael's ability to actually use it. The backend is done and tested. The frontend gap is the only thing preventing end-to-end commission testing (which is already on the priority list in active.md).

Phase 1 + 2 should ship in one session. Phase 3 can follow immediately or defer until the commission flow is validated.
