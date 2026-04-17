# Commission UX Fixes — Research Findings

## Binding Problem

Two UX problems degrade the commission workflow:
1. Project-scoped findings paths get mangled in the UI, making them un-clickable and hard to read
2. When the steward pauses for input, there's no way to provide that input — Resume just restarts blindly
3. Related: the steward's "8 open questions" never actually appear in the thread (only a count summary)

## Issue 1: Findings Path Mangling

### Root Cause Chain

**Go `filepath.Join` on Windows** produces backslash paths in `projectRelPath()`:
- `scripts/brain/internal/pipeline/context.go` lines 209-220
- `scripts/brain/internal/pipeline/research.go` lines 322-330
- Result: `projects\cpuchip.net\.spec\scratch\pull-down-all-of-the-images...\main.md`

**MarkdownIt** interprets backslashes as escape characters:
- `\.` → `.`, `\_` → `_`, etc.
- The raw backslash path stored in the session message gets eaten by markdown parsing

**The normalization regex** in `useMarkdown.ts` (line 35-39) doesn't include `projects` in its prefix list:
```ts
/(\.(spec|github)|study|scripts|docs|...)(\\[\w._-]+)+/g
```
- Only handles paths starting with known workspace directories
- Project-prefixed paths like `projects\cpuchip.net\...` aren't caught

**The linkification regex** (line 26) also doesn't include `projects`:
```ts
/(?:^|\s|["'(>\x60])((\.spec|study|...|public)\/[\w./_-]+...)/g
```

### Best Fix Location

Fix at the **source** in Go: normalize `scratchPath` to forward slashes before storing in the session message. This way:
- MarkdownIt doesn't get confused by backslashes
- The existing normalization + linkification regex just need `projects` added to their prefix lists
- All future path communications are clean regardless of OS

Specifically:
1. `projectRelPath()` should use `path.Join` (POSIX) instead of `filepath.Join` (OS-native), OR normalize the result with `filepath.ToSlash()`
2. The `scratchPath` construction in `runResearch()` should also use forward slashes
3. Add `projects` to both regex prefix lists in `useMarkdown.ts`

Actually, the cleanest fix: **just use `filepath.ToSlash()`** on the final `scratchPath` before it goes into any message. Don't change `filepath.Join` everywhere (it's correct for actual file I/O). Only normalize for display/message purposes.

## Issue 2: Questions Not Appearing in Thread

### Root Cause

`extractQuestionSummary()` in `research.go` lines 451-517 **only extracts a count summary**:
```
**8 open questions** for you about Architecture, Deployment. Your answers will drive the planning phase.
```

The actual question text is only in the scratch file on disk. The user sees "8 open questions" but has to manually navigate to the scratch file to read them.

### Fix

Modify `extractQuestionSummary()` to also return the actual question text (or a separate function). Include the questions in the session message, perhaps in a collapsible format or just listed.

Options:
- **Include full questions** in the message — clearest for the user
- **Include abbreviated questions** (first line only, no sub-bullets) — keeps messages compact
- **Include full questions with a max** — e.g., first 10, with "see scratch file for all N"

Recommendation: include the full question text. These are typically 5-15 questions, each one line. That's a reasonable message length. The user needs to see them to provide answers.

## Issue 3: Resume With Input Dialog

### Current Flow

1. Commission surfaces a concern → status = "paused", message posted
2. User sees a green "▶ Resume" button
3. Clicking Resume calls `PUT /api/commissions/{id}/resume` with **no body**
4. Backend flips status to "active", posts "Commission resumed", spawns goroutine
5. No mechanism to pass feedback/direction

### What We Need

When the steward pauses for input, the user should be able to:
1. See what the steward is asking (already shown in the thread)
2. Provide direction/answers before resuming
3. Optionally resume without input (just "keep going")

### Design

**Frontend:**
- When "Resume" is clicked, show a dialog with:
  - The surface reason (from the last system message about surfacing)
  - A textarea for feedback/direction
  - "Resume with Feedback" button (primary)
  - "Resume Without Input" button (secondary/subtle)
- If feedback is provided, it gets posted as a session message AND passed to the resume API

**Backend:**
- `ResumeCommission(id string)` → `ResumeCommission(id string, feedback string)`
- If feedback is non-empty:
  - Post it as a `human` message to the thread
  - Store it somewhere the commission goroutine can pick it up
- The resumed goroutine should check for recent human messages as context

Actually, the simplest approach:
1. Frontend posts the feedback as a session message (role: "human") before calling resume
2. The resume API stays simple (no body needed)
3. The commission's pipeline stages already read session messages for context (the AI sees the full thread)

This means the feedback naturally flows into the AI's context without special plumbing. The dialog is purely a UX convenience — making it easy to type feedback before resuming.

**Even simpler:** Just make the Resume button open a dialog. The dialog has the textarea and two buttons. If the user types feedback, POST it as a session message, THEN call resume. If they click "Resume" without feedback, just call resume directly.

## Summary

Three fixes, decreasing complexity:
1. **Path normalization** — `filepath.ToSlash()` in Go + add `projects` to frontend regex (small)
2. **Questions in thread** — expand `extractQuestionSummary()` to include actual question text (small)
3. **Resume with input dialog** — new Vue dialog component, POST feedback before resume (medium)
