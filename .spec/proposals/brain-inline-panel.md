# Brain Inline Panel: Reply + Close from Slide-Out

**Status:** planned
**Binding problem:** When the review/nudge agent asks a clarifying question on an entry, the user must click "Open →" and navigate to the full EntryDetailView just to type a reply. The slide-out panel shows the conversation but has no input. Similarly, there's no way to close/dismiss an entry with a personal note (e.g., "already completed" or "going a different direction") without routing it through an agent.

## Success Criteria

1. User can type and send a reply from the slide-out panel on ProjectDetailView — same `api.reply()` call as EntryDetailView
2. User can close/dismiss an entry with an optional reason note — stores the note as a session message, sets route_status to "dismissed" or "complete"
3. Reply auto-advance still triggers from panel replies (existing backend behavior, no change needed)
4. Panel conversation updates after reply without reopening

## Constraints

- **No new API endpoints.** Everything needed already exists: `api.reply()`, `api.markComplete()`, `api.updateEntry()`. The close-with-reason needs one new endpoint OR a two-step call (reply + markComplete).
- **Match EntryDetailView patterns.** The reply textarea + Ctrl+Enter + send button pattern is already implemented there.
- **Panel stays lightweight.** Don't turn the slide-out into a full entry editor. Reply and close are the two inline actions.

## Proposed Approach

### Part A: Inline Reply in Slide-Out Panel

Add to `ProjectDetailView.vue`:

**Script:**
- `panelReplyText` ref (string)
- `panelReplying` ref (boolean)
- `sendPanelReply()` async function — calls `api.reply(selectedEntry.id, panelReplyText)`, then reloads messages, clears input

**Template — insert after the conversation messages div, before "Agent Output":**
```html
<!-- Reply input (only when entry has active route) -->
<div v-if="selectedEntry.route_status && selectedEntry.route_status !== 'complete' && selectedEntry.route_status !== 'dismissed'" class="mt-3">
  <div class="flex gap-2">
    <textarea
      v-model="panelReplyText"
      @keydown.ctrl.enter="sendPanelReply"
      placeholder="Reply..."
      rows="2"
      class="flex-1 bg-gray-950 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-200 placeholder-gray-600 focus:outline-none focus:ring-2 focus:ring-sky-500 resize-none"
    />
    <button
      @click="sendPanelReply"
      :disabled="!panelReplyText.trim() || panelReplying"
      class="px-3 py-1.5 text-sm bg-sky-600 text-white rounded-lg hover:bg-sky-500 disabled:opacity-40 self-end"
    >Send</button>
  </div>
  <p class="text-xs text-gray-600 mt-1">Ctrl+Enter to send</p>
</div>
```

### Part B: Close with Reason

**New backend endpoint:** `POST /api/entries/{id}/close`
- Accepts `{ reason: string }` (optional)
- If reason provided: stores it as a session message with role "human" and prefix "[Closed] "
- Sets maturity_notes to reason (if provided)
- Sets route_status to "complete"
- Returns `{ entry_id, status: "closed" }`

This is distinct from `markComplete` because it stores the *why*.

**Frontend — add to slide-out panel actions section:**
```html
<button
  @click="openCloseDialog(selectedEntry!.id)"
  class="px-3 py-1.5 text-sm bg-gray-700 text-gray-300 rounded-lg hover:bg-gray-600 transition-colors"
>✕ Close</button>
```

**Close dialog** (reuse the feedback dialog pattern):
- Title: "Close Entry"
- Textarea placeholder: "Why are you closing this? (optional)"
- Submit calls the new close endpoint
- Refreshes the entry list

### Part C: Frontend API

Add to `api.ts`:
```typescript
closeEntry(entryId: string, reason?: string) {
  return request<{ entry_id: string; status: string }>(`/entries/${encodeURIComponent(entryId)}/close`, {
    method: 'POST',
    body: JSON.stringify({ reason }),
  })
},
```

### Part D: Nudge Bot Controls in Scheduled Tasks

**Problem:** The review nudge bot fires 4x/day (hours 7, 11, 15, 19) via a hardcoded goroutine in `server.go`. It doesn't appear in the Scheduled Tasks UI. The user can't pause it, can't see what it's doing, and can't control when it fires. Each nudge creates a new Copilot SDK session in the VS Code sidebar. It costs 0.33 premium requests per entry nudged (up to 10 entries per wake = 3.3 requests per wake = 13.2/day worst case).

**Current code:** `StartReviewLoop(DefaultReviewConfig())` called at server start (`server.go:51`). Config is hardcoded — no persistence, no API to modify.

**Proposed:**

1. **Surface the nudge bot as a "system" scheduled task** — on server start, if no "review-nudge" task exists in the scheduled_tasks table, auto-create one with status="active", schedule matching the current config. This makes it visible in Scheduled Tasks UI alongside user-created tasks.

2. **Pause/resume from Scheduled Tasks tab** — pausing the system task sets a flag that the review loop checks before each scan. No goroutine restart needed — just skip the scan if paused.

3. **Show nudge history in run history** — each scan that produces nudges logs a task run with the count and entry titles. Empty scans don't log (too noisy).

4. **Backend changes:**
   - Add `reviewEnabled` atomic bool to Pipeline struct, checked in `runReviewScan`
   - New endpoint `POST /api/review/toggle` (or reuse scheduled task pause/resume for the system task)
   - `GET /api/review/status` returns current config, enabled state, last scan time, next scan time
   - Modify `StartReviewLoop` to read enabled state before each scan

5. **Frontend — Scheduled Tasks tab:**
   - The review-nudge system task shows with a ⚙️ badge to distinguish from user tasks
   - Pause/resume works the same as any other scheduled task
   - Run history shows nudge scans: "Nudged 3 entries: [title1], [title2], [title3]"

**Why this matters:** Michael flagged that it "fires off a few rounds of haiku every 4 hours and I'm not around every 4 hours" and he's "not sure if it actually moves things along." Making it visible and pausable lets him evaluate its value without removing it entirely.

## Phased Delivery

**Phase 1 (one session):**
1. Add reply textarea + send to slide-out panel
2. Add close endpoint to backend
3. Add close button + dialog to slide-out panel
4. Add `closeEntry` to api.ts
5. Rebuild frontend, test

**Phase 2 (separate session):**
1. Surface review nudge bot as system scheduled task
2. Add pause/resume flow
3. Add nudge scan logging to task runs
4. Show in Scheduled Tasks tab with run history

## Verification

**Phase 1:**
- [ ] Open project board → click entry with "Review" badge → type reply in panel → message appears in conversation → auto-advance triggers if applicable
- [ ] Open project board → click any entry → click Close → type reason → entry shows as complete → reason visible in conversation history
- [ ] Close without reason → entry still closes, no empty message stored
- [ ] Ctrl+Enter sends reply from panel
- [ ] Panel refreshes conversation after send without needing to close/reopen

**Phase 2:**
- [ ] Scheduled Tasks tab shows "Review Nudge Bot" as a system task with ⚙️ badge
- [ ] Pausing the system task stops nudge scans (next wake hour skips)
- [ ] Resuming re-enables scans
- [ ] Run history shows nudge scan results: entry count and titles
- [ ] Empty scans don't clutter run history

## Costs

- **Phase 1 Backend:** ~30 lines (one new handler + route registration)
- **Phase 1 Frontend:** ~60 lines (reply refs, sendPanelReply function, close dialog, template additions)
- **Phase 2 Backend:** ~80 lines (atomic bool, system task auto-creation, toggle endpoint, scan logging)
- **Phase 2 Frontend:** ~20 lines (system task badge, no new UI beyond existing scheduled task patterns)
- **Risk:** Low — Phase 1 builds on existing patterns. Phase 2 bridges two existing systems (review loop + scheduled tasks).
- **Time:** Two focused sessions, one per phase
