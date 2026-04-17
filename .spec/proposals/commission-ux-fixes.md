# Commission UX Fixes

*Status: Shipped (2026-04-15)*
*Research: [.spec/scratch/commission-ux-fixes/main.md](../../.spec/scratch/commission-ux-fixes/main.md)*

## Problem

The commission workflow has three UX gaps that degrade the experience when the steward surfaces for human input:

1. **Path mangling:** Project-scoped findings paths (e.g., `projects\cpuchip.net\.spec\scratch\...`) are stored with Windows backslashes. MarkdownIt interprets backslashes as escape characters, stripping them. The frontend regex that normalizes and linkifies paths doesn't include `projects` as a known prefix — so even if backslashes survived, the path wouldn't become clickable.

2. **Questions invisible:** `extractQuestionSummary()` reports "8 open questions for you" but never includes the actual question text. The user has to manually open the scratch file to see what's being asked.

3. **No input on resume:** When the steward pauses for input, the Resume button just flips the status. There's no dialog to provide feedback/answers — the steward restarts blind. The user can type in the reply box, but that's non-obvious and disconnected from the resume action.

## Success Criteria

- [x] Project-scoped paths display correctly (forward slashes) and are clickable links
- [x] Open questions from research appear as actual text in the session thread
- [x] Clicking Resume on a paused commission shows a dialog with a textarea for optional feedback
- [ ] Feedback typed in the resume dialog is posted as a session message before resuming
- [ ] Resuming without feedback still works (no regression)

## Approach

### Fix 1: Path Normalization (Backend + Frontend)

**Backend** — `scripts/brain/internal/pipeline/research.go`:
- After constructing `scratchPath` (line ~332), normalize to forward slashes for display:
  ```go
  displayPath := filepath.ToSlash(scratchPath)
  ```
- Use `displayPath` in the session message (line ~428), keep `scratchPath` for file I/O
- Same treatment in `projectRelPath()` return value — or better, normalize at the one call site in `runResearch`

**Frontend** — `scripts/brain/frontend/src/composables/useMarkdown.ts`:
- Add `projects` to the backslash normalization regex (line 37):
  ```ts
  /(projects[\\/][\w._-]+[\\/])?(\.(spec|github)|study|scripts|...)(\\[\w._-]+)+/g
  ```
- Add `projects` to the `FILE_PATH_RE` linkification regex (line 26):
  ```ts
  /(?:^|\s|["'(>\x60])((projects\/[\w._-]+\/)?(\.spec|study|...|public)\/[\w./_-]+...)/g
  ```
  
The backend fix is primary (prevents mangling at the source). The frontend regex update is defense-in-depth (catches any other paths we missed).

### Fix 2: Include Questions in Thread

**Backend** — `scripts/brain/internal/pipeline/research.go`:
- Modify `extractQuestionSummary()` to return the actual question text, not just a count
- Format: include the summary line ("8 open questions...") followed by the actual numbered questions
- Cap at a reasonable limit (e.g., first 20 questions) with a "see scratch file for all N" note if truncated

The function currently extracts the questions into a `[]string` slice (line ~488) but then only uses `len(questions)`. The fix is to append them to the returned summary.

### Fix 3: Resume with Input Dialog

**Frontend** — new component + changes to EntryDetailView:

**New: `ResumeDialog.vue`** (`scripts/brain/frontend/src/components/ResumeDialog.vue`):
- Props: `show` (boolean), `surfaceReason` (string — the concern text)
- Emits: `resume(feedback: string)`, `cancel`
- Template:
  - `<div role="dialog">` overlay (same pattern as the fixed dialogs from Phase 2)
  - Shows the surface reason text
  - Textarea for feedback/direction
  - Two buttons: "Resume with Feedback" (primary, amber) and "Resume" (secondary, subtle)
  - Escape key closes

**Modified: `EntryDetailView.vue`**:
- Import `ResumeDialog`
- New state: `showResumeDialog` (boolean), `surfaceReason` (string)
- When Resume is clicked: set `surfaceReason` from the last system message containing "Surfacing for your input", show dialog
- On dialog `resume(feedback)`: if feedback, POST session message (role: "human"), then call `resumeCommission()`, close dialog
- On dialog `cancel`: close dialog

**Backend** — no changes needed. The feedback flows naturally:
1. Frontend POSTs feedback as a session message via `POST /api/entries/{id}/reply`
2. Frontend calls `PUT /commissions/{id}/resume`
3. The commission goroutine restarts, runs the next pipeline stage
4. The pipeline stage's AI context includes the full session thread → sees the human's feedback
5. The gate evaluator also sees the feedback in the thread context

**Verified safe:** The reply handler calls `tryReplyAutoAdvance()`, but that only fires when `entry.AgentRoute == "review"` — commission entries won't match. The `route_status` update to `"your_turn"` is harmless since `commissionSurface` already sets that. So the POST purely primes the session context — no unintended pipeline actions.

This is the simplest design because the existing AI context mechanism already reads session messages. No special "feedback" field needed on the resume API.

## Phased Delivery

### Phase 1: Path Fix (small, backend + frontend)
- `filepath.ToSlash()` on `scratchPath` in research.go message formatting
- Add `projects` to both regex patterns in `useMarkdown.ts`
- **Verification:** Commission a project entry, check that findings path renders correctly and is clickable

### Phase 2: Questions in Thread (small, backend only)
- Expand `extractQuestionSummary()` to include question text
- **Verification:** Commission an entry, verify questions appear in the "Research complete" message

### Phase 3: Resume Dialog (medium, frontend only)
- New `ResumeDialog.vue` component
- Wire into EntryDetailView's Resume button
- Post feedback as session message before resuming
- **Verification:** Pause a commission, click Resume, type feedback, verify it appears in thread, verify commission resumes and the AI references the feedback

## Costs and Risks

**Effort:** Small. Phase 1-2 are ~10 lines of Go each. Phase 3 is a new dialog component (~80 lines) plus wiring (~20 lines).

**Risks:**
- Path regex changes could mis-match edge cases (project names with special characters). Mitigated by the backend `ToSlash` fix being the primary defense.
- Including all questions in the thread could make long messages if the AI generates 30+ questions. Mitigated by the cap.
- The resume dialog relies on the AI reading the feedback from the thread context. If the pipeline doesn't include recent messages in its prompt, feedback would be invisible. **Verify:** check that `RetryAdvance` and `EvaluateGate` include session messages in their context.

## Recommendation

**Build.** All three fixes are small, clearly scoped, and directly improve a workflow that's already in production use (the cpuchip.net commission shown in the screenshot). Phase 1-2 could ship in one session, Phase 3 in the same or the next.
