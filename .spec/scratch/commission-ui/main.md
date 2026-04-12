# Commission UI — Research & Inventory

**Created:** 2026-04-11
**For:** [.spec/proposals/commission-ui.md](../../proposals/commission-ui.md)

---

## Binding Problem

The steward commission backend is fully wired — API endpoints, goroutine state machine, gate evaluation, budget tracking, pause/resume/revoke — but there's **zero frontend**. Michael can't commission a steward without manual curl calls. The whole purpose of the commission model is delegated autonomy, and the delegation mechanism itself requires API expertise to invoke.

---

## Backend Inventory

### Commission API (all wired in server.go)

| Method | Route | What it does |
|--------|-------|-------------|
| POST | `/api/commissions` | Create — requires `entry_id`, optional `intent`, `authority`, `model`, `max_cost` |
| GET | `/api/commissions` | List all commissions |
| GET | `/api/commissions/{id}` | Get commission with decisions |
| PUT | `/api/commissions/{id}/pause` | Pause active commission |
| PUT | `/api/commissions/{id}/resume` | Resume paused commission |
| PUT | `/api/commissions/{id}/revoke` | Revoke commission permanently |

### Create request shape

```json
{
  "entry_id": "uuid (required)",
  "intent": "string — human's goal/instructions",
  "authority": "advance_and_execute | advance_only",
  "model": "claude-opus-4.6 (default)",
  "max_cost": 50.0
}
```

Defaults applied by `CreateCommission`: authority → `advance_and_execute`, model → `claude-opus-4.6`, max_cost → `50.0`.

### Commission struct (from store/types.go)

Key fields: ID, EntryID, ProjectID, Intent, Scope ("single"), Authority, Model, MaxCost, CostUsed, Status (active/paused/completed/revoked/failed), StartedAt, CreatedAt. Also loads `[]CommissionDecision` — log of gate decisions.

### State machine

```
raw → research → researched → plan → planned → scenarios → specced
                                                                ↓
                                          [execute if authority=advance_and_execute]
                                                                ↓
                                          executing → poll → verified → DONE
```

At each gate: EvaluateGate() → advance | revise | surface. If surface → pauses commission, sets `route_status="your_turn"`.

---

## Frontend Inventory

### What exists

| Pattern | Implementation |
|---------|---------------|
| Modal dialogs | `<Teleport to="body"><dialog>` with backdrop click dismiss |
| Entry creation | CaptureView — textarea, auto-title, notebook checkbox, NO project selector |
| Project assignment | EntryDetailView edit form — `setEntryProject()` API call |
| Pipeline actions on project board | Advance/Revise/Defer/Execute/Cancel/Verify/Complete buttons per-card |
| Toast system | Per-view `ref()` + `setTimeout(3s)`, fixed top-right, color variants |
| Loading states | `ref()` flags, disabled buttons during submission |
| Execute dialog | Modal with feedback textarea + model info + cost estimate |

### What's missing for commission

| Item | Impact |
|------|--------|
| `Commission` TypeScript interface | Can't type commission data |
| Commission API functions in api.ts | Can't call commission endpoints |
| "Create entry for project" button | Must create in CaptureView then assign separately |
| Commission trigger button | No way to invoke commission from UI |
| Commission form (intent, authority, budget) | No way to configure commission |
| Commission status badge on entries | Can't see which entries have active commissions |
| Commission list/management view | Can't see or control all commissions |
| Commission controls (pause/resume/revoke) | Can't manage in-flight commissions |

---

## Existing Dialog Patterns (for consistency)

All dialogs in ProjectDetailView follow this exact pattern:

```vue
<Teleport to="body">
  <dialog :open="showXxxDialog" class="fixed inset-0 z-40 flex items-center justify-center bg-transparent" v-if="showXxxDialog">
    <div class="fixed inset-0 bg-black/50" @click="showXxxDialog = false" />
    <div class="relative bg-gray-900 border border-gray-700 rounded-xl p-6 shadow-xl max-w-md mx-auto w-full">
      <!-- heading, form fields, buttons -->
    </div>
  </dialog>
</Teleport>
```

Buttons: primary action (colored) + Cancel (gray). Loading flag disables buttons during submission.

---

## Design Decisions

### "Create entry in project" — inline on ProjectDetailView
The project board has no way to add new entries. CaptureView creates orphan entries. A `+ New Entry` button on the project board that opens a dialog with title/body + auto-assigns to the project is the natural home. Matches the project-centric workflow.

### Commission trigger — two paths
1. **From project board card** — "📜 Commission" button on entry cards (like existing Advance/Execute buttons)
2. **From entry detail** — commission button in pipeline gates section
3. **Standalone commission dialog** — for commissioning any entry from anywhere

Path 1 and 2 are the most natural. Path 3 is useful for triage but lower priority.

### Commission form — progressive disclosure
Simple by default (intent + go), with expandable advanced options (authority, model, budget). Most commissions should be one click + one sentence.

### Commission status — badge on entry
Entry cards and detail view should show a small badge when a commission is active. The session messages already stream progress, but a visual indicator prevents accidentally starting a second pipeline action on a commissioned entry.
